import { initializeFaro } from '@grafana/faro-web-sdk';
import { createErrorFingerprint } from './fingerprint';
import { TelemetryTokenClient } from './token';
import type { ExceptionEvent, FaroTransportEvent } from './types';

export interface InitGatedFaroOptions {
  /** Cloudflare Worker ingest URL (never a public Alloy / Sentry DSN). */
  workerUrl: string;
  /** Backend endpoint that mints session-scoped JWTs. */
  tokenUrl: string;
  app: { name: string; version: string };
  environment?: string;
  /** Extra Faro options forwarded to initializeFaro (except url / beforeSend). */
  faroOptions?: Record<string, unknown>;
}

/**
 * Initialize Grafana Faro against the JWT-gated Worker proxy.
 *
 * - Fetches a short-lived JWT from your backend
 * - Attaches it as Authorization on every transport to the Worker URL
 * - Awaits Web Crypto fingerprints so Loki labels match the Worker
 */
export async function initGatedFaro(opts: InitGatedFaroOptions) {
  const tokens = new TelemetryTokenClient({ tokenUrl: opts.tokenUrl });
  await tokens.getToken();

  const originalFetch = globalThis.fetch.bind(globalThis);
  globalThis.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
    if (url.startsWith(opts.workerUrl)) {
      const token = await tokens.getToken();
      const headers = new Headers(
        init?.headers ?? (input instanceof Request ? input.headers : undefined),
      );
      headers.set('Authorization', `Bearer ${token}`);
      return originalFetch(input, { ...init, headers });
    }
    return originalFetch(input, init);
  };

  return initializeFaro({
    url: opts.workerUrl,
    app: opts.app,
    ...(opts.faroOptions ?? {}),
    // Faro accepts Promise-returning beforeSend in recent SDK versions.
    beforeSend: async (event: FaroTransportEvent) => {
      if (event.type === 'exception') {
        const fingerprint = await createErrorFingerprint(event as ExceptionEvent);
        event.meta = {
          ...event.meta,
          labels: {
            ...(event.meta?.labels ?? {}),
            error_fingerprint: fingerprint,
            release: opts.app.version,
            environment: opts.environment ?? 'production',
          },
        };
      }
      return event;
    },
  } as Parameters<typeof initializeFaro>[0]);
}

export { createErrorFingerprint, normaliseMessage } from './fingerprint';
export { TelemetryTokenClient } from './token';
