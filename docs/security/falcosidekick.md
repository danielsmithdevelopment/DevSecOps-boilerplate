# Falcosidekick

> Alert fan-out proxy for Falco — routes events to webhooks, chat, and response tools.

## In this repo

| Item | Value |
|------|-------|
| Image | `falcosecurity/falcosidekick:2.31.0` |
| Port | `2801` |
| Compose file | `docker/observability/docker-compose.security.yaml` |

## Quick start

Starts automatically with the security overlay. Receives Falco alerts and forwards to configured outputs.

```bash
docker compose -f docker/observability/docker-compose.security.yaml up -d falcosidekick
curl http://localhost:2801/health
```

## Configuration

Environment in `docker-compose.security.yaml`:

| Variable | Description |
|----------|-------------|
| `FALCOSIDEKICK_CUSTOMFIELDS` | `source:falco` |
| `TALON_ADDRESS` | `http://falco-talon:2803` — forward to [Falco Talon](falco-talon.md) |

Add outputs (Slack, PagerDuty, etc.) via Falcosidekick env vars — see official docs.

## Making changes

1. Edit environment in `docker-compose.security.yaml`.
2. Restart: `docker compose -f docker-compose.security.yaml restart falcosidekick`.
3. Trigger a test Falco rule to verify delivery.

## Integration

- Receives from [Falco](falco.md)
- Forwards to [Falco Talon](falco-talon.md) for automated response

## Official resources

- [Falcosidekick docs](https://github.com/falcosecurity/falcosidekick)
- [Outputs reference](https://github.com/falcosecurity/falcosidekick#outputs)
