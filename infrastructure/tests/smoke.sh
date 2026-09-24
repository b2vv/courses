#!/bin/sh
set -eu
docker compose exec -T api curl -fsS http://127.0.0.1:8080/readyz > /dev/null
docker compose exec -T media-worker curl -fsS http://127.0.0.1:8081/readyz > /dev/null
docker compose exec -T upload curl -fsS http://127.0.0.1:8082/readyz > /dev/null
docker compose exec -T asr-worker curl -fsS http://127.0.0.1:8083/readyz > /dev/null
ffmpeg_version="$(docker compose exec -T media-worker ffmpeg -version | sed -n '1p')"
printf '%s\n' "$ffmpeg_version"
case "$ffmpeg_version" in
  'ffmpeg version 5.1.'*) ;;
  *) echo "Unexpected worker ffmpeg version" >&2; exit 1 ;;
esac
# C20: asr-worker image carries ffmpeg, Python with faster-whisper, model weights; runs as 10001
asr_ffmpeg="$(docker compose exec -T asr-worker ffmpeg -version | sed -n '1p')"
case "$asr_ffmpeg" in
  'ffmpeg version 5.1.'*) ;;
  *) echo "Unexpected asr-worker ffmpeg version: $asr_ffmpeg" >&2; exit 1 ;;
esac
asr_uid="$(docker compose exec -T asr-worker id -u)"
case "$asr_uid" in
  10001*) ;;
  *) echo "asr-worker runs as uid $asr_uid, expected 10001" >&2; exit 1 ;;
esac
docker compose exec -T asr-worker /opt/whisper/bin/python3 -c "import faster_whisper" > /dev/null
docker compose exec -T asr-worker sh -eu -c 'test -f "$WHISPER_MODEL_DIR/model.bin"'
docker compose exec -T asr-worker sh -eu <<'SH'
audio="/tmp/asr-model-smoke-$$.wav"
speech_audio="/tmp/asr-speech-smoke-$$.wav"
trap 'rm -f "$audio" "$speech_audio"' EXIT
if ! ffmpeg -nostdin -hide_banner -loglevel error \
  -f lavfi -i 'sine=frequency=440:sample_rate=16000:duration=1' \
  -ac 1 -ar 16000 -c:a pcm_s16le -f wav -y "$audio"; then
  echo "ASR engine smoke failed: could not generate the 16 kHz mono input" >&2
  exit 1
fi
if ! result="$(/opt/whisper/bin/python3 /opt/whisper/transcribe.py \
  --model "$WHISPER_MODEL_DIR" --audio "$audio")"; then
  echo "ASR engine smoke failed: transcribe.py could not load or run the model" >&2
  exit 1
fi
if ! printf '%s\n' "$result" | /opt/whisper/bin/python3 -c '
import json
import sys

def reject_non_json_number(value):
    raise ValueError(f"{value} is not a valid JSON number")

result = json.load(sys.stdin, parse_constant=reject_non_json_number)
speech_seconds = result.get("speech_seconds")
if isinstance(speech_seconds, bool) or not isinstance(speech_seconds, (int, float)):
    raise SystemExit("speech_seconds is not numeric")
'; then
  echo "ASR engine smoke failed: stdout is not valid JSON with numeric speech_seconds" >&2
  exit 1
fi
if ! ffmpeg -nostdin -hide_banner -loglevel error \
  -f lavfi -i "flite=text='Welcome to the course. Today we will learn how video transcripts work.':voice=slt" \
  -ar 16000 -ac 1 -y "$speech_audio"; then
  echo "ASR engine smoke failed: could not generate the speech input" >&2
  exit 1
fi
if ! speech_result="$(/opt/whisper/bin/python3 /opt/whisper/transcribe.py \
  --model "$WHISPER_MODEL_DIR" --audio "$speech_audio")"; then
  echo "ASR engine smoke failed: transcribe.py could not recognise speech" >&2
  exit 1
fi
if ! printf '%s\n' "$speech_result" | /opt/whisper/bin/python3 -c '
import json
import sys

def reject_non_json_number(value):
    raise ValueError(f"{value} is not a valid JSON number")

result = json.load(sys.stdin, parse_constant=reject_non_json_number)
speech_seconds = result.get("speech_seconds")
segments = result.get("segments")
if isinstance(speech_seconds, bool) or not isinstance(speech_seconds, (int, float)) or speech_seconds < 1:
    raise SystemExit("speech_seconds must be at least 1")
if not isinstance(segments, list) or not any(isinstance(segment, dict) and segment.get("text", "").strip() for segment in segments):
    raise SystemExit("speech result must contain a non-empty segment")
'; then
  echo "ASR engine smoke failed: speech JSON did not contain a non-empty segment" >&2
  exit 1
fi
SH
asr_status="$(docker compose exec -T asr-worker curl -fsS http://127.0.0.1:8083/status)"
printf 'asr-worker status: %s\n' "$asr_status"
docker compose exec -T frontend node --input-type=module <<'JS'
import {writeFile} from 'node:fs/promises';
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
const status=await fetch('http://media-worker:8081/status').then(r=>r.json());
if(status.stage!=='media'||!Array.isArray(status.handlers)||status.handlers.length!==1||status.handlers[0]!=='media:transcode') throw Error('Unexpected worker status');
const metrics=await fetch('http://media-worker:8081/metrics').then(r=>r.text());
// The claim is that the worker REGISTERS every label combination at startup, rather than
// creating a series lazily on first use - a counter that appears only after the first failure
// blinds a dashboard exactly when it is needed. Matching a trailing ' 0' stated that through
// the value instead, which is true only on a worker that has never run a job: on CI the stack
// is always fresh, so the hidden precondition held there and failed every local `make check`
// that followed an e2e run. The trailing space keeps this anchored to the end of the labels.
for(const series of [
  'courses_transcoding_job_duration_seconds_count{outcome="ready"} ',
  'courses_transcoding_job_duration_seconds_count{outcome="failed"} ',
  'courses_transcoding_job_duration_seconds_count{outcome="cancelled"} ',
  'courses_transcoding_failures_total{reason="unsupported_source"} ',
  'courses_transcoding_failures_total{reason="probe_failed"} ',
  'courses_transcoding_failures_total{reason="render_failed"} ',
  'courses_transcoding_failures_total{reason="publish_failed"} ',
  'courses_transcoding_failures_total{reason="storage_unavailable"} ',
]) {
  if(!metrics.includes(series)) throw Error(`Worker metric series is not registered: ${series}`);
}
// C15: asr queue visible in courses_queue_depth as a label value, media still present
const asrMetrics=await fetch('http://asr-worker:8083/metrics').then(r=>r.text());
if(!asrMetrics.includes('courses_queue_depth{queue="asr"')) throw Error('C15: courses_queue_depth missing queue="asr" label');
if(!metrics.includes('courses_queue_depth{queue="media"')) throw Error('C15: courses_queue_depth lost queue="media" after asr-worker added');
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
await writeFile('/tmp/courses-smoke-media-cookie',adminCookie,{mode:0o600});
// G2: compression is enabled for /api/ only, and both halves of that are asserted. The header
// is the contract - that nginx compresses JSON when the client offers gzip, and leaves it alone
// when the client does not - while `Vary` is what keeps a shared cache from handing a gzip body
// to a client that never asked.
//
// The page route carries the other half. It streams with `proxy_buffering off`, so compressing
// it would buffer the stream to compress it, and the assertion that it is NOT compressed is
// what proves the directives sit in `location ^~ /api/` rather than in `http{}`. A placement
// this test cannot see would look identical in the config and behave differently here.
//
// Deliberately absent: the media half. At this point in the gate no HLS exists on the volume,
// so every /media/ request answers 403, and an assertion that a 403 carries no
// `Content-Encoding` passes under any configuration at all - green while blind. It belongs where
// real bytes exist, which is e2e.
const compressed = await fetch(origin+'/api/v1/courses?limit=20', {headers:{'Accept-Encoding':'gzip'}});
if(compressed.status!==200) throw Error(`gzip probe: expected 200, got ${compressed.status}`);
if(compressed.headers.get('content-encoding')!=='gzip') throw Error(`/api/ did not compress JSON for a client offering gzip: content-encoding=${compressed.headers.get('content-encoding')}`);
if(!(compressed.headers.get('vary')??'').toLowerCase().includes('accept-encoding')) throw Error('/api/ compressed without Vary: Accept-Encoding, so a shared cache can serve gzip to a client that did not ask');
const identity = await fetch(origin+'/api/v1/courses?limit=20', {headers:{'Accept-Encoding':'identity'}});
if(identity.status!==200) throw Error(`identity probe: expected 200, got ${identity.status}`);
if(identity.headers.get('content-encoding')!==null) throw Error(`/api/ compressed for a client that asked for identity: ${identity.headers.get('content-encoding')}`);
const page = await fetch(origin+'/', {headers:{'Accept-Encoding':'gzip'}});
if(page.status!==200) throw Error(`page probe: expected 200, got ${page.status}`);
if(page.headers.get('content-encoding')!==null) throw Error('the streamed page route is compressed, so gzip is not confined to location ^~ /api/');
console.log('PASS gzip on /api/ JSON with Vary, absent for identity, and absent on the streamed page route');
console.log('PASS readiness, upload proxy/body cap, private media/metrics, worker handler/metrics and API, login and SSR rate limits');
JS

redis_stopped=false
media_session_created=true
wait_for_media_policy() {
  docker compose exec -T frontend node --input-type=module <<'JS'
import {readFile} from 'node:fs/promises';
const mediaPath='http://nginx:8080/media/p/10000000000040008000000000000001/hls/10000000-0000-4000-8000-000000000001/v0/segment_000.m4s';
const cookie=await readFile('/tmp/courses-smoke-media-cookie','utf8');
const deadline=Date.now()+30000;
let lastStatus='no response';
while(Date.now()<deadline) {
  try {
    lastStatus=(await fetch(mediaPath,{headers:{Cookie:cookie}})).status;
    if(lastStatus===403) {
      console.log('PASS Redis recovery restored media authorization policy (503 -> 403)');
      process.exit(0);
    }
  } catch(error) {
    lastStatus=error instanceof Error?error.message:String(error);
  }
  await new Promise(resolve=>setTimeout(resolve,250));
}
throw Error(`Redis recovery did not restore media policy within 30s; last result: ${lastStatus}`);
JS
}
restore_redis() {
  docker compose start redis > /dev/null
  wait_for_media_policy
}
close_media_session() {
  docker compose exec -T frontend node --input-type=module <<'JS'
import {readFile,unlink} from 'node:fs/promises';
const cookiePath='/tmp/courses-smoke-media-cookie';
let cookie;
try {
  cookie=await readFile(cookiePath,'utf8');
} catch(error) {
  if(error && typeof error==='object' && 'code' in error && error.code==='ENOENT') process.exit(0);
  throw error;
}
let logoutError;
try {
  const response=await fetch('http://api:8080/api/v1/auth/logout',{method:'POST',headers:{Cookie:cookie,Origin:'http://localhost:8080'}});
  if(response.status!==204) logoutError=Error(`admin smoke logout: expected 204, got ${response.status}`);
} catch(error) {
  logoutError=error;
}
await unlink(cookiePath);
if(logoutError) throw logoutError;
JS
}
restore_smoke_state_on_exit() {
  original_status=$?
  trap - EXIT HUP INT TERM
  set +e
  restore_status=0
  session_status=0
  if [ "$redis_stopped" = true ]; then
    restore_redis
    restore_status=$?
  fi
  if [ "$media_session_created" = true ]; then
    close_media_session
    session_status=$?
  fi
  if [ "$original_status" -ne 0 ]; then
    exit "$original_status"
  fi
  if [ "$restore_status" -ne 0 ]; then
    exit "$restore_status"
  fi
  exit "$session_status"
}

trap restore_smoke_state_on_exit EXIT HUP INT TERM
redis_stopped=true
docker compose stop redis > /dev/null
docker compose exec -T frontend node --input-type=module <<'JS'
import {readFile} from 'node:fs/promises';
const mediaPath='http://nginx:8080/media/p/10000000000040008000000000000001/hls/10000000-0000-4000-8000-000000000001/v0/segment_000.m4s';
const cookie=await readFile('/tmp/courses-smoke-media-cookie','utf8');
// This request uses an unknown playbackId, so no media exists regardless of Redis state.
// Proof of "no media bytes" moved to e2e T16 by user decision 2026-09-12.
const response=await fetch(mediaPath,{headers:{Cookie:cookie}});
if(response.status!==503) throw Error(`media with Redis unavailable: expected 503, got ${response.status}`);
if(response.headers.get('retry-after')===null) throw Error('media with Redis unavailable: Retry-After is missing');
console.log('PASS unavailable Redis returns 503 with Retry-After');
JS
restore_redis
redis_stopped=false
close_media_session
media_session_created=false
trap - EXIT HUP INT TERM

docker compose exec -T frontend node --input-type=module <<'JS'
const origin='http://nginx:8080';
const mediaPath='/media/p/10000000000040008000000000000001/hls/10000000-0000-4000-8000-000000000001/v0/segment_000.m4s';
// The media zone admits a burst of 240 and refills at 120 requests/second. Sending
// three bursts at once requires four seconds of refill to avoid a rejection.
const statuses=await Promise.all(Array.from({length:720},()=>fetch(origin+mediaPath).then(response=>response.status)));
// Requiring both statuses proves the limit sheds only excess requests before authorization instead of rejecting every request.
if(!statuses.includes(429)) throw Error('Media rate limit not enforced');
if(!statuses.includes(403)) throw Error('Media rate limit did not admit a request to authorization');
console.log('PASS media rate limit');
JS
