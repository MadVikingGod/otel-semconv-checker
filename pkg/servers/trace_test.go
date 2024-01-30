// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	pbCollectorTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	resource "go.opentelemetry.io/proto/otlp/resource/v1"
	trace "go.opentelemetry.io/proto/otlp/trace/v1"
)

func TestTraceServerExport(t *testing.T) {
	defaultServer := &TraceServer{
		matches: []matchDef{newTestMatchDef([]string{"test"}, nil)},
	}
	testCases := []struct {
		name       string
		traceAttrs []attribute.KeyValue
		scopeAttrs []attribute.KeyValue
		resAttrs   []attribute.KeyValue
		server     *TraceServer
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
			server: &TraceServer{
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
			req := newRequest(tc.traceAttrs, tc.scopeAttrs, tc.resAttrs)
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

func newRequest(traceAttrs, scopeAttrs, resAttrs []attribute.KeyValue) *pbCollectorTrace.ExportTraceServiceRequest {
	tAttrs := createKeyValues(traceAttrs)
	sAttrs := createKeyValues(scopeAttrs)
	rAttrs := createKeyValues(resAttrs)

	res := &resource.Resource{
		Attributes: rAttrs,
	}
	if len(resAttrs) == 0 {
		res = nil
	}

	return &pbCollectorTrace.ExportTraceServiceRequest{
		ResourceSpans: []*trace.ResourceSpans{
			{
				Resource: res,
				ScopeSpans: []*trace.ScopeSpans{
					{
						Scope: &common.InstrumentationScope{
							Name:       "TestScope",
							Attributes: sAttrs,
						},
						Spans: []*trace.Span{
							{
								Name:       "test",
								Attributes: tAttrs,
							},
						}},
				},
			},
		},
	}
}

func createKeyValues(attrs []attribute.KeyValue) []*common.KeyValue {
	keyValues := make([]*common.KeyValue, len(attrs))
	for i, attr := range attrs {
		keyValues[i] = &common.KeyValue{
			Key:   string(attr.Key),
			Value: createValue(attr.Value.Emit()),
		}
	}
	if len(keyValues) == 0 {
		keyValues = nil
	}
	return keyValues
}

func createValue(value string) *common.AnyValue {
	return &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: value}}
}
