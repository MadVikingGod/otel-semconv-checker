package servers

type Config struct {
	ServerAddress   string `mapstructure:"server_address"`
	Resource        Match
	Trace           []Match
	Metrics         []Match
	Log             []Match
	ReportUnmatched bool `mapstructure:"report_unmatched"`
	OneShot         bool `mapstructure:"one_shot"`
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
server_address: localhost:4317
`
