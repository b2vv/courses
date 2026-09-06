#!/bin/sh
set -eu
docker compose exec -T api curl -fsS http://127.0.0.1:8080/readyz > /dev/null
docker compose exec -T media-worker curl -fsS http://127.0.0.1:8081/readyz > /dev/null
docker compose exec -T upload curl -fsS http://127.0.0.1:8082/readyz > /dev/null
ffmpeg_version="$(docker compose exec -T media-worker ffmpeg -version | sed -n '1p')"
printf '%s\n' "$ffmpeg_version"
case "$ffmpeg_version" in
  'ffmpeg version 5.1.'*) ;;
  *) echo "Unexpected worker ffmpeg version" >&2; exit 1 ;;
esac
docker compose exec -T frontend node --input-type=module <<'JS'
const origin = 'http://nginx:8080';
for (const [path, expected] of [['/healthz',200],['/',200],['/metrics',404],['/readyz',404]]) {
  const response = await fetch(origin+path);
  if(response.status!==expected) throw Error(`${path}: expected ${expected}, got ${response.status}`);
}
const unknownVideoId='10000000-0000-4000-8000-000000000001';
const unknownUpload=origin+'/uploads/'+unknownVideoId;
const head=await fetch(unknownUpload,{method:'HEAD',headers:{'Tus-Resumable':'1.0.0'}});
if(head.status!==404) throw Error(`/uploads/{unknown}: expected 404, got ${head.status}`);
const uploadPatchCount=async()=>{
  const metrics=await fetch('http://upload:8082/metrics').then(r=>r.text());
  return [...metrics.matchAll(/^courses_http_requests_total\{([^}]*)\}\s+([^\s]+)$/gm)]
    .filter(([,labels])=>labels.includes('method="PATCH"')&&labels.includes('route="/uploads/{videoId}"'))
    .reduce((total,[,,value])=>total+Number(value),0);
};
const patchCountBefore=await uploadPatchCount();
const oversized=await fetch(unknownUpload,{
  method:'PATCH',
  headers:{'Content-Type':'application/offset+octet-stream','Content-Length':String(32*1024*1024),'Tus-Resumable':'1.0.0','Upload-Offset':'0'},
  body:new Uint8Array(32*1024*1024),
});
if(oversized.status!==413) throw Error(`oversized upload chunk: expected 413, got ${oversized.status}`);
const patchCountAfter=await uploadPatchCount();
if(patchCountAfter!==patchCountBefore) throw Error(`oversized upload chunk reached upload service: counter moved from ${patchCountBefore} to ${patchCountAfter}`);
const adminLogin=await fetch(origin+'/api/v1/auth/login',{
  method:'POST',
  headers:{'Content-Type':'application/json',Origin:'http://localhost:8080'},
  body:JSON.stringify({email:'admin@courseforge.local',password:'courseforge-admin-demo'}),
});
if(adminLogin.status!==200) throw Error(`admin smoke login: expected 200, got ${adminLogin.status}`);
const adminCookie=adminLogin.headers.get('set-cookie')?.split(';',1)[0];
if(!adminCookie?.startsWith('courseforge_session=')) throw Error('admin smoke login did not set the session cookie');
for(const path of ['/media','/media/','/media/x/master.m3u8']) {
  for(const headers of [undefined,{Cookie:adminCookie}]) {
    const response=await fetch(origin+path,{headers});
    if(response.status!==403) throw Error(`${path}: expected 403, got ${response.status}`);
  }
}
const adminLogout=await fetch(origin+'/api/v1/auth/logout',{method:'POST',headers:{Cookie:adminCookie,Origin:'http://localhost:8080'}});
if(adminLogout.status!==204) throw Error(`admin smoke logout: expected 204, got ${adminLogout.status}`);
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
console.log('PASS readiness, upload proxy/body cap, private media/metrics, worker handler/metrics and API, login and SSR rate limits');
JS
