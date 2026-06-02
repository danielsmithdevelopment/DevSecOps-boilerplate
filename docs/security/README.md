# Security documentation

Guides aligned with the [ClawQL security curriculum](https://docs.clawql.com/security/best-practices) and this repo's CI controls.

| Document | Purpose |
| -------- | ------- |
| [VULN_SLA.md](VULN_SLA.md) | Vulnerability triage, SLAs, and waiver process (`osv-scanner.toml`, `.trivyignore`) |
| [THREAT_MODEL.md](THREAT_MODEL.md) | STRIDE threat model for task-runner and wallet-auth |
| [QUARTERLY_REVIEW.md](QUARTERLY_REVIEW.md) | Quarterly posture checklist |

## CI and supply chain

- **PR CI** (`.github/workflows/ci.yml`): Gitleaks → lint → Go/Python/frontend → OSV + Syft SBOM + Trivy
- **CodeQL** (`.github/workflows/codeql.yml`): SAST for Go, Python, TypeScript
- **Docker publish** (`.github/workflows/docker-publish.yml`): build → Trivy → GHCR push → Cosign sign/verify
- **TruffleHog** (`.github/workflows/trufflehog-scheduled.yml`): weekly git history scan

## Local developer setup

```bash
pip install pre-commit && pre-commit install
cp task-runner/.env.example task-runner/.env
cp docker/grafana-stack/.env.example docker/grafana-stack/.env
```

## Kubernetes examples

See [`kubernetes/examples/`](../kubernetes/examples/) for sample NetworkPolicy and Kyverno Cosign verification policies.
