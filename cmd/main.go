// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/madvikinggod/otel-semconv-checker/pkg/servers"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
)

var config = flag.String("cfg", "config.yaml", "The config file to use.")

func main() {
	flag.Parse()

	cfg := servers.Config{}

	viper.SetConfigFile(*config)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		viper.SetConfigType("yaml")
		_ = viper.ReadConfig(strings.NewReader(servers.DefaultConfig))
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		return
	}

	lis, err := net.Listen("tcp", cfg.ServerAddress)
	if err != nil {
		slog.Error("failed to listen", "address", cfg.ServerAddress, "error", err)
		return
	}

	grpcServer := grpc.NewServer()

	err = servers.RegisterTraceServer(grpcServer, cfg)
	if err != nil {
		slog.Error("failed to register trace server", "error", err)
		return
	}
	err = servers.RegisterMetricsServer(grpcServer, cfg)
	if err != nil {
		slog.Error("failed to register metrics server", "error", err)
		return
	}
	err = servers.RegisterLogServer(grpcServer, cfg)
	if err != nil {
		slog.Error("failed to register log server", "error", err)
		return
	}

	slog.Info("starting server", "address", cfg.ServerAddress)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
		return
	}
}
