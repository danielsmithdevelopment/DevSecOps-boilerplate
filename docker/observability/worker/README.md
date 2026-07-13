# Telemetry ingest Worker (JWT-gated Faro proxy)

Cloudflare Worker that replaces a static public Sentry DSN / open Faro collector URL with an **ephemeral-token-gated proxy**.

## What it does

1. Validates a short-lived HS256 JWT (`Authorization: Bearer …`)
2. Checks project + origin claims against configured allowlists
3. Rate-limits by session (`sub`) per rolling minute
4. Validates Faro payload shape / size
5. Computes a Loki-friendly `error_fingerprint` (Web Crypto SHA-256)
6. Optionally symbolicates stack frames from an R2 source-map bucket
7. Forwards clean events to a private Grafana Alloy Faro receiver
8. Returns **204** on every failure path (no useful signal to attackers)

## Quick start

```bash
cd docker/observability/worker
npm install
cp ../../.env.example ../../.env   # optional — Worker secrets are separate

# Local secrets (never commit)
echo "dev-signing-key-change-me-32chars!!" | npx wrangler secret put JWT_SIGNING_KEY --local

# Point ALLOY_INGEST_URL / ALLOWED_ORIGINS / PROJECT_ID in wrangler.toml
npx wrangler dev
```

Deploy:

```bash
npx wrangler secret put JWT_SIGNING_KEY
npx wrangler deploy
```

## Backend token issuance

Your application backend must issue JWTs signed with the same `JWT_SIGNING_KEY`:

```json
{
  "sub": "session_abc123",
  "project": "frontend-prod",
  "origin": "https://app.example.com",
  "iat": 1717600000,
  "exp": 1717603600
}
```

See `../faro/` for the browser-side token fetch + Faro `beforeSend` hook.

## What this covers / doesn't

| Covered | Not covered |
|---------|-------------|
| Third-party attacker with a scraped Worker URL and no JWT | Malicious / compromised end-user replaying their own JWT until expiry |
| Flooding a shared static DSN | Social-engineering an agent with data from a *valid* session |

Pair with the alert rule `ClientErrorNoServerSpan` and Tetragon policies under `../configs/security/`.
