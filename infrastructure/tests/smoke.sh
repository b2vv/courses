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
if(!burst.includes(429)) throw Error('Rate limit not enforced');
console.log('PASS readiness, proxy, private media/metrics, worker status and rate limit');
JS
