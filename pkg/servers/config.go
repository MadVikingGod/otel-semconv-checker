// SPDX-License-Identifier: Apache-2.0

package servers

type Config struct {
	ServerAddress   string `mapstructure:"server_address"`
	Resource        Match
	Trace           []Match
	Metrics         []Match
	Log             []Match
	ReportUnmatched bool `mapstructure:"report_unmatched"`
	DisableError    bool `mapstructure:"disable_error"`
}

type Match struct {
	SemanticVersion  string `mapstructure:"semantic_version"`
	Match            string
	MatchAttributes  map[string]string `mapstructure:"match_attributes"`
	Groups           []string
	Ignore           []string
	Include          []string
	ReportAdditional bool `mapstructure:"report_additional"`
}

var DefaultConfig = `---
resource:
  match_attributes:
    service.name: ""
  groups:
  - host
  - os
  ignore:
  -
  report_additional: true
trace:
- match: http.server.*
  groups:
  - trace.http.server
  ignore:
  -
  report_additional: true
metric:
log:
report_unmatched: true
server_address: 0.0.0.0:4317
disable_error: false
`
