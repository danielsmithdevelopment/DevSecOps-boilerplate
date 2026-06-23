# GitHub Actions CI

> PR and push pipeline with `ci-gate` branch protection.

## In this repo

| Path | Purpose |
|------|---------|
| `.github/workflows/ci.yml` | Main CI workflow |
| Branch protection | `main` requires **`ci-gate`** |

## Pipeline overview

```
secret-scan (Gitleaks)
  └── quick (yamllint, ruff, packer fmt, shellcheck)
        ├── terraform (fmt + tflint)
        ├── ansible (ansible-lint)
        ├── go (5 modules: lint + test)
        ├── frontend (npm audit + build)
        └── compose-images → trivy-compose-images (22 images)

go + terraform + ansible + frontend
  ├── supply-chain (OSV + SBOM + compose image OSV)
  └── trivy-fs

go → trivy-docker (3 Dockerfiles)

All → ci-gate
```

## Jobs

| Job | What it checks |
|-----|----------------|
| `secret-scan` | [Gitleaks](gitleaks.md) on full git history |
| `quick` | YAML, Python, Packer format, ShellCheck, no `:latest` in compose |
| `go` | gofmt, golangci-lint, `go test` (5 modules) |
| `frontend` | `npm audit` + build (wallet-auth) |
| `terraform` | `terraform fmt` + tflint |
| `ansible` | ansible-lint |
| `supply-chain` | [OSV-Scanner](osv-scanner.md) + SBOM + compose image OSV |
| `trivy-fs` | [Trivy](trivy.md) filesystem (vuln, secret, misconfig) |
| `trivy-docker` | Trivy on built images (strict) |
| `trivy-compose-images` | Trivy on 22 pinned compose images (report-only) |
| `ci-gate` | Fails if any required job fails |

## Quick start

CI runs automatically on PRs and pushes to `main`. To reproduce locally:

```bash
# Go
cd task-runner && go test ./... && golangci-lint run ./...

# YAML
yamllint -c .yamllint.yml .

# Secrets
pip install pre-commit && pre-commit run gitleaks --all-files
```

## Making changes

1. Edit `.github/workflows/ci.yml`.
2. Add new jobs to `ci-gate` `needs` and `check` steps.
3. Pin action versions (no floating `@main`).
4. Test on a PR before merging workflow changes.

## Integration

- Complements [CodeQL](codeql.md), [Docker publish](docker-publish.md), [Dependabot](dependabot.md)
- Scan policy: [VULN_SLA.md](../security/VULN_SLA.md)

## Official resources

- [GitHub Actions docs](https://docs.github.com/en/actions)
- [Branch protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)
