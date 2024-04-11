// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"strings"
	"testing"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/otel/attribute"
)

func TestPTraceExport(t *testing.T) {
	svs, err := semconv.ParseSemanticVersion()
	require.NoError(t, err)
	defaultServer := &trace{
		trace: match.NewTraces([]Match{
			{
				Match:   `.*`,
				Include: []string{"test"},
			},
		}, svs),
	}

	testCases := []struct {
		name       string
		traceAttrs []attribute.KeyValue
		scopeAttrs []attribute.KeyValue
		resAttrs   []attribute.KeyValue
		server     *trace
		hasError   bool
	}{
		{
			name: "matches",
			traceAttrs: []attribute.KeyValue{
				attribute.String("test", "test"),
			},
		},
		{
			name:       "no attributes",
			traceAttrs: []attribute.KeyValue{},
			hasError:   true,
		},
		{
			name: "no matches",
			traceAttrs: []attribute.KeyValue{
				attribute.String("notTest", "test"),
			},
			hasError: true,
		},
		{
			name: "extra attributes",
			traceAttrs: []attribute.KeyValue{
				attribute.String("test", "test"),
				attribute.String("notTest", "test"),
			},
		},
		{
			name: "Disable Error",
			traceAttrs: []attribute.KeyValue{
				attribute.String("notTest", "test"),
			},
			server: &trace{
				trace: match.NewTraces([]match.Match{
					{
						Match:   `.*`,
						Include: []string{"test"},
					},
				}, svs),
				disableError: true,
			},
		},
		{
			name: "Match Scope Attrs",
			scopeAttrs: []attribute.KeyValue{
				attribute.String("test", "test"),
			},
		},
		{
			name: "Match Resource Attrs",
			resAttrs: []attribute.KeyValue{
				attribute.String("test", "test"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := defaultServer
			if tc.server != nil {
				ts = tc.server
			}
			req := newPdataRequest(tc.traceAttrs, tc.scopeAttrs, tc.resAttrs)
			_, err := ts.Export(context.Background(), req)
			if tc.hasError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				errMsg := err.Error()
				if !strings.Contains(errMsg, "TestScope") {
					t.Errorf("expected error to contain %q, got %q", "TestScope", errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %q", err.Error())
				}
			}
		})
	}
}

func newPdataRequest(traceAttrs, scopeAttrs, resAttrs []attribute.KeyValue) ptraceotlp.ExportRequest {
	tAttrs := testCreateMap(traceAttrs)
	sAttrs := testCreateMap(scopeAttrs)
	rAttrs := testCreateMap(resAttrs)

	req := ptraceotlp.NewExportRequest()
	req.Traces().ResourceSpans().AppendEmpty()
	rs := req.Traces().ResourceSpans().At(0)
	if len(resAttrs) != 0 {
		rs.Resource().Attributes().FromRaw(rAttrs)
	}
	rs.ScopeSpans().AppendEmpty()
	ss := rs.ScopeSpans().At(0)
	ss.Scope().SetName("TestScope")
	if len(sAttrs) != 0 {
		ss.Scope().Attributes().FromRaw(sAttrs)
	}
	ss.Spans().AppendEmpty()
	span := ss.Spans().At(0)
	span.SetName("test")
	if len(tAttrs) != 0 {
		span.Attributes().FromRaw(tAttrs)
	}

	return req
}

func testCreateMap(attrs []attribute.KeyValue) map[string]any {
	m := map[string]any{}
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.AsString()
	}
	return m
}
