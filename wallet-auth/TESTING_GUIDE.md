# Dual Authentication Testing Guide

This guide will help you test the wallet and email authentication features.

## Prerequisites

1. **Services Running**: Ensure all Docker services are up:
   ```bash
   docker-compose ps
   ```

2. **Environment Variables**: Make sure you have:
   - `VITE_WALLETCONNECT_PROJECT_ID` set (for wallet authentication)
   - AWS SES credentials (optional, for email sending - if not set, email features will be disabled but won't crash)

3. **Access Points**:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

## Test Scenarios

### 1. Wallet Authentication (Existing Flow)

1. Navigate to http://localhost:3000
2. Click "Wallet" tab (should be selected by default)
3. Click "Connect Wallet"
4. Select your wallet and approve connection
5. Click "Sign In"
6. You should see your wallet address and last login time

**Expected Result**: Successfully authenticated with wallet

### 2. Email Authentication (New Flow)

1. Navigate to http://localhost:3000
2. Click "Email" tab
3. Enter your email address
4. Click "Send Magic Link"
5. Check your email (if AWS SES is configured) or check backend logs for the verification token
6. Click the verification link in the email OR manually navigate to:
   ```
   http://localhost:3000/email-verify?token=<token_from_logs>
   ```
7. You should be redirected to the home page and see your email address

**Expected Result**: Successfully authenticated with email

**Note**: If AWS SES is not configured, the backend will log a warning but the endpoint will still work. You can check the backend logs to see the verification token:
```bash
docker-compose logs backend | grep -i "verification\|token"
```

### 3. Adding Wallet to Email User

1. Sign in with email (follow Test Scenario 2)
2. You should see an "Add Wallet" section
3. Click "Add Wallet"
4. Connect your wallet
5. Sign the SIWE message
6. Your wallet should be linked to your account

**Expected Result**: User now has both email and wallet linked

### 4. Adding Email to Wallet User

1. Sign in with wallet (follow Test Scenario 1)
2. You should see an "Add Email" section
3. Click "Add Email"
4. Enter your email address
5. Click "Send Verification Email"
6. Verify the email using the magic link (same as Test Scenario 2)
7. Your email should be linked to your account

**Expected Result**: User now has both wallet and email linked

### 5. Testing API Endpoints Directly

You can test the API endpoints using curl:

#### Email Signup
```bash
curl -X POST http://localhost:8080/auth/email/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

#### Verify Email (replace TOKEN with actual token)
```bash
curl http://localhost:8080/auth/email/verify?token=TOKEN
```

#### Get User Info (requires JWT token)
```bash
curl http://localhost:8080/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Add Wallet to Email User (requires JWT token)
```bash
curl -X POST http://localhost:8080/auth/wallet/add \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address":"0x..."}'
```

## Troubleshooting

### Backend Issues

1. **Check backend logs**:
   ```bash
   docker-compose logs backend
   ```

2. **Check database connection**:
   ```bash
   docker-compose exec postgres psql -U walletauth -d walletauth -c "SELECT * FROM users LIMIT 5;"
   ```

3. **Check if schema was migrated**:
   ```bash
   docker-compose exec postgres psql -U walletauth -d walletauth -c "\d users"
   ```
   You should see columns: `email`, `email_verified`, `email_verification_token`, etc.

### Frontend Issues

1. **Check frontend logs**:
   ```bash
   docker-compose logs frontend
   ```

2. **Check browser console** for JavaScript errors

3. **Verify environment variables** are set in docker-compose.yml

### Email Issues

1. **If AWS SES is not configured**:
   - The backend will log warnings but continue to work
   - You can extract verification tokens from backend logs
   - Email sending will be disabled

2. **To configure AWS SES**:
   - Set `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `SES_FROM_EMAIL` in your `.env` file
   - Restart the backend: `docker-compose restart backend`

## Database Verification

To verify users are being created correctly:

```bash
# Connect to database
docker-compose exec postgres psql -U walletauth -d walletauth

# Check users table structure
\d users

# View all users
SELECT id, wallet_address, email, email_verified, created_at FROM users;

# Check for users with both methods
SELECT id, wallet_address, email, email_verified FROM users 
WHERE wallet_address IS NOT NULL AND email IS NOT NULL;
```

## Common Issues

1. **"user not found" error**: 
   - Make sure the user was created during the challenge/signup phase
   - Check address normalization (should be lowercase)

2. **Email verification not working**:
   - Check if token is expired (24 hour expiry)
   - Verify token in database matches the one in URL

3. **Wallet linking fails**:
   - Ensure email is verified first (for email-first users)
   - Check SIWE message format and signature

4. **JWT token issues**:
   - Tokens expire after 24 hours
   - Make sure `JWT_SECRET` is set consistently

## Next Steps

After testing, you can:
1. Configure AWS SES for production email sending
2. Set up proper domain and URI for SIWE messages
3. Add rate limiting for email signup
4. Implement email verification resend with cooldown
5. Add logging and monitoring
