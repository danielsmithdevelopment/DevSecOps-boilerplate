# Faro initialization (JWT-gated)

Browser helpers that wire [Grafana Faro](https://grafana.com/docs/grafana-cloud/monitor-applications/frontend-observability/) to the Cloudflare Worker under `../worker/` instead of a public collector URL.

## Usage

```ts
import { initGatedFaro } from './faro-init';

await initGatedFaro({
  workerUrl: 'https://telemetry-ingest.your-account.workers.dev/collect',
  tokenUrl: '/api/telemetry/token',
  app: { name: 'frontend', version: import.meta.env.VITE_APP_VERSION },
  environment: import.meta.env.MODE,
});
```

Your backend `POST /api/telemetry/token` must authenticate the session and return:

```json
{ "token": "<jwt>", "expires_at": 1717603600 }
```

JWT claims must match the Worker config (`project`, `origin`, `sub`, `exp`).

## What gets labeled

On every `exception` event, `beforeSend` attaches:

| Label | Purpose |
|-------|---------|
| `error_fingerprint` | Sentry-style grouping key for Loki streams |
| `release` | App version for source-map lookup / regression |
| `environment` | e.g. production / staging |

## Local / no-Worker fallback

For pure local compose (open Alloy Faro on `:8027`) you can still point Faro directly:

```ts
import { initializeFaro } from '@grafana/faro-web-sdk';

initializeFaro({
  url: 'http://localhost:8027/collect',
  app: { name: 'frontend', version: 'dev' },
});
```

Do **not** leave an unauthenticated collector exposed on the public internet.
