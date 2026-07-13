import { createErrorFingerprint } from './fingerprint';
import { validatePayload } from './schema';
import { fetchSourceMap, symbolicateFrame } from './symbolication';
import type { Env, ExceptionEvent, JwtClaims } from './types';

/** Silent drop — never leak why validation failed to an attacker. */
const DROP = new Response(null, { status: 204 });

const rateBuckets = new Map<string, { count: number; resetAt: number }>();

function b64urlToBytes(input: string): Uint8Array {
  const pad = '='.repeat((4 - (input.length % 4)) % 4);
  const b64 = (input + pad).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(b64);
  const out = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
  return out;
}

function bytesToB64url(bytes: ArrayBuffer | Uint8Array): string {
  const arr = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
  let s = '';
  for (const b of arr) s += String.fromCharCode(b);
  return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

async function importHmacKey(secret: string): Promise<CryptoKey> {
  return crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['verify'],
  );
}

async function verifyJwt(token: string, secret: string): Promise<JwtClaims | null> {
  const parts = token.split('.');
  if (parts.length !== 3) return null;
  const [headerB64, payloadB64, sigB64] = parts as [string, string, string];

  let header: { alg?: string; typ?: string };
  try {
    header = JSON.parse(new TextDecoder().decode(b64urlToBytes(headerB64)));
  } catch {
    return null;
  }
  if (header.alg !== 'HS256') return null;

  const key = await importHmacKey(secret);
  const data = new TextEncoder().encode(`${headerB64}.${payloadB64}`);
  const sig = b64urlToBytes(sigB64);
  const valid = await crypto.subtle.verify('HMAC', key, sig, data);
  if (!valid) return null;

  let claims: JwtClaims;
  try {
    claims = JSON.parse(new TextDecoder().decode(b64urlToBytes(payloadB64)));
  } catch {
    return null;
  }

  const now = Math.floor(Date.now() / 1000);
  if (!claims.exp || claims.exp < now) return null;
  if (!claims.sub || !claims.project || !claims.origin) return null;
  return claims;
}

function checkRateLimit(sessionId: string, limit: number): boolean {
  const now = Date.now();
  const bucket = rateBuckets.get(sessionId);
  if (!bucket || now >= bucket.resetAt) {
    rateBuckets.set(sessionId, { count: 1, resetAt: now + 60_000 });
    return true;
  }
  if (bucket.count >= limit) return false;
  bucket.count += 1;
  return true;
}

function parseAllowedOrigins(raw: string): Set<string> {
  return new Set(
    raw
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
  );
}

async function enrichExceptions(body: unknown, env: Env): Promise<unknown> {
  const enrichOne = async (event: Record<string, unknown>): Promise<Record<string, unknown>> => {
    if (event.type !== 'exception') return event;

    const exceptionEvent = event as unknown as ExceptionEvent;
    const fingerprint = await createErrorFingerprint(exceptionEvent);
    const labels: Record<string, string> = {
      ...(exceptionEvent.meta?.labels ?? {}),
      error_fingerprint: fingerprint,
    };

    if (env.SOURCEMAPS && exceptionEvent.payload.exceptions?.[0]?.stacktrace?.frames) {
      const release = labels.release ?? exceptionEvent.meta?.app?.version ?? 'unknown';
      const frames = exceptionEvent.payload.exceptions[0].stacktrace.frames;
      const symbolicated = [];
      for (const frame of frames) {
        const filename = frame.filename ?? frame.abs_path;
        if (!filename) {
          symbolicated.push(frame);
          continue;
        }
        const map = await fetchSourceMap(env.SOURCEMAPS, release, filename);
        symbolicated.push(map ? symbolicateFrame(frame, map) : frame);
      }
      exceptionEvent.payload.exceptions[0].stacktrace.frames = symbolicated;
    }

    return {
      ...exceptionEvent,
      meta: { ...exceptionEvent.meta, labels },
    };
  };

  if (Array.isArray(body)) {
    return Promise.all(body.map((e) => enrichOne(e as Record<string, unknown>)));
  }
  if (body && typeof body === 'object' && Array.isArray((body as { events?: unknown[] }).events)) {
    const envelope = body as { events: Record<string, unknown>[] };
    return {
      ...envelope,
      events: await Promise.all(envelope.events.map((e) => enrichOne(e))),
    };
  }
  return enrichOne(body as Record<string, unknown>);
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // CORS preflight for browser SDKs
    if (request.method === 'OPTIONS') {
      const origin = request.headers.get('Origin') ?? '';
      const allowed = parseAllowedOrigins(env.ALLOWED_ORIGINS);
      if (!allowed.has(origin) && !allowed.has('*')) return DROP;
      return new Response(null, {
        status: 204,
        headers: {
          'Access-Control-Allow-Origin': origin,
          'Access-Control-Allow-Methods': 'POST, OPTIONS',
          'Access-Control-Allow-Headers': 'Authorization, Content-Type',
          'Access-Control-Max-Age': '86400',
        },
      });
    }

    if (request.method !== 'POST') return DROP;

    const auth = request.headers.get('Authorization') ?? '';
    const token = auth.startsWith('Bearer ') ? auth.slice(7) : '';
    if (!token || !env.JWT_SIGNING_KEY) return DROP;

    const claims = await verifyJwt(token, env.JWT_SIGNING_KEY);
    if (!claims) return DROP;

    if (claims.project !== env.PROJECT_ID) return DROP;

    const allowed = parseAllowedOrigins(env.ALLOWED_ORIGINS);
    if (!allowed.has(claims.origin) && !allowed.has('*')) return DROP;

    const reqOrigin = request.headers.get('Origin');
    if (reqOrigin && reqOrigin !== claims.origin && !allowed.has('*')) return DROP;

    const rateLimit = Number.parseInt(env.RATE_LIMIT_PER_MINUTE || '60', 10);
    if (!checkRateLimit(claims.sub, rateLimit)) return DROP;

    const maxBytes = Number.parseInt(env.MAX_BODY_BYTES || '65536', 10);
    const raw = await request.arrayBuffer();
    let parsed: unknown;
    try {
      parsed = JSON.parse(new TextDecoder().decode(raw));
    } catch {
      return DROP;
    }

    const schema = validatePayload(parsed, maxBytes, raw.byteLength);
    if (!schema.ok) return DROP;

    const enriched = await enrichExceptions(schema.value, env);

    const upstream = await fetch(env.ALLOY_INGEST_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Telemetry-Session': claims.sub,
        'X-Telemetry-Project': claims.project,
        'X-Telemetry-Origin': claims.origin,
      },
      body: JSON.stringify(enriched),
    });

    // Still return 204 to clients — don't surface Alloy errors to the browser.
    if (!upstream.ok) {
      console.error('alloy_forward_failed', upstream.status);
    }

    const origin = reqOrigin && (allowed.has(reqOrigin) || allowed.has('*')) ? reqOrigin : '';
    return new Response(null, {
      status: 204,
      headers: origin
        ? {
            'Access-Control-Allow-Origin': origin,
            Vary: 'Origin',
          }
        : undefined,
    });
  },
};

/** Helper for backend token issuance tests — not used at the edge. */
export async function signDevJwt(
  claims: JwtClaims,
  secret: string,
): Promise<string> {
  const header = bytesToB64url(new TextEncoder().encode(JSON.stringify({ alg: 'HS256', typ: 'JWT' })));
  const payload = bytesToB64url(new TextEncoder().encode(JSON.stringify(claims)));
  const key = await crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  );
  const sig = await crypto.subtle.sign(
    'HMAC',
    key,
    new TextEncoder().encode(`${header}.${payload}`),
  );
  return `${header}.${payload}.${bytesToB64url(sig)}`;
}
