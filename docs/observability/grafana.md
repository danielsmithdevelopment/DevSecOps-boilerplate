# Grafana

> Visualization, dashboards, Explore, and correlations across metrics, logs, traces, and profiles.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/grafana:11.4.0` |
| Datasources | `docker/observability/configs/grafana/provisioning/datasources/datasources.yaml` |
| Dashboards | `docker/observability/configs/grafana/dashboards/` |
| Dashboard provisioning | `configs/grafana/provisioning/dashboards/dashboards.yaml` |
| Port | `3000` |
| Compose service | `grafana` |

## Quick start

```bash
# Default credentials from .env
open http://localhost:3000
# user: admin  password: GF_SECURITY_ADMIN_PASSWORD from .env
```

## Configuration

### Environment (`.env`)

| Variable | Description |
|----------|-------------|
| `GF_SECURITY_ADMIN_USER` | Admin username (`admin`) |
| `GF_SECURITY_ADMIN_PASSWORD` | Admin password |

### Provisioned datasources

Auto-configured: Prometheus, Mimir, Loki, Tempo, Pyroscope.

### Plugins

Installed via `GF_INSTALL_PLUGINS`:

- `grafana-pyroscope-app` — profiling UI
- `grafana-oncall-app` — OnCall integration

### Feature toggles

`GF_FEATURE_TOGGLES_ENABLE=traceToProfiles,traceqlEditor`

## Making changes

1. **Datasources:** edit `configs/grafana/provisioning/datasources/datasources.yaml`.
2. **Dashboards:** add JSON to `configs/grafana/dashboards/` and register in `dashboards.yaml`.
3. **UI changes:** persist to provisioning files (not only in-container edits).
4. Restart: `docker compose restart grafana`.

### OnCall plugin

After first boot: Plugins → Grafana OnCall → Enable. Configure Alertmanager integration in OnCall UI.

## Integration

- Queries all observability backends (see [observability hub](README.md))
- OnCall plugin connects to [OnCall engine](oncall.md)
- Faro RUM data visible via [Loki](loki.md) / [Tempo](tempo.md)

## Official resources

- [Grafana documentation](https://grafana.com/docs/grafana/latest/)
- [Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)
- [Dashboard JSON model](https://grafana.com/docs/grafana/latest/dashboards/json-model/)
