# Trivy

> Comprehensive scanner for vulnerabilities, secrets, and misconfigurations.

## In this repo

| Path | Purpose |
|------|---------|
| `.trivyignore` | Vulnerability waivers |
| `.github/workflows/ci.yml` | fs, docker, compose image scans |
| `.github/workflows/docker-publish.yml` | Image scan before push |

## Scan modes

| Job | Target | Severity | Gate |
|-----|--------|----------|------|
| `trivy-fs` | Entire repository | CRITICAL, HIGH, MEDIUM | **Strict** |
| `trivy-docker` | 3 built Dockerfiles | CRITICAL, HIGH, MEDIUM | **Strict** |
| `trivy-compose-images` | 22 pinned compose images | CRITICAL, HIGH, MEDIUM (OS packages only) | **Report-only** |

All scans use `ignore-unfixed: true`.

## Quick start

```bash
# Filesystem
docker run --rm -v $(pwd):/repo aquasec/trivy:0.58.1 fs \
  --severity CRITICAL,HIGH,MEDIUM --exit-code 1 /repo

# Image
docker run --rm aquasec/trivy:0.58.1 image \
  --severity CRITICAL,HIGH,MEDIUM grafana/loki:3.3.2
```

## Configuration

### Waivers (`.trivyignore`)

```
CVE-2024-12345
```

Document rationale per [VULN_SLA.md](../security/VULN_SLA.md).

### Compose image policy

Third-party pinned images are scanned for visibility but do not block `ci-gate` (`exit-code: "0"`). Repo code and built images remain strictly gated.

## Making changes

1. Fix findings by upgrading dependencies or base images.
2. Add waivers to `.trivyignore` with PR documentation.
3. Image list: `scripts/list-compose-images.sh`.

## Integration

- [GitHub Actions CI](github-actions-ci.md)
- [Docker publish](docker-publish.md)
- Complements [OSV-Scanner](osv-scanner.md)

## Official resources

- [Trivy documentation](https://aquasecurity.github.io/trivy/)
- [Trivy GitHub Action](https://github.com/aquasecurity/trivy-action)
- [Vulnerability DB](https://github.com/aquasecurity/trivy-db)
