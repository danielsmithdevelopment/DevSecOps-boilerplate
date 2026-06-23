# Alertmanager

> Routes firing alerts from Mimir/Prometheus to notification receivers.

## In this repo

| Item | Value |
|------|-------|
| Image | `prom/alertmanager:v0.28.1` |
| Config | `docker/observability/configs/alertmanager/alertmanager.yaml` |
| Port | `9093` |
| Compose service | `alertmanager` |

## Quick start

```bash
curl http://localhost:9093/-/healthy
open http://localhost:9093
```

## Configuration

`alertmanager.yaml` defines:

- **Route tree** — group by alert, default receiver
- **Receivers** — webhook to [OnCall](oncall.md) engine

Example receiver:

```yaml
receivers:
  - name: oncall
    webhook_configs:
      - url: http://oncall-engine:8080/integrations/v1/alertmanager/
```

## Making changes

1. Edit `configs/alertmanager/alertmanager.yaml`.
2. Add receivers for Slack, PagerDuty, email, etc.
3. Restart: `docker compose restart alertmanager`.
4. Test with a manual alert or by triggering a Mimir rule.

## Integration

- Receives alerts from [Mimir](mimir.md) ruler and [Prometheus](prometheus.md)
- Default webhook → [Grafana OnCall](oncall.md)

## Official resources

- [Alertmanager docs](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [Configuration](https://prometheus.io/docs/alerting/latest/configuration/)
- [Notification templates](https://prometheus.io/docs/alerting/latest/notifications/)
