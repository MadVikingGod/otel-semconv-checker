// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"log/slog"
	"regexp"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
)

type matchDef struct {
	name   *regexp.Regexp
	attrs  map[string]string
	semVer *string
	group  []string
	ignore []string

	reportAdditional bool
}

func newMatchDef(m Match, g map[string]semconv.Group) matchDef {
	semver := new(string)
	if m.SemanticVersion != "" {
		*semver = m.SemanticVersion
	}
	var reg *regexp.Regexp
	if m.Match != "" {
		reg = regexp.MustCompile(m.Match)
	}
	groups := []semconv.Group{}
	for _, group := range m.Groups {
		groups = append(groups, g[group])
	}
	attrs := map[string]string{}
	for _, attr := range m.MatchAttributes {
		attrs[attr.Name] = attr.Value
	}
	return matchDef{
		name:             reg,
		semVer:           semver,
		attrs:            attrs,
		group:            append(semconv.GetAttributes(groups...), m.Include...),
		ignore:           m.Ignore,
		reportAdditional: m.ReportAdditional,
	}
}

func (m matchDef) isMatch(name string, attrs []*v1.KeyValue) bool {
	return m.isNameMatch(name) && m.isAttrMatch(attrs)
}

func (m matchDef) isNameMatch(name string) bool {
	if m.name != nil {
		return m.name.MatchString(name)
	}
	return true
}

func (m matchDef) isAttrMatch(attrs []*v1.KeyValue) bool {
	if len(m.attrs) == 0 {
		return true
	}
	for key, val := range m.attrs {
		found := false
		for _, attr := range attrs {
			if attr.Key == key && (val == "" || attr.Value.GetStringValue() == val) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (m matchDef) compareAttributes(log *slog.Logger, attrs ...[]*v1.KeyValue) int {
	missing, extra := semconv.Compare(m.group, attrs...)
	missing, extra = filter(missing, m.ignore), filter(extra, m.ignore)

	m.logAttributes(log, missing, extra)

	return len(missing)
}

func (m matchDef) logAttributes(log *slog.Logger, missing, extra []string) {
	if len(missing) > 0 {
		log.Info("missing attributes",
			slog.Any("attributes", missing),
		)
	}
	if len(extra) > 0 && m.reportAdditional {
		log.Info("extra attributes",
			slog.Any("attributes", extra),
		)
	}
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
