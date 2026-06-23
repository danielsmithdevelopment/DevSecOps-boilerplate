# grafana-stack app

> Go sample application demonstrating metrics, traces, and Postgres integration with the shared observability stack.

## What it is

A minimal Go HTTP service that writes data to PostgreSQL and exports Prometheus metrics and OpenTelemetry traces. It exists to show how an application plugs into the consolidated observability platform.

## In this repo

| Path | Purpose |
|------|---------|
| `docker/grafana-stack/cmd/main.go` | Application entry point |
| `docker/grafana-stack/internal/otelsetup/` | OpenTelemetry initialization |
| `docker/grafana-stack/docker-compose.yaml` | App + Postgres + observability include |
| `docker/grafana-stack/Dockerfile` | Container build |
| `docker/grafana-stack/.env.example` | Environment template |
| `docker/grafana-stack/prometheus-app.yml` | App-specific Prometheus scrape config |

## Quick start

```bash
cd docker/grafana-stack
cp .env.example .env
docker compose up -d
```

| URL | Service |
|-----|---------|
| http://localhost:8265 | grafana-stack app |
| http://localhost:3000 | Grafana |
| http://localhost:9090 | Prometheus |

## Configuration

### Environment variables

| Variable | Description |
|----------|-------------|
| `DB_HOST` | PostgreSQL host (`postgres`) |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | Database credentials |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP gRPC endpoint (`otel-collector:4317`) |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` for local dev |
| `OTEL_SERVICE_NAME` | `grafana-stack-app` |

### Compose integration

```yaml
include:
  - path: ../observability/docker-compose.yaml
```

Prometheus config is overridden with `prometheus-app.yml` to scrape the app on port 8265.

## Making changes

1. Edit `cmd/main.go` or `internal/otelsetup/`.
2. Test: `cd docker/grafana-stack && go test ./...`
3. Update scrape targets in `prometheus-app.yml` if ports or paths change.
4. Rebuild: `docker compose build` in `docker/grafana-stack/`.

> Legacy standalone configs (`loki-config.yaml`, `tempo.yaml`, etc.) in this directory are superseded by `docker/observability/configs/`. Do not edit them.

## Integration

- Shares the [observability stack](../observability/README.md) via compose `include`.
- Published to GHCR as `devsecops-boilerplate/grafana-stack` via [Docker publish](../cicd/docker-publish.md).

## Official resources

- [Go](https://go.dev/doc/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [pgx](https://github.com/jackc/pgx) — PostgreSQL driver
