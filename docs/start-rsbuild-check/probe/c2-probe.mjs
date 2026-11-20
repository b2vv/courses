// M1 C2 probe: raw HTTP/1.1 over net to nginx, delay-service control. Diagnostic only.
// usage: node c2-probe.mjs <runId> <identity|gzip> [nginxHost=nginx] [delayHost=m1-delay]
import net from 'node:net';
import zlib from 'node:zlib';

const [run, enc = 'identity', nginxHost = 'nginx', delayHost = 'm1-delay'] = process.argv.slice(2);
if (!/^[a-z0-9-]{1,40}$/.test(run ?? '')) throw new Error('bad run id');
const T0 = performance.now();
const now = () => Math.round(performance.now() - T0);
const ev = { dispatch: null, headers: null, shell: null, late: null, end: null, release: null, held: null };
const out = { run, enc, status: null, contentType: null, contentEncoding: null, errors: [], events: ev };
const SHELL = 'data-m1-shell="1"';
const LATE = `data-m1-late="late-${run}"`;
let text = '';
let lateBeforeRelease = false;
let shellBeforeRelease = false;

async function control(method, path) {
  const ctl = new AbortController();
  const t = setTimeout(() => ctl.abort(), 2000);
  try {
    const r = await fetch(`http://${delayHost}:9000${path}?run=${run}`, { method, signal: ctl.signal });
    if (r.status !== 200) throw new Error(`${path} status ${r.status}`);
    const body = await r.json();
    const isInt = (v) => Number.isInteger(v) && v >= 0;
    const valid =
      path === '/release'
        ? isInt(body?.released)
        : isInt(body?.count) && isInt(body?.held) && typeof body?.released === 'boolean';
    if (!valid) throw new Error(`${path} invalid body ${JSON.stringify(body).slice(0, 80)}`);
    return body;
  } finally {
    clearTimeout(t);
  }
}

const socket = net.connect(8080, nginxHost);
let gunzip = null;
const onText = (s) => {
  text += s;
  if (ev.shell === null && text.includes(SHELL)) ev.shell = now();
  if (ev.late === null && text.includes(LATE)) {
    ev.late = now();
    if (ev.release === null) lateBeforeRelease = true;
  }
};
// incremental HTTP/1.1 response parser (headers + chunked or close-delimited body)
let raw = Buffer.alloc(0);
let headersDone = false;
let chunked = false;
let bodyDone = false;
const feedBody = (buf) => (gunzip ? gunzip.write(buf) : onText(buf.toString('utf8')));
const parse = () => {
  if (!headersDone) {
    const i = raw.indexOf('\r\n\r\n');
    if (i < 0) return;
    const head = raw.subarray(0, i).toString('latin1').split('\r\n');
    raw = raw.subarray(i + 4);
    headersDone = true;
    ev.headers = now();
    out.status = Number(head[0].split(' ')[1]);
    const h = Object.fromEntries(head.slice(1).map((l) => [l.slice(0, l.indexOf(':')).toLowerCase(), l.slice(l.indexOf(':') + 1).trim()]));
    out.contentType = h['content-type'] ?? null;
    out.contentEncoding = h['content-encoding'] ?? null;
    out.transferEncoding = h['transfer-encoding'] ?? null;
    chunked = /chunked/i.test(h['transfer-encoding'] ?? '');
    if (out.contentEncoding === 'gzip') {
      gunzip = zlib.createGunzip();
      gunzip.on('data', (d) => onText(d.toString('utf8')));
      gunzip.on('error', (e) => out.errors.push(`gunzip: ${e.message}`));
    }
  }
  if (!chunked) {
    if (raw.length) feedBody(raw), (raw = Buffer.alloc(0));
    return;
  }
  for (;;) {
    const i = raw.indexOf('\r\n');
    if (i < 0) return;
    const size = parseInt(raw.subarray(0, i).toString('latin1'), 16);
    if (Number.isNaN(size)) return out.errors.push('bad chunk size');
    if (size === 0) {
      bodyDone = true;
      ev.end = now();
      gunzip?.end();
      return;
    }
    if (raw.length < i + 2 + size + 2) return;
    feedBody(raw.subarray(i + 2, i + 2 + size));
    raw = raw.subarray(i + 2 + size + 2);
  }
};
socket.on('data', (d) => {
  raw = Buffer.concat([raw, d]);
  parse();
});
socket.on('end', () => {
  if (!chunked && headersDone && ev.end === null) ((bodyDone = true), (ev.end = now()), gunzip?.end());
});
socket.on('error', (e) => out.errors.push(`socket: ${e.message}`));
socket.on('connect', () => {
  ev.dispatch = now();
  socket.write(
    `GET /m1-stream?run=${run} HTTP/1.1\r\nHost: localhost:8080\r\nAccept: text/html\r\nAccept-Encoding: ${enc === 'gzip' ? 'gzip' : 'identity'}\r\nConnection: close\r\n\r\n`,
  );
});

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
// Hard run cap: every wait in the main flow checks it, so no loop outlives 15 s (fail closed).
const CAP_MS = 15000;
const checkCap = () => {
  if (now() > CAP_MS) throw new Error('15s cap exceeded');
};
const cap = setTimeout(() => socket.destroy(new Error('15s cap')), CAP_MS + 500);
let released = false;
try {
  try {
    // t0 = first state with held == 1 (deadline 3 s after dispatch)
    while (ev.held === null) {
      checkCap();
      if (ev.dispatch === null && now() > 3000) throw new Error('NOT_TESTED: connect not within 3s');
      if (ev.dispatch !== null && now() - ev.dispatch > 3000) throw new Error('NOT_TESTED: no hold within 3s');
      try {
        const s = await control('GET', '/state');
        if (s.held === 1) ev.held = now();
        if (s.held > 1) throw new Error(`duplicate hold ${s.held}`);
      } catch (e) {
        // a duplicate hold, a malformed control body or the cap fail closed; only a timeout/network blip retries
        if (/duplicate|invalid body|cap exceeded|status/.test(String(e.message))) throw e;
        out.errors.push(`state: ${e.message}`);
      }
      if (ev.held === null) await sleep(50);
    }
    await sleep(Math.max(0, ev.held + 1000 - now()));
    checkCap();
    shellBeforeRelease = ev.shell !== null;
    out.lateBeforeRelease = ev.late !== null;
    ev.release = now();
    released = true;
    out.releaseAck = await control('POST', '/release');
    const deadline = ev.release + 5000;
    while (now() < Math.min(deadline, CAP_MS) && !(ev.late !== null && bodyDone)) await sleep(20);
  } finally {
    if (!released) {
      try {
        out.releaseAck = await control('POST', '/release');
      } catch (e) {
        out.errors.push(`release(finally): ${e.message}`);
      }
    }
    try {
      out.finalState = await control('GET', '/state');
    } catch (e) {
      out.errors.push(`final state: ${e.message}`);
    }
  }
} catch (e) {
  out.errors.push(e.message);
} finally {
  clearTimeout(cap);
  socket.destroy();
}
out.shellBeforeRelease = shellBeforeRelease;
out.bodyBytes = text.length;
out.bodyComplete = bodyDone && /<\/html>\s*$/i.test(text);
const a = out.status === 200 && /text\/html/.test(out.contentType ?? '');
const b = shellBeforeRelease;
const c = out.lateBeforeRelease === false && !lateBeforeRelease;
const d = ev.late !== null && ev.release !== null && ev.late - ev.release <= 3000 && out.bodyComplete && ev.end - ev.release <= 5000;
const e = out.finalState?.count === 1 && out.releaseAck?.released === 1;
out.criteria = { a, b, c, d, e };
out.verdict = a && b && c && d && e && out.errors.length === 0 ? 'PASS' : 'FAIL';
console.log(JSON.stringify(out));
