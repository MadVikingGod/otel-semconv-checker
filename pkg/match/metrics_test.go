// SPDX-License-Identifier: Apache-2.0

package match

import (
	"context"
	"path"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_Match(t *testing.T) {
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
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testResource",
				},
				{
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testScope",
				},
				{
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testMetric",
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
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testScope",
				},
				{
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testMetric",
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
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testResource",
					Attributes:  []string{"user_agent.original"},
				},
				{
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testScope",
					Attributes:  []string{"user_agent.original"},
				},
				{
					Type:        "Metric",
					ServiceName: "match peer",
					Name:        "testMetric",
					Attributes:  []string{"user_agent.original"},
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
					Type:            "Metric",
					ServiceName:     "match peer",
					Name:            "testResource",
					Attributes:      []string{"peer.service"},
					ExtraAttributes: []string{"NotPeerService", "service.name"},
				},
				{
					Type:            "Metric",
					ServiceName:     "match peer",
					Name:            "testScope",
					Attributes:      []string{"peer.service"},
					ExtraAttributes: []string{"NotPeerService", "service.name"},
				},
				{
					Type:            "Metric",
					ServiceName:     "match peer",
					Name:            "testMetric",
					Attributes:      []string{"peer.service"},
					ExtraAttributes: []string{"NotPeerService", "service.name"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := golden.ReadMetrics(path.Join("testdata", "metrics", tt.reqFile))
			require.NoError(t, err)

			m, err := NewMetrics(tt.config)
			require.NoError(t, err)
			gotFound, gotNotFound := m.Match(context.Background(), req)
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
