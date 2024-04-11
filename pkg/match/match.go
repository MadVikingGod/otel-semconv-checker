// SPDX-License-Identifier: Apache-2.0

package match

import (
	"fmt"
	"maps"
	"regexp"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type Match struct {
	SemanticVersion string `mapstructure:"semantic_version"`
	Match           string
	MatchAttributes []Attribute `mapstructure:"match_attributes"`

	Groups           []string
	Ignore           []string
	Include          []string
	ReportAdditional bool `mapstructure:"report_additional"`
}

type Attribute struct {
	Name  string
	Value string
}

type matchDef struct {
	name   *regexp.Regexp
	attrs  map[string]string
	semVer *string
	group  []string
	group2 strMap
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
	group2 := map[string]struct{}{}
	for _, attr := range append(semconv.GetAttributes(groups...), m.Include...) {
		group2[attr] = struct{}{}
	}

	return matchDef{
		name:             reg,
		semVer:           semver,
		attrs:            attrs,
		group:            append(semconv.GetAttributes(groups...), m.Include...),
		group2:           group2,
		ignore:           m.Ignore,
		reportAdditional: m.ReportAdditional,
	}
}

func (m matchDef) isMatch(name string, attrs pcommon.Map) bool {
	return m.isNameMatch(name) && m.isAttrMatch(attrs)
}

func (m matchDef) isNameMatch(name string) bool {
	if m.name != nil {
		return m.name.MatchString(name)
	}
	return true
}

func (m matchDef) isAttrMatch(attrs pcommon.Map) bool {
	if len(m.attrs) == 0 {
		return true
	}
	for key, val := range m.attrs {
		sval, found := attrs.Get(key)
		if !found {
			return false
		}
		if val == "" {
			continue
		}
		if val != sval.AsString() {
			return false
		}
	}
	return true
}

func (m matchDef) compareAttributes(attrs ...pcommon.Map) ([]string, []string) {
	missing, extra := m.comparePdata(attrs...)
	missing, extra = filter(missing, m.ignore), filter(extra, m.ignore)

	miss := make([]string, 0, len(missing))
	for k := range missing {
		miss = append(miss, k)
	}
	ext := make([]string, 0, len(extra))
	for k := range extra {
		ext = append(ext, k)
	}

	return miss, ext
}

type strMap = map[string]struct{}

func (m matchDef) comparePdata(attributes ...pcommon.Map) (strMap, strMap) {
	missing := maps.Clone(m.group2)
	extra := strMap{}
	for _, m := range attributes {
		m.Range(func(k string, _ pcommon.Value) bool {
			extra[k] = struct{}{}
			return true
		})
	}

	for k := range missing {
		if _, f := extra[k]; f {
			delete(missing, k)
		}
	}
	if m.reportAdditional {
		for k := range extra {
			if _, f := m.group2[k]; f {
				delete(missing, k)
			}
		}
	} else {
		extra = strMap{}
	}

	return missing, extra
}

func filter(input strMap, removed []string) strMap {
	for _, k := range removed {
		delete(input, k)
	}
	return input
}

type Response struct {
	Type string

	ResourceSchema string
	ServiceName    string

	ScopeName    string
	ScopeVersion string
	ScopeURL     string

	Name            string
	Attributes      []string
	ExtraAttributes []string
}

func (r Response) GetScopeName() string {
	return fmt.Sprintf("%s/%s", r.ScopeName, r.Name)
}
