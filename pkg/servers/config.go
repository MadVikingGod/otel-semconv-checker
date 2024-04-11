// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"log/slog"
	"os"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
)

type Config struct {
	ServerAddress   string `mapstructure:"server_address"`
	Resource        Match
	Trace           []Match
	Metrics         []Match
	Log             []Match
	ReportUnmatched bool `mapstructure:"report_unmatched"`
	DisableError    bool `mapstructure:"disable_error"`
}

type Match = match.Match

var DefaultConfig = `---
resource:
  match_attributes:
    - name: "service.name"
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

var logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
