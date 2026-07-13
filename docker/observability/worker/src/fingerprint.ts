import type { ExceptionEvent } from './types';

/**
 * Strip dynamic values so "User 12345 not found" and "User 67890 not found"
 * collapse to the same fingerprint — Sentry-style grouping for Loki labels.
 */
export function normaliseMessage(msg: string): string {
  return msg
    .replace(/\b[0-9a-f]{8,}\b/gi, '<hash>')
    .replace(/\b\d+\b/g, '<n>')
    .replace(/https?:\/\/\S+/g, '<url>');
}

/**
 * Web Crypto SHA-256 fingerprint. Workers have no Node `crypto` module; use
 * `crypto.subtle` so hashing stays inside the CPU budget during error storms.
 */
export async function createErrorFingerprint(event: ExceptionEvent): Promise<string> {
  const err = event.payload.exceptions?.[0];
  const topFrame = err?.stacktrace?.frames?.at(-1);

  const raw = [
    err?.type ?? 'UnknownError',
    normaliseMessage(err?.value ?? ''),
    topFrame?.function ?? '',
    topFrame?.filename ?? '',
  ].join('|');

  const encoded = new TextEncoder().encode(raw);
  const hashBuffer = await crypto.subtle.digest('SHA-256', encoded);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
    .slice(0, 16);
}
