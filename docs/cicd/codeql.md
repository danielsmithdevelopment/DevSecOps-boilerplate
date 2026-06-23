# CodeQL

> GitHub-native static analysis for Go, Python, and TypeScript.

## In this repo

| Path | Purpose |
|------|---------|
| `.github/workflows/codeql.yml` | CodeQL workflow |

## Languages analyzed

| Language | Paths |
|----------|-------|
| Go | All Go modules |
| Python | `pulumi/aws/prod/function/` |
| JavaScript/TypeScript | `wallet-auth/frontend/` |

## Quick start

CodeQL runs automatically on PRs and weekly on `main`. Results appear in the **Security** tab → Code scanning alerts.

## Configuration

Default CodeQL queries for each language. To customize:

1. Add a `.github/codeql/codeql-config.yml`
2. Specify query packs or paths to exclude

## Making changes

1. Fix alerts in the Security tab or PR checks.
2. Suppress false positives via `// codeql[query-id]` comments (document why).
3. Add new languages by extending the matrix in `codeql.yml`.

## Integration

- Runs alongside [GitHub Actions CI](github-actions-ci.md) (separate workflow)
- Complements [Trivy](trivy.md) (dependency/vuln) and [Gitleaks](gitleaks.md) (secrets)

## Official resources

- [CodeQL docs](https://codeql.github.com/docs/)
- [Code scanning](https://docs.github.com/en/code-security/code-scanning)
- [CodeQL for Go](https://codeql.github.com/docs/codeql-language-guides/codeql-for-go/)
