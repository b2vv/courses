#!/bin/sh
set -eu
docker compose exec -T api curl -fsS http://127.0.0.1:8080/readyz > /dev/null
docker compose exec -T media-worker curl -fsS http://127.0.0.1:8081/readyz > /dev/null
docker compose exec -T frontend node --input-type=module <<'JS'
const origin = 'http://nginx:8080';
for (const [path, expected] of [['/healthz',200],['/',200],['/media/example/master.m3u8',403],['/metrics',404],['/readyz',404]]) {
  const response = await fetch(origin+path);
  if(response.status!==expected) throw Error(`${path}: expected ${expected}, got ${response.status}`);
}
const status=await fetch('http://media-worker:8081/status').then(r=>r.json());
if(!Array.isArray(status.handlers)||status.handlers.length!==0) throw Error('Unexpected worker handlers');
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
console.log('PASS readiness, proxy, private media/metrics, worker status and API, login and SSR rate limits');
JS
