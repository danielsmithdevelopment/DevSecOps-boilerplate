# Distributed Task Runner

## Summary

The distributed task runner is a golang-based server that clusters together with any number of other running replicas. Together they execute tasks on predetermined schedules and aggregate results in an easy to understand format. When a task is completed, its results are stored in a global datastore and can be retrieved later by id.

## Task Definition and Execution

- Tasks are defined in YAML format
- Initially supports HTTP request tasks with extensibility for other task types
- Tasks can have dependencies on other tasks
- Maximum task execution time is 30 seconds
- Tasks that fail to complete within the time limit are considered failed

## Scheduling

- Supports cron syntax for recurring task schedules
- Tasks can be manually invoked via API in addition to their scheduled runs
- Supports both one-time and recurring task definitions

## Clustering and High Availability

- Designed to handle node failures gracefully
- Failed tasks are automatically re-executed on available nodes
- Supports up to 100 nodes in a cluster
- Load-based task distribution to maintain even load across nodes
- No leader election mechanism required

## Data Storage

- PostgreSQL used as the primary datastore
- Task results retained for 90 days
- Stored metadata includes:
  - Task ID
  - Creation timestamp
  - Creator's user ID
  - Task result version
  - Additional relevant task metadata
- Supports task result versioning

## API and Interface

- REST API endpoints:
  - POST /task/create
  - PUT /task/update
  - DELETE /task/delete
  - POST /task/invoke
  - GET /task/status
- JWT-based authentication for all endpoints
- No web UI required

## Monitoring and Observability

- Prometheus for metrics collection
- Grafana for visualization
- OpenTelemetry for distributed tracing
- Loki for log aggregation
- Tempo for trace storage
- Alerting system for failed tasks

## Performance Requirements

- Maximum 10 concurrent tasks per node
- No specific throughput or latency requirements

## Security

- Mutual TLS authentication between nodes
- Encrypted secrets for tasks
- Decryption at runtime
- HTTP-based inter-node communication