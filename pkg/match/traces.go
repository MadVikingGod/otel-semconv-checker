package match

import (
	"context"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

type Traces struct {
	matches []matchDef
}

func NewTraces(cfg []Match, svs map[string]semconv.SemanticVersion) *Traces {
	matches := []matchDef{}
	for _, match := range cfg {
		groups, ok := svs[match.SemanticVersion]
		if !ok {
			match.SemanticVersion = semconv.DefaultVersion
			groups = svs[match.SemanticVersion]
		}
		matches = append(matches, newMatchDef(match, groups.Groups))
	}

	return &Traces{matches: matches}
}

func (t *Traces) Match(_ context.Context, req ptraceotlp.ExportRequest) (found []Response, notFound []Response) {
	traces := req.Traces()
	rs := traces.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		resp := Response{
			Type: "Trace",
		}
		r := rs.At(i)
		if schema := r.SchemaUrl(); schema != "" {
			resp.ResourceSchema = schema
		}

		if attr := r.Resource().Attributes(); attr.Len() > 0 {
			v, found := attr.Get("service.name")
			if found {
				resp.ServiceName = v.AsString()
			}
		}

		ss := r.ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			resp := resp
			scope := ss.At(j)
			is := scope.Scope()
			if name := is.Name(); name != "" {
				resp.ScopeName = name
			}
			if version := is.Version(); version != "" {
				resp.ScopeVersion = version
			}
			if url := scope.SchemaUrl(); url != "" {
				resp.ScopeURL = url
			}

			spans := scope.Spans()
			for k := 0; k < spans.Len(); k++ {
				resp := resp
				span := spans.At(k)
				doesMatch := false
				name := span.Name()
				resp.Name = name
				for _, match := range t.matches {

					if !match.isMatch(name, span.Attributes()) {
						continue
					}

					missing, extra := match.compareAttributes(span.Attributes(), scope.Scope().Attributes(), r.Resource().Attributes())
					resp.Attributes = append(resp.Attributes, missing...)
					resp.ExtraAttributes = append(resp.ExtraAttributes, extra...)
					doesMatch = true
				}

				if doesMatch {
					if len(resp.Attributes) > 0 || len(resp.ExtraAttributes) > 0 {
						found = append(found, resp)
					}
				} else {
					notFound = append(notFound, resp)
				}
			}
		}
	}
	return found, notFound
}
