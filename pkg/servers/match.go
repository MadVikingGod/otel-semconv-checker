// SPDX-License-Identifier: Apache-2.0

packageservers

import (
	"regexp"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
)

type matchDef struct {
	name   *regexp.Regexp
	group  []string
	ignore []string
}

func newMatchDef(m Match, g map[string]semconv.Group) matchDef {
	reg := regexp.MustCompile(m.Match)
	groups := []semconv.Group{}
	for _, group := range m.Groups {
		groups = append(groups, g[group])
	}
	return matchDef{
		name:   reg,
		group:  semconv.GetAttributes(groups...),
		ignore: m.Ignore,
	}
}
