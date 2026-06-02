# Distributed Task Runner

A golang-based distributed task runner that clusters together with any number of other running replicas to execute tasks on predetermined schedules and aggregate results.

## Features

- Task scheduling using cron syntax
- HTTP task support with extensibility for other task types
- Task dependencies
- Distributed execution across multiple nodes
- PostgreSQL storage for tasks and results
- JWT-based authentication
- Prometheus metrics
- Grafana dashboards
- OpenTelemetry tracing
- Loki log aggregation
- Tempo trace storage
- Health checks for all services
- Automatic service recovery with restart policies

## Prerequisites

- Docker
- Docker Compose
- Go 1.21 or later (for development)

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/task-runner.git
   cd task-runner
   ```

2. Start the services:
   ```bash
   docker-compose up -d
   ```

3. Access the services:
   - Task Runner API (5 instances):
     - Instance 1: http://localhost:8080
     - Instance 2: http://localhost:8081
     - Instance 3: http://localhost:8082
     - Instance 4: http://localhost:8083
     - Instance 5: http://localhost:8084
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090
   - Loki: http://localhost:3100
   - Tempo: http://localhost:3200
   - PostgreSQL: localhost:5432

## API Endpoints

### Authentication
- `POST /auth` - Get JWT token
  ```json
  {
    "username": "your-username",
    "password": "your-password"
  }
  ```

### Tasks
- `POST /task/create` - Create a new task
  ```json
  {
    "name": "example-task",
    "type": "http",
    "schedule": "*/5 * * * *",
    "config": {
      "http": {
        "url": "https://api.example.com",
        "method": "GET",
        "headers": {
          "Authorization": "Bearer token"
        }
      }
    }
  }
  ```

- `PUT /task/update` - Update an existing task
- `DELETE /task/delete/:id` - Delete a task
- `POST /task/invoke/:id` - Manually invoke a task
- `GET /task/status/:id` - Get task status and results
- `GET /health` - Health check endpoint

## Task Configuration

Tasks are defined in YAML format with the following structure:

```yaml
name: example-task
type: http
schedule: "*/5 * * * *"
config:
  http:
    url: https://api.example.com
    method: GET
    headers:
      Authorization: Bearer token
```

## Development

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run tests:
   ```bash
   go test ./...
   ```

3. Build the application:
   ```bash
   go build -o task-runner ./cmd/server
   ```

## Configuration

The task runner can be configured using environment variables:

- `DB_CONN` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT token generation
- `INSTANCE_ID` - Unique identifier for each task runner instance
- `ADDR` - Server address (default: :8080)

## Monitoring

### Metrics

The task runner exposes the following Prometheus metrics:

- `task_execution_total` - Total number of task executions
- `task_execution_duration_seconds` - Task execution duration
- `task_execution_status` - Task execution status
- `task_schedule_total` - Total number of scheduled tasks

### Logging

Logs are collected by Loki and can be viewed in Grafana.

### Tracing

Distributed tracing is provided by OpenTelemetry and can be viewed in Grafana with Tempo.

## Security

- JWT-based authentication for API endpoints
- PostgreSQL for secure data storage
- Environment variable based configuration
- Health checks for all services
- Automatic service recovery

## License

MIT 