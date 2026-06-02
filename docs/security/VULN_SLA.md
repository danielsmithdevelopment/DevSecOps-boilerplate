# Vulnerability management SLA

This repo blocks **Critical** and **High** findings in CI (OSV, Trivy fs/image, `npm audit`). Medium findings in Trivy filesystem scans are also reported; address or waive with documented rationale.

## Response targets

| Severity | Action | Target |
| -------- | ------ | ------ |
| Critical | Fix or waive with expiry | 7 days |
| High | Fix or waive with expiry | 30 days |
| Medium | Track in backlog | 90 days |

## Waivers

1. Add an entry to `osv-scanner.toml` (`[[IgnoredVulns]]`) or `.trivyignore` with CVE/OSV ID.
2. Document **reason**, **risk acceptance**, and **expiry date** in the PR.
3. Review waivers in the [quarterly review](QUARTERLY_REVIEW.md).

## Dependency updates

- **Dependabot** opens weekly PRs for Go, npm, pip, Terraform, and GitHub Actions.
- Merge dependency PRs only when **`ci-gate`** passes.

## Image promotion

Images published to GHCR by `docker-publish.yml` must pass OSV/Trivy repo gates, image scan, and Cosign verify before `latest` moves.
