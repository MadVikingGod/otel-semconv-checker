// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"fmt"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterTraceServer(g *grpc.Server, cfg Config, svs map[string]semconv.SemanticVersion) {
	traceMatch := match.NewTraces(cfg.Trace, svs)

	ptraceotlp.RegisterGRPCServer(g, &trace{

		trace:           traceMatch,
		reportUnmatched: cfg.ReportUnmatched,
		disableError:    cfg.DisableError,
	})
}

// t *trace ptraceotlp.GRPCServer
type trace struct {
	ptraceotlp.UnimplementedGRPCServer

	trace           *match.Traces
	reportUnmatched bool
	disableError    bool
}

var _ ptraceotlp.GRPCServer = &trace{}

// Export is called every time a new request is received.
// For performance reasons, it is recommended to keep this RPC
// alive for the entire life of the application.
func (t *trace) Export(_ context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	found, notFound := t.trace.Match(context.TODO(), req)

	t.log("missing Attributes", found)
	if t.reportUnmatched {
		t.log("not matched", notFound)
	}

	if len(found) > 0 && !t.disableError {
		resp := ptraceotlp.NewExportResponse()
		resp.PartialSuccess().SetRejectedSpans(int64(len(found)))
		resp.PartialSuccess().SetErrorMessage("missing attributes")
		names := make([]string, 0, len(found))
		for _, r := range found {
			names = append(names, r.GetScopeName())
		}
		return resp, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))

	}

	return ptraceotlp.NewExportResponse(), nil
}

func (t *trace) log(msg string, responses []match.Response) {
	for _, resp := range responses {
		logger.Info(msg,
			"type", resp.Type,
			"resource.schema", resp.ResourceSchema,
			"service.name", resp.ServiceName,
			"scope.name", resp.ScopeName,
			"scope.version", resp.ScopeVersion,
			"scope.schema", resp.ScopeURL,
			"name", resp.Name,
			"missing.attributes", resp.Attributes,
			"extra.attributes", resp.ExtraAttributes,
		)
	}
}
