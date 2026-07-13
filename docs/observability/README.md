# Observability stack

> Shared Grafana LGTM+ platform with OpenTelemetry, profiling, RUM, load testing, and incident routing.

## Architecture

```
Browser ──Faro + JWT──► Cloudflare Worker ──► Alloy Faro (:8027)
                                                ├──► Loki / Tempo
Apps ──OTLP──► OTel Collector ──┬──► Tempo (traces)
                                ├──► Mimir  (metrics)
                                └──► Loki   (logs)

Browser (local only) ──Faro──► Alloy ────────┬──► Tempo / Loki
Docker logs ──► Alloy ───────────────────────┘
Security syslog ──► Alloy :1516 ─────────────► Loki

Beyla (eBPF) ──OTLP──► OTel Collector

Prometheus ──remote_write──► Mimir ──rules──► Alertmanager ──► OnCall
Loki ruler (LogQL) ─────────────────────────► Alertmanager
```

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
| http://localhost:4040 | Pyroscope |
| http://localhost:3100 | Loki API |
| http://localhost:3200 | Tempo API |
| http://localhost:9009 | Mimir API |

### With an application

```bash
cd task-runner && cp .env.example .env && docker compose up -d
./scripts/e2e-observability-stack.sh
```

### Security overlay (Linux + eBPF)

```bash
docker compose -f docker/observability/docker-compose.yaml up -d
docker compose -f docker/observability/docker-compose.security.yaml up -d
```

## Components

| Tool | Doc | Image |
|------|-----|-------|
| Prometheus | [prometheus.md](prometheus.md) | `prom/prometheus:v3.2.1` |
| Mimir | [mimir.md](mimir.md) | `grafana/mimir:2.14.3` |
| Loki | [loki.md](loki.md) | `grafana/loki:3.3.2` |
| Tempo | [tempo.md](tempo.md) | `grafana/tempo:2.6.1` |
| Grafana | [grafana.md](grafana.md) | `grafana/grafana:11.4.0` |
| Pyroscope | [pyroscope.md](pyroscope.md) | `grafana/pyroscope:1.13.1` |
| OTel Collector | [opentelemetry-collector.md](opentelemetry-collector.md) | `otel/opentelemetry-collector-contrib:0.116.1` |
| Alloy | [alloy.md](alloy.md) | `grafana/alloy:v1.17.0` |
| Beyla | [beyla.md](beyla.md) | `grafana/beyla:2.0.5` |
| Alertmanager | [alertmanager.md](alertmanager.md) | `prom/alertmanager:v0.28.1` |
| OnCall | [oncall.md](oncall.md) | `grafana/oncall:v1.16.11` |
| k6 | [k6.md](k6.md) | `grafana/k6:0.57.0` |

## Config layout

```
docker/observability/
├── docker-compose.yaml           # core stack
├── docker-compose.security.yaml  # Falco, Talon, Tetragon, Wazuh
├── worker/                       # Cloudflare Worker (JWT-gated Faro proxy)
├── faro/                         # browser Faro init + fingerprinting
├── .env.example
└── configs/
    ├── alloy/ loki/ tempo/ mimir/ pyroscope/
    ├── prometheus/ alertmanager/
    ├── otel-collector/
    ├── grafana/                  # datasources + dashboards + alerting
    ├── k6/                       # loadtest + synthetic-check
    └── security/                 # falco, falco-talon, tetragon, wazuh
```

## Compose includes

These app stacks pull in the full observability platform:

- `task-runner/docker-compose.yml`
- `docker/grafana-stack/docker-compose.yaml`
- `development/go/worker-cluster/docker-compose.yaml`

Each overrides only the Prometheus scrape config for app-specific targets.

## Making changes

1. Edit configs under `docker/observability/configs/`.
2. Bump image tags in `docker-compose.yaml` (no `:latest` — CI enforces pinned tags).
3. Restart affected services: `docker compose up -d <service>`.
4. Run E2E: `./scripts/e2e-observability-stack.sh`.
5. CI scans all pinned images — see [Trivy](../cicd/trivy.md) and [OSV-Scanner](../cicd/osv-scanner.md).

## Official resources

- [Grafana observability documentation](https://grafana.com/docs/)
- [OpenTelemetry](https://opentelemetry.io/docs/)
- [Prometheus](https://prometheus.io/docs/)
