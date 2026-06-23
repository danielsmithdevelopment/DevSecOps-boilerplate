# Falco

> Runtime threat detection using kernel rules and eBPF.

## In this repo

| Item | Value |
|------|-------|
| Image | `falcosecurity/falco:0.41.0` |
| Config | `docker/observability/configs/security/falco/falco.yaml` |
| Custom rules | `configs/security/falco/rules.d/custom.yaml` |
| Port | `8765` (metrics/health) |
| Compose file | `docker/observability/docker-compose.security.yaml` |

Part of the **optional security overlay** — not started by default.

## Quick start

```bash
docker compose -f docker/observability/docker-compose.yaml up -d
docker compose -f docker/observability/docker-compose.security.yaml up -d
```

Requires **Linux** with kernel modules and `privileged: true`.

## Configuration

- `falco.yaml` — gRPC output, JSON logging, syscall source
- `rules.d/custom.yaml` — add custom detection rules
- Alerts forwarded to [Falcosidekick](falcosidekick.md) via gRPC

## Making changes

1. Add rules under `configs/security/falco/rules.d/`.
2. Test: `docker compose -f docker-compose.security.yaml restart falco`
3. View alerts: Falcosidekick logs or [Falco Talon](falco-talon.md) actions
4. E2E: `./scripts/e2e-observability-stack.sh --security` (Linux only)

## Integration

```
Falco → Falcosidekick → Falco Talon (automated response)
```

Shares the `observability` Docker network with the main stack.

## Official resources

- [Falco documentation](https://falco.org/docs/)
- [Falco rules](https://github.com/falcosecurity/rules)
- [Falco on Docker](https://falco.org/docs/getting-started/installation/#docker)
