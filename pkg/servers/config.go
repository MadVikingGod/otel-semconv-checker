// SPDX-License-Identifier: Apache-2.0

packageservers

type Config struct {
	ServerAddress   string `mapstructure:"server_address"`
	Resource        Match
	Trace           []Match
	Metrics         []Match
	Log             []Match
	ReportUnmatched bool `mapstructure:"report_unmatched"`
}

type Match struct {
	SemanticVersion  string `mapstructure:"semantic_version"`
	Match            string
	Groups           []string
	Ignore           []string
	ReportAdditional bool `mapstructure:"report_additional"`
}

var DefaultConfig = `---
resource:
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
`
