# Wallet Auth Development Notes

## Overview

This document summarizes the development and debugging process for the wallet-auth project, a Sign-In with Ethereum (SIWE) implementation using WalletConnect and a Go backend with React frontend.

## Project Structure

```
wallet-auth/
├── backend/          # Go backend API
│   ├── cmd/server/   # Main server entry point
│   └── internal/
│       ├── api/      # HTTP handlers (auth.go, middleware.go)
│       ├── siwe/     # SIWE message generation and verification
│       ├── storage/  # PostgreSQL storage layer
│       └── models/   # Data models
├── frontend/         # React + Vite frontend
│   └── src/
│       ├── hooks/    # useAuth hook for WalletConnect integration
│       ├── components/ # React components
│       └── services/ # API client
└── docker-compose.yml # Service orchestration
```

## Implementation Details

### Backend (Go)

#### Key Components

1. **SIWE Verifier** (`internal/siwe/verifier.go`)
   - Generates SIWE challenge messages (EIP-4361 format)
   - Verifies EIP-191 signatures
   - Normalizes addresses to lowercase for consistency

2. **Storage Layer** (`internal/storage/postgres.go`)
   - Case-insensitive address lookup using `LOWER(TRIM(wallet_address))`
   - User creation and nonce management
   - Address normalization on insert and lookup

3. **API Handlers** (`internal/api/auth.go`)
   - `/auth/challenge` - Generates SIWE challenge message with nonce
   - `/auth/verify` - Verifies signature and returns JWT token
   - `/auth/me` - Returns authenticated user data (protected route)

#### Address Normalization

All wallet addresses are normalized to lowercase throughout the system:
- In `ExtractAddressFromMessage()` - normalizes address from SIWE message
- In `VerifySignature()` - normalizes recovered address
- In `GetOrCreateUser()` - normalizes before database operations
- In `GetUserByAddress()` - normalizes input and uses case-insensitive SQL query

#### Database Schema

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP,
    nonce VARCHAR(32) NOT NULL
);
```

### Frontend (React + TypeScript)

#### WalletConnect Integration

- Uses `@walletconnect/ethereum-provider` for wallet connections
- Requires `VITE_WALLETCONNECT_PROJECT_ID` environment variable
- Handles wallet connection, message signing, and authentication flow

#### Authentication Flow

1. User clicks "Connect Wallet"
2. WalletConnect modal opens for wallet selection
3. User connects wallet, address is stored in state
4. User clicks "Sign In"
5. Frontend requests challenge from `/auth/challenge`
6. User signs SIWE message with wallet
7. Frontend sends signature to `/auth/verify`
8. Backend verifies signature and returns JWT token
9. Frontend stores token and loads user data from `/auth/me`
10. User sees welcome screen with wallet address

## Issues Encountered and Solutions

### Issue 1: Docker Daemon Not Running

**Problem**: `Cannot connect to the Docker daemon`

**Solution**: Installed and started Docker Desktop (not just docker-compose client)

### Issue 2: Obsolete Docker Compose Version

**Problem**: Warning about `version: '3.8'` being obsolete

**Solution**: Removed the `version` line from `docker-compose.yml`

### Issue 3: Missing go.sum During Build

**Problem**: `go.sum` not found during Docker build

**Solution**: Modified Dockerfile to copy `go.mod` and `go.sum` first, then run `go mod tidy && go mod download` before copying source code

### Issue 4: Frontend Build Issues

**Problems**:
- WalletConnect package version conflicts
- TypeScript errors with `EthereumProvider` type
- Missing `vite/client` type definitions

**Solutions**:
- Updated WalletConnect packages to compatible versions
- Fixed type definition: `Awaited<ReturnType<typeof EthereumProvider.init>>`
- Added `vite/client` to `types` in `tsconfig.json`

### Issue 5: Environment Variables Not Available in Frontend

**Problem**: `VITE_WALLETCONNECT_PROJECT_ID` not available at runtime

**Solution**: 
- Vite environment variables are embedded at build time
- Added build args to `frontend/Dockerfile` and `docker-compose.yml` to pass `VITE_WALLETCONNECT_PROJECT_ID` during build

### Issue 6: "User Not Found" Error

**Problem**: Persistent "user not found" errors during authentication

**Root Cause**: The error was actually coming from `/auth/me` endpoint, not `/auth/verify`. The flow was:
1. `/auth/verify` succeeded and returned JWT token
2. Frontend called `/auth/me` to load user data
3. `/auth/me` failed because wallet address wasn't being extracted correctly from JWT

**Solution**: 
- Added detailed error logging to `/auth/me` endpoint
- Fixed address normalization in JWT token claims
- Ensured `wallet_address` is properly set in JWT and extracted in middleware

**Key Learning**: When debugging authentication flows, check ALL endpoints in the chain, not just the first one that appears to fail.

### Issue 7: Docker Build Cache Issues

**Problem**: Code changes not appearing after rebuild

**Solution**: Used `docker-compose build --no-cache backend` to force a clean rebuild

## Key Technical Decisions

### Address Normalization

All Ethereum addresses are normalized to lowercase for consistency:
- Prevents case-sensitivity issues
- Ensures database lookups work regardless of address format
- Standard practice in Ethereum applications

### Case-Insensitive Database Queries

SQL queries use `LOWER(TRIM(wallet_address))` for lookups:
```sql
WHERE LOWER(TRIM(wallet_address)) = LOWER(TRIM($1))
```

This ensures addresses match regardless of:
- Case differences (0xABC vs 0xabc)
- Whitespace differences
- Checksummed vs non-checksummed addresses

### Nonce Management

- New nonce generated for each challenge request
- Nonce stored in database and verified during signature verification
- Prevents replay attacks

### Error Handling

- Detailed error messages during development
- Address normalization errors include full context
- Database errors include the address being looked up

## Environment Variables

### Backend

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT signing
- `DOMAIN` - Domain for SIWE messages (default: localhost)
- `URI` - URI for SIWE messages (default: http://localhost:3000)
- `CHAIN_ID` - Ethereum chain ID (default: 1)

### Frontend

- `VITE_API_URL` - Backend API URL (default: http://localhost:8080)
- `VITE_WALLETCONNECT_PROJECT_ID` - WalletConnect project ID (required)

## Running the Project

1. **Set up environment variables**:
   ```bash
   cp env.example .env
   # Edit .env with your WalletConnect project ID
   ```

2. **Start services**:
   ```bash
   docker-compose up -d
   ```

3. **Access the application**:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

4. **View logs**:
   ```bash
   docker-compose logs -f backend
   docker-compose logs -f frontend
   ```

## Testing the Flow

1. Open http://localhost:3000
2. Click "Connect Wallet"
3. Select a wallet from WalletConnect modal
4. Click "Sign In"
5. Approve the signature request in your wallet
6. You should see the welcome screen with your wallet address

## Debugging Tips

### Check Backend Logs

```bash
docker-compose logs backend --tail 100
```

### Check Database

```bash
docker-compose exec postgres psql -U walletauth -d walletauth -c "SELECT * FROM users;"
```

### Force Clean Rebuild

If code changes aren't appearing:

```bash
docker-compose build --no-cache backend
docker-compose restart backend
```

### Test API Directly

```bash
# Get challenge
curl "http://localhost:8080/auth/challenge?address=0x..."

# Verify signature (requires valid message and signature)
curl -X POST http://localhost:8080/auth/verify \
  -H "Content-Type: application/json" \
  -d '{"message":"...","signature":"..."}'
```

## Security Considerations

1. **JWT Secret**: Use a strong, random secret in production
2. **Nonce**: Each challenge generates a unique nonce to prevent replay attacks
3. **Signature Verification**: EIP-191 signature verification ensures message authenticity
4. **Address Normalization**: Prevents address format attacks
5. **CORS**: Currently allows all origins (`*`) - restrict in production

## Future Improvements

1. Add rate limiting to prevent abuse
2. Implement token refresh mechanism
3. Add session management
4. Restrict CORS to specific origins
5. Add comprehensive error handling and logging
6. Implement user profile management
7. Add support for multiple chains
8. Implement account linking (multiple addresses per user)

## References

- [EIP-4361: Sign-In with Ethereum](https://eips.ethereum.org/EIPS/eip-4361)
- [EIP-191: Signed Data Standard](https://eips.ethereum.org/EIPS/eip-191)
- [WalletConnect Documentation](https://docs.walletconnect.com/)
- [SIWE Message Format](https://github.com/spruceid/siwe)

## Notes

- The project uses PostgreSQL for user storage
- All addresses are stored in lowercase for consistency
- The frontend uses Vite for fast development and builds
- Docker Compose orchestrates all services
- The backend uses Gin for HTTP routing
- JWT tokens expire after 24 hours
