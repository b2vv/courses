import { serve } from 'srvx';
import { staticMiddleware } from 'srvx/static';
import entry from './dist/server/index.js';

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
      const response = await entry.fetch(request);
      response.headers.set('Cache-Control', 'private, no-store');
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
