// SPDX-License-Identifier: Apache-2.0

package match

import (
	"context"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/collector/pdata/plog"
)

type Logs struct {
	matches []matchDef
}

func NewLogs(cfg []Match, opt ...Option) (*Logs, error) {
	matchCfg, err := newConfig(opt...)
	if err != nil {
		return nil, err
	}

	matches := []matchDef{}
	for _, match := range cfg {
		groups, ok := matchCfg.SemanticVersions[match.SemanticVersion]
		if !ok {
			match.SemanticVersion = semconv.DefaultVersion
			groups = matchCfg.SemanticVersions[match.SemanticVersion]
		}
		matches = append(matches, newMatchDef(match, groups.Groups))
	}

	return &Logs{matches: matches}, nil
}

func (t *Logs) Match(_ context.Context, logs plog.Logs) (found []Response, notFound []Response) {
	rs := logs.ResourceLogs()
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

		sl := r.ScopeLogs()
		for j := 0; j < sl.Len(); j++ {
			resp := resp
			scope := sl.At(j)
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

			logRecords := scope.LogRecords()
			for k := 0; k < logRecords.Len(); k++ {
				resp := resp
				lr := logRecords.At(k)
				doesMatch := false
				name := lr.Body().AsString()
				resp.Name = name
				for _, match := range t.matches {

					if !match.isMatch(name, lr.Attributes()) {
						continue
					}

					missing, extra := match.compareAttributes(lr.Attributes(), scope.Scope().Attributes(), r.Resource().Attributes())
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
