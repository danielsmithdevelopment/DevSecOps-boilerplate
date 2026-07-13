# Wazuh

> SIEM and host intrusion detection (simplified single-node dev layout).

## In this repo

| Item | Value |
|------|-------|
| Images | `wazuh/wazuh-indexer:4.11.2`, `wazuh-manager:4.11.2`, `wazuh-dashboard:4.11.2` |
| Port | `5601` (dashboard) |
| Compose file | `docker/observability/docker-compose.security.yaml` |
| Manager config | `docker/observability/configs/security/wazuh/ossec.conf` |

> This is a **dev/boilerplate layout**. Production deployments should use [wazuh-docker](https://github.com/wazuh/wazuh-docker) multi-node architecture.

## Quick start

```bash
docker compose -f docker/observability/docker-compose.security.yaml up -d
open http://localhost:5601
```

Default credentials are set in compose (change for any non-local use).

## Configuration (`ossec.conf`)

The mounted config focuses on the capabilities used in the Sentry-replacement stack:

| Capability | What is configured |
|------------|--------------------|
| **File Integrity Monitoring** | Realtime FIM on `/etc`, `/usr/bin`, `/usr/sbin`, Vault/SSL key material paths |
| **Log analysis** | `auth.log`, `syslog`, optional Vault audit JSON |
| **Vulnerability detection** | Enabled with 60m feed update interval |
| **Remote syslog** | Agent TCP `1514` + UDP syslog `514` |
| **Active response** | Present but commented out — enable after confirming host firewall tooling |

Compose mounts the file via Wazuh's config-mount path:

```yaml
volumes:
  - ./configs/security/wazuh/ossec.conf:/wazuh-config-mount/etc/ossec.conf:ro
```

## Making changes

1. Edit `configs/security/wazuh/ossec.conf`.
2. Restart: `docker compose -f docker-compose.security.yaml restart wazuh-manager`.
3. Add Wazuh agents on hosts to report to the manager.
4. For production, migrate to official wazuh-docker templates.

## Integration

- Runs alongside the [observability stack](../observability/README.md) on the `observability` network
- Complements [Falco](falco.md) / [Tetragon](tetragon.md) runtime detection with log analysis and FIM
- Alloy syslog for Falco-style events uses **`:1516`** so it does not collide with Wazuh agent port `:1514`

## Official resources

- [Wazuh documentation](https://documentation.wazuh.com/)
- [Wazuh Docker deployment](https://documentation.wazuh.com/current/deployment-options/docker/index.html)
- [Wazuh ruleset](https://documentation.wazuh.com/current/user-manual/ruleset/index.html)
- [FIM documentation](https://documentation.wazuh.com/current/user-manual/capabilities/file-integrity/index.html)
