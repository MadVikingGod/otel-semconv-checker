package match

import (
	"context"
	"testing"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

func TestTraces_Match(t *testing.T) {
	tests := []struct {
		name         string
		config       []Match
		req          testSpanReq
		wantFound    []Response
		wantNotFound []Response
	}{
		{
			name: "Matches Everything",
			config: []Match{
				{
					Match:  `test.*`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
			},
			req: testMultiReq("match peer", map[string]any{
				"peer.service": "res",
			}),
		},
		{
			name: "Matches Nothing",
			config: []Match{
				{
					Match:  `NotInTests`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
				{
					Match:  `DifferentNotMatch`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
			},
			req: testMultiReq("match peer", map[string]any{
				"peer.service": "res",
			}),
			wantNotFound: []Response{
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testResource",
				},
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testScope",
				},
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testSpan",
				},
			},
		},
		{
			name: "Matches Nothing",
			config: []Match{
				{
					Match:  `testResource`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
			},
			req: testMultiReq("match peer", map[string]any{
				"peer.service": "res",
			}),
			wantNotFound: []Response{
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testScope",
				},
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testSpan",
				},
			},
		},
		{
			name: "Missing peer.service",
			config: []Match{
				{
					Match:            `test.*`,
					Groups:           []string{"peer"}, // contains "peer.service"
					ReportAdditional: true,
				},
			},
			req: testMultiReq("match peer", map[string]any{
				"NotPeerService": "res",
			}),
			wantFound: []Response{
				{
					Type:            "Trace",
					ServiceName:     "match peer",
					Name:            "testResource",
					Attributes:      []string{"peer.service"},
					ExtraAttributes: []string{"NotPeerService", "service.name"},
				},
				{
					Type:            "Trace",
					ServiceName:     "match peer",
					Name:            "testScope",
					Attributes:      []string{"peer.service"},
					ExtraAttributes: []string{"NotPeerService", "service.name"},
				},
				{
					Type:            "Trace",
					ServiceName:     "match peer",
					Name:            "testSpan",
					Attributes:      []string{"peer.service"},
					ExtraAttributes: []string{"NotPeerService", "service.name"},
				},
			},
		},
		{
			name: "Matches One of the matches",
			config: []Match{
				{
					Match:  `NotInTests`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
				{
					Match:  `test.*`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
			},
			req: testMultiReq("match peer", map[string]any{
				"peer.service": "res",
			}),
		},
		{
			name: "Matches Multiple",
			config: []Match{
				{
					Match:  `test.*`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
				{
					Match:  `test.*`,
					Groups: []string{"attributes.user_agent"}, // contains "user_agent.original"
				},
			},
			req: testMultiReq("match peer", map[string]any{
				"peer.service": "res",
			}),
			wantFound: []Response{
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testResource",
					Attributes:  []string{"user_agent.original"},
				},
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testScope",
					Attributes:  []string{"user_agent.original"},
				},
				{
					Type:        "Trace",
					ServiceName: "match peer",
					Name:        "testSpan",
					Attributes:  []string{"user_agent.original"},
				},
			},
		},
	}

	svs, err := semconv.ParseSemanticVersion()
	require.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTraces(tt.config, svs)
			gotFound, gotNotFound := tr.Match(context.TODO(), tt.req.Req())
			assert.ElementsMatch(t, tt.wantFound, gotFound)
			assert.ElementsMatch(t, tt.wantNotFound, gotNotFound)
		})
	}
}

type testSpanReq struct {
	resSpans []testResSpans
}

type testResSpans struct {
	serviceName  string
	resourceAttr map[string]any
	scopes       []testScopeSpans
}

type testScopeSpans struct {
	scopeName  string
	scopeAttrs map[string]any
	traces     []testSpans
}
type testSpans struct {
	name  string
	attrs map[string]any
}

func (t testSpanReq) Req() ptraceotlp.ExportRequest {
	req := ptraceotlp.NewExportRequest()
	trace := req.Traces()
	for i, tRes := range t.resSpans {
		trace.ResourceSpans().AppendEmpty()
		rs := trace.ResourceSpans().At(i)

		res := rs.Resource()
		if tRes.resourceAttr == nil {
			tRes.resourceAttr = map[string]any{}
		}
		tRes.resourceAttr["service.name"] = tRes.serviceName
		res.Attributes().FromRaw(tRes.resourceAttr)

		scopes := rs.ScopeSpans()

		for j, tscp := range tRes.scopes {
			scopes.AppendEmpty()
			scp := scopes.At(j)
			scp.Scope().SetName(tscp.scopeName)
			scp.Scope().Attributes().FromRaw(tscp.scopeAttrs)

			spans := scp.Spans()
			for k, tTrace := range tscp.traces {
				spans.AppendEmpty()
				span := spans.At(k)
				span.SetName(tTrace.name)
				span.Attributes().FromRaw(tTrace.attrs)
			}
		}
	}

	return req
}

func testMultiReq(name string, attrs map[string]any) testSpanReq {
	return testSpanReq{
		resSpans: []testResSpans{
			{
				serviceName:  name,
				resourceAttr: attrs,
				scopes: []testScopeSpans{
					{
						traces: []testSpans{
							{
								name: "testResource",
							},
						},
					},
				},
			},
			{
				serviceName: name,
				scopes: []testScopeSpans{
					{
						scopeAttrs: attrs,
						traces: []testSpans{
							{
								name: "testScope",
							},
						},
					},
				},
			},
			{
				serviceName: name,
				scopes: []testScopeSpans{
					{
						traces: []testSpans{
							{
								name:  "testSpan",
								attrs: attrs,
							},
						},
					},
				},
			},
		},
	}

}
