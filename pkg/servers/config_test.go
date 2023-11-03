package servers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Config{}
	err := yaml.Unmarshal([]byte(DefaultConfig), &cfg)
	assert.NoError(t, err)

	assert.NotPanics(t, func() {
		NewTraceService(cfg, nil)
		NewMetricsService(cfg, nil)
	})
}
