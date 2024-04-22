// SPDX-License-Identifier: Apache-2.0

package match

import (
	"context"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type Metrics struct {
	matches []matchDef
}

func NewMetrics(cfg []Match, svs map[string]semconv.SemanticVersion) *Metrics {
	matches := []matchDef{}
	for _, match := range cfg {
		groups, ok := svs[match.SemanticVersion]
		if !ok {
			match.SemanticVersion = semconv.DefaultVersion
			groups = svs[match.SemanticVersion]
		}
		matches = append(matches, newMatchDef(match, groups.Groups))
	}

	return &Metrics{matches: matches}
}

func (m *Metrics) Match(_ context.Context, metrics pmetric.Metrics) (found []Response, notFound []Response) {
	rs := metrics.ResourceMetrics()
	for i := 0; i < rs.Len(); i++ {
		resp := Response{
			Type: "Metric",
		}
		r := rs.At(i)
		if schema := r.SchemaUrl(); schema != "" {
			resp.ResourceSchema = schema
		}

		rAttr := r.Resource().Attributes()
		if rAttr.Len() > 0 {
			v, found := rAttr.Get("service.name")
			if found {
				resp.ServiceName = v.AsString()
			}
		}

		sm := r.ScopeMetrics()
		for j := 0; j < sm.Len(); j++ {
			resp := resp
			scope := sm.At(j)
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
			sAttr := scope.Scope().Attributes()

			metrics := scope.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				resp := resp
				metric := metrics.At(k)
				name := metric.Name()
				resp.Name = name

				fnd := []Response{}
				nfnd := []Response{}
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					dps := getAttributes(metric.Gauge().DataPoints())
					fnd, nfnd = m.matchDataPoints(resp, name, dps, sAttr, rAttr)
				case pmetric.MetricTypeSum:
					dps := getAttributes(metric.Sum().DataPoints())
					fnd, nfnd = m.matchDataPoints(resp, name, dps, sAttr, rAttr)
				case pmetric.MetricTypeHistogram:
					dps := getAttributes(metric.Histogram().DataPoints())
					fnd, nfnd = m.matchDataPoints(resp, name, dps, sAttr, rAttr)
				case pmetric.MetricTypeExponentialHistogram:
					dps := getAttributes(metric.ExponentialHistogram().DataPoints())
					fnd, nfnd = m.matchDataPoints(resp, name, dps, sAttr, rAttr)
				case pmetric.MetricTypeSummary:
					dps := getAttributes(metric.Summary().DataPoints())
					fnd, nfnd = m.matchDataPoints(resp, name, dps, sAttr, rAttr)

				}
				if len(fnd) > 0 {
					found = append(found, fnd...)
				}
				if len(nfnd) > 0 {
					notFound = append(notFound, nfnd...)
				}
			}
		}
	}
	return found, notFound

}

type attributeGetter interface {
	Attributes() pcommon.Map
}
type slice[V attributeGetter] interface {
	Len() int
	At(int) V
}

func getAttributes[G slice[V], V attributeGetter](s G) []pcommon.Map {
	attrs := make([]pcommon.Map, s.Len())
	for i := 0; i < s.Len(); i++ {
		attrs[i] = s.At(i).Attributes()
	}
	return attrs
}

func (m *Metrics) matchDataPoints(base Response, name string, dps []pcommon.Map, extraAttrs ...pcommon.Map) ([]Response, []Response) {
	found := []Response{}
	notFound := []Response{}
	for _, dpAttr := range dps {
		resp := base
		doesMatch := false
		for _, match := range m.matches {

			if !match.isMatch(name, dpAttr) {
				continue
			}

			missing, extra := match.compareAttributes(append(extraAttrs, dpAttr)...)
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
	return found, notFound
}
