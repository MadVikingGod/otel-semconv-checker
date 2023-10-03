package servers

import (
	"context"
	"log/slog"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	pbTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	pbResource "go.opentelemetry.io/proto/otlp/resource/v1"
)

type TraceServer struct {
	pbTrace.UnimplementedTraceServiceServer

	resourceVersion string
	resourceGroups  []string
	resourceIgnore  []string
}

func NewTraceService(cfg Config, g map[string]semconv.Group) *TraceServer {
	resourceGroups := []semconv.Group{}
	for _, group := range cfg.Resource.Groups {
		resourceGroups = append(resourceGroups, g[group])
	}
	return &TraceServer{
		resourceVersion: semconv.Version,
		resourceGroups:  semconv.Combine(resourceGroups...),
		resourceIgnore:  cfg.Resource.Ignore,
	}
}

func (s *TraceServer) Export(ctx context.Context, req *pbTrace.ExportTraceServiceRequest) (*pbTrace.ExportTraceServiceResponse, error) {
	if req == nil {
		return nil, nil
	}
	for _, r := range req.ResourceSpans {
		if r.SchemaUrl != s.resourceVersion {
			slog.Info("incorrect resource version",
				slog.String("version", r.SchemaUrl),
				slog.String("expected", s.resourceVersion),
			)
		}
		checkResource(s.resourceGroups, s.resourceIgnore, r.Resource)
		for j, scope := range r.ScopeSpans {
			slog.Info("scope",
				slog.Int("index", j),
				slog.Any("scope", scope.Scope),
				slog.String("schema_url", scope.SchemaUrl),
			)
			for k, span := range scope.Spans {
				slog.Info("span",
					slog.Int("index", k),
					slog.String("name", span.Name),
					slog.Any("attributes", span.Attributes),
				)
			}
		}
	}
	return &pbTrace.ExportTraceServiceResponse{}, nil
}

func filter(input, removed []string) []string {
	output := []string{}
OUTER:
	for _, in := range input {
		for _, rem := range removed {
			if in == rem {
				continue OUTER
			}
		}
		output = append(output, in)
	}
	return output
}

func checkResource(rg, ignore []string, r *pbResource.Resource) {
	if r != nil {
		missing, extra := semconv.Compare(rg, r.Attributes)
		missing, extra = filter(missing, ignore), filter(extra, ignore)
		if len(missing) > 0 {
			slog.Info("missing attributes",
				slog.Any("attributes", missing),
			)
		}
		if len(extra) > 0 {
			slog.Info("extra attributes",
				slog.Any("attributes", extra),
			)
		}
	}
}
