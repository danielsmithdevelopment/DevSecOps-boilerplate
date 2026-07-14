# Documentation

Single source of truth for every tool and component in this repository.

## Applications

Sample workloads that demonstrate DevSecOps patterns (observability, auth, distributed execution).

| Tool | Description |
|------|-------------|
| [task-runner](applications/task-runner.md) | Distributed cron task runner with JWT auth and PostgreSQL |
| [grafana-stack app](applications/grafana-stack.md) | Go data-ingestion demo wired to the observability stack |
| [worker-cluster](applications/worker-cluster.md) | Fiber HTTP demo app and k6 load-test target |
| [wallet-auth](applications/wallet-auth.md) | SIWE wallet + email auth (Go backend, React frontend) |
| [telemetry-ingest-worker](applications/telemetry-ingest-worker.md) | JWT-gated Cloudflare Worker + Faro helpers (closes public DSN vector) |

## Observability

Shared Grafana LGTM+ platform in `docker/observability/`. See the [observability hub](observability/README.md) for architecture and quick start.

| Tool | Role |
|------|------|
| [Prometheus](observability/prometheus.md) | Metrics scraping and short-term storage |
| [Mimir](observability/mimir.md) | Long-term metrics storage and alerting rules |
| [Loki](observability/loki.md) | Log aggregation |
| [Tempo](observability/tempo.md) | Distributed trace storage |
| [Grafana](observability/grafana.md) | Dashboards, Explore, correlations |
| [Pyroscope](observability/pyroscope.md) | Continuous profiling |
| [OpenTelemetry Collector](observability/opentelemetry-collector.md) | OTLP ingress for traces, metrics, logs |
| [Grafana Alloy](observability/alloy.md) | Docker log shipping, Faro RUM, secondary OTLP |
| [Beyla](observability/beyla.md) | eBPF auto-instrumentation |
| [Alertmanager](observability/alertmanager.md) | Alert routing to OnCall |
| [Grafana OnCall](observability/oncall.md) | On-call schedules and escalations |
| [k6](observability/k6.md) | Load testing with Mimir remote-write |

[E2E verification script](scripts/e2e-observability.md) — automated health checks for the full stack.

## Runtime security (optional overlay)

Linux/eBPF tooling in `docker/observability/docker-compose.security.yaml`.

| Tool | Role |
|------|------|
| [Falco](security/falco.md) | Runtime threat detection |
| [Falcosidekick](security/falcosidekick.md) | Falco alert fan-out |
| [Falco Talon](security/falco-talon.md) | Automated response actions |
| [Tetragon](security/tetragon.md) | Cilium eBPF process/network observability |
| [Wazuh](security/wazuh.md) | SIEM/HIDS (dev layout) |

## Security program

| Document | Purpose |
|----------|---------|
| [Security overview](security/README.md) | CI, supply chain, local setup |
| [Vulnerability SLA](security/VULN_SLA.md) | Triage, SLAs, waivers |
| [Threat model](security/THREAT_MODEL.md) | STRIDE for task-runner and wallet-auth |
| [Quarterly review](security/QUARTERLY_REVIEW.md) | Posture checklist |

## Infrastructure as Code

| Tool | Target |
|------|--------|
| [Packer](iac/packer.md) | Golden images (Linode VM, Docker Ubuntu base) |
| [Ansible](iac/ansible.md) | Configuration management for Packer builds |
| [Terraform](iac/terraform.md) | Linode cloud provisioning |
| [Pulumi](iac/pulumi.md) | AWS serverless (API Gateway + Lambda) |

## CI/CD and supply chain

| Tool | Role |
|------|------|
| [GitHub Actions CI](cicd/github-actions-ci.md) | Lint, test, OSV, Trivy, ci-gate |
| [Docker publish](cicd/docker-publish.md) | GHCR images with Cosign signing |
| [OSV-Scanner](cicd/osv-scanner.md) | Open-source vulnerability scanning |
| [Trivy](cicd/trivy.md) | Container and filesystem scanning |
| [CodeQL](cicd/codeql.md) | SAST for Go, Python, TypeScript |
| [Gitleaks](cicd/gitleaks.md) | Secret scanning |
| [Dependabot](cicd/dependabot.md) | Automated dependency updates |
| [pre-commit](cicd/pre-commit.md) | Local git hooks |

## Kubernetes examples

Reference policies (not deployed by default).

| Example | Purpose |
|---------|---------|
| [NetworkPolicy](kubernetes/networkpolicy.md) | Default-deny egress for task-runner |
| [Kyverno](kubernetes/kyverno.md) | Cosign image signature verification |
