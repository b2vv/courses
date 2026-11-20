// M1 delay service: holds GET /hold?run= until POST /release?run=. Diagnostic only.
import { createServer } from 'node:http';

const runs = new Map(); // run -> { count, held: [res], released: boolean, ack: number }
const RUN = /^[a-z0-9-]{1,40}$/;
const get = (run) => {
  let state = runs.get(run);
  if (!state) runs.set(run, (state = { count: 0, held: [], released: false, ack: 0 }));
  return state;
};
const json = (res, status, body) => {
  res.writeHead(status, { 'content-type': 'application/json' });
  res.end(JSON.stringify(body));
};
const late = (res, run) => {
  res.writeHead(200, { 'content-type': 'text/plain' });
  res.end(`late-${run}`);
};

const server = createServer((req, res) => {
  const url = new URL(req.url, 'http://m1-delay');
  const run = url.searchParams.get('run') ?? '';
  if (url.pathname === '/healthz') return json(res, 200, { ok: true });
  if (!RUN.test(run)) return json(res, 400, { error: 'bad run' });
  const state = get(run);
  if (req.method === 'GET' && url.pathname === '/hold') {
    state.count += 1;
    if (state.released) return late(res, run); // terminal latch
    state.held.push(res);
    res.on('close', () => {
      const i = state.held.indexOf(res);
      if (i >= 0) state.held.splice(i, 1);
    });
    return;
  }
  if (req.method === 'GET' && url.pathname === '/state')
    return json(res, 200, { count: state.count, held: state.held.length, released: state.released });
  if (req.method === 'POST' && url.pathname === '/release') {
    if (!state.released) {
      state.released = true;
      state.ack = state.held.length;
      for (const held of state.held.splice(0)) late(held, run);
    }
    return json(res, 200, { released: state.ack });
  }
  return json(res, 404, { error: 'not found' });
});

const shutdown = () => {
  for (const state of runs.values())
    for (const held of state.held.splice(0)) {
      held.writeHead(503, { 'content-type': 'text/plain' });
      held.end('shutdown');
    }
  server.close(() => process.exit(0));
  setTimeout(() => process.exit(0), 2000).unref();
};
process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);
server.listen(9000, '0.0.0.0');
