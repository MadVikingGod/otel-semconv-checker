package servers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterTraceServer(g *grpc.Server, cfg Config, svs map[string]semconv.SemanticVersion) {
	_, found := svs[cfg.Resource.SemanticVersion]
	if cfg.Resource.SemanticVersion != "" && !found {
		cfg.Resource.SemanticVersion = semconv.DefaultVersion
	}

	matches := []matchDef{}
	for _, match := range cfg.Trace {
		groups, ok := svs[match.SemanticVersion]
		if !ok {
			match.SemanticVersion = semconv.DefaultVersion
			groups = svs[match.SemanticVersion]
		}
		matches = append(matches, newMatchDef(match, groups.Groups))
	}

	ptraceotlp.RegisterGRPCServer(g, &pTrace{

		matches:         matches,
		reportUnmatched: cfg.ReportUnmatched,
		disableError:    cfg.DisableError,
	})
}

// t *pTrace ptraceotlp.GRPCServer
type pTrace struct {
	ptraceotlp.UnimplementedGRPCServer

	matches         []matchDef
	reportUnmatched bool

	disableError bool
}

var _ ptraceotlp.GRPCServer = &pTrace{}

// Export is called every time a new request is received.
// For performance reasons, it is recommended to keep this RPC
// alive for the entire life of the application.
func (t *pTrace) Export(_ context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	traces := req.Traces()
	rs := traces.ResourceSpans()
	count := 0
	names := []string{}
	for i := 0; i < rs.Len(); i++ {
		r := rs.At(i)
		log := logger.With("type", "trace")
		if schema := r.SchemaUrl(); schema != "" {
			log = log.With("resource.schema", schema)
		}

		if attr := r.Resource().Attributes(); attr.Len() > 0 {
			v, found := attr.Get("service.name")
			if found {
				log = log.With("service.name", v.AsString())
			}
		}

		ss := r.ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			scope := ss.At(j)
			is := scope.Scope()
			if name := is.Name(); name != "" {
				log = log.With(slog.String("scope.name", name))
			}
			if version := is.Version(); version != "" {
				log = log.With(slog.String("scope.version", version))
			}
			if url := scope.SchemaUrl(); url != "" {
				log = log.With(slog.String("scope.schema", url))
			}

			spans := scope.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				found := false
				name := span.Name()
				log := log.With(slog.String("name", name))
				for _, match := range t.matches {

					if !match.isPMatch(name, span.Attributes()) {
						continue
					}

					missing := match.comparePAttributes(log, span.Attributes(), scope.Scope().Attributes(), r.Resource().Attributes())
					found = true
					count += missing
					if missing > 0 {
						names = append(names, fmt.Sprintf("%s/%s", scope.Scope().Name(), span.Name()))
					}
				}
				if !found && t.reportUnmatched {
					log.Info("unmatched span")
				}
			}
		}
	}

	if count > 0 && !t.disableError {
		resp := ptraceotlp.NewExportResponse()
		resp.PartialSuccess().SetRejectedSpans(int64(count))
		resp.PartialSuccess().SetErrorMessage("missing attributes")
		return resp, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))

	}

	return ptraceotlp.NewExportResponse(), nil
}
