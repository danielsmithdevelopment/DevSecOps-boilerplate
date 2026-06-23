# Wazuh

> SIEM and host intrusion detection (simplified single-node dev layout).

## In this repo

| Item | Value |
|------|-------|
| Images | `wazuh/wazuh-indexer:4.11.2`, `wazuh-manager:4.11.2`, `wazuh-dashboard:4.11.2` |
| Port | `5601` (dashboard) |
| Compose file | `docker/observability/docker-compose.security.yaml` |

> This is a **dev/boilerplate layout**. Production deployments should use [wazuh-docker](https://github.com/wazuh/wazuh-docker) multi-node architecture.

## Quick start

```bash
docker compose -f docker/observability/docker-compose.security.yaml up -d
open http://localhost:5601
```

Default credentials are set in compose (change for any non-local use).

## Configuration

Services in `docker-compose.security.yaml`:

- **wazuh-indexer** — OpenSearch backend
- **wazuh-manager** — analysis engine and agent management
- **wazuh-dashboard** — Kibana-based UI

## Making changes

1. Edit compose environment for passwords and cluster settings.
2. Add Wazuh agents on hosts to report to the manager.
3. For production, migrate to official wazuh-docker templates.

## Integration

- Runs alongside the [observability stack](../observability/README.md) on the `observability` network
- Complements [Falco](falco.md) / [Tetragon](tetragon.md) runtime detection with log analysis and FIM

## Official resources

- [Wazuh documentation](https://documentation.wazuh.com/)
- [Wazuh Docker deployment](https://documentation.wazuh.com/current/deployment-options/docker/index.html)
- [Wazuh ruleset](https://documentation.wazuh.com/current/user-manual/ruleset/index.html)
