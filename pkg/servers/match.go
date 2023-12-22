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
	semVer *string
	group  []string
	ignore []string
}

func newMatchDef(m Match, g map[string]semconv.Group) matchDef {
	semver := new(string)
	if m.SemanticVersion != "" {
		*semver = m.SemanticVersion
	}
	reg := regexp.MustCompile(m.Match)
	groups := []semconv.Group{}
	for _, group := range m.Groups {
		groups = append(groups, g[group])
	}
	return matchDef{
		name:   reg,
		semVer: semver,
		group:  append(semconv.GetAttributes(groups...), m.Include...),
		ignore: m.Ignore,
	}
}

func (m matchDef) match(log *slog.Logger, attrs []*v1.KeyValue, schemaUrl string) (int, bool) {
	missing, extra := semconv.Compare(m.group, attrs)
	missing, extra = filter(missing, m.ignore), filter(extra, m.ignore)

	if m.semVer != nil && *m.semVer != schemaUrl {
		log = log.With(slog.String("schema", schemaUrl))
	}

	logAttributes(log, missing, extra)
	return len(missing), true
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
