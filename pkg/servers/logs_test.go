// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	pbCollectorLogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	logs "go.opentelemetry.io/proto/otlp/logs/v1"
	resource "go.opentelemetry.io/proto/otlp/resource/v1"
)

func TestLogsServerExport(t *testing.T) {
	defaultServer := &LogServer{
		matches: []matchDef{newTestMatchDef([]string{"test"}, nil)},
	}
	testCases := []struct {
		name       string
		traceAttrs []attribute.KeyValue
		scopeAttrs []attribute.KeyValue
		resAttrs   []attribute.KeyValue
		server     *LogServer
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
			server: &LogServer{
				matches:      []matchDef{newTestMatchDef([]string{"test"}, nil)},
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
			req := newLogsRequest(tc.traceAttrs, tc.scopeAttrs, tc.resAttrs)
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

func newLogsRequest(logsAttrs, scopeAttrs, resAttrs []attribute.KeyValue) *pbCollectorLogs.ExportLogsServiceRequest {
	tAttrs := createKeyValues(logsAttrs)
	sAttrs := createKeyValues(scopeAttrs)
	rAttrs := createKeyValues(resAttrs)

	res := &resource.Resource{
		Attributes: rAttrs,
	}
	if len(resAttrs) == 0 {
		res = nil
	}

	return &pbCollectorLogs.ExportLogsServiceRequest{
		ResourceLogs: []*logs.ResourceLogs{
			{
				Resource: res,
				ScopeLogs: []*logs.ScopeLogs{
					{
						Scope: &common.InstrumentationScope{
							Name:       "TestScope",
							Attributes: sAttrs,
						},
						LogRecords: []*logs.LogRecord{
							{
								Body:       createValue("test"),
								Attributes: tAttrs,
							},
						}},
				},
			},
		},
	}
}
