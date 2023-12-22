// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	pbCollectorMetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
	pbMetrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pbCollectorMetrics.UnimplementedMetricsServiceServer

	resourceVersion string
	resourceGroups  []string
	resourceIgnore  []string
	matches         []matchDef
	reportUnmatched bool
}

func NewMetricsService(cfg Config, svs map[string]semconv.SemanticVersion) *MetricsServer {
	if cfg.Resource.SemanticVersion == "" {
		cfg.Resource.SemanticVersion = semconv.DefaultVersion
	}

	resourceGroups := []semconv.Group{}
	for _, group := range cfg.Resource.Groups {
		resourceGroups = append(resourceGroups, svs[cfg.Resource.SemanticVersion].Groups[group])
	}
	matches := []matchDef{}
	for _, match := range cfg.Metrics {
		groups, ok := svs[match.SemanticVersion]
		if !ok {
			match.SemanticVersion = semconv.DefaultVersion
		}
		matches = append(matches, newMatchDef(match, groups.Groups))
	}

	return &MetricsServer{
		resourceVersion: semconv.DefaultVersion,
		resourceGroups:  semconv.GetAttributes(resourceGroups...),
		resourceIgnore:  cfg.Resource.Ignore,
		matches:         matches,
		reportUnmatched: cfg.ReportUnmatched,
	}
}

func (s *MetricsServer) Export(ctx context.Context, req *pbCollectorMetrics.ExportMetricsServiceRequest) (*pbCollectorMetrics.ExportMetricsServiceResponse, error) {
	if req == nil {
		return nil, nil
	}
	log := slog.With("type", "metrics")
	count := 0
	names := []string{}
	for _, r := range req.ResourceMetrics {
		if r.SchemaUrl != s.resourceVersion {
			log.Info("incorrect resource version",
				slog.String("section", "resource"),
				slog.String("version", r.SchemaUrl),
				slog.String("expected", s.resourceVersion),
			)
		}
		missing, extra := checkResource(s.resourceGroups, s.resourceIgnore, r.Resource)
		logAttributes(log.With(
			slog.String("section", "resource"),
			slog.String("version", r.SchemaUrl),
		), missing, extra)

		for _, scope := range r.ScopeMetrics {
			log := log.With(slog.String("section", "metric"))
			if scope.SchemaUrl != s.resourceVersion {
				log.Info("incorrect scope version",
					slog.String("schemaUrl", scope.SchemaUrl),
					slog.String("expected", s.resourceVersion),
					slog.Any("scope", scope.Scope),
				)
				// count++
			}
			if scope.Scope != nil {
				log = log.With(slog.String("scope.name", scope.Scope.Name))
			}
			fmt.Println(len(scope.Metrics))
			for _, metric := range scope.Metrics {
				found := false
				log := log.With(slog.String("name", metric.Name))
				for _, match := range s.matches {
					missing, matched := checkMetric(log, match, metric, scope.GetSchemaUrl())
					found = found || matched
					count += missing
					if missing > 0 {
						names = append(names, scope.Scope.GetName())
					}
				}
				if !found && s.reportUnmatched {
					log.Info("unmatched metric")
				}
			}
		}
	}

	if count > 0 {
		return &pbCollectorMetrics.ExportMetricsServiceResponse{
			PartialSuccess: &pbCollectorMetrics.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(count),
				ErrorMessage:       "missing attributes",
			},
		}, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))
	}

	return &pbCollectorMetrics.ExportMetricsServiceResponse{}, nil
}

func checkMetric(log *slog.Logger, match matchDef, metric *pbMetrics.Metric, schemaUrl string) (int, bool) {
	name := metric.GetName()
	if !match.name.MatchString(name) {
		return 0, false
	}
	log = log.With(slog.String("name", name))

	switch d := metric.Data.(type) {
	case *pbMetrics.Metric_Gauge:
		return checkDataPoints(log, match, d.Gauge, schemaUrl)
	case *pbMetrics.Metric_Sum:
		return checkDataPoints(log, match, d.Sum, schemaUrl)
	case *pbMetrics.Metric_Histogram:
		return checkDataPoints(log, match, d.Histogram, schemaUrl)
	case *pbMetrics.Metric_Summary:
		return checkDataPoints(log, match, d.Summary, schemaUrl)
	case *pbMetrics.Metric_ExponentialHistogram:
		return checkDataPoints(log, match, d.ExponentialHistogram, schemaUrl)
	default:
		log.Warn("Unsupported metric type: %t", metric.Data)
	}
	return 0, false
}

func checkDataPoints[T attributeGetter, D dataPointGetter[T]](log *slog.Logger, match matchDef, metric D, schemaUrl string) (int, bool) {
	found := false
	count := 0
	for _, p := range metric.GetDataPoints() {
		missing, matched := match.match(log, p.GetAttributes(), schemaUrl)
		found = found || matched
		count += missing
	}
	return count, found
}

type attributeGetter interface {
	GetAttributes() []*v1.KeyValue
}

var _ attributeGetter = &pbMetrics.NumberDataPoint{}
var _ attributeGetter = &pbMetrics.HistogramDataPoint{}
var _ attributeGetter = &pbMetrics.SummaryDataPoint{}
var _ attributeGetter = &pbMetrics.ExponentialHistogramDataPoint{}

type dataPointGetter[T attributeGetter] interface {
	GetDataPoints() []T
}

var _ dataPointGetter[*pbMetrics.NumberDataPoint] = &pbMetrics.Gauge{}
var _ dataPointGetter[*pbMetrics.NumberDataPoint] = &pbMetrics.Sum{}
var _ dataPointGetter[*pbMetrics.HistogramDataPoint] = &pbMetrics.Histogram{}
var _ dataPointGetter[*pbMetrics.SummaryDataPoint] = &pbMetrics.Summary{}
var _ dataPointGetter[*pbMetrics.ExponentialHistogramDataPoint] = &pbMetrics.ExponentialHistogram{}
