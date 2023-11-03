package servers

import (
  "github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	pbResource "go.opentelemetry.io/proto/otlp/resource/v1"
)

func checkResource(rg, ignore []string, r *pbResource.Resource) (missing, extra []string) {
	if r != nil {
		missing, extra := semconv.Compare(rg, r.Attributes)
		missing, extra = filter(missing, ignore), filter(extra, ignore)
		return missing, extra
	}
	return nil, nil
}
