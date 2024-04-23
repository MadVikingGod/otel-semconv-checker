// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"strings"
	"testing"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/otel/attribute"
)

func TestLogsServerExport(t *testing.T) {

	m, err := match.NewLogs([]Match{
		{
			Match:   `.*`,
			Include: []string{"test"},
		},
	})
	require.NoError(t, err)
	defaultServer := &log{
		log: m,
	}

	testCases := []struct {
		name       string
		traceAttrs []attribute.KeyValue
		scopeAttrs []attribute.KeyValue
		resAttrs   []attribute.KeyValue
		server     *log
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
			server: &log{
				log:          m,
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
			req := newPdataLogsRequest(t, tc.traceAttrs, tc.scopeAttrs, tc.resAttrs)
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

func newPdataLogsRequest(t *testing.T, traceAttrs, scopeAttrs, resAttrs []attribute.KeyValue) plogotlp.ExportRequest {
	tAttrs := testCreateMap(traceAttrs)
	sAttrs := testCreateMap(scopeAttrs)
	rAttrs := testCreateMap(resAttrs)

	req := plogotlp.NewExportRequest()
	req.Logs().ResourceLogs().AppendEmpty()
	rl := req.Logs().ResourceLogs().At(0)
	if len(resAttrs) != 0 {
		err := rl.Resource().Attributes().FromRaw(rAttrs)
		assert.NoError(t, err)
	}
	rl.ScopeLogs().AppendEmpty()
	sl := rl.ScopeLogs().At(0)
	sl.Scope().SetName("TestScope")
	if len(sAttrs) != 0 {
		err := sl.Scope().Attributes().FromRaw(sAttrs)
		assert.NoError(t, err)
	}
	sl.LogRecords().AppendEmpty()
	lr := sl.LogRecords().At(0)
	lr.Body().SetStr("test")
	if len(tAttrs) != 0 {
		err := lr.Attributes().FromRaw(tAttrs)
		assert.NoError(t, err)
	}

	return req
}
