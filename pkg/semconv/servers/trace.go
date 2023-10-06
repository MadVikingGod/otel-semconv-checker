package servers

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	pbCollectorTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	pbResource "go.opentelemetry.io/proto/otlp/resource/v1"
	pbTrace "go.opentelemetry.io/proto/otlp/trace/v1"
)

type TraceServer struct {
	pbCollectorTrace.UnimplementedTraceServiceServer

	resourceVersion string
	resourceGroups  []string
	resourceIgnore  []string
	matches         []traceMatch
	reportUnmatched bool
}

type traceMatch struct {
	match  *regexp.Regexp
	group  []string
	ignore []string
}

func NewTraceService(cfg Config, g map[string]semconv.Group) *TraceServer {
	resourceGroups := []semconv.Group{}
	for _, group := range cfg.Resource.Groups {
		resourceGroups = append(resourceGroups, g[group])
	}
	matches := []traceMatch{}
	for _, match := range cfg.Trace {
		reg := regexp.MustCompile(match.Match)
		groups := []semconv.Group{}
		for _, group := range match.Groups {
			groups = append(groups, g[group])
		}
		matches = append(matches, traceMatch{
			match:  reg,
			group:  semconv.Combine(groups...),
			ignore: match.Ignore,
		})
	}

	return &TraceServer{
		resourceVersion: semconv.Version,
		resourceGroups:  semconv.Combine(resourceGroups...),
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
	for _, r := range req.ResourceSpans {
		if r.SchemaUrl != s.resourceVersion {
			log.Info("incorrect resource version",
				slog.String("version", r.SchemaUrl),
				slog.String("expected", s.resourceVersion),
			)
		}
		checkResource(s.resourceGroups, s.resourceIgnore, r.Resource)

		for _, scope := range r.ScopeSpans {
			if scope.SchemaUrl != s.resourceVersion {
				log.Info("incorrect scope version",
					slog.String("version", scope.SchemaUrl),
					slog.String("expected", s.resourceVersion),
					slog.Any("scope", scope.Scope),
				)
			}
			log := log
			if scope.Scope != nil {
				log = slog.With(slog.String("scope.name", scope.Scope.Name))
			}
			fmt.Println(len(scope.Spans))
			for _, span := range scope.Spans {
				found := false
				for _, match := range s.matches {
					if match.match.MatchString(span.Name) {
						found = true
						checkSpan(match.group, match.ignore, span, log)
					}
				}
				if !found && s.reportUnmatched {
					log.Info("unmatched span",
						slog.String("name", span.Name),
					)
				}
			}
		}
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

func checkSpan(ag, ignore []string, s *pbTrace.Span, log *slog.Logger) {
	if s != nil {
		missing, extra := semconv.Compare(ag, s.Attributes)
		missing, extra = filter(missing, ignore), filter(extra, ignore)
		if len(missing) > 0 {
			log.Info("missing attributes",
				slog.String("name", s.Name),
				slog.Any("attributes", missing),
			)
		}
		if len(extra) > 0 {
			log.Info("extra attributes",
				slog.String("name", s.Name),
				slog.Any("attributes", extra),
			)
		}
	}
}
