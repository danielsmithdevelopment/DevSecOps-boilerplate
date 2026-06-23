# Beyla

> eBPF-based auto-instrumentation for HTTP/gRPC metrics and traces without code changes.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/beyla:2.0.5` |
| Port | `8080` (instrumented app port via `BEYLA_OPEN_PORT`) |
| Compose service | `beyla` |

Runs `privileged: true` with `pid: host` — requires Linux and kernel debug access.

## Quick start

Beyla auto-discovers processes listening on `BEYLA_OPEN_PORT` (default `8080`):

```bash
# In .env
BEYLA_OPEN_PORT=8080
```

Traces and metrics appear in Grafana via [OTel Collector](opentelemetry-collector.md) → [Tempo](tempo.md) / [Mimir](mimir.md).

## Configuration

| Variable | Description |
|----------|-------------|
| `BEYLA_OPEN_PORT` | Port to instrument |
| `BEYLA_PRINT_TRACES` | `false` — log traces to stdout |
| `BEYLA_PROMETHEUS_PORT` | Prometheus scrape port on Beyla |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://otel-collector:4317` |

## Making changes

1. Set `BEYLA_OPEN_PORT` in `docker/observability/.env` to match your app's listen port.
2. On macOS Docker Desktop, eBPF features may be limited — Beyla is best tested on Linux.
3. Restart: `docker compose restart beyla`.

## Integration

- Exports OTLP to [OpenTelemetry Collector](opentelemetry-collector.md)
- Complements application-level instrumentation (does not replace it for custom spans)

## Official resources

- [Grafana Beyla docs](https://grafana.com/docs/beyla/latest/)
- [Beyla configuration](https://grafana.com/docs/beyla/latest/configure/)
