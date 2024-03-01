from opentelemetry import trace
from logging import getLogger
from time import sleep
from requests import get
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import (
   OTLPSpanExporter
)
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import SimpleSpanProcessor
from typing import Union, TypeVar
from typing import Sequence as TypingSequence
from opentelemetry.sdk.trace import ReadableSpan
from opentelemetry.sdk.metrics.export import MetricsData
from grpc import (
    RpcError,
    StatusCode,
)
from google.rpc.error_details_pb2 import RetryInfo
from opentelemetry.exporter.otlp.proto.common._internal import (
    _create_exp_backoff_generator,
)
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.context import (
    _SUPPRESS_INSTRUMENTATION_KEY,
    attach,
    detach,
    set_value,
)
from docker import from_env
from pytest import fixture

logger = getLogger(__name__)

ExportResultT = TypeVar("ExportResultT")


class SimpleTestSpanProcessor(SimpleSpanProcessor):

    def on_end(self, span: ReadableSpan) -> None:
        if not span.context.trace_flags.sampled:
            return
        token = attach(set_value(_SUPPRESS_INSTRUMENTATION_KEY, True))
        self.span_exporter.export((span,))
        detach(token)


class OTLPTestSpanExporter(OTLPSpanExporter):

    def _export(
        self, data: Union[TypingSequence[ReadableSpan], MetricsData]
    ) -> ExportResultT:

        # After the call to shutdown, subsequent calls to Export are
        # not allowed and should return a Failure result.
        if self._shutdown:
            logger.warning("Exporter already shutdown, ignoring batch")
            return self._result.FAILURE

        # FIXME remove this check if the export type for traces
        # gets updated to a class that represents the proto
        # TracesData and use the code below instead.
        # logger.warning(
        #     "Transient error %s encountered while exporting %s, retrying in %ss.",  # noqa
        #     error.code(),
        #     data.__class__.__name__,
        #     delay,
        # )
        max_value = 64
        # expo returns a generator that yields delay values which grow
        # exponentially. Once delay is greater than max_value, the yielded
        # value will remain constant.
        for delay in _create_exp_backoff_generator(max_value=max_value):
            if delay == max_value or self._shutdown:
                return self._result.FAILURE

            with self._export_lock:
                try:
                    result = self._client.Export(
                        request=self._translate_data(data),
                        metadata=self._headers,
                        timeout=self._timeout,
                    )
                    result

                    return self._result.SUCCESS

                except RpcError as error:

                    if error.code() in [
                        StatusCode.CANCELLED,
                        StatusCode.DEADLINE_EXCEEDED,
                        StatusCode.RESOURCE_EXHAUSTED,
                        StatusCode.ABORTED,
                        StatusCode.OUT_OF_RANGE,
                        StatusCode.UNAVAILABLE,
                        StatusCode.DATA_LOSS,
                    ]:

                        retry_info_bin = dict(error.trailing_metadata()).get(
                            "google.rpc.retryinfo-bin"
                        )
                        if retry_info_bin is not None:
                            retry_info = RetryInfo()
                            retry_info.ParseFromString(retry_info_bin)
                            delay = (
                                retry_info.retry_delay.seconds
                                + retry_info.retry_delay.nanos / 1.0e9
                            )

                        logger.warning(
                            (
                                "Transient error %s encountered while "
                                "exporting "
                                "%s to %s, retrying in %ss."
                            ),
                            error.code(),
                            self._exporting,
                            self._endpoint,
                            delay,
                        )
                        sleep(delay)
                        continue
                    else:
                        raise
                        logger.error(
                            "Failed to export %s to %s, error code: %s",
                            self._exporting,
                            self._endpoint,
                            error.code(),
                            exc_info=error.code() == StatusCode.UNKNOWN,
                        )

                    if error.code() == StatusCode.OK:
                        return self._result.SUCCESS

                    return self._result.FAILURE

        return self._result.FAILURE


resource = Resource(attributes={
    SERVICE_NAME: "your-service-name"
})

provider = TracerProvider(resource=resource)
processor = SimpleTestSpanProcessor(
    OTLPTestSpanExporter(insecure=True, endpoint="0.0.0.0:4318")
)
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

tracer = trace.get_tracer("tracer_namesdfdsf")


@fixture
def create_kill_container():

    container = from_env().containers.run(
        "ghcr.io/madvikinggod/semantic-convention-checker:0.0.8",
        ports={"4318/tcp": 4318},
        volumes=[
            "/home/tigre/github/ocelotl/otel-semconv-checker/python/"
            "config.yaml:/config.yaml"
        ],
        detach=True
    )
    yield
    container.kill()


def test_requests(create_kill_container):

    RequestsInstrumentor().instrument()

    try:
        get(
            "http://localhost:8082/server_request",
            params={"param": "hello"},
            headers={"a": "b"},
        )
    except Exception as error:
        raise
        the_error = error
        the_error
        pass


def test_manual():

    span_name_prefix = "http.server."

    with tracer.start_as_current_span(f"{span_name_prefix}2") as span:
        print("done")

    processor.on_end(span)
