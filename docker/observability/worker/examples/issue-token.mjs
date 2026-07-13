#!/usr/bin/env node
/**
 * Example: mint a short-lived HS256 JWT for the telemetry ingest Worker.
 * Wire this behind your real session auth (cookie / OAuth) — never expose the signing key.
 *
 *   JWT_SIGNING_KEY=... PROJECT_ID=frontend-prod ORIGIN=https://app.example.com \
 *     node examples/issue-token.mjs
 */
import { createHmac, randomUUID } from 'node:crypto';

function b64url(input) {
  return Buffer.from(input)
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

function signHs256(payload, secret) {
  const header = b64url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = b64url(JSON.stringify(payload));
  const sig = createHmac('sha256', secret)
    .update(`${header}.${body}`)
    .digest('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
  return `${header}.${body}.${sig}`;
}

const secret = process.env.JWT_SIGNING_KEY;
if (!secret) {
  console.error('JWT_SIGNING_KEY is required');
  process.exit(1);
}

const now = Math.floor(Date.now() / 1000);
const claims = {
  sub: process.env.SESSION_ID || `session_${randomUUID()}`,
  project: process.env.PROJECT_ID || 'frontend-prod',
  origin: process.env.ORIGIN || 'https://app.example.com',
  iat: now,
  exp: now + Number(process.env.TOKEN_TTL_SECONDS || 3600),
};

const token = signHs256(claims, secret);
console.log(JSON.stringify({ token, expires_at: claims.exp, claims }, null, 2));
