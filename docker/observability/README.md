# Observability stack

Full Grafana LGTM+ platform with OpenTelemetry, profiling, frontend telemetry, load testing, incident response, and optional runtime security tooling.

> **Full documentation:** [docs/observability/](../../docs/observability/README.md) — one page per component with configuration and official resource links.

## Components

| Layer | Services | Purpose |
|-------|----------|---------|
| **Metrics** | Prometheus, Mimir | Scrape + long-term metrics storage |
| **Logs** | Loki, Alloy | Log aggregation (Alloy replaces Promtail) |
| **Traces** | Tempo, OTel Collector | Distributed tracing |
| **Profiles** | Pyroscope | Continuous profiling |
| **Visualization** | Grafana | Dashboards, Explore, correlations |
| **Auto-instrumentation** | Beyla | eBPF HTTP/gRPC metrics & traces |
| **Frontend RUM** | Alloy Faro receiver (`:8027`) | Browser telemetry from Grafana Faro SDK |
| **Load testing** | k6 (`--profile loadtest`) | Prometheus remote-write to Mimir |
| **Alerting** | Alertmanager, Mimir rules | Route alerts to OnCall |
| **Incidents** | Grafana OnCall OSS | On-call schedules & escalations |
| **Security** | Falco, Falco Talon, Tetragon, Wazuh | Runtime detection & response (optional overlay) |

## Quick start

```bash
cd docker/observability
cp .env.example .env
docker compose up -d
```

| URL | Service |
|-----|---------|
| http://localhost:3000 | Grafana (`admin` / `.env` password) |
| http://localhost:9090 | Prometheus |
| http://localhost:9093 | Alertmanager |
| http://localhost:8090 | OnCall engine |
| http://localhost:8027 | Faro collector (Alloy) |
| http://localhost:5601 | Wazuh dashboard (security overlay) |

### With task-runner

```bash
cd task-runner
cp .env.example .env
docker compose up -d
docker compose --profile loadtest run --rm k6
```

### Security overlay (Linux + eBPF)

```bash
docker compose -f docker/observability/docker-compose.yaml up -d
docker compose -f docker/observability/docker-compose.security.yaml up -d
```

**Falco → Falcosidekick → Falco Talon** automates response to runtime threats. **Tetragon** adds Cilium eBPF process/network observability. **Wazuh** provides SIEM/HIDS (simplified dev layout).

## Frontend telemetry (Faro)

Point the [Grafana Faro Web SDK](https://grafana.com/docs/grafana-cloud/monitor-applications/frontend-observability/) at Alloy:

```javascript
import { initializeFaro } from '@grafana/faro-web-sdk';

initializeFaro({
  url: 'http://localhost:8027/collect',
  app: { name: 'wallet-auth-frontend', version: '1.0.0' },
});
```

## OnCall setup

After first boot, enable the OnCall plugin in Grafana (Plugins → Grafana OnCall → Enable). Configure an Alertmanager integration in OnCall and verify webhook delivery from Alertmanager.

> OnCall OSS is archived (2026-03-24). Suitable for dev/boilerplate; use Grafana Cloud IRM for production.

## Pinned images

All images use explicit version tags (no `:latest`). See `docker-compose.yaml` for the full list.

## Architecture

```
Apps ──OTLP──► OTel Collector ──┬──► Tempo (traces)
                                ├──► Mimir  (metrics)
                                └──► Loki   (logs)

Browser ──Faro──► Alloy ────────┬──► Tempo / Loki
Docker logs ──► Alloy ──────────┘

Beyla (eBPF) ──OTLP──► OTel Collector

Prometheus ──remote_write──► Mimir ──rules──► Alertmanager ──► OnCall

Falco ──► Falcosidekick ──► Falco Talon (response actions)
Tetragon ──events──► stdout / metrics (:2112)
Wazuh manager ──► indexer ──► dashboard
```

## Config layout

```
docker/observability/
├── docker-compose.yaml           # core stack
├── docker-compose.security.yaml  # Falco, Talon, Tetragon, Wazuh
├── configs/
│   ├── alloy/                    # logs, Faro, secondary OTLP ingress
│   ├── loki/ tempo/ mimir/ pyroscope/
│   ├── prometheus/ alertmanager/
│   ├── otel-collector/
│   ├── grafana/                  # datasources + dashboards
│   ├── k6/
│   └── security/                 # falco, falco-talon, tetragon
```

Legacy per-app configs under `task-runner/configs/` and `docker/grafana-stack/` are superseded by this shared stack; app compose files only override Prometheus scrape targets.
