import { ingestClientLog } from '../../../adapters/api/generated/courses';
import type { ClientLogRequest } from '../../../adapters/api/generated/models/clientLogRequest';
import type { ClientLogReporter } from '../ports/client-log';

// A per-page-load storm guard. The nginx clientlog zone throttles the server
// side; this bounds the browser side so an error inside a render loop cannot
// fire thousands of POSTs from one tab. It is a different concern from the
// transport throttle, not a duplicate of it.
const sessionCeiling = 50;

const messageCap = 2000;
const stackCap = 8000;

interface Described {
  message: string;
  name?: string;
  stack?: string;
}

// describeUnknown turns any thrown/rejected value — an Error, a string, a bare
// `undefined`, or an arbitrary object — into a well-formed event. Each distinct
// value yields a distinct message, so distinct failures never collapse into one
// (there is no dedup; channel separation guarantees one event per failure).
export function describeUnknown(value: unknown): Described {
  if (value instanceof Error) {
    const described: Described = {
      message: truncate(value.message || value.name || 'Error', messageCap),
    };
    if (value.name) {
      described.name = value.name;
    }
    if (value.stack) {
      described.stack = truncate(value.stack, stackCap);
    }
    return described;
  }
  if (typeof value === 'string') {
    return { message: truncate(value || 'empty string rejection', messageCap) };
  }
  if (value === null) {
    return { message: 'null rejection' };
  }
  if (value === undefined) {
    return { message: 'undefined rejection' };
  }
  let serialized: string;
  try {
    serialized = JSON.stringify(value) ?? String(value);
  } catch {
    serialized = Object.prototype.toString.call(value);
  }
  return { message: truncate(`non-error rejection: ${serialized}`, messageCap) };
}

// createClientLogReporter builds the sink. It is created once, in the client
// entry, so its ceiling counter is a per-page-load closure, not a module-level
// singleton (SSR renders many requests in one process and never runs the client
// entry, so this never leaks across requests).
export function createClientLogReporter(): ClientLogReporter {
  let sent = 0;
  return {
    capture(error: unknown): void {
      if (sent >= sessionCeiling) {
        return;
      }
      sent += 1;
      const described = describeUnknown(error);
      const body: ClientLogRequest = {
        level: 'error',
        message: described.message,
        route: currentRoute(),
        release: process.env.APP_RELEASE ?? 'dev',
        userAgent: coarseUserAgent(),
      };
      if (described.name) {
        body.errorName = described.name;
      }
      if (described.stack) {
        body.errorStack = described.stack;
      }
      // Fire-and-forget: the server scrubs §17 and returns 204; any failure
      // (429/5xx/offline) is swallowed and never re-captured.
      void ingestClientLog(body).catch(() => {});
    },
  };
}

function currentRoute(): string {
  if (typeof window === 'undefined') {
    return 'unknown';
  }
  return window.location.pathname || 'unknown';
}

// coarseUserAgent reports the browser family only — no version, no platform
// detail — so the log line carries enough to group by browser without becoming
// a fingerprint.
function coarseUserAgent(): string {
  if (typeof navigator === 'undefined') {
    return 'unknown';
  }
  const ua = navigator.userAgent;
  if (/Edg\//.test(ua)) {
    return 'Edge';
  }
  if (/OPR\/|Opera/.test(ua)) {
    return 'Opera';
  }
  if (/Firefox\//.test(ua)) {
    return 'Firefox';
  }
  if (/Chrome\//.test(ua)) {
    return 'Chrome';
  }
  if (/Safari\//.test(ua)) {
    return 'Safari';
  }
  return 'Other';
}

function truncate(value: string, max: number): string {
  return value.length <= max ? value : value.slice(0, max);
}
