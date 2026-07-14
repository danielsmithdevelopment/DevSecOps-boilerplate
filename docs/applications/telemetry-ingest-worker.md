# Telemetry ingest Worker

> Cloudflare Worker that replaces a static public Sentry DSN / open Faro URL with an ephemeral JWT-gated proxy.

## In this repo

| Item | Path |
|------|------|
| Worker source | `docker/observability/worker/` |
| Faro browser helpers | `docker/observability/faro/` |
| Token mint example | `docker/observability/worker/examples/issue-token.mjs` |

## What it closes

Anyone can scrape a public DSN from a JS bundle and inject fake events. When AI agents treat those events as instructions (Nutrient-class prompt injection), an unauthenticated ingest URL becomes an attack surface. This Worker requires a short-lived HS256 JWT on every request, rate-limits by session, validates payload shape, fingerprints exceptions for Loki grouping, optionally symbolicates from R2, and **silently drops** failures with HTTP 204.

## Quick start

```bash
cd docker/observability/worker
npm install
# edit wrangler.toml: ALLOY_INGEST_URL, ALLOWED_ORIGINS, PROJECT_ID
echo "dev-signing-key-change-me-32chars!!" | npx wrangler secret put JWT_SIGNING_KEY
npx wrangler dev
```

Browser:

```ts
import { initGatedFaro } from '../faro/src/faro-init';

await initGatedFaro({
  workerUrl: 'http://127.0.0.1:8787/collect',
  tokenUrl: '/api/telemetry/token',
  app: { name: 'frontend', version: 'dev' },
});
```

## Covers / does not cover

| Covers | Does not cover |
|--------|----------------|
| Third party with no JWT | Compromised end-user replaying their own JWT until expiry |
| Flooding an open collector | Agent trust-boundary failures (needs Langfuse alerts + Tetragon) |

## Related

- [Alloy Faro receiver](../observability/alloy.md)
- [Loki fingerprint alerts](../observability/loki.md)
- [Tetragon package-manager policy](../security/tetragon.md)
- [Wazuh FIM](../security/wazuh.md)
