# Falco Talon

> Automated response engine for Falco alerts — executes actions like killing pods or pausing containers.

## In this repo

| Item | Value |
|------|-------|
| Image | `falcosecurity/falco-talon:0.2.1` |
| Rules | `docker/observability/configs/security/falco-talon/rules.yaml` |
| Port | `2803` |
| Compose file | `docker/observability/docker-compose.security.yaml` |

## Quick start

```bash
docker compose -f docker/observability/docker-compose.security.yaml up -d falco-talon
```

Receives events from [Falcosidekick](falcosidekick.md) at `TALON_ADDRESS`.

## Configuration

`rules.yaml` defines matchers and actions:

```yaml
# Example structure — see configs/security/falco-talon/rules.yaml
- action: ...
  match:
    rules: [...]
```

Actions can target Docker containers in this dev layout (Kubernetes actions require a K8s cluster).

## Making changes

1. Edit `configs/security/falco-talon/rules.yaml`.
2. Start with dry-run / log-only actions before enabling destructive responses.
3. Restart: `docker compose -f docker-compose.security.yaml restart falco-talon`.

## Integration

```
Falco → Falcosidekick → Falco Talon
```

## Official resources

- [Falco Talon docs](https://github.com/falcosecurity/falco-talon)
- [Rules reference](https://github.com/falcosecurity/falco-talon/blob/main/docs/rules.md)
