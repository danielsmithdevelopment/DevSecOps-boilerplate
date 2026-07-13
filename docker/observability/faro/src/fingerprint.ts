import type { ExceptionEvent } from './types';

/**
 * Strip dynamic values so similar exceptions share one Loki stream label.
 */
export function normaliseMessage(msg: string): string {
  return msg
    .replace(/\b[0-9a-f]{8,}\b/gi, '<hash>')
    .replace(/\b\d+\b/g, '<n>')
    .replace(/https?:\/\/\S+/g, '<url>');
}

/**
 * Web Crypto SHA-256 fingerprint — works in modern browsers (same algorithm
 * as docker/observability/worker/src/fingerprint.ts).
 */
export async function createErrorFingerprint(event: ExceptionEvent): Promise<string> {
  const err = event.payload?.exceptions?.[0];
  const frames = err?.stacktrace?.frames ?? [];
  const topFrame = frames[frames.length - 1];

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
