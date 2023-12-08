// SPDX-License-Identifier: Apache-2.0

packagemain

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbTraceCollector "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	pbCommon "go.opentelemetry.io/proto/otlp/common/v1"
	pbResource "go.opentelemetry.io/proto/otlp/resource/v1"

	pbTrace "go.opentelemetry.io/proto/otlp/trace/v1"
)

func main() {

	conn, err := grpc.Dial("localhost:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pbTraceCollector.NewTraceServiceClient(conn)

	_, err = client.Export(context.Background(), traceReq)
	if err != nil {
		panic(err)
	}
}

var traceReq = &pbTraceCollector.ExportTraceServiceRequest{
	ResourceSpans: []*pbTrace.ResourceSpans{
		{
			Resource: &pbResource.Resource{
				Attributes: []*pbCommon.KeyValue{
					StringKV("resource.name", "resource-1"),
					StringKV("host.id", "host-1"),
					StringKV("host.type", "stuff"),
				},
			},
			SchemaUrl: "https://opentelemetry.io/schemas/1.21.0",
			ScopeSpans: []*pbTrace.ScopeSpans{
				{
					Scope: &pbCommon.InstrumentationScope{
						Name:    "scope-1",
						Version: "v1",
						Attributes: []*pbCommon.KeyValue{
							StringKV("scope.name", "scope-1"),
						},
					},
					Spans: []*pbTrace.Span{
						{
							Name: "empty",
							Attributes: []*pbCommon.KeyValue{
								StringKV("span.name", "empty"),
							},
						},
						{
							Name: "http.server.response",
							Attributes: []*pbCommon.KeyValue{
								StringKV("http.route", "something"),
							},
						},
					},
				},
			},
		},
		{
			Resource: &pbResource.Resource{
				Attributes: []*pbCommon.KeyValue{
					StringKV("resource.Extra", "resource-2"),
				},
			},
			SchemaUrl:  "https://opentelemetry.io/schemas/1.18.0",
			ScopeSpans: []*pbTrace.ScopeSpans{},
		},
	},
}

func StringKV(key, value string) *pbCommon.KeyValue {
	return &pbCommon.KeyValue{
		Key: key,
		Value: &pbCommon.AnyValue{
			Value: &pbCommon.AnyValue_StringValue{
				StringValue: value,
			},
		},
	}
}
