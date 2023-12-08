// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	pbCollectorTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	pbTrace "go.opentelemetry.io/proto/otlp/trace/v1"
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
	if cfg.Resource.SemanticVersion == "" {
		cfg.Resource.SemanticVersion = semconv.DefaultVersion
	}

	resourceGroups := []semconv.Group{}
	for _, group := range cfg.Resource.Groups {
		resourceGroups = append(resourceGroups, svs[cfg.Resource.SemanticVersion].Groups[group])
	}
	matches := []matchDef{}
	for _, match := range cfg.Trace {
		if match.SemanticVersion == "" {
			match.SemanticVersion = semconv.DefaultVersion
		}
		matches = append(matches, newMatchDef(match, svs[match.SemanticVersion].Groups))
	}

	return &TraceServer{
		resourceVersion: semconv.DefaultVersion,
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
			if scope.SchemaUrl != s.resourceVersion {
				log.Info("incorrect scope version",
					slog.String("schemaUrl", scope.SchemaUrl),
					slog.String("expected", s.resourceVersion),
					slog.Any("scope", scope.Scope),
				)
				// count++
			}
			if scope.Scope != nil {
				log = log.With(slog.String("scope.name", scope.Scope.Name))
			}
			fmt.Println(len(scope.Spans))
			for _, span := range scope.Spans {
				found := false
				log := log.With(slog.String("name", span.Name))
				for _, match := range s.matches {
					if match.name.MatchString(span.Name) {
						found = true
						missing, extra := checkSpan(match.group, match.ignore, span)
						logAttributes(log, missing, extra)
						count += len(missing)
						names = append(names, scope.Scope.Name)
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

func checkSpan(ag, ignore []string, s *pbTrace.Span) (missing []string, extra []string) {
	if s != nil {
		missing, extra := semconv.Compare(ag, s.Attributes)
		missing, extra = filter(missing, ignore), filter(extra, ignore)
		return missing, extra
	}
	return nil, nil
}
