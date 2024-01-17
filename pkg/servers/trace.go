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

	resourceVersion string
	resourceGroups  []string
	resourceIgnore  []string
	matches         []matchDef
	reportUnmatched bool
}

func NewTraceService(cfg Config, svs map[string]semconv.SemanticVersion) *TraceServer {
	_, found := svs[cfg.Resource.SemanticVersion]
	if cfg.Resource.SemanticVersion == "" || !found {
		cfg.Resource.SemanticVersion = semconv.DefaultVersion
	}

	resourceGroups := []semconv.Group{}
	for _, group := range cfg.Resource.Groups {
		resourceGroups = append(resourceGroups, svs[cfg.Resource.SemanticVersion].Groups[group])
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

	return &TraceServer{
		resourceVersion: cfg.Resource.SemanticVersion,
		resourceGroups:  semconv.GetAttributes(resourceGroups...),
		resourceIgnore:  cfg.Resource.Ignore,
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
		if r.SchemaUrl != s.resourceVersion {
			log.Info("incorrect resource version",
				slog.String("section", "resource"),
				slog.String("version", r.SchemaUrl),
				slog.String("expected", s.resourceVersion),
			)
		}
		missing, extra := checkResource(s.resourceGroups, s.resourceIgnore, r.Resource)
		logAttributes(log.With(
			slog.String("section", "resource"),
			slog.String("version", r.SchemaUrl),
		), missing, extra)

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

					missing, matched := match.matchAttributes(log, span.GetAttributes())
					found = found || matched
					count += missing
					if missing > 0 {
						names = append(names, scope.Scope.GetName())
					}
				}
				if !found && s.reportUnmatched {
					log.Info("unmatched span")
				}
			}
		}
	}

	if count > 0 {
		return &pbCollectorTrace.ExportTraceServiceResponse{
			PartialSuccess: &pbCollectorTrace.ExportTracePartialSuccess{
				RejectedSpans: int64(count),
				ErrorMessage:  "missing attributes",
			},
		}, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))
	}

	return &pbCollectorTrace.ExportTraceServiceResponse{}, nil
}
