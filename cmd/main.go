package main

import (
	"context"
	"log/slog"
	"net"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv/servers"
	pbLog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	pbMetric "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	pbTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func main() {

	g, err := semconv.ParseGroups()
	if err != nil {
		slog.Error("failed to parse groups", "error", err)
		return
	}

	lis, err := net.Listen("tcp", "localhost:4317")
	if err != nil {
		slog.Error("failed to listen", "address", "localhost:4317", "error", err)
		return
	}

	cfg := servers.Config{
		Resource: servers.Match{
			Groups: []string{"host", "os"},
			Ignore: []string{"host.id", "host.name", "resource.name"},
		},
	}
	grpcServer := grpc.NewServer()
	pbTrace.RegisterTraceServiceServer(grpcServer, servers.NewTraceService(cfg, g))
	pbMetric.RegisterMetricsServiceServer(grpcServer, &metricServer{g: g})
	pbLog.RegisterLogsServiceServer(grpcServer, &logServer{g: g})

	slog.Info("starting server", "address", "localhost:4317")
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
		return
	}
}

type metricServer struct {
	pbMetric.UnimplementedMetricsServiceServer
	g map[string]semconv.Group
}

func (s *metricServer) Export(ctx context.Context, req *pbMetric.ExportMetricsServiceRequest) (*pbMetric.ExportMetricsServiceResponse, error) {
	return nil, nil
}

type logServer struct {
	pbLog.UnimplementedLogsServiceServer
	g map[string]semconv.Group
}

func (s *logServer) Export(ctx context.Context, req *pbLog.ExportLogsServiceRequest) (*pbLog.ExportLogsServiceResponse, error) {
	return nil, nil
}
