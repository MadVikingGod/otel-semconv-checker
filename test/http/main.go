// SPDX-License-Identifier: Apache-2.0

packagemain

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/status"
)

func main() {

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
		fmt.Println("ERROR:", cause)
		if _, ok := status.FromError(cause); ok {
			os.Exit(101)
		}
	}))

	shutdown, err := setupOTelSDK("localhost:4317")

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
