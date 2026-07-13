# Tetragon

> Cilium eBPF-based security observability — process execution and network events, with optional enforcement.

## In this repo

| Item | Value |
|------|-------|
| Image | `quay.io/cilium/tetragon:v1.4.0` |
| Policies | `docker/observability/configs/security/tetragon/` |
| Metrics port | `2112` |
| Compose file | `docker/observability/docker-compose.security.yaml` |

### Policies

| File | Mode | Purpose |
|------|------|---------|
| `tracing-policy.yaml` | Monitor (`Post`) | Baseline shell exec observability |
| `policy-no-package-install.yaml` | Monitor (`Post`) | Detect `npm`/`npx`/`pip`/`yarn`/`pnpm` outside `ci`/`build` |
| `policy-no-exfil.yaml` | Monitor (`Post`) | Detect `curl`/`wget`/`nc`/`socat` outside allowlisted namespaces |

> **Start in Monitor mode.** Change `matchActions.action` from `Post` to `Sigkill` only after reviewing a week of events — overly broad enforcement kills legitimate processes.

## Quick start

```bash
docker compose -f docker/observability/docker-compose.security.yaml up -d tetragon
```

Requires Linux with eBPF support and `privileged: true`.

## Making changes

1. Edit YAML under `configs/security/tetragon/`.
2. Restart: `docker compose -f docker-compose.security.yaml restart tetragon`.
3. View events: `docker compose logs tetragon`
4. Metrics: `curl localhost:2112/metrics`

## Integration

- Complements [Falco](falco.md) (rule-based detection) with Kubernetes-aware eBPF policies
- Pairs with Langfuse / Loki alert `UnexpectedAgentToolUse` and `PackageManagerRuntimeEvent`
- Part of the security overlay on the `observability` network

## Official resources

- [Tetragon docs](https://tetragon.io/docs/)
- [Tracing policies](https://tetragon.io/docs/concepts/tracing-policy/)
- [Cilium project](https://cilium.io/)
