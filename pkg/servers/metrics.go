// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"fmt"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterMetricsServer(g *grpc.Server, cfg Config) error {
	metricMatch, err := match.NewMetrics(cfg.Metrics)
	if err != nil {
		return err
	}

	pmetricotlp.RegisterGRPCServer(g, &metric{
		metric:          metricMatch,
		reportUnmatched: cfg.ReportUnmatched,
		disableError:    cfg.DisableError,
	})
	return nil
}

type metric struct {
	pmetricotlp.UnimplementedGRPCServer

	metric          *match.Metrics
	reportUnmatched bool
	disableError    bool
}

var _ pmetricotlp.GRPCServer = &metric{}

func (m *metric) Export(_ context.Context, req pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	found, notFound := m.metric.Match(context.TODO(), req.Metrics())

	logResponse("missing Attributes", found)
	if m.reportUnmatched {
		logResponse("not matched", notFound)
	}

	if len(found) > 0 && !m.disableError {
		resp := pmetricotlp.NewExportResponse()
		resp.PartialSuccess().SetRejectedDataPoints(int64(len(found)))
		resp.PartialSuccess().SetErrorMessage("missing attributes")
		names := make([]string, 0, len(found))
		for _, r := range found {
			names = append(names, r.GetScopeName())
		}
		return resp, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))

	}

	return pmetricotlp.NewExportResponse(), nil
}
