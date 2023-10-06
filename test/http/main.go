package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func main() {
	serviceName := "e2e.http"
	serviceVersion := "v0.0.1"
	// Setup resource.
	res, err := newResource(serviceName, serviceVersion)
	if err != nil {
		panic(err)
	}

	shutdown, err := setupOTelSDK(res)

	mux := http.NewServeMux()
	mux.Handle("/echo", otelhttp.NewHandler(http.HandlerFunc(hello), "http.server.echo"))
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(hello), "http.server.root"))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()
	client.Transport = otelhttp.NewTransport(client.Transport, otelhttp.WithSpanNameFormatter(formatter))

	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		panic(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	resp, err = client.Get(ts.URL + "/echo")
	if err != nil {
		panic(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	shutdown(context.Background())
}

func formatter(operation string, r *http.Request) string {
	return strings.Join([]string{"http.client", r.Method, operation}, ".")
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Hello World")
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	res, err := resource.New(context.Background(),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	return resource.Merge(resource.Default(), res)
}
