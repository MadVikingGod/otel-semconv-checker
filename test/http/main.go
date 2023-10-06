package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	shutdown, err := setupOTelSDK()

	mux := http.NewServeMux()
	mux.Handle("/echo", otelhttp.NewHandler(http.HandlerFunc(hello), "http.server.echo"))
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(hello), "http.server.root"))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()
	client.Transport = otelhttp.NewTransport(client.Transport)

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

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Hello World")
}
