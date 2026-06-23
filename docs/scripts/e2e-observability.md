# E2E observability verification

> Automated health checks for the task-runner + observability stack.

## In this repo

| Path | Purpose |
|------|---------|
| `scripts/e2e-observability-stack.sh` | Full stack verification script |

## Quick start

```bash
./scripts/e2e-observability-stack.sh
```

With security overlay:

```bash
./scripts/e2e-observability-stack.sh --security
```

The script boots `task-runner` compose (which includes observability), optionally the security overlay, generates traffic, and probes all backends.

## What it checks

- Service health endpoints (task-runner, Grafana, Prometheus, etc.)
- Prometheus scrape targets
- Mimir, Loki, Tempo API readiness
- Grafana datasource connectivity
- Alloy Faro collector
- k6 load test profile (optional)
- OnCall engine
- Falco/Talon (with `--security`, Linux only)

## Configuration

| Flag | Description |
|------|-------------|
| `--security` | Also start and verify security overlay |
| `--skip-up` | Assume stack is already running |
| `--timeout` | Per-check timeout (default 120s) |

## Making changes

1. Edit `scripts/e2e-observability-stack.sh` to add checks for new services.
2. Run ShellCheck locally: `shellcheck -S warning scripts/e2e-observability-stack.sh`
3. CI runs ShellCheck in the [quick lint](../cicd/github-actions-ci.md) job.

## Integration

- Uses [task-runner](../applications/task-runner.md) compose as the entry point
- Validates the full [observability stack](../observability/README.md)

## Official resources

- [Docker Compose profiles](https://docs.docker.com/compose/profiles/)
- [curl](https://curl.se/docs/manual.html)
