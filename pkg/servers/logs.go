// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"fmt"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterLogServer(g *grpc.Server, cfg Config) error {
	logMatch, err := match.NewLogs(cfg.Log)
	if err != nil {
		return err
	}

	plogotlp.RegisterGRPCServer(g, &log{

		log:             logMatch,
		reportUnmatched: cfg.ReportUnmatched,
		disableError:    cfg.DisableError,
	})
	return nil
}

type log struct {
	plogotlp.UnimplementedGRPCServer

	log             *match.Logs
	reportUnmatched bool
	disableError    bool
}

var _ plogotlp.GRPCServer = &log{}

func (t *log) Export(_ context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	found, notFound := t.log.Match(context.TODO(), req.Logs())

	logResponse("missing Attributes", found)
	if t.reportUnmatched {
		logResponse("not matched", notFound)
	}

	if len(found) > 0 && !t.disableError {
		resp := plogotlp.NewExportResponse()
		resp.PartialSuccess().SetRejectedLogRecords(int64(len(found)))
		resp.PartialSuccess().SetErrorMessage("missing attributes")
		names := make([]string, 0, len(found))
		for _, r := range found {
			names = append(names, r.GetScopeName())
		}
		return resp, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))

	}

	return plogotlp.NewExportResponse(), nil
}
