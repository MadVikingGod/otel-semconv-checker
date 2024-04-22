// SPDX-License-Identifier: Apache-2.0

package sccconnector

import (
	"context"
	"testing"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/connector/connectortest"
	"go.opentelemetry.io/collector/consumer/consumertest"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

func Test_connectorImp_ConsumeTraces(t *testing.T) {

	tests := []struct {
		name    string
		config  Config
		args    ptrace.Traces
		wantlen int
	}{
		{
			name: "match everything",
			config: Config{
				Trace: []match.Match{
					{
						Match:   `.*`,
						Include: []string{"test"},
					},
				},
			},
			args: testMultiTrace(map[string]any{
				"test": "foo",
			}),
			wantlen: 0,
		},
		{
			name: "Missing Metrics",
			config: Config{
				Trace: []match.Match{
					{
						Match:   `.*`,
						Include: []string{"foo"},
					},
				},
			},
			args: testMultiTrace(map[string]any{
				"test": "foo",
			}),
			wantlen: 3,
		},
		{
			name: "Not found",
			config: Config{
				Trace: []match.Match{
					{
						Match:   `thisdoesnotexist`,
						Include: []string{"foo"},
					},
				},
				ReportUnmatched: true,
			},
			args: testMultiTrace(map[string]any{
				"test": "foo",
			}),
			wantlen: 3,
		},
		{
			name: "report extras",
			config: Config{
				Trace: []match.Match{
					{
						Match:            `.*`,
						Include:          []string{"test"},
						ReportAdditional: true,
					},
				},
			},
			args: testMultiTrace(map[string]any{
				"test":  "foo",
				"extra": "bar",
			}),
			wantlen: 3,
		},
	}
	f := NewFactory()
	ctx := context.Background()
	set := connectortest.NewNopCreateSettings()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logsSink := new(consumertest.LogsSink)
			c, err := f.CreateTracesToLogs(ctx, set, &tt.config, logsSink)
			assert.NoError(t, err)
			assert.NotNil(t, c)

			err = c.ConsumeTraces(context.Background(), tt.args)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantlen, logsSink.LogRecordCount())
		})
	}
}

func testMultiTrace(attrs map[string]any) ptrace.Traces {
	traces := ptrace.NewTraces()
	traces.ResourceSpans().EnsureCapacity(3)
	// attributes in the Resrouce
	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().FromRaw(attrs)
	scope := rs.ScopeSpans().AppendEmpty()
	scope.Scope().SetName("resource Scope")
	spans := scope.Spans().AppendEmpty()
	spans.SetName("resource span")
	// attributes in the Scope
	rs = traces.ResourceSpans().AppendEmpty()
	scope = rs.ScopeSpans().AppendEmpty()
	scope.Scope().SetName("scope Scope")
	scope.Scope().Attributes().FromRaw(attrs)
	spans = scope.Spans().AppendEmpty()
	spans.SetName("scope span")
	// attributes in the Span
	rs = traces.ResourceSpans().AppendEmpty()
	scope = rs.ScopeSpans().AppendEmpty()
	spans = scope.Spans().AppendEmpty()
	spans.SetName("span")
	spans.Attributes().FromRaw(attrs)

	return traces
}
