# pre-commit

> Local git hooks for fast feedback before push.

## In this repo

| Path | Purpose |
|------|---------|
| `.pre-commit-config.yaml` | Hook definitions |

## Hooks enabled

| Hook | Tool |
|------|------|
| `gitleaks` | [Gitleaks](gitleaks.md) secret scan |
| `trailing-whitespace` | Whitespace cleanup |
| `end-of-file-fixer` | EOF newline |
| `check-yaml` | YAML syntax |
| `check-added-large-files` | Block large files |

## Quick start

```bash
pip install pre-commit
pre-commit install
pre-commit run --all-files   # first-time full scan
```

Hooks run automatically on `git commit`.

## Making changes

1. Edit `.pre-commit-config.yaml` to add hooks (pin versions).
2. Run `pre-commit autoupdate` periodically.
3. Keep hooks fast — heavy scans (Trivy, OSV) stay in CI.

## Integration

- Mirrors [Gitleaks](gitleaks.md) CI check locally
- Documented in [security README](../security/README.md)

## Official resources

- [pre-commit framework](https://pre-commit.com/)
- [Supported hooks](https://pre-commit.com/hooks.html)
