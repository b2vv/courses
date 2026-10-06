import { randomBytes } from 'node:crypto';
import { serve } from 'srvx';
import { staticMiddleware } from 'srvx/static';
import entry from './dist/server/index.js';

const contentSecurityPolicy = (nonce) =>
  `default-src 'self'; script-src 'self' 'nonce-${nonce}'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self' blob:; worker-src 'self' blob:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'`;
// Enforcing in production: the full e2e run (playback, ASR, media) reported zero
// CSP violations under report-only, so the directive is complete. The e2e sets
// CSP_REPORT_ONLY=true because @axe-core/playwright and addStyleTag inject their
// own script/style without the nonce — enforcing would block the test tooling, not
// any app resource. Real users never see those injections, so production enforces.
const contentSecurityPolicyHeader =
  process.env.CSP_REPORT_ONLY === 'true'
    ? 'Content-Security-Policy-Report-Only'
    : 'Content-Security-Policy';

const assets = staticMiddleware({
  dir: new URL('./dist/client', import.meta.url).pathname,
  dotfiles: false,
});
const server = serve({
  hostname: '0.0.0.0',
  port: process.env.PORT ?? 3000,
  gracefulShutdown: { gracefulTimeout: 10, forceTimeout: 12 },
  maxRequestBodySize: 64 * 1024,
  middleware: [assets],
  async fetch(request) {
    const path = new URL(request.url).pathname;
    if (path === '/healthz') return Response.json({ status: 'ok', service: 'frontend' });
    if (path.startsWith('/media/') || path.startsWith('/api/'))
      return new Response('Not found', { status: 404 });
    const started = performance.now();
    try {
      const nonce = randomBytes(32).toString('base64');
      const response = await entry.fetch(request, { context: { nonce } });
      response.headers.set('Cache-Control', 'private, no-store');
      response.headers.append('Vary', 'Cookie');
      response.headers.set(contentSecurityPolicyHeader, contentSecurityPolicy(nonce));
      process.stdout.write(
        JSON.stringify({
          level: 'info',
          service: 'frontend',
          event: 'http_request',
          status: response.status,
          durationMs: Math.round(performance.now() - started),
        }) + '\n',
      );
      return response;
    } catch {
      process.stderr.write(
        JSON.stringify({ level: 'error', service: 'frontend', event: 'render_failed' }) + '\n',
      );
      return new Response('Rendering unavailable', {
        status: 503,
        headers: { 'Cache-Control': 'no-store' },
      });
    }
  },
});
await server.ready();
