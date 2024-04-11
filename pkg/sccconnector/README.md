This is a connector for the collector to continiously monitor telemetry's adhearence to semantic convetion.

To use use a colelctor with this built in, and add this to your confing:
```
connectors:
  scc:

service:
  pipelines:
    # This is your normal pipeline, add scc to it.
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp, scc]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp, scc]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp, scc]
    # This is your scc logs data
    logs/2:
      receivers: [scc]
      exporters: [otlp]
```