// SPDX-License-Identifier: Apache-2.0

packagesemconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE ALL THESE ARE DEPENDANT ON THE SEMCONV.  THEY MAY CHANGE WITH THE SEMCONV.
func TestGetAttributes(t *testing.T) {
	groups, err := ParseGroups("src/v1.21.0")
	require.NoError(t, err)
	tests := []struct {
		name   string
		groups []string
		want   []string
	}{
		{
			name:   "Resolve References",
			groups: []string{"attributes.http.common"},
			want: []string{
				"http.request.method",
				"http.response.status_code",
				"error.type",
				"network.protocol.name",
				"network.protocol.version",
			},
		},
		{
			name:   "Resolve Extends",
			groups: []string{"trace.http.common"},
			want: []string{
				"http.request.method",
				"http.response.status_code",
				"error.type",
				"network.protocol.name",
				"network.protocol.version",
				"http.request.method_original",
				"http.request.body.size",
				"http.response.body.size",
				"http.request.header",
				"http.response.header",
				"network.transport",
				"network.type",
				"user_agent.original",
				"http.request.method",
			},
		},
		{
			name:   "Combine groups",
			groups: []string{"host", "os"},
			// https://opentelemetry.io/docs/specs/otel/resource/semantic_conventions/host/
			// https://opentelemetry.io/docs/specs/otel/resource/semantic_conventions/os/
			want: []string{
				"host.id",
				"host.name",
				"host.type",
				"host.arch",
				"host.image.name",
				"host.image.id",
				"host.image.version",
				"os.type",
				"os.description",
				"os.name",
				"os.version",
				"os.build_id",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grps := []Group{}
			for _, g := range tt.groups {
				grps = append(grps, groups[g])
			}
			got := GetAttributes(grps...)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
