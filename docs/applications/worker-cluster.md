# worker-cluster

> Fiber HTTP demo application used as a load-test target and observability example.

## What it is

A lightweight Go HTTP server built with [Fiber](https://gofiber.io/) that exposes `/metrics` and OTLP traces. It is the default target for k6 load tests in the observability stack.

## In this repo

| Path | Purpose |
|------|---------|
| `development/go/worker-cluster/main.go` | Application entry point |
| `development/go/worker-cluster/internal/otelsetup/` | OpenTelemetry setup |
| `development/go/worker-cluster/docker-compose.yaml` | App + observability include |
| `development/go/worker-cluster/Dockerfile` | Container build |
| `development/go/worker-cluster/configs/prometheus-app.yml` | Prometheus scrape config |
| `development/go/worker-cluster/loadtest.js` | Legacy k6 script (superseded by `docker/observability/configs/k6/`) |

## Quick start

```bash
cd development/go/worker-cluster
docker compose up -d
```

| URL | Service |
|-----|---------|
| http://localhost:8265 | worker-cluster app |
| http://localhost:3000 | Grafana |

Run k6 load test (from task-runner or observability stack):

```bash
docker compose --profile loadtest run --rm k6
```

Set `K6_TARGET_URL=http://worker-cluster:8265` to target this app.

## Configuration

### Environment variables

| Variable | Description |
|----------|-------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel-collector:4317` |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` |
| `OTEL_SERVICE_NAME` | `worker-cluster` |

### Compose integration

Includes `../../../docker/observability/docker-compose.yaml` and overrides Prometheus with `configs/prometheus-app.yml`.

## Making changes

1. Edit `main.go` or add routes/handlers.
2. Test: `cd development/go/worker-cluster && go test ./...`
3. Update `configs/prometheus-app.yml` if metrics path or port changes.
4. Rebuild: `docker compose build`.

## Integration

- Default k6 target when `K6_TARGET_URL` points at this service.
- Published to GHCR as `devsecops-boilerplate/worker-cluster`.

## Official resources

- [Fiber](https://docs.gofiber.io/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
