# Loki

> Log aggregation system optimized for label-based queries.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/loki:3.3.2` |
| Config | `docker/observability/configs/loki/loki.yaml` |
| Alert rules | `docker/observability/configs/loki/rules/fake/` |
| Port | `3100` |
| Compose service | `loki` |

## Quick start

```bash
curl http://localhost:3100/ready
```

Query logs in Grafana → Explore → Loki datasource.

## Configuration

`loki.yaml` uses:

- **Schema:** TSDB v13 with filesystem storage (single-node dev)
- **Retention:** `retention_period: 744h` (31 days) with **matching** `chunk_store_config.max_look_back_period: 744h`
- **Auth:** disabled for local dev; ruler tenant directory is `fake`
- **Ruler:** LogQL alerts for frontend fingerprints, agent tool use, and package-manager events

### Retention footgun

If `max_look_back_period` is shorter than `retention_period`, queries for older-but-still-retained data **silently return empty results**. Always keep them equal.

## Making changes

1. Edit `configs/loki/loki.yaml`.
2. Edit LogQL alert groups under `configs/loki/rules/fake/`.
3. Restart: `docker compose restart loki`.
4. After schema changes, you may need to clear the `loki-data` volume.

## Integration

- Log sources: Alloy (Docker + Faro + syslog `:1516`), OTel Collector (app OTLP logs)
- Grafana datasource: `http://loki:3100`
- Correlate with [Tempo](tempo.md) traces via Grafana Explore
- Frontend error grouping labels (`error_fingerprint`) come from the [Worker](../applications/telemetry-ingest-worker.md) / [Faro helpers](../../docker/observability/faro/)

## Official resources

- [Grafana Loki docs](https://grafana.com/docs/loki/latest/)
- [LogQL](https://grafana.com/docs/loki/latest/query/)
- [Loki configuration reference](https://grafana.com/docs/loki/latest/configure/)
