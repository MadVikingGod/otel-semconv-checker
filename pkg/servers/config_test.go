// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assert.Contains(t, cfg.Resource.MatchAttributes, Attribute{Name: "service.name"})

	assert.NotPanics(t, func() {
		NewTraceService(cfg, nil)
		NewMetricsService(cfg, nil)
	})
}
