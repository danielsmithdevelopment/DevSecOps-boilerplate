# Kyverno

> Kubernetes policy engine — Cosign image signature verification example.

## In this repo

| Path | Purpose |
|------|---------|
| `kubernetes/examples/kyverno-verify-images.yaml` | ClusterPolicy requiring GHCR Cosign signatures |
| `kubernetes/examples/README.md` | Brief index |

## What it does

Blocks pod creation unless the container image is signed by this repo's [Docker publish](../cicd/docker-publish.md) workflow (Cosign keyless via GitHub OIDC).

## Quick start

```bash
# Install Kyverno first: https://kyverno.io/docs/installation/
kubectl apply -f kubernetes/examples/kyverno-verify-images.yaml
```

Test with a signed image:

```bash
kubectl run test --image=ghcr.io/danielsmithdevelopment/devsecops-boilerplate/task-runner:latest
```

## Configuration

The policy verifies:

- Registry: `ghcr.io/danielsmithdevelopment/devsecops-boilerplate/*`
- Certificate issuer: `https://token.actions.githubusercontent.com`
- Certificate identity matches the GitHub repo

Update `match` rules for your registry and repo.

## Making changes

1. Edit `kyverno-verify-images.yaml` for your GHCR org/repo.
2. Add policies for resource limits, labels, or [NetworkPolicy](networkpolicy.md) enforcement.
3. Test in a dev cluster before enforcing in production.

## Integration

- Enforces signatures from [Docker publish](../cicd/docker-publish.md)
- Complements [Trivy](trivy.md) scanning at build time

## Official resources

- [Kyverno documentation](https://kyverno.io/docs/)
- [Verify images with Cosign](https://kyverno.io/docs/writing-policies/verify-images/)
- [Sigstore policy controller](https://docs.sigstore.dev/policy-controller/overview/)
