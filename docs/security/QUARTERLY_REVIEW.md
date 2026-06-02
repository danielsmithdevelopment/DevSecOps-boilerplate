# Quarterly security review

Use this checklist each quarter. Record evidence links (CI run URLs, SBOM artifacts, restore test logs).

## Access and secrets

- [ ] Rotate `JWT_SECRET` and database credentials in non-dev environments
- [ ] Review GitHub org members and repo collaborators
- [ ] Confirm branch protection requires `ci-gate` on `main`
- [ ] Review Dependabot open PRs and merge or dismiss with rationale

## Supply chain

- [ ] Download latest `sbom-repository-cyclonedx` artifact from CI; spot-check critical deps
- [ ] Confirm `docker-publish` Cosign signatures verify for promoted GHCR digests
- [ ] Review `osv-scanner.toml` and `.trivyignore` waivers — remove expired entries
- [ ] Run TruffleHog scheduled workflow manually if weekly job missed

## Infrastructure

- [ ] `terraform plan` for drift in staging/production workspaces
- [ ] Ansible playbooks still apply cleanly on golden images
- [ ] Restore test from Postgres backup (task-runner or wallet-auth DB)

## Observability and IR

- [ ] Grafana dashboards load; alert routes tested
- [ ] Sample trace from app → OTel → Tempo visible in Grafana
- [ ] Tabletop exercise: one failed CI gate and one simulated secret leak

## Sign-off

| Role | Name | Date |
| ---- | ---- | ---- |
| Owner | | |
| Reviewer | | |
