import { describe, expect, it } from 'vitest';
import { createErrorFingerprint, normaliseMessage } from './fingerprint';
import { validatePayload } from './schema';
import type { ExceptionEvent } from './types';

describe('normaliseMessage', () => {
  it('collapses numbers, hashes, and urls', () => {
    expect(normaliseMessage('User 12345 not found at https://x.test/a')).toBe(
      'User <n> not found at <url>',
    );
    expect(normaliseMessage('id=abcdef12 deadbeef')).toBe('id=<hash> <hash>');
  });
});

describe('createErrorFingerprint', () => {
  it('groups errors that differ only by dynamic message values', async () => {
    const frame = {
      function: 'fetchUser',
      filename: 'UserProfile.tsx',
      lineno: 47,
      colno: 12,
    };
    const a: ExceptionEvent = {
      type: 'exception',
      payload: {
        exceptions: [
          { type: 'Error', value: 'User 12345 not found', stacktrace: { frames: [frame] } },
        ],
      },
    };
    const b: ExceptionEvent = {
      type: 'exception',
      payload: {
        exceptions: [
          { type: 'Error', value: 'User 67890 not found', stacktrace: { frames: [frame] } },
        ],
      },
    };
    expect(await createErrorFingerprint(a)).toBe(await createErrorFingerprint(b));
    expect((await createErrorFingerprint(a)).length).toBe(16);
  });
});

describe('validatePayload', () => {
  it('accepts a single exception event', () => {
    const result = validatePayload({ type: 'exception', payload: {} }, 1024, 32);
    expect(result.ok).toBe(true);
  });

  it('rejects unknown event types and oversized bodies', () => {
    expect(validatePayload({ type: 'malware' }, 1024, 16).ok).toBe(false);
    expect(validatePayload({ type: 'exception' }, 10, 100).ok).toBe(false);
  });
});
