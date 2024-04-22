// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"testing"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	common "go.opentelemetry.io/proto/otlp/common/v1"
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

	assert.NotPanics(t, func() {
		RegisterTraceServer(grpcServer, cfg, nil)
		RegisterMetricsServer(grpcServer, cfg, nil)
	})
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
