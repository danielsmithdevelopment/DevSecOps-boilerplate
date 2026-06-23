# Dependabot

> Automated dependency update pull requests.

## In this repo

| Path | Purpose |
|------|---------|
| `.github/dependabot.yml` | Ecosystem configuration |

## Ecosystems monitored

| Ecosystem | Path |
|-----------|------|
| Go modules | `/`, `task-runner/`, `wallet-auth/backend/`, etc. |
| npm | `wallet-auth/frontend/` |
| pip | `pulumi/aws/prod/function/` |
| Terraform | `terraform/linode/` |
| GitHub Actions | `.github/workflows/` |

## Quick start

Dependabot opens PRs weekly. Merge only when **`ci-gate`** passes.

```bash
gh pr list --author app/dependabot
```

## Configuration

Edit `.github/dependabot.yml` to adjust:

- `schedule.interval` — weekly default
- `open-pull-requests-limit`
- `groups` — batch related updates

## Making changes

1. Review Dependabot PRs like any other PR.
2. If an update breaks CI, pin or waive per [VULN_SLA.md](../security/VULN_SLA.md).
3. Add new ecosystems when adding `go.mod`, `package.json`, etc.

## Integration

- All PRs must pass [GitHub Actions CI](github-actions-ci.md)
- Complements [OSV-Scanner](osv-scanner.md) and [Trivy](trivy.md)

## Official resources

- [Dependabot docs](https://docs.github.com/en/code-security/dependabot)
- [Configuration reference](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)
