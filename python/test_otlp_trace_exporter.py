from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import (
   OTLPSpanExporter
)
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import SimpleSpanProcessor

resource = Resource(attributes={
    SERVICE_NAME: "your-service-name"
})

provider = TracerProvider(resource=resource)
processor = SimpleSpanProcessor(
    OTLPSpanExporter(endpoint="0.0.0.0:4317")
)
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

tracer = trace.get_tracer("tracer_name")


def test_case():

    with tracer.start_as_current_span("0"):
        with tracer.start_as_current_span("1"):
            with tracer.start_as_current_span("2") as span:
                print("done")

    processor.on_end(span)
