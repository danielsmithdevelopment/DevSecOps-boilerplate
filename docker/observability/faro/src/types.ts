export interface StackFrame {
  function?: string;
  filename?: string;
  lineno?: number;
  colno?: number;
}

export interface ExceptionEvent {
  type: 'exception';
  payload: {
    exceptions?: Array<{
      type?: string;
      value?: string;
      stacktrace?: { frames?: StackFrame[] };
    }>;
  };
  meta?: {
    labels?: Record<string, string>;
  };
}

export type FaroTransportEvent = ExceptionEvent | { type: string; meta?: { labels?: Record<string, string> } };
