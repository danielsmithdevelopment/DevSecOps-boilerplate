# Prometheus

> Short-term metrics scraping; remote-writes to Mimir for long-term storage.

## In this repo

| Item | Value |
|------|-------|
| Image | `prom/prometheus:v3.2.1` |
| Config | `docker/observability/configs/prometheus/prometheus.yaml` |
| Port | `9090` |
| Compose service | `prometheus` |

App stacks override the config volume with app-specific scrape files:

- `task-runner/configs/prometheus.yml`
- `docker/grafana-stack/prometheus-app.yml`
- `development/go/worker-cluster/configs/prometheus-app.yml`

## Quick start

```bash
curl http://localhost:9090/-/healthy
curl http://localhost:9090/api/v1/targets
```

## Configuration

Key settings in `prometheus.yaml`:

- **Scrape jobs:** observability stack self-monitoring + app `/metrics` endpoints
- **Remote write:** `http://mimir:9009/api/v1/push` — all scraped metrics forwarded to Mimir
- **Rule files:** `configs/mimir/rules/` mounted for evaluation

### Adding a scrape target

In an app's `prometheus-*.yml`, add a job:

```yaml
- job_name: my-app
  static_configs:
    - targets: ['my-app:8080']
```

Keep the `remote_write` and stack monitoring jobs from the base config.

## Making changes

1. Edit base config: `docker/observability/configs/prometheus/prometheus.yaml`
2. For app-specific targets, edit the app's override file (not the base).
3. Reload: `docker compose restart prometheus` or `curl -X POST localhost:9090/-/reload` (if enabled).

## Integration

- Scrapes → [Mimir](mimir.md) via remote_write
- Grafana queries Prometheus as a datasource (short-term) and Mimir (long-term)
- Alert rules evaluated against Mimir/Prometheus data → [Alertmanager](alertmanager.md)

## Official resources

- [Prometheus documentation](https://prometheus.io/docs/)
- [Remote write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)
- [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/)
