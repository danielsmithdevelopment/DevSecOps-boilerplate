# Grafana Alloy

> Telemetry collector for Docker logs, Grafana Faro RUM, syslog security events, and secondary OTLP ingress.

## In this repo

| Item | Value |
|------|-------|
| Image | `grafana/alloy:v1.17.0` |
| Config | `docker/observability/configs/alloy/config.alloy` |
| Ports | `12345` (UI), `8027` (Faro), `4319/4320` (OTLP), `1516` (syslog) |
| Compose service | `alloy` |

Requires Docker socket mount for container log discovery.

## Quick start

### Faro (frontend RUM)

**Production:** send Faro to the [JWT-gated Worker](../applications/telemetry-ingest-worker.md); the Worker forwards to Alloy.

**Local only** (do not expose `:8027` publicly):

```javascript
import { initializeFaro } from '@grafana/faro-web-sdk';

initializeFaro({
  url: 'http://localhost:8027/collect',
  app: { name: 'my-frontend', version: '1.0.0' },
});
```

### Alloy UI

http://localhost:12345 — component graph and config status.

## Configuration

`config.alloy` (River syntax) configures:

- **`loki.source.docker`** — ship container logs to Loki
- **`loki.source.syslog`** — security events on TCP `:1516` (avoids Wazuh `:1514`)
- **`faro.receiver`** — browser telemetry on `:8027`
- **`otelcol.receiver.otlp`** — secondary OTLP on 4319/4320
- **`otelcol.processor.transform`** — Delta → Cumulative metric temporality for Mimir
- **`prometheus.relabel`** — drop high-cardinality labels (`container_id`, `pod_template_hash`, `*_id`)
- Exporters to Loki, Tempo, Mimir

## Making changes

1. Edit `configs/alloy/config.alloy` (River language).
2. Validate syntax: `docker compose exec alloy alloy fmt /etc/alloy/config.alloy`
3. Restart: `docker compose restart alloy`.
4. Check UI at `:12345` for component health.

> Use stable `faro.receiver`, not the experimental `otelcol.receiver.faro`.

## Integration

- Docker logs → [Loki](loki.md)
- Faro RUM → [Loki](loki.md) + [Tempo](tempo.md)
- Replaces Promtail for log shipping in this stack

## Official resources

- [Grafana Alloy docs](https://grafana.com/docs/alloy/latest/)
- [Alloy River reference](https://grafana.com/docs/alloy/latest/reference/)
- [Grafana Faro Web SDK](https://grafana.com/docs/grafana-cloud/monitor-applications/frontend-observability/)
