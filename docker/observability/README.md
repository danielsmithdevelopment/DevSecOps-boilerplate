# Observability stack

Full Grafana LGTM+ platform with OpenTelemetry, profiling, frontend telemetry, load testing, incident response, optional runtime security tooling, and a **JWT-gated Cloudflare Worker** that closes the static public DSN / open Faro collector attack vector.

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
| **Frontend RUM** | Faro → Worker → Alloy (`:8027`) | Browser telemetry via short-lived JWT proxy |
| **Ingest proxy** | [`worker/`](./worker/) | Cloudflare Worker: JWT, rate limit, fingerprint, symbolication |
| **Browser SDK helpers** | [`faro/`](./faro/) | Faro init + fingerprint + token refresh |
| **Load / synthetic** | k6 (`--profile loadtest` / `synthetic`) | Prometheus remote-write to Mimir |
| **Alerting** | Alertmanager, Loki/Mimir rules, Grafana OnCall | Route pages by severity |
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
| http://localhost:8027 | Faro collector (Alloy — keep private; public traffic goes through the Worker) |
| http://localhost:5601 | Wazuh dashboard (security overlay) |

### JWT-gated frontend ingest (production pattern)

```bash
cd worker && npm install
# configure wrangler.toml ALLOY_INGEST_URL / ALLOWED_ORIGINS / PROJECT_ID
echo "$JWT_SIGNING_KEY" | npx wrangler secret put JWT_SIGNING_KEY
npx wrangler deploy
```

In the browser, use the Faro helpers:

```ts
import { initGatedFaro } from './faro/src/faro-init';

await initGatedFaro({
  workerUrl: 'https://telemetry-ingest.workers.dev/collect',
  tokenUrl: '/api/telemetry/token',
  app: { name: 'frontend', version: '1.0.0' },
});
```

Local compose without a Worker can still point Faro at Alloy directly (not for public internet):

```javascript
import { initializeFaro } from '@grafana/faro-web-sdk';

initializeFaro({
  url: 'http://localhost:8027/collect',
  app: { name: 'wallet-auth-frontend', version: '1.0.0' },
});
```

### With task-runner

```bash
cd task-runner
cp .env.example .env
docker compose up -d
docker compose --profile loadtest run --rm k6
# or: docker compose --profile synthetic run --rm k6-synthetic
```

### Security overlay (Linux + eBPF)

```bash
docker compose -f docker/observability/docker-compose.yaml up -d
docker compose -f docker/observability/docker-compose.security.yaml up -d
```

**Falco → Falcosidekick → Falco Talon** automates response to runtime threats. **Tetragon** ships Monitor-mode policies (`policy-no-package-install.yaml`, `policy-no-exfil.yaml`) — flip `Post` → `Sigkill` after baselining. **Wazuh** mounts `configs/security/wazuh/ossec.conf` for FIM + syslog + vulnerability detection.

## OnCall setup

After first boot, enable the OnCall plugin in Grafana (Plugins → Grafana OnCall → Enable). Configure an Alertmanager integration in OnCall and verify webhook delivery from Alertmanager.

> OnCall OSS is archived (2026-03-24). Suitable for dev/boilerplate; use Grafana Cloud IRM for production.

## Pinned images

All images use explicit version tags (no `:latest`). See `docker-compose.yaml` for the full list.

## Architecture

```
Browser ──Faro + JWT──► Cloudflare Worker ──► Alloy Faro (:8027)
                              │                 ├──► Loki (errors + fingerprints)
                              │                 └──► Tempo
Apps ──OTLP──► OTel Collector / Alloy ──┬──► Tempo
                                        ├──► Mimir  (via delta→cumulative + labeldrop)
                                        └──► Loki
Beyla (eBPF) ──OTLP──► Collector
Falco / Tetragon / Wazuh ──► Alloy syslog :1516 / stdout ──► Loki
Prometheus ──remote_write──► Mimir ──rules──► Alertmanager ──► OnCall / PagerDuty
```

## Config layout

```
docker/observability/
├── docker-compose.yaml           # core stack
├── docker-compose.security.yaml  # Falco, Talon, Tetragon, Wazuh
├── worker/                       # Cloudflare Worker (JWT-gated ingest)
├── faro/                         # browser Faro init + fingerprinting
└── configs/
    ├── alloy/                    # logs, Faro, OTLP, cardinality controls
    ├── loki/                     # retention footgun fix + LogQL alert rules
    ├── tempo/ mimir/ pyroscope/
    ├── prometheus/ alertmanager/
    ├── otel-collector/
    ├── grafana/                  # datasources, dashboards, contact points
    ├── k6/                       # loadtest.js + synthetic-check.js
    └── security/                 # falco, falco-talon, tetragon, wazuh/ossec.conf
```

Legacy per-app configs under `task-runner/configs/` and `docker/grafana-stack/` are superseded by this shared stack; app compose files only override Prometheus scrape targets.
