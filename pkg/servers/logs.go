// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/madvikinggod/otel-semconv-checker/pkg/semconv"
	pbCollectorLogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogServer struct {
	pbCollectorLogs.UnimplementedLogsServiceServer
	resource        matchDef
	matches         []matchDef
	reportUnmatched bool

	disableError bool
}

var _ pbCollectorLogs.LogsServiceServer = &LogServer{}

func NewLogService(cfg Config, svs map[string]semconv.SemanticVersion) *LogServer {
	_, found := svs[cfg.Resource.SemanticVersion]
	if cfg.Resource.SemanticVersion != "" && !found {
		cfg.Resource.SemanticVersion = semconv.DefaultVersion
	}

	resSemVer := cfg.Resource.SemanticVersion
	if !found {
		resSemVer = semconv.DefaultVersion
	}
	resource := newMatchDef(cfg.Resource, svs[resSemVer].Groups)

	matches := []matchDef{}
	for _, match := range cfg.Log {
		groups, ok := svs[match.SemanticVersion]
		if !ok {
			match.SemanticVersion = semconv.DefaultVersion
			groups = svs[match.SemanticVersion]
		}
		matches = append(matches, newMatchDef(match, groups.Groups))
	}

	return &LogServer{
		resource:        resource,
		matches:         matches,
		reportUnmatched: cfg.ReportUnmatched,
		disableError:    cfg.DisableError,
	}
}

func (s *LogServer) Export(ctx context.Context, req *pbCollectorLogs.ExportLogsServiceRequest) (*pbCollectorLogs.ExportLogsServiceResponse, error) {
	if req == nil {
		return nil, nil
	}
	count := 0
	names := []string{}
	for _, r := range req.ResourceLogs {
		log := slog.With("type", "log")
		if schema := r.GetSchemaUrl(); schema != "" {
			log = log.With("resource.schema", schema)
		}
		if attr := r.Resource.GetAttributes(); len(attr) > 0 {
			name := ""
			for _, kv := range attr {
				if kv.Key == "service.name" {
					name = kv.Value.GetStringValue()
				}
				if name == "" {
					name = kv.String()
				}
			}
			if name != "" {
				log = log.With("service.name", name)
			}
		}

		for _, scope := range r.ScopeLogs {
			log := log.With(slog.String("section", "logs"))

			if scope := scope.GetScope(); scope != nil {
				if name := scope.GetName(); name != "" {
					log = log.With(slog.String("scope.name", name))
				}
				if version := scope.GetVersion(); version != "" {
					log = log.With(slog.String("scope.version", version))
				}
			}
			if url := scope.GetSchemaUrl(); url != "" {
				log = log.With(slog.String("scope.schema", url))
			}

			for _, record := range scope.LogRecords {
				found := false
				name := record.GetBody().String()
				if len(name) > 100 {
					name = name[:100]
				}
				log := log.With(slog.String("name", name))
				for _, match := range s.matches {
					if !match.isMatch(record.GetBody().String(), record.GetAttributes()) {
						continue
					}

					missing := match.compareAttributes(log, record.GetAttributes(), scope.GetScope().GetAttributes(), r.GetResource().GetAttributes())
					found = true
					count += missing
					if missing > 0 {
						names = append(names, fmt.Sprintf("%s/%s", scope.Scope.GetName(), name))
					}
				}
				if !found && s.reportUnmatched {
					log.Info("unmatched log")
				}
			}
		}
	}

	if count > 0 && !s.disableError {
		return &pbCollectorLogs.ExportLogsServiceResponse{
			PartialSuccess: &pbCollectorLogs.ExportLogsPartialSuccess{
				RejectedLogRecords: int64(count),
				ErrorMessage:       "missing attributes",
			},
		}, status.Error(codes.FailedPrecondition, fmt.Sprintf("missing attributes: %v", names))
	}

	return &pbCollectorLogs.ExportLogsServiceResponse{}, nil
}
