# task-runner

> Distributed cron task runner with JWT auth, PostgreSQL storage, and full observability.

## What it is

A Go service that schedules and executes HTTP tasks across a cluster of replicas. Tasks use cron syntax, support dependencies, and store results in PostgreSQL. Five replicas run by default to demonstrate distributed execution.

## In this repo

| Path | Purpose |
|------|---------|
| `task-runner/cmd/server/main.go` | Application entry point |
| `task-runner/internal/` | API, scheduler, worker, storage, OTel setup |
| `task-runner/docker-compose.yml` | 5 replicas + shared observability stack |
| `task-runner/Dockerfile` | Multi-stage build, non-root user |
| `task-runner/.env.example` | Environment template |
| `task-runner/configs/prometheus.yml` | App-specific Prometheus scrape jobs |
| `scripts/create_test_task.sh` | Create a sample task via API |

## Quick start

```bash
cd task-runner
cp .env.example .env
docker compose up -d
```

| URL | Service |
|-----|---------|
| http://localhost:8080–8084 | Task runner API (5 instances) |
| http://localhost:3000 | Grafana |
| http://localhost:9090 | Prometheus |

Verify the stack:

```bash
./scripts/e2e-observability-stack.sh
```

## Configuration

### Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_CONN` | PostgreSQL connection string | `postgres://taskrunner:taskrunner@postgres:5432/taskrunner?sslmode=disable` |
| `JWT_SECRET` | JWT signing key | Set in `.env` |
| `INSTANCE_ID` | Unique replica identifier | `task-runner-1` … `task-runner-5` |
| `ADDR` | Listen address | `:8080` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP gRPC endpoint | `otel-collector:4317` |
| `OTEL_SERVICE_NAME` | Trace/metric service name | `task-runner` |

### Compose integration

`task-runner/docker-compose.yml` includes the shared observability stack:

```yaml
include:
  - path: ../docker/observability/docker-compose.yaml
    env_file: .env
```

Prometheus scrape config is overridden via `configs/prometheus.yml` to add task-runner targets while keeping stack self-monitoring.

## API

Authenticate first:

```bash
curl -X POST http://localhost:8080/auth \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth` | POST | Get JWT token |
| `/task/create` | POST | Create task |
| `/task/update` | PUT | Update task |
| `/task/delete/:id` | DELETE | Delete task |
| `/task/invoke/:id` | POST | Manual invoke |
| `/task/status/:id` | GET | Status and results |
| `/health` | GET | Health check |
| `/metrics` | GET | Prometheus metrics |

### Task example

```json
{
  "name": "example-task",
  "type": "http",
  "schedule": "*/5 * * * *",
  "config": {
    "http": {
      "url": "https://api.example.com",
      "method": "GET"
    }
  }
}
```

## Making changes

1. Edit Go code under `task-runner/internal/` or `cmd/server/`.
2. Run tests: `cd task-runner && go test ./...`
3. Lint locally: `golangci-lint run ./...` (see [GitHub Actions CI](../cicd/github-actions-ci.md)).
4. Rebuild: `docker compose build` in `task-runner/`.
5. To add Prometheus scrape targets, edit `configs/prometheus.yml` (not the shared stack config).

## Integration

- **Observability:** OTLP traces/metrics → [OpenTelemetry Collector](../observability/opentelemetry-collector.md); logs → [Loki](../observability/loki.md) via Docker logging; metrics scraped by [Prometheus](../observability/prometheus.md) → [Mimir](../observability/mimir.md).
- **Alerting:** Mimir rules in `docker/observability/configs/mimir/rules/task-runner.yaml`.
- **CI:** Built and scanned in [Docker publish](../cicd/docker-publish.md) → `ghcr.io/danielsmithdevelopment/devsecops-boilerplate/task-runner`.
- **Kubernetes:** Example [NetworkPolicy](../kubernetes/networkpolicy.md) restricts egress.

## Official resources

- [Go](https://go.dev/doc/)
- [robfig/cron](https://github.com/robfig/cron) — cron scheduling
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Prometheus client_golang](https://github.com/prometheus/client_golang)
