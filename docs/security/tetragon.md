# Tetragon

> Cilium eBPF-based security observability — process execution and network events.

## In this repo

| Item | Value |
|------|-------|
| Image | `quay.io/cilium/tetragon:v1.4.0` |
| Policy | `docker/observability/configs/security/tetragon/tracing-policy.yaml` |
| Metrics port | `2112` |
| Compose file | `docker/observability/docker-compose.security.yaml` |

## Quick start

```bash
docker compose -f docker/observability/docker-compose.security.yaml up -d tetragon
```

Requires Linux with eBPF support and `privileged: true`.

## Configuration

`tracing-policy.yaml` defines which process/network events to capture. Edit to add:

- Process exec monitoring
- File access tracing
- Network connection logging

## Making changes

1. Edit `configs/security/tetragon/tracing-policy.yaml`.
2. Restart: `docker compose -f docker-compose.security.yaml restart tetragon`.
3. View events: `docker compose logs tetragon`
4. Metrics: `curl localhost:2112/metrics`

## Integration

- Complements [Falco](falco.md) (rule-based detection) with granular eBPF observability
- Part of the security overlay on the `observability` network

## Official resources

- [Tetragon docs](https://tetragon.io/docs/)
- [Tracing policies](https://tetragon.io/docs/concepts/tracing-policy/)
- [Cilium project](https://cilium.io/)
