// SPDX-License-Identifier: Apache-2.0

package match

import (
	"context"
	"path"
	"testing"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraces_Match(t *testing.T) {
	tests := []struct {
		name         string
		config       []Match
		reqFile      string
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
			reqFile: "matches.yaml",
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
			reqFile: "matches.yaml",
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
			name: "Matches Some",
			config: []Match{
				{
					Match:  `testResource`,
					Groups: []string{"peer"}, // contains "peer.service"
				},
			},
			reqFile: "matches.yaml",
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
			reqFile: "missing.yaml",
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
			reqFile: "matches.yaml",
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
			reqFile: "matches.yaml",
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
			req, err := golden.ReadTraces(path.Join("testdata", "trace", tt.reqFile))
			require.NoError(t, err)

			gotFound, gotNotFound := tr.Match(context.TODO(), req)
			// Sort slice fields.  [Extra]Attributes are populated by a map and can be in random order.
			testSortResps(tt.wantFound)
			testSortResps(gotFound)
			assert.ElementsMatch(t, tt.wantFound, gotFound, "Founds do not match")
			testSortResps(tt.wantNotFound)
			testSortResps(gotNotFound)
			assert.ElementsMatch(t, tt.wantNotFound, gotNotFound, "NotFounds do not match")
		})
	}
}
