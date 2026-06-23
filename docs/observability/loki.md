# Loki

> Log aggregation system optimized for label-based queries.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/loki:3.3.2` |
| Config | `docker/observability/configs/loki/loki.yaml` |
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
- **Retention:** configured in `limits_config`
- **Auth:** disabled for local dev

Container logs reach Loki via [Alloy](alloy.md) (Docker log discovery) and via [OTel Collector](opentelemetry-collector.md) (OTLP logs).

## Making changes

1. Edit `configs/loki/loki.yaml`.
2. For retention or storage backend changes, update `storage_config` and `limits_config`.
3. Restart: `docker compose restart loki`.
4. After schema changes, you may need to clear the `loki-data` volume.

## Integration

- Log sources: Alloy (Docker + Faro), OTel Collector (app OTLP logs)
- Grafana datasource: `http://loki:3100`
- Correlate with [Tempo](tempo.md) traces via Grafana Explore

## Official resources

- [Grafana Loki docs](https://grafana.com/docs/loki/latest/)
- [LogQL](https://grafana.com/docs/loki/latest/query/)
- [Loki configuration reference](https://grafana.com/docs/loki/latest/configure/)
