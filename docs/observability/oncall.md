# Grafana OnCall

> Open-source on-call management and alert routing (OSS, archived 2026-03-24).

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/oncall:v1.16.11` |
| Port | `8090` (engine UI) |
| Compose services | `oncall-engine`, `oncall-celery`, `oncall-redis`, `oncall-rabbitmq`, `oncall-db` |

Suitable for dev/boilerplate. Use [Grafana Cloud IRM](https://grafana.com/products/cloud/irm/) for production.

## Quick start

```bash
open http://localhost:8090
```

After stack boot:

1. Enable OnCall plugin in [Grafana](grafana.md): Plugins → Grafana OnCall → Enable
2. Create an Alertmanager integration in OnCall
3. Verify [Alertmanager](alertmanager.md) webhook delivery

## Configuration

| Variable | Description |
|----------|-------------|
| `ONCALL_SECRET_KEY` | Django secret (≥32 chars) |
| `ONCALL_BASE_URL` | `http://localhost:8090` |
| `DATABASE_TYPE` | `sqlite3` (hobby/dev mode) |
| `BROKER_TYPE` | `redis` |

Set in `docker/observability/.env` (see `.env.example`).

## Making changes

1. Update `.env` for `ONCALL_SECRET_KEY` and `ONCALL_BASE_URL`.
2. For production-like setup, switch `DATABASE_TYPE` to PostgreSQL and use external Redis/RabbitMQ.
3. Restart: `docker compose restart oncall-engine oncall-celery`.

## Integration

- Receives alerts from [Alertmanager](alertmanager.md)
- Grafana plugin for schedules, escalations, and incident UI

## Official resources

- [Grafana OnCall docs](https://grafana.com/docs/oncall/latest/)
- [Alertmanager integration](https://grafana.com/docs/oncall/latest/configure/integrations/references/alertmanager/)
- [OSS archive notice](https://github.com/grafana/oncall)
