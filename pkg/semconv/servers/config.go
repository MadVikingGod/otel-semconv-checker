package servers

type Config struct {
	ServerAddress string `mapstructure:"server_address"`
	Resource      Match
	Trace         []Match
	Metric        []Match
	Log           []Match
}

type Match struct {
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
  - "host.id"
  - "host.name"
  - "resource.name"
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
`
