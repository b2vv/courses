// Pure builders for the SSR process's structured log lines. server.mjs writes
// the success line to stdout and the failure line to stderr. They live in a
// copied-verbatim .mjs beside server.mjs (the Dockerfile ships both) so they are
// unit-testable without starting the server. Both carry source:"ssr" — the
// field promtail turns into a Loki label, splitting SSR from the api/worker/
// client lines — plus the runtime release and the request route.

function release() {
  return process.env.APP_RELEASE ?? 'dev';
}

export function ssrRequestLine({ status, durationMs, route }) {
  return (
    JSON.stringify({
      level: 'info',
      service: 'frontend',
      source: 'ssr',
      event: 'http_request',
      status,
      durationMs,
      route,
      release: release(),
    }) + '\n'
  );
}

export function ssrRenderFailedLine({ route }) {
  return (
    JSON.stringify({
      level: 'error',
      service: 'frontend',
      source: 'ssr',
      event: 'render_failed',
      route,
      release: release(),
    }) + '\n'
  );
}
