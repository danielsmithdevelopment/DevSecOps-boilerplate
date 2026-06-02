# Sign-In with Wallet

A complete implementation of wallet-based authentication using WalletConnect and EIP-4361 Sign-In with Ethereum (SIWE). This project provides a Go backend API and React frontend that enables users to sign in using any crypto wallet.

## Features

- **WalletConnect Integration**: Support for all WalletConnect-compatible wallets (MetaMask, WalletConnect mobile app, Coinbase Wallet, etc.)
- **Multi-Chain Support**: Works with Ethereum mainnet and other EVM-compatible chains
- **EIP-4361 SIWE**: Secure authentication using Sign-In with Ethereum standard
- **JWT Authentication**: Token-based authentication for protected routes
- **PostgreSQL Storage**: User data stored in PostgreSQL database
- **Docker Support**: Easy deployment with Docker Compose

## Architecture

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   React     │────────▶│  Go Backend  │────────▶│ PostgreSQL  │
│  Frontend   │         │     API      │         │  Database   │
└─────────────┘         └──────────────┘         └─────────────┘
      │                         │
      │                         │
      ▼                         ▼
┌─────────────┐         ┌──────────────┐
│ WalletConnect│         │ SIWE Verifier │
│     SDK     │         │  (go-ethereum)│
└─────────────┘         └──────────────┘
```

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 20+ (for local development)
- WalletConnect Project ID ([Get one here](https://cloud.walletconnect.com/))

## Quick Start

### Using Docker Compose

1. Clone the repository:
   ```bash
   cd wallet-auth
   ```

2. Create a `.env` file:
   ```bash
   cp env.example .env
   ```

3. Edit `.env` and set your WalletConnect Project ID:
   ```
   VITE_WALLETCONNECT_PROJECT_ID=your-project-id-here
   JWT_SECRET=your-secret-key-min-32-characters
   ```

4. Start all services:
   ```bash
   docker-compose up -d
   ```

5. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - PostgreSQL: localhost:5432

### Local Development

#### Backend

1. Navigate to backend directory:
   ```bash
   cd backend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set environment variables:
   ```bash
   export DATABASE_URL="postgres://walletauth:walletauth@localhost:5432/walletauth?sslmode=disable"
   export JWT_SECRET="your-secret-key-min-32-characters"
   export DOMAIN="localhost"
   export URI="http://localhost:3000"
   export CHAIN_ID=1
   ```

4. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

#### Frontend

1. Navigate to frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Create `.env` file:
   ```
   VITE_API_URL=http://localhost:8080
   VITE_WALLETCONNECT_PROJECT_ID=your-project-id-here
   ```

4. Start development server:
   ```bash
   npm run dev
   ```

## API Endpoints

### Public Endpoints

#### `GET /health`
Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

#### `GET /auth/challenge?address={wallet_address}`
Generate a SIWE challenge message for the given wallet address.

**Parameters:**
- `address` (query): Ethereum wallet address (0x...)

**Response:**
```json
{
  "message": "localhost wants you to sign in with your Ethereum account:\n0x1234...\n\nURI: http://localhost:3000\nVersion: 1\nChain ID: 1\nNonce: abc123...\nIssued At: 2024-01-01T00:00:00Z",
  "nonce": "abc123..."
}
```

#### `POST /auth/verify`
Verify a signed SIWE message and return a JWT token.

**Request Body:**
```json
{
  "message": "localhost wants you to sign in...",
  "signature": "0x1234..."
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "wallet_address": "0x1234..."
  }
}
```

### Protected Endpoints

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

#### `GET /auth/me`
Get the current authenticated user.

**Response:**
```json
{
  "id": "uuid",
  "wallet_address": "0x1234...",
  "created_at": "2024-01-01T00:00:00Z",
  "last_login": "2024-01-01T00:00:00Z"
}
```

## Authentication Flow

1. **User clicks "Connect Wallet"**
   - Frontend initializes WalletConnect SDK
   - User selects and connects their wallet

2. **User clicks "Sign In"**
   - Frontend requests a challenge from backend (`GET /auth/challenge`)
   - Backend generates a SIWE message with a unique nonce
   - Frontend requests signature from wallet using `personal_sign`
   - User approves signature in their wallet

3. **Verification**
   - Frontend sends signed message to backend (`POST /auth/verify`)
   - Backend verifies signature using EIP-191
   - Backend checks nonce validity
   - Backend returns JWT token

4. **Authenticated State**
   - Frontend stores JWT token in localStorage
   - Token is included in Authorization header for protected routes
   - Token expires after 24 hours

## Database Schema

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP,
    nonce VARCHAR(32) NOT NULL
);
```

## Environment Variables

### Backend

- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for signing JWT tokens (min 32 characters)
- `DOMAIN`: Domain for SIWE messages (e.g., "example.com")
- `URI`: URI for SIWE messages (defaults to `https://{DOMAIN}`)
- `CHAIN_ID`: Chain ID for SIWE messages (default: 1 for Ethereum mainnet)
- `PORT`: Server port (default: 8080)

### Frontend

- `VITE_API_URL`: Backend API URL
- `VITE_WALLETCONNECT_PROJECT_ID`: WalletConnect Cloud project ID

## Security Considerations

1. **SIWE Message Verification**: All signatures are verified using EIP-191 standard
2. **Nonce Management**: Each challenge includes a unique nonce stored in the database
3. **JWT Expiration**: Tokens expire after 24 hours
4. **CORS**: Configure CORS to restrict access to your frontend domain in production
5. **Rate Limiting**: Consider implementing rate limiting on auth endpoints
6. **HTTPS**: Always use HTTPS in production

## Supported Wallets

- MetaMask
- WalletConnect Mobile App
- Coinbase Wallet
- Trust Wallet
- Rainbow Wallet
- And all other WalletConnect-compatible wallets

## Supported Chains

- Ethereum Mainnet (Chain ID: 1)
- Polygon (Chain ID: 137)
- Arbitrum (Chain ID: 42161)
- Optimism (Chain ID: 10)
- BSC (Chain ID: 56)
- And other EVM-compatible chains

## Troubleshooting

### WalletConnect Modal Not Appearing

- Ensure `VITE_WALLETCONNECT_PROJECT_ID` is set correctly
- Check browser console for errors
- Verify your WalletConnect project is active

### Signature Verification Fails

- Ensure the wallet address matches the one in the SIWE message
- Check that the nonce hasn't expired (nonces are single-use)
- Verify the chain ID matches your configured chain

### Database Connection Issues

- Ensure PostgreSQL is running
- Check `DATABASE_URL` is correct
- Verify database credentials

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
