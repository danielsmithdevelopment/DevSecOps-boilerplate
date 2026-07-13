export interface Env {
  JWT_SIGNING_KEY: string;
  ALLOY_INGEST_URL: string;
  ALLOWED_ORIGINS: string;
  PROJECT_ID: string;
  RATE_LIMIT_PER_MINUTE: string;
  MAX_BODY_BYTES: string;
  /** Optional R2 binding for release source maps. */
  SOURCEMAPS?: R2Bucket;
}

export interface JwtClaims {
  sub: string;
  project: string;
  origin: string;
  iat: number;
  exp: number;
}

export interface StackFrame {
  function?: string;
  filename?: string;
  lineno?: number;
  colno?: number;
  abs_path?: string;
}

export interface ExceptionPayload {
  type?: string;
  value?: string;
  stacktrace?: { frames?: StackFrame[] };
}

export interface ExceptionEvent {
  type: 'exception';
  payload: {
    exceptions?: ExceptionPayload[];
  };
  meta?: {
    labels?: Record<string, string>;
    app?: { name?: string; version?: string };
  };
}

export type FaroEvent = ExceptionEvent | { type: string; [key: string]: unknown };
