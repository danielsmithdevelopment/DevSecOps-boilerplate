# OpenTelemetry Collector

> Central OTLP ingress; routes traces, metrics, and logs to Tempo, Mimir, and Loki.

## In this repo

| Item | Value |
|------|-------|
| Image | `otel/opentelemetry-collector-contrib:0.116.1` |
| Config | `docker/observability/configs/otel-collector/config.yaml` |
| Ports | `14317` (gRPC OTLP), `14318` (HTTP OTLP) — mapped from 4317/4318 |
| Compose service | `otel-collector` |

Host-mapped ports avoid conflicts when app compose files also expose OTLP.

## Quick start

Apps connect inside the Docker network:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
OTEL_EXPORTER_OTLP_INSECURE=true
OTEL_SERVICE_NAME=my-service
```

From the host (for local dev outside Docker):

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:14317
```

## Configuration

`config.yaml` defines:

- **Receivers:** OTLP gRPC and HTTP
- **Processors:** batch, memory limiter
- **Exporters:** Tempo (traces), Mimir/Prometheus (metrics), Loki (logs)

## Making changes

1. Edit `configs/otel-collector/config.yaml`.
2. Add receivers/exporters for new signal types or backends.
3. Restart: `docker compose restart otel-collector`.
4. Validate: [OTel Collector config validator](https://opentelemetry.io/docs/collector/configuration/).

## Integration

- Primary telemetry path for [task-runner](../applications/task-runner.md), [grafana-stack](../applications/grafana-stack.md), [worker-cluster](../applications/worker-cluster.md)
- [Beyla](beyla.md) exports to this collector
- Secondary OTLP ingress available via [Alloy](alloy.md) on ports 4319/4320

## Official resources

- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [Collector contrib components](https://github.com/open-telemetry/opentelemetry-collector-contrib)
- [OTLP specification](https://opentelemetry.io/docs/specs/otlp/)
