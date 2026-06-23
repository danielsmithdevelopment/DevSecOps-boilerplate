# OSV-Scanner

> Open-source vulnerability scanning for dependencies and container images.

## In this repo

| Path | Purpose |
|------|---------|
| `osv-scanner.toml` | Waiver configuration |
| `.github/workflows/ci.yml` | Recursive repo scan + compose image scan |
| `.github/workflows/docker-publish.yml` | Pre-publish gate |

## What it scans

| Target | Command | Gate |
|--------|---------|------|
| Repository dependencies | `osv-scanner scan --recursive .` | **Strict** — fails on findings |
| Compose images (22) | `osv-scanner scan image <image>` | **Report-only** — warns, does not fail |

## Quick start

```bash
go install github.com/google/osv-scanner/v2/cmd/osv-scanner@latest

# Repo scan
osv-scanner scan --recursive --config=osv-scanner.toml .

# Image scan
osv-scanner scan image grafana/loki:3.3.2
```

## Configuration

### Waivers (`osv-scanner.toml`)

```toml
[[IgnoredVulns]]
id = "CVE-XXXX-YYYY"
reason = "..."
expires = "2026-12-31"
```

Follow the process in [VULN_SLA.md](../security/VULN_SLA.md).

## Making changes

1. Fix vulnerabilities by bumping dependencies (or [Dependabot](dependabot.md) PRs).
2. For accepted risk, add waivers to `osv-scanner.toml` with expiry.
3. Image list comes from `scripts/list-compose-images.sh`.

## Integration

- Part of [GitHub Actions CI](github-actions-ci.md) `supply-chain` job
- Runs before [Docker publish](docker-publish.md)

## Official resources

- [OSV-Scanner docs](https://google.github.io/osv-scanner/)
- [OSV database](https://osv.dev/)
- [Image scanning](https://google.github.io/osv-scanner/usage/scan-image)
