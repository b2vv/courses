#!/bin/sh
set -eu
docker compose exec -T api curl -fsS http://127.0.0.1:8080/readyz > /dev/null
docker compose exec -T media-worker curl -fsS http://127.0.0.1:8081/readyz > /dev/null
ffmpeg_version="$(docker compose exec -T media-worker ffmpeg -version | sed -n '1p')"
printf '%s\n' "$ffmpeg_version"
case "$ffmpeg_version" in
  'ffmpeg version 5.1.'*) ;;
  *) echo "Unexpected worker ffmpeg version" >&2; exit 1 ;;
esac
docker compose exec -T frontend node --input-type=module <<'JS'
const origin = 'http://nginx:8080';
for (const [path, expected] of [['/healthz',200],['/',200],['/media/example/master.m3u8',403],['/metrics',404],['/readyz',404]]) {
  const response = await fetch(origin+path);
  if(response.status!==expected) throw Error(`${path}: expected ${expected}, got ${response.status}`);
}
const status=await fetch('http://media-worker:8081/status').then(r=>r.json());
if(status.stage!=='media'||!Array.isArray(status.handlers)||status.handlers.length!==1||status.handlers[0]!=='media:transcode') throw Error('Unexpected worker status');
const metrics=await fetch('http://media-worker:8081/metrics').then(r=>r.text());
for(const series of [
  'courses_transcoding_job_duration_seconds_count{outcome="ready"} 0',
  'courses_transcoding_job_duration_seconds_count{outcome="failed"} 0',
  'courses_transcoding_job_duration_seconds_count{outcome="cancelled"} 0',
  'courses_transcoding_failures_total{reason="unsupported_source"} 0',
  'courses_transcoding_failures_total{reason="probe_failed"} 0',
  'courses_transcoding_failures_total{reason="render_failed"} 0',
  'courses_transcoding_failures_total{reason="publish_failed"} 0',
  'courses_transcoding_failures_total{reason="storage_unavailable"} 0',
]) {
  if(!metrics.includes(series)) throw Error(`Missing zero-value worker metric: ${series}`);
}
const burst=await Promise.all(Array.from({length:60},()=>fetch(origin+'/api/v1/not-found').then(r=>r.status)));
if(!burst.includes(429)) throw Error('API rate limit not enforced');
// Login is metered far below the API rate because each attempt costs an
// Argon2id verification, paid even when the account does not exist.
const login=Array.from({length:30},()=>fetch(origin+'/api/v1/auth/login',{method:'POST',headers:{'Content-Type':'application/json',Origin:origin},body:JSON.stringify({email:'absent@courseforge.local',password:'x'})}).then(r=>r.status));
const logins=await Promise.all(login);
if(!logins.includes(429)) throw Error('Login rate limit not enforced');
if(logins.filter(s=>s===429).length<logins.length-10) throw Error('Login budget is not tighter than the API budget');
// SSR reaches the API on the internal network, so the page routes need their
// own budget or the catalog is an unmetered path into PostgreSQL.
const pages=await Promise.all(Array.from({length:400},()=>fetch(origin+'/').then(r=>r.status)));
if(!pages.includes(429)) throw Error('SSR rate limit not enforced');
console.log('PASS readiness, proxy, private media/metrics, worker handler/metrics and API, login and SSR rate limits');
JS
