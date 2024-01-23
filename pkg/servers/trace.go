// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	pbCollectorTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TraceServer struct {
	pbCollectorTrace.UnimplementedTraceServiceServer

	resource        matchDef
	matches         []matchDef
	reportUnmatched bool

	disableError bool
}

func NewTraceService(cfg Config, svs map[string]semconv.SemanticVersion) *TraceServer {
	_, found := svs[cfg.Resource.SemanticVersion]
	if cfg.Resource.SemanticVersion != "" && !found {
		cfg.Resource.SemanticVersion = semconv.DefaultVersion
	}

	resSemVer := cfg.Resource.SemanticVersion
	if !found {
		resSemVer = semconv.DefaultVersion
	}
	resource := newMatchDef(cfg.Resource, svs[resSemVer].Groups)

	matches := []matchDef{}
	for _, match := range cfg.Trace {
		groups, ok := svs[match.SemanticVersion]
		if !ok {
			match.SemanticVersion = semconv.DefaultVersion
			groups = svs[match.SemanticVersion]
		}
		matches = append(matches, newMatchDef(match, groups.Groups))
	}

	return &TraceServer{
		resource:        resource,
		matches:         matches,
		reportUnmatched: cfg.ReportUnmatched,
	}
}

func (s *TraceServer) Export(ctx context.Context, req *pbCollectorTrace.ExportTraceServiceRequest) (*pbCollectorTrace.ExportTraceServiceResponse, error) {
	if req == nil {
		return nil, nil
	}
	log := slog.With("type", "trace")
	count := 0
	names := []string{}
	for _, r := range req.ResourceSpans {
		if s.resource.semVer != nil && *s.resource.semVer != "" && r.SchemaUrl != *s.resource.semVer {
			log.Info("incorrect resource version",
				slog.String("section", "resource"),
				slog.String("version", r.SchemaUrl),
				slog.String("expected", *s.resource.semVer),
			)
		}
		if r.Resource != nil {
			log := log.With(
				slog.String("section", "resource"),
				slog.String("version", r.SchemaUrl),
			)

			s.resource.compareAttributes(log, r.Resource.Attributes)
		}

		for _, scope := range r.ScopeSpans {
			log := log.With(slog.String("section", "span"))
			if scope.Scope != nil {
				log = log.With(slog.String("scope.name", scope.Scope.Name))
			}
			if url := scope.GetSchemaUrl(); url != "" {
				log = log.With(slog.String("schema", url))
			}

			for _, span := range scope.Spans {
				found := false
				name := span.GetName()
				log := log.With(slog.String("name", name))
				for _, match := range s.matches {
					if !match.isMatch(name, span.GetAttributes()) {
						continue
					}

					missing := match.compareAttributes(log, span.GetAttributes(), scope.GetScope().GetAttributes(), r.GetResource().GetAttributes())
					found = true
					count += missing
					if missing > 0 {
						names = append(names, fmt.Sprintf("%s/%s", scope.Scope.GetName(), span.Name))
					}
				}
				if !found && s.reportUnmatched {
					log.Info("unmatched span")
				}
			}
		}
	}

	if count > 0 && !s.disableError {
		return &pbCollectorTrace.ExportTraceServiceResponse{
			PartialSuccess: &pbCollectorTrace.ExportTracePartialSuccess{
				RejectedSpans: int64(count),
				ErrorMessage:  "missing attributes",
			},
		}, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))
	}

	return &pbCollectorTrace.ExportTraceServiceResponse{}, nil
}
