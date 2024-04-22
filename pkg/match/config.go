// SPDX-License-Identifier: Apache-2.0

package match

import (
	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	SemanticVersions map[string]semconv.SemanticVersion
	TracerProvider   trace.TracerProvider
	MeterProvider    metric.MeterProvider
}

func newConfig(opts ...Option) (*config, error) {
	svs, err := semconv.ParseSemanticVersion()
	if err != nil {
		return nil, err
	}

	cfg := &config{
		SemanticVersions: svs,
		TracerProvider:   otel.GetTracerProvider(),
		MeterProvider:    otel.GetMeterProvider(),
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg, nil
}

type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (f optionFunc) apply(cfg *config) {
	f(cfg)
}

func WithSemanticConventionGroups(versions map[string]semconv.SemanticVersion) Option {
	return optionFunc(func(cfg *config) {
		cfg.SemanticVersions = versions
	})
}

func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(cfg *config) {
		cfg.TracerProvider = provider
	})
}

func WithMeterProvider(provider metric.MeterProvider) Option {
	return optionFunc(func(cfg *config) {
		cfg.MeterProvider = provider
	})
}
