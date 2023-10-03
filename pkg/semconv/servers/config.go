package servers

type Config struct {
	ServerAddress string
	Resource      Match
	Trace         []Match
	Metric        []Match
	Log           []Match
}

type Match struct {
	Match            string
	Groups           []string
	Ignore           []string
	ReportAdditional bool
}

/*
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
- match: xyz.regex
  groups:
  -
  ignore:
  -
  report_additional: true
metric:
log:
report_unmatched: true
*/
