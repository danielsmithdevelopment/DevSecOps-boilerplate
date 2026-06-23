# Security documentation

Guides aligned with the [ClawQL security curriculum](https://docs.clawql.com/security/best-practices) and this repo's CI controls.

## Program documents

| Document | Purpose |
| -------- | ------- |
| [VULN_SLA.md](VULN_SLA.md) | Vulnerability triage, SLAs, and waiver process (`osv-scanner.toml`, `.trivyignore`) |
| [THREAT_MODEL.md](THREAT_MODEL.md) | STRIDE threat model for task-runner and wallet-auth |
| [QUARTERLY_REVIEW.md](QUARTERLY_REVIEW.md) | Quarterly posture checklist |

## Runtime security tools

Optional overlay — see [observability security compose](../../docker/observability/docker-compose.security.yaml).

| Tool | Doc |
|------|-----|
| Falco | [falco.md](falco.md) |
| Falcosidekick | [falcosidekick.md](falcosidekick.md) |
| Falco Talon | [falco-talon.md](falco-talon.md) |
| Tetragon | [tetragon.md](tetragon.md) |
| Wazuh | [wazuh.md](wazuh.md) |

## CI and supply chain

| Tool | Doc |
|------|-----|
| GitHub Actions CI | [../cicd/github-actions-ci.md](../cicd/github-actions-ci.md) |
| Docker publish + Cosign | [../cicd/docker-publish.md](../cicd/docker-publish.md) |
| OSV-Scanner | [../cicd/osv-scanner.md](../cicd/osv-scanner.md) |
| Trivy | [../cicd/trivy.md](../cicd/trivy.md) |
| CodeQL | [../cicd/codeql.md](../cicd/codeql.md) |
| Gitleaks | [../cicd/gitleaks.md](../cicd/gitleaks.md) |
| Dependabot | [../cicd/dependabot.md](../cicd/dependabot.md) |
| pre-commit | [../cicd/pre-commit.md](../cicd/pre-commit.md) |

**TruffleHog** (`.github/workflows/trufflehog-scheduled.yml`): weekly git history scan — no dedicated page; see [Gitleaks](gitleaks.md) for secret scanning.

## Local developer setup

```bash
pip install pre-commit && pre-commit install
cp task-runner/.env.example task-runner/.env
cp docker/observability/.env.example docker/observability/.env
```

## Kubernetes examples

| Example | Doc |
|---------|-----|
| NetworkPolicy | [../kubernetes/networkpolicy.md](../kubernetes/networkpolicy.md) |
| Kyverno Cosign verify | [../kubernetes/kyverno.md](../kubernetes/kyverno.md) |

Full documentation index: [../README.md](../README.md)
