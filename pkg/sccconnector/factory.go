// SPDX-License-Identifier: Apache-2.0

package sccconnector

import (
	"context"

	"github.com/madvikinggod/otel-semconv-checker/pkg/match"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// this is the name used to refer to the connector in the config.yaml
const (
	typeStr      = "scc"
	scopeName    = "github.com/madvikinggod/otel-semconv-checker/pkg/sccconnector"
	scopeVersion = "v0.0.16"
	schemaUrl    = "https://opentelemetry.io/schemas/1.24.0"
)

// NewFactory creates a factory for example connector.
func NewFactory() connector.Factory {
	// OpenTelemetry connector factory to make a factory for connectors
	return connector.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		connector.WithTracesToLogs(createTracesToLogsConnector, component.StabilityLevelAlpha),
		connector.WithMetricsToLogs(createMetricToLogsConnector, component.StabilityLevelAlpha),
		connector.WithLogsToLogs(createLogsToLogsConnector, component.StabilityLevelAlpha))
}

func createTracesToLogsConnector(ctx context.Context, params connector.CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (connector.Traces, error) {
	c, err := newConnector(params, cfg)
	if err != nil {
		return nil, err
	}
	c.logsConsumer = nextConsumer
	return c, nil
}
func createMetricToLogsConnector(ctx context.Context, params connector.CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (connector.Metrics, error) {
	c, err := newConnector(params, cfg)
	if err != nil {
		return nil, err
	}
	c.logsConsumer = nextConsumer
	return c, nil
}
func createLogsToLogsConnector(ctx context.Context, params connector.CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (connector.Logs, error) {
	c, err := newConnector(params, cfg)
	if err != nil {
		return nil, err
	}
	c.logsConsumer = nextConsumer
	return c, nil
}

// schema for connector
type connectorImp struct {
	tracesMatch  *match.Traces
	metricsMatch *match.Metrics
	logsMatch    *match.Logs

	logsConsumer    consumer.Logs
	logger          *zap.Logger
	reportUnmatched bool

	resource pcommon.Resource

	component.StartFunc
	component.ShutdownFunc
}

// newConnector is a function to create a new connector
func newConnector(params connector.CreateSettings, config component.Config) (*connectorImp, error) {
	params.Logger.Info("Building sccconnector connector")
	cfg := config.(*Config)
	tracesMatch, err := match.NewTraces(cfg.Trace)
	if err != nil {
		return nil, err
	}
	metricsMatch, err := match.NewMetrics(cfg.Metrics)
	if err != nil {
		return nil, err
	}
	logsMatch, err := match.NewLogs(cfg.Log)
	if err != nil {
		return nil, err
	}

	return &connectorImp{
		tracesMatch:     tracesMatch,
		metricsMatch:    metricsMatch,
		logsMatch:       logsMatch,
		logger:          params.Logger,
		resource:        params.Resource,
		reportUnmatched: cfg.ReportUnmatched,
	}, nil
}

var _ connector.Traces = &connectorImp{}

// Capabilities implements the consumer interface.
func (c *connectorImp) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

var matchedBody pcommon.Value = pcommon.NewValueStr("matched")
var notMatchedBody pcommon.Value = pcommon.NewValueStr("not matched")

func (c *connectorImp) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	found, notFound := c.tracesMatch.Match(ctx, td)
	return c.sendLogs(ctx, found, notFound)
}
func (c *connectorImp) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	found, notFound := c.metricsMatch.Match(ctx, md)
	return c.sendLogs(ctx, found, notFound)
}
func (c *connectorImp) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	found, notFound := c.logsMatch.Match(ctx, ld)
	return c.sendLogs(ctx, found, notFound)
}

func (c *connectorImp) sendLogs(ctx context.Context, found, notFound []match.Response) error {
	if len(found) == 0 && len(notFound) == 0 {
		return nil
	}
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	c.resource.CopyTo(rl.Resource())
	scope := rl.ScopeLogs().AppendEmpty()
	scope.Scope().SetName(scopeName)
	scope.Scope().SetVersion(scopeVersion)
	scope.SetSchemaUrl(schemaUrl)
	scope.LogRecords().EnsureCapacity(len(found) + len(notFound))
	for _, resp := range found {
		lr := scope.LogRecords().AppendEmpty()
		matchedBody.CopyTo(lr.Body())
		putAttributes(resp, lr.Attributes())
	}
	if c.reportUnmatched {
		for _, resp := range notFound {
			lr := scope.LogRecords().AppendEmpty()
			notMatchedBody.CopyTo(lr.Body())
			putAttributes(resp, lr.Attributes())
		}
	}

	return c.logsConsumer.ConsumeLogs(ctx, logs)
}

func putAttributes(resp match.Response, dst pcommon.Map) {
	dst.EnsureCapacity(9)
	dst.PutStr("type", resp.Type)
	if resp.ResourceSchema != "" {
		dst.PutStr("resource.Schema", resp.ResourceSchema)
	}
	if resp.ServiceName != "" {
		resp.ServiceName = "_NONE"
	}
	dst.PutStr("service.name", resp.ServiceName)
	if resp.ScopeName != "" {
		dst.PutStr("scope.name", resp.ScopeName)
	}
	if resp.ScopeVersion != "" {
		dst.PutStr("scope.version", resp.ScopeVersion)
	}
	if resp.ScopeURL != "" {
		dst.PutStr("scope.schema_url", resp.ScopeURL)
	}
	if resp.Name != "" {
		dst.PutStr("name", resp.Name)
	}
	if len(resp.Attributes) > 0 {
		slice := dst.PutEmptySlice("missing_attributes")
		slice.EnsureCapacity(len(resp.Attributes))
		for _, attr := range resp.Attributes {
			slice.AppendEmpty().SetStr(attr)
		}
	}
	if len(resp.ExtraAttributes) > 0 {
		slice := dst.PutEmptySlice("extra_attributes")
		slice.EnsureCapacity(len(resp.ExtraAttributes))
		for _, attr := range resp.ExtraAttributes {
			slice.AppendEmpty().SetStr(attr)
		}
	}
}
