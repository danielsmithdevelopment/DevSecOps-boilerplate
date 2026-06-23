# wallet-auth

> Sign-In with Ethereum (SIWE) and email authentication — Go backend, React frontend.

## What it is

A complete wallet-based auth demo using WalletConnect and EIP-4361 SIWE, plus optional email auth via AWS SES. JWT tokens protect API routes. This app runs standalone (no observability compose include) but can send Faro RUM telemetry to Alloy.

## In this repo

| Path | Purpose |
|------|---------|
| `wallet-auth/backend/cmd/server/main.go` | Go API entry point |
| `wallet-auth/backend/internal/` | API, SIWE, storage, email |
| `wallet-auth/frontend/src/` | React + Vite + Tailwind UI |
| `wallet-auth/docker-compose.yml` | Postgres, backend, frontend |
| `wallet-auth/env.example` | Environment template |
| `wallet-auth/Makefile` | Dev shortcuts |
| `wallet-auth/DEVELOPMENT_NOTES.md` | Implementation notes |
| `wallet-auth/TESTING_GUIDE.md` | Test scenarios |

## Quick start

```bash
cd wallet-auth
cp env.example .env
# Set VITE_WALLETCONNECT_PROJECT_ID and JWT_SECRET (min 32 chars)
docker compose up -d
```

| URL | Service |
|-----|---------|
| http://localhost:3000 | React frontend |
| http://localhost:8080 | Go backend API |
| localhost:5432 | PostgreSQL |

### Local development

**Backend:** `cd backend && go run ./cmd/server`  
**Frontend:** `cd frontend && npm ci && npm run dev`

## Configuration

Key variables in `.env`:

| Variable | Description |
|----------|-------------|
| `VITE_WALLETCONNECT_PROJECT_ID` | WalletConnect Cloud project ID |
| `JWT_SECRET` | JWT signing key (≥32 characters) |
| `DATABASE_URL` | PostgreSQL connection string |
| `AWS_*` | SES credentials for email auth (optional) |
| `CORS_ORIGINS` | Allowed frontend origins |

See `env.example` for the full list.

## API overview

| Endpoint | Description |
|----------|-------------|
| `POST /api/auth/nonce` | Get SIWE nonce |
| `POST /api/auth/verify` | Verify SIWE signature, return JWT |
| `POST /api/auth/email/request` | Request email magic link |
| `GET /api/auth/email/verify` | Verify email token |
| `GET /api/user/me` | Protected user profile |

Full API details: `wallet-auth/README.md`.

## Frontend telemetry (optional)

Point [Grafana Faro](https://grafana.com/docs/grafana-cloud/monitor-applications/frontend-observability/) at Alloy when the observability stack is running:

```javascript
import { initializeFaro } from '@grafana/faro-web-sdk';

initializeFaro({
  url: 'http://localhost:8027/collect',
  app: { name: 'wallet-auth-frontend', version: '1.0.0' },
});
```

## Making changes

1. Backend: edit `wallet-auth/backend/internal/`, run `go test ./...`.
2. Frontend: edit `wallet-auth/frontend/src/`, run `npm run build`.
3. Update `env.example` when adding new env vars.
4. See [TESTING_GUIDE.md](../../wallet-auth/TESTING_GUIDE.md) for auth flow testing.

## Integration

- Threat model: [THREAT_MODEL.md](../security/THREAT_MODEL.md)
- CI: Go module `wallet-auth/backend` + `npm audit` on frontend
- Not included in [Docker publish](../cicd/docker-publish.md) (no GHCR image yet)

## Official resources

- [EIP-4361 SIWE](https://eips.ethereum.org/EIPS/eip-4361)
- [WalletConnect](https://docs.walletconnect.com/)
- [go-ethereum](https://geth.ethereum.org/docs)
- [Grafana Faro Web SDK](https://grafana.com/docs/grafana-cloud/monitor-applications/frontend-observability/)
