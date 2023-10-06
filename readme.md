# OTEL Semantic Convention Checker

A tool to check that your instrumentation is emitting the all the current semantic convention attributes.

## How It Works

otel-semconv-checker runs as a service similar to a collector.  When it receives telemetry, only traces currently, it will compare the attributes in the trace to the semantic convention groups and report which are missing.

## Getting Started

### Run The Server

Make changes ad needed to `config.yaml`, then run 

```bash
$ go run ./cmd
2023/10/06 10:14:35 INFO starting server address=localhost:4317
```

### Run the instrumentation

Configure your instrumentation, or collector, to point at the server. Or use one of the built in e2e tests

```bash
$ cd test
test/$ go run ./http
```

### Get the results

You should see log lines from the server:

```log
2023/10/06 10:14:35 INFO starting server address=localhost:4317
2023/10/06 10:16:37 INFO missing attributes type=trace section=resource version=https://opentelemetry.io/schemas/1.21.0 attributes="[host.type host.arch host.image.name host.image.id host.image.version os.name os.version os.build_id]"
2023/10/06 10:16:37 INFO extra attributes type=trace section=resource version=https://opentelemetry.io/schemas/1.21.0 attributes="[process.command_args process.executable.path telemetry.sdk.language telemetry.sdk.name process.runtime.description process.runtime.version service.name service.version telemetry.sdk.version process.executable.name process.owner process.pid process.runtime.name]"
2023/10/06 10:16:37 INFO incorrect scope version type=trace section=scope schemaUrl="" expected=https://opentelemetry.io/schemas/1.21.0 scope="name:\"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp\"  version:\"0.45.0\""
4
2023/10/06 10:16:37 INFO missing attributes type=trace section=span scope.name=go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp name=http.server.root attributes="[http.route server.address server.port url.scheme server.address server.port url.scheme server.address server.port server.socket.address server.socket.port client.address client.port client.socket.address client.socket.port url.path url.query url.scheme]"
2023/10/06 10:16:37 INFO extra attributes type=trace section=span scope.name=go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp name=http.server.root attributes="[net.sock.peer.port http.user_agent http.wrote_bytes http.status_code http.scheme net.host.name net.host.port net.sock.peer.addr http.method http.flavor]"
2023/10/06 10:16:37 INFO unmatched span type=trace section=span scope.name=go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp name=http.client.GET.
2023/10/06 10:16:37 INFO missing attributes type=trace section=span scope.name=go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp name=http.server.echo attributes="[http.route server.address server.port url.scheme server.address server.port url.scheme server.address server.port server.socket.address server.socket.port client.address client.port client.socket.address client.socket.port url.path url.query url.scheme]"
2023/10/06 10:16:37 INFO extra attributes type=trace section=span scope.name=go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp name=http.server.echo attributes="[net.host.port net.sock.peer.addr http.wrote_bytes http.status_code http.method http.scheme net.host.name http.flavor net.sock.peer.port http.user_agent]"
2023/10/06 10:16:37 INFO unmatched span type=trace section=span scope.name=go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp name=http.client.GET.
```