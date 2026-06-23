# Pyroscope

> Continuous profiling for CPU, memory, and other resource usage.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/pyroscope:1.13.1` |
| Config | `docker/observability/configs/pyroscope/pyroscope.yaml` |
| Port | `4040` |
| Compose service | `pyroscope` |

## Quick start

```bash
curl http://localhost:4040/ready
```

View profiles in Grafana (Pyroscope plugin) or Explore → Profiles.

## Configuration

`pyroscope.yaml` sets filesystem storage:

```yaml
storage:
  filesystem:
    dir: /data/pyroscope
```

## Making changes

1. Edit `configs/pyroscope/pyroscope.yaml`.
2. To profile Go apps, add Pyroscope SDK or use Grafana's `traceToProfiles` correlation.
3. Restart: `docker compose restart pyroscope`.

## Integration

- Grafana datasource and `grafana-pyroscope-app` plugin
- `GF_FEATURE_TOGGLES_ENABLE=traceToProfiles` links [Tempo](tempo.md) traces to profiles

## Official resources

- [Grafana Pyroscope docs](https://grafana.com/docs/pyroscope/latest/)
- [Pyroscope Go SDK](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/go/)
