# Gitleaks

> Secret scanning for git repositories.

## In this repo

| Path | Purpose |
|------|---------|
| `.gitleaks.toml` | Rules and allowlists |
| `.github/workflows/ci.yml` | `secret-scan` job (first in pipeline) |
| `.pre-commit-config.yaml` | Local pre-commit hook |

## Quick start

```bash
# CI equivalent
docker run --rm -v $(pwd):/repo zricethezav/gitleaks:latest \
  detect --source /repo --config /repo/.gitleaks.toml

# Pre-commit
pip install pre-commit && pre-commit install
pre-commit run gitleaks --all-files
```

## Configuration

`.gitleaks.toml`:

- **`useDefault = true`** — standard secret patterns
- **`allowlist.paths`** — `.env.example`, `docs/security/`, etc.
- **`allowlist.regexes`** — documented dev placeholders (`your-secret-key-here`, etc.)

## Making changes

1. Never commit real secrets — use `.env.example` templates.
2. Add allowlist entries only for documented placeholders.
3. If a finding is a false positive, add to `allowlist` with a comment explaining why.

## Integration

- First job in [GitHub Actions CI](github-actions-ci.md) — blocks all downstream jobs
- Weekly [TruffleHog](https://github.com/danielsmithdevelopment/DevSecOps-boilerplate/blob/main/.github/workflows/trufflehog-scheduled.yml) supplements git history scanning

## Official resources

- [Gitleaks](https://github.com/gitleaks/gitleaks)
- [Configuration](https://github.com/gitleaks/gitleaks#configuration)
