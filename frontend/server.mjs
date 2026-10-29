import { randomBytes } from 'node:crypto';
import { serve } from 'srvx';
import { staticMiddleware } from 'srvx/static';
import entry from './dist/server/index.js';
import { ssrRenderFailedLine, ssrRequestLine } from './server-log.mjs';

// form-action does NOT fall back to default-src, so without it a native <form> submit is
// unrestricted (T-F2). The checkout posts a native cross-origin form to WayForPay
// (purchase OrderStatusView.submitCheckoutForm); the allowed origins mirror the client-side
// checkout validator (modules/purchase/adapters/purchase-repository.ts): the production pay host is
// always allowed, and the offline e2e/stub host is added only when PUBLIC_WAYFORPAY_PAY_HOST is set
// — an unset value (production) cannot widen the allowlist. The port is '*' because it lives in the
// api's WAYFORPAY_PAY_URL, not here, and the validator matches on hostname alone.
const rawWayforpayPayHost = process.env.PUBLIC_WAYFORPAY_PAY_HOST?.trim() ?? '';
// Only a bare hostname may be interpolated into the policy: a stray space or ';' in a
// misconfigured deploy env would otherwise split the CSP into bogus directives. A value that is not
// a hostname is ignored (prod branch), so a bad env fails safe-closed — the e2e host is simply not
// added — rather than widening or corrupting the policy (SA13-C review N2).
const wayforpayPayHost = /^[a-z0-9.-]+$/i.test(rawWayforpayPayHost) ? rawWayforpayPayHost : '';
const formAction = wayforpayPayHost
  ? `form-action 'self' https://secure.wayforpay.com http://${wayforpayPayHost}:*`
  : `form-action 'self' https://secure.wayforpay.com`;
const contentSecurityPolicy = (nonce) =>
  `default-src 'self'; script-src 'self' 'nonce-${nonce}'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self' blob:; worker-src 'self' blob:; connect-src 'self'; object-src 'none'; base-uri 'self'; ${formAction}; frame-ancestors 'none'`;
// CSP enforces in production. Both app entries configure Zod jitless before schemas
// load, avoiding its new Function probe under script-src 'self' 'nonce-…'. The
// 2026-10-19 runtime close-out found that violation despite clean report-only runs.
// E2e uses CSP_REPORT_ONLY=true for axe/addStyleTag injections without the nonce.
// The 503/render_failed catch below covers shell-level throws and non-render
// entry.fetch rejections. Start 1.169.3 wraps route components in Suspense: route
// render errors resolve as HTTP 200 with the error deferred to the client.
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
        ssrRequestLine({
          status: response.status,
          durationMs: Math.round(performance.now() - started),
          route: path,
        }),
      );
      return response;
    } catch {
      process.stderr.write(ssrRenderFailedLine({ route: path }));
      return new Response('Rendering unavailable', {
        status: 503,
        headers: { 'Cache-Control': 'no-store' },
      });
    }
  },
});
await server.ready();
