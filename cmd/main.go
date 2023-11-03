package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"github.com/madvikinggod/otel-semconv-checker/pkg/servers"
	"github.com/spf13/viper"
	pbLog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	pbMetric "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	pbTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

var config = flag.String("cfg", "config.yaml", "The config file to use.")
var oneshot = flag.Bool("one", false, "The server will only receive one message, and exit 100 if it any attributes are missing.")

func main() {
	flag.Parse()

	svs, err := semconv.ParseSemanticVersion()
	if err != nil {
		slog.Error("failed to parse groups", "error", err)
		return
	}

	cfg := servers.Config{}

	viper.SetConfigFile(*config)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		viper.SetConfigType("yaml")
		viper.ReadConfig(strings.NewReader(servers.DefaultConfig))
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		return
	}

	if *oneshot {
		cfg.OneShot = true
	}

	lis, err := net.Listen("tcp", cfg.ServerAddress)
	if err != nil {
		slog.Error("failed to listen", "address", cfg.ServerAddress, "error", err)
		return
	}

	grpcServer := grpc.NewServer()
	pbTrace.RegisterTraceServiceServer(grpcServer, servers.NewTraceService(cfg, svs))
	pbMetric.RegisterMetricsServiceServer(grpcServer, servers.NewMetricsService(cfg, svs))
	pbLog.RegisterLogsServiceServer(grpcServer, &logServer{g: nil})

	slog.Info("starting server", "address", cfg.ServerAddress)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
		return
	}
}

type logServer struct {
	pbLog.UnimplementedLogsServiceServer
	g map[string]semconv.Group
}

func (s *logServer) Export(ctx context.Context, req *pbLog.ExportLogsServiceRequest) (*pbLog.ExportLogsServiceResponse, error) {
	return nil, nil
}
