# Docker publish

> Build, scan, push, and Cosign-sign container images to GHCR.

## In this repo

| Path | Purpose |
|------|---------|
| `.github/workflows/docker-publish.yml` | Publish workflow |

## Images published

| Image | Dockerfile |
|-------|------------|
| `ghcr.io/danielsmithdevelopment/devsecops-boilerplate/task-runner` | `task-runner/Dockerfile` |
| `ghcr.io/danielsmithdevelopment/devsecops-boilerplate/grafana-stack` | `docker/grafana-stack/Dockerfile` |
| `ghcr.io/danielsmithdevelopment/devsecops-boilerplate/worker-cluster` | `development/go/worker-cluster/Dockerfile` |

Tags: `latest` and `sha-<commit>`.

## Trigger

Runs on push to `main` when Dockerfiles or the workflow change. Also available via `workflow_dispatch`.

## Pipeline

```
repo-supply-chain (OSV + Trivy fs)
  → build + push to GHCR
  → Trivy image scan
  → Cosign sign (keyless OIDC)
  → Cosign verify
```

## Quick start

### Verify a published image

```bash
cosign verify \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'https://github.com/danielsmithdevelopment/DevSecOps-boilerplate/' \
  ghcr.io/danielsmithdevelopment/devsecops-boilerplate/task-runner:latest
```

### Pull

```bash
docker pull ghcr.io/danielsmithdevelopment/devsecops-boilerplate/task-runner:latest
```

## Making changes

1. Edit Dockerfiles or `docker-publish.yml`.
2. Ensure images pass [CI Trivy scans](trivy.md) before merge.
3. [Kyverno](../kubernetes/kyverno.md) example policy enforces Cosign signatures in Kubernetes.

## Integration

- Images must pass repo [OSV](osv-scanner.md) and [Trivy](trivy.md) gates
- Signed with Cosign keyless (GitHub OIDC)
- Referenced in `kubernetes/examples/kyverno-verify-images.yaml`

## Official resources

- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Cosign](https://docs.sigstore.dev/cosign/overview/)
- [Sigstore keyless signing](https://docs.sigstore.dev/cosign/signing/signing_with_github_actions/)
