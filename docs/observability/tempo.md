# Tempo

> Distributed trace backend for OpenTelemetry and Jaeger-compatible queries.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/tempo:2.6.1` |
| Config | `docker/observability/configs/tempo/tempo.yaml` |
| Port | `3200` |
| Compose service | `tempo` |

Runs as `user: "0"` in compose to avoid permission issues on the blocks volume.

## Quick start

```bash
curl http://localhost:3200/ready
```

View traces in Grafana → Explore → Tempo.

## Configuration

`tempo.yaml` configures:

- **Receivers:** OTLP (from OTel Collector and Alloy)
- **Storage:** local filesystem under `/tmp/tempo`
- **Metrics generator:** optional span metrics for service graphs

## Making changes

1. Edit `configs/tempo/tempo.yaml`.
2. Adjust retention in `compactor` / `storage` sections.
3. Restart: `docker compose restart tempo`.

Apps send traces via OTLP:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
OTEL_EXPORTER_OTLP_INSECURE=true
```

## Integration

- Receives from [OTel Collector](opentelemetry-collector.md) and [Alloy](alloy.md)
- [Beyla](beyla.md) sends eBPF-derived traces via OTel Collector
- Grafana datasource with trace-to-log ([Loki](loki.md)) and trace-to-metrics ([Mimir](mimir.md)) links

## Official resources

- [Grafana Tempo docs](https://grafana.com/docs/tempo/latest/)
- [TraceQL](https://grafana.com/docs/tempo/latest/traceql/)
- [OpenTelemetry traces](https://opentelemetry.io/docs/concepts/signals/traces/)
