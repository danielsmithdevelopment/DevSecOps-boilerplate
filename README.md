# DevSecOps-boilerplate

Boilerplate automation for DevSecOps: golden images (Packer), configuration management (Ansible), infrastructure as code (Terraform, Pulumi), observability (Grafana LGTM+ with JWT-gated Faro ingest), runtime security, and GitHub Actions CI/CD with supply-chain gates.

**Documentation:** [docs/README.md](docs/README.md) — one page per tool with quick start, configuration, and official resource links.

**Observability starter:** [docker/observability/](docker/observability/) — Compose stack, Cloudflare Worker (`worker/`), Faro helpers (`faro/`), Tetragon/Falco/Wazuh overlay, and pre-built dashboards.

## Quick links

| Area | Start here |
|------|------------|
| Observability stack | [docs/observability/README.md](docs/observability/README.md) |
| JWT-gated telemetry Worker | [docs/applications/telemetry-ingest-worker.md](docs/applications/telemetry-ingest-worker.md) |
| Task runner | [docs/applications/task-runner.md](docs/applications/task-runner.md) |
| Security program | [docs/security/README.md](docs/security/README.md) |
| CI / supply chain | [docs/cicd/github-actions-ci.md](docs/cicd/github-actions-ci.md) |
