package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type lc struct {
	t testing.TB
}

func (l *lc) Accept(log testcontainers.Log) {
	l.t.Log(string(log.Content))
}

func TestHttp(t *testing.T) {
	wd, err := os.Getwd()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/madvikinggod/semantic-convention-checker:0.0.1",
		ExposedPorts: []string{"4317/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("4317/tcp"),
			wait.ForLog("INFO starting server address="),
		),

		Mounts: []testcontainers.ContainerMount{
			{
				Source: testcontainers.GenericBindMountSource{
					HostPath: fmt.Sprintf("%s/config.yaml", wd),
				},
				Target: "/config.yaml",
			},
		},
	}

	scc, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Logger:           testcontainers.TestLogger(t),
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, scc.Terminate(ctx))
	}()
	scc.FollowOutput(&lc{t: t})
	err = scc.StartLogProducer(ctx)
	require.NoError(t, err)

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		t.Errorf("Missing attributes: %v", err)
		rc, err := scc.Logs(context.Background())
		if err != nil {
			return
		}
		stdout, err := io.ReadAll(rc)
		if err != nil {
			return
		}
		t.Log(stdout)
	}))

	endpoint, err := scc.Endpoint(ctx, "")
	assert.NoError(t, err)

	t.Log(endpoint)

	tp, err := newTraceProvider(endpoint)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(hello), "http.server.root", otelhttp.WithTracerProvider(tp)))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()
	client.Transport = otelhttp.NewTransport(client.Transport, otelhttp.WithSpanNameFormatter(formatter), otelhttp.WithTracerProvider(tp))

	resp, err := client.Get(ts.URL + "/")
	require.NoError(t, err)

	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	err = tp.ForceFlush(context.Background())
	require.NoError(t, err)

	time.Sleep(5 * time.Second)
}
