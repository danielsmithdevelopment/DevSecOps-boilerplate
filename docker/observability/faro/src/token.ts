export interface TelemetryToken {
  token: string;
  expiresAt: number; // unix seconds
}

export interface TokenClientOptions {
  /** Backend endpoint that returns `{ token, expires_at }` after session auth. */
  tokenUrl: string;
  /** Refresh this many seconds before expiry (default 120). */
  refreshSkewSeconds?: number;
  fetchImpl?: typeof fetch;
}

/**
 * Fetch and refresh short-lived JWTs for the ingest Worker.
 * Tokens live only in memory — never localStorage.
 */
export class TelemetryTokenClient {
  private token: TelemetryToken | null = null;
  private inflight: Promise<string> | null = null;
  private readonly refreshSkew: number;
  private readonly fetchImpl: typeof fetch;

  constructor(private readonly opts: TokenClientOptions) {
    this.refreshSkew = opts.refreshSkewSeconds ?? 120;
    this.fetchImpl = opts.fetchImpl ?? fetch.bind(globalThis);
  }

  async getToken(): Promise<string> {
    const now = Math.floor(Date.now() / 1000);
    if (this.token && this.token.expiresAt - this.refreshSkew > now) {
      return this.token.token;
    }
    if (this.inflight) return this.inflight;

    this.inflight = this.refresh()
      .then((t) => {
        this.inflight = null;
        return t;
      })
      .catch((err) => {
        this.inflight = null;
        throw err;
      });

    return this.inflight;
  }

  private async refresh(): Promise<string> {
    const res = await this.fetchImpl(this.opts.tokenUrl, {
      method: 'POST',
      credentials: 'include',
      headers: { Accept: 'application/json' },
    });
    if (!res.ok) {
      throw new Error(`telemetry_token_http_${res.status}`);
    }
    const body = (await res.json()) as {
      token?: string;
      expires_at?: number;
      expiresAt?: number;
    };
    if (!body.token) {
      throw new Error('telemetry_token_missing');
    }
    const expiresAt = body.expires_at ?? body.expiresAt ?? Math.floor(Date.now() / 1000) + 3600;
    this.token = { token: body.token, expiresAt };
    return body.token;
  }

  clear(): void {
    this.token = null;
  }
}
