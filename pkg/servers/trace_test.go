// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	pbCollectorTrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	trace "go.opentelemetry.io/proto/otlp/trace/v1"
)

func TestTraceServerExport(t *testing.T) {
	defaultServer := &TraceServer{
		matches: []matchDef{newTestMatchDef([]string{"test"}, nil)},
	}
	testCases := []struct {
		name     string
		attrs    []attribute.KeyValue
		server   *TraceServer
		hasError bool
	}{
		{
			name: "matches",
			attrs: []attribute.KeyValue{
				attribute.String("test", "test"),
			},
		},
		{
			name:     "no attributes",
			attrs:    []attribute.KeyValue{},
			hasError: true,
		},
		{
			name: "no matches",
			attrs: []attribute.KeyValue{
				attribute.String("notTest", "test"),
			},
			hasError: true,
		},
		{
			name: "extra attributes",
			attrs: []attribute.KeyValue{
				attribute.String("test", "test"),
				attribute.String("notTest", "test"),
			},
		},
		{
			name: "Disable Error",
			attrs: []attribute.KeyValue{
				attribute.String("notTest", "test"),
			},
			server: &TraceServer{
				matches:      []matchDef{newTestMatchDef([]string{"test"}, nil)},
				disableError: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := defaultServer
			if tc.server != nil {
				ts = tc.server
			}
			req := newRequest(tc.attrs)
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

func newRequest(attrs []attribute.KeyValue) *pbCollectorTrace.ExportTraceServiceRequest {
	attributes := make([]*common.KeyValue, len(attrs))
	for i, attr := range attrs {
		attributes[i] = &common.KeyValue{
			Key:   string(attr.Key),
			Value: createValue(attr.Value.Emit()),
		}
	}
	if len(attributes) == 0 {
		attributes = nil
	}

	return &pbCollectorTrace.ExportTraceServiceRequest{
		ResourceSpans: []*trace.ResourceSpans{
			{
				ScopeSpans: []*trace.ScopeSpans{
					{
						Scope: &common.InstrumentationScope{
							Name: "TestScope",
						},
						Spans: []*trace.Span{
							{
								Name:       "test",
								Attributes: attributes,
							},
						}},
				},
			},
		},
	}
}

func createValue(value string) *common.AnyValue {
	return &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: value}}
}
