# Mimir

> Horizontally scalable long-term metrics storage and alerting rule evaluation.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/mimir:2.14.3` |
| Config | `docker/observability/configs/mimir/mimir.yaml` |
| Rules | `docker/observability/configs/mimir/rules/` |
| Port | `9009` (HTTP API, Prometheus-compatible) |
| Compose service | `mimir` |

## Quick start

```bash
curl http://localhost:9009/ready
curl 'http://localhost:9009/prometheus/api/v1/query?query=up'
```

## Configuration

`mimir.yaml` configures:

- **Filesystem storage** for blocks (dev/single-node layout)
- **Ruler** for alerting rule evaluation
- **Alertmanager** integration for firing alerts
- **Ingestion limits:** `ingestion_rate: 10000`, `ingestion_burst_size: 200000`, `compactor_blocks_retention_period: 90d`, `max_label_names_per_series: 30`

Watch `cortex_discarded_samples_total` from day one — Beyla + high-cardinality labels can hit the rate limit silently. Use the **High cardinality** Grafana dashboard.

### Alert rules

Rules live in `configs/mimir/rules/` (PromQL). LogQL frontend/security alerts live under `configs/loki/rules/` — see [Loki](loki.md).

## Making changes

1. Edit `configs/mimir/mimir.yaml` for storage, limits, or replication settings.
2. Add rule files under `configs/mimir/rules/`.
3. Restart: `docker compose restart mimir`.
4. Verify in Grafana → Alerting or `curl localhost:9009/prometheus/api/v1/rules`.

## Integration

- Receives metrics from [Prometheus](prometheus.md) remote_write and [OTel Collector](opentelemetry-collector.md)
- k6 remote-writes load test metrics directly to Mimir
- Firing alerts → [Alertmanager](alertmanager.md) → [OnCall](oncall.md)
- Grafana datasource: `http://mimir:9009/prometheus`

## Official resources

- [Grafana Mimir docs](https://grafana.com/docs/mimir/latest/)
- [Mimir alerting](https://grafana.com/docs/mimir/latest/operators-guide/configure-alertmanager/)
- [Recording and alerting rules](https://grafana.com/docs/mimir/latest/operators-guide/configure/rules/)
