# NetworkPolicy

> Kubernetes egress restrictions for task-runner workloads.

## In this repo

| Path | Purpose |
|------|---------|
| `kubernetes/examples/networkpolicy-task-runner.yaml` | Default-deny egress with allowlists |
| `kubernetes/examples/README.md` | Brief index |

## What it does

- **Default deny** all egress from `task-runner` pods
- **Allow** DNS (UDP/TCP 53)
- **Allow** PostgreSQL (5432)
- **Allow** OTLP to observability namespace (4317, 4318)

## Quick start

```bash
kubectl apply -f kubernetes/examples/networkpolicy-task-runner.yaml
```

Adjust namespace labels and selectors to match your deployment.

## Configuration

Key fields to customize:

```yaml
metadata:
  namespace: task-runner
spec:
  podSelector:
    matchLabels:
      app: task-runner
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: observability
      ports:
        - port: 4317
```

## Making changes

1. Edit allowlist rules for your database host, OTel endpoint, or external APIs.
2. Test with `kubectl exec` and network probes.
3. Combine with [Kyverno](kyverno.md) image verification.

## Integration

- Protects [task-runner](../applications/task-runner.md) in Kubernetes
- Allows telemetry to [OpenTelemetry Collector](../observability/opentelemetry-collector.md)

## Official resources

- [Kubernetes NetworkPolicies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [NetworkPolicy recipes](https://github.com/ahmetb/kubernetes-network-policy-recipes)
