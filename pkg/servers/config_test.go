// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"testing"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Config{}
	tmpCfg := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(DefaultConfig), &tmpCfg)
	require.NoError(t, err)
	err = mapstructure.Decode(tmpCfg, &cfg)
	assert.NoError(t, err)

	assert.Equal(t, cfg.ServerAddress, "0.0.0.0:4317")
	assert.GreaterOrEqual(t, len(cfg.Trace), 1)
	assert.Equal(t, cfg.Trace[0].Match, "http.server.*")
	assert.GreaterOrEqual(t, len(cfg.Trace[0].Groups), 1)

	assert.Contains(t, cfg.Resource.MatchAttributes, match.Attribute{Name: "service.name"})

	grpcServer := grpc.NewServer()

	err = RegisterTraceServer(grpcServer, cfg)
	require.NoError(t, err)
	err = RegisterMetricsServer(grpcServer, cfg)
	require.NoError(t, err)

}
