package servers

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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
	oneShot         bool
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
			group:  semconv.GetAttributes(groups...),
			ignore: match.Ignore,
		})
	}

	return &TraceServer{
		resourceVersion: semconv.Version,
		resourceGroups:  semconv.GetAttributes(resourceGroups...),
		resourceIgnore:  cfg.Resource.Ignore,
		matches:         matches,
		reportUnmatched: cfg.ReportUnmatched,
		oneShot:         cfg.OneShot,
	}
}

func (s *TraceServer) Export(ctx context.Context, req *pbCollectorTrace.ExportTraceServiceRequest) (*pbCollectorTrace.ExportTraceServiceResponse, error) {
	if req == nil {
		return nil, nil
	}
	log := slog.With("type", "trace")
	count := 0
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
					if match.match.MatchString(span.Name) {
						found = true
						missing, extra := checkSpan(match.group, match.ignore, span)
						logAttributes(log, missing, extra)
						count += len(missing)
					}
				}
				if !found && s.reportUnmatched {
					log.Info("unmatched span")
				}
			}
		}
	}

	if s.oneShot {
		if count > 0 {
			os.Exit(100)
		}
		os.Exit(0)
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

func checkResource(rg, ignore []string, r *pbResource.Resource) (missing, extra []string) {
	if r != nil {
		missing, extra := semconv.Compare(rg, r.Attributes)
		missing, extra = filter(missing, ignore), filter(extra, ignore)
		return missing, extra
	}
	return nil, nil
}

func checkSpan(ag, ignore []string, s *pbTrace.Span) (missing []string, extra []string) {
	if s != nil {
		missing, extra := semconv.Compare(ag, s.Attributes)
		missing, extra = filter(missing, ignore), filter(extra, ignore)
		return missing, extra
	}
	return nil, nil
}

func logAttributes(log *slog.Logger, missing, extra []string) {
	if len(missing) > 0 {
		log.Info("missing attributes",
			slog.Any("attributes", missing),
		)
	}
	if len(extra) > 0 {
		log.Info("extra attributes",
			slog.Any("attributes", extra),
		)
	}
}
