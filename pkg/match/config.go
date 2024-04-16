// SPDX-License-Identifier: Apache-2.0

package match

import (
	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	semanticVersionMap map[string]semconv.SemanticVersion
	traceProvider      trace.TracerProvider
	meterProvider      metric.MeterProvider
}
