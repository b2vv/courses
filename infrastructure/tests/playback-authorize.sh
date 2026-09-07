#!/bin/sh
set -eu

run_hex="$(od -An -N6 -tx1 /dev/urandom | tr -d ' \n')"
playback_id="$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')"
course_id=10000000-0000-4000-8000-000000000001
module_id="60000000-0000-4000-8001-$run_hex"
lesson_id="60000000-0000-4000-8002-$run_hex"
video_id="60000000-0000-4000-8003-$run_hex"
admin_id="40000000-0000-4000-8004-$run_hex"
admin_email="playback-$run_hex@courseforge.local"
http_port="${HTTP_PORT:-8080}"
origin="http://127.0.0.1:${http_port}"
cookie_jar="$(mktemp /tmp/courses-t1-cookie.XXXXXX)"
chmod 600 "$cookie_jar"

restore() {
  set +e
  docker compose start postgres redis api nginx media-worker >/dev/null 2>&1
  attempts=0
  until docker compose exec -T postgres sh -c 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >/dev/null 2>&1 || [ "$attempts" -ge 30 ]; do
    attempts=$((attempts + 1))
    sleep 1
  done
  attempts=0
  until docker compose exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli ping' >/dev/null 2>&1 || [ "$attempts" -ge 30 ]; do
    attempts=$((attempts + 1))
    sleep 1
  done
  docker compose exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli DEL "$1" >/dev/null' sh "playback:owner:$admin_id" >/dev/null 2>&1
  docker compose exec -T media-worker sh -c 'rm -rf "$1"' sh "/media/hls/$video_id" >/dev/null 2>&1
  docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >/dev/null 2>&1 <<SQL
DELETE FROM video_assets WHERE id='$video_id';
DELETE FROM course_lessons WHERE id='$lesson_id';
DELETE FROM course_modules WHERE id='$module_id';
DELETE FROM users WHERE id='$admin_id';
SQL
  rm -f "$cookie_jar"
}
trap restore EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

wait_for_exec() {
  service="$1"
  shift
  attempts=0
  until docker compose exec -T "$service" "$@" >/dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ "$attempts" -ge 30 ]; then
      echo "Timed out waiting for $service" >&2
      exit 1
    fi
    sleep 1
  done
}

assert_status() {
  name="$1"
  expected="$2"
  shift 2
  actual="$(curl -sS -o /dev/null -w '%{http_code}' "$@")"
  if [ "$actual" != "$expected" ]; then
    echo "$name: expected $expected, got $actual" >&2
    exit 1
  fi
  printf '%s status=%s\n' "$name" "$actual"
}

docker compose exec -T nginx nginx -t
docker compose exec -T nginx sh -c 'test -z "$(find /var/cache/nginx -mindepth 1 -print -quit)"'
printf '%s\n' 'B37 nginx_config=valid cache_entries=0'

docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <<SQL
INSERT INTO users(id,email,display_name,role,password_hash)
VALUES ('$admin_id','$admin_email','Playback benchmark admin','admin','\$argon2id\$v=19\$m=65536,t=3,p=2\$YWRtaW4tZGVtby1zZWVkIQ\$obXgLD59DMQsNm2cOgKWcl050QIrNheSnkrYE3HUvT0');
INSERT INTO course_modules(id,course_id,title,position)
VALUES ('$module_id','$course_id','Playback benchmark',999)
ON CONFLICT (id) DO UPDATE SET course_id=EXCLUDED.course_id,title=EXCLUDED.title,position=EXCLUDED.position,updated_at=now();
INSERT INTO course_lessons(id,module_id,title,description,position)
VALUES ('$lesson_id','$module_id','Authorize benchmark','',0)
ON CONFLICT (id) DO UPDATE SET module_id=EXCLUDED.module_id,title=EXCLUDED.title,description=EXCLUDED.description,position=EXCLUDED.position,updated_at=now();
INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at)
VALUES ('$video_id','$lesson_id',1,'ready',1,1,'benchmark.mp4',ARRAY['360p'],now()+interval '1 day',now(),now())
ON CONFLICT (id) DO UPDATE SET lesson_id=EXCLUDED.lesson_id,state='ready',declared_bytes=1,received_bytes=1,renditions=ARRAY['360p'],evicted_at=NULL,updated_at=now();
SQL

docker compose exec -T media-worker sh <<SH
set -eu
directory=/media/hls/$video_id
mkdir -p "\$directory/v0"
printf '#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1\nv0/index.m3u8\n' > "\$directory/master.m3u8"
{
  printf '#EXTM3U\n#EXT-X-TARGETDURATION:6\n#EXT-X-MEDIA-SEQUENCE:0\n'
  index=0
  while [ "\$index" -lt 10 ]; do
    printf '#EXTINF:6.0,\nsegment_%03d.m4s\n' "\$index"
    printf 'x' > "\$directory/v0/segment_\$(printf '%03d' "\$index").m4s"
    index=\$((index + 1))
  done
  printf '#EXT-X-ENDLIST\n'
} > "\$directory/v0/index.m3u8"
SH

ownership="{\"playbackId\":\"$playback_id\",\"videoId\":\"$video_id\",\"lessonId\":\"$lesson_id\"}"
docker compose exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli SET "$1" "$2" EX 900 >/dev/null' sh "playback:owner:$admin_id" "$ownership"

docker compose exec -T frontend node --input-type=module <<JS
import { performance } from 'node:perf_hooks';

const edge = 'http://nginx:8080';
const api = 'http://api:8080';
const playbackId = '$playback_id';
const videoId = '$video_id';
const allowedPath = \`/media/p/\${playbackId}/hls/\${videoId}/v0/segment_000.m4s\`;

const login = await fetch(edge + '/api/v1/auth/login', {
  method: 'POST',
  headers: {'Content-Type':'application/json', Origin:'http://localhost:8080'},
  body: JSON.stringify({email:'$admin_email', password:'courseforge-admin-demo'}),
});
if (login.status !== 200) throw Error(\`login: expected 200, got \${login.status}\`);
const cookie = login.headers.get('set-cookie')?.split(';', 1)[0];
if (!cookie?.startsWith('courseforge_session=')) throw Error('login returned no session cookie');

const metric = async (name) => {
  const body = await fetch(api + '/metrics').then((response) => response.text());
  const line = body.split('\n').find((row) => row.startsWith(name + ' '));
  if (line === undefined) throw Error(\`missing metric \${name}\`);
  return Number(line.slice(name.length + 1));
};
const authorizeCount = async () => {
  const body = await fetch(api + '/metrics').then((response) => response.text());
  return [...body.matchAll(/^courses_http_requests_total\{([^}]*)\}\s+([^\s]+)$/gm)]
    .filter(([, labels]) => labels.includes('method="GET"') && labels.includes('route="GET /internal/playback/authorize"'))
    .reduce((total, [, , value]) => total + Number(value), 0);
};
const databaseQueryCount = async () => {
  const body = await fetch(api + '/metrics').then((response) => response.text());
  return [...body.matchAll(/^courses_database_query_duration_seconds_count\{[^}]*\}\s+([^\s]+)$/gm)]
    .reduce((total, [, value]) => total + Number(value), 0);
};
const percentile = (values, fraction) => {
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(sorted.length * fraction) - 1];
};
const run = async (requests, expected, pauseBetweenBatchesMs = 0) => {
  const latencies = [];
  const statuses = new Map();
  const started = performance.now();
  for (let offset = 0; offset < requests.length; offset += 20) {
    const batch = requests.slice(offset, offset + 20);
    await Promise.all(batch.map(async (request) => {
      const before = performance.now();
      const response = await fetch(request.url, {headers:request.headers});
      await response.arrayBuffer();
      latencies.push(performance.now() - before);
      statuses.set(response.status, (statuses.get(response.status) ?? 0) + 1);
    }));
    if (pauseBetweenBatchesMs > 0 && offset + 20 < requests.length) {
      await new Promise((resolve) => setTimeout(resolve, pauseBetweenBatchesMs));
    }
  }
  const elapsedSeconds = (performance.now() - started) / 1000;
  const result = {
		requests: requests.length,
    concurrency: 20,
    p50_ms: Number(percentile(latencies, .50).toFixed(3)),
    p95_ms: Number(percentile(latencies, .95).toFixed(3)),
    p99_ms: Number(percentile(latencies, .99).toFixed(3)),
    rps: Number((requests.length / elapsedSeconds).toFixed(3)),
    statuses: Object.fromEntries([...statuses.entries()].sort(([a], [b]) => a - b)),
  };
  if (statuses.size !== 1 || statuses.get(expected) !== requests.length) {
		throw Error(\`expected \${requests.length} x \${expected}, got \${JSON.stringify(result.statuses)}\`);
  }
  return result;
};

const authorizeURL = api + '/internal/playback/authorize';
const m1Requests = Array.from({length:400}, () => ({
  url: authorizeURL,
  headers:{Cookie:cookie, 'X-Playback-Id':playbackId, 'X-Video-Id':videoId},
}));
const m1 = await run(m1Requests, 204);
console.log('M1 ' + JSON.stringify(m1));

const condition3Before = await metric('courses_playback_authorization_database_checks_total');
const databaseBefore = await databaseQueryCount();
const randomRequests = Array.from({length:400}, (_, index) => {
  const randomPlayback = (BigInt(index) + 0x20000000000000000000000000000000n).toString(16).padStart(32, '0');
	return {url:authorizeURL, headers:{Cookie:cookie, 'X-Playback-Id':randomPlayback, 'X-Video-Id':videoId}};
});
const m2 = await run(randomRequests, 403);
const databaseAfter = await databaseQueryCount();
const condition3After = await metric('courses_playback_authorization_database_checks_total');
const condition3Queries = condition3After - condition3Before;
const allPostgresQueries = databaseAfter - databaseBefore;
console.log('M2 ' + JSON.stringify({...m2, all_postgres_queries:allPostgresQueries, condition3_postgres_queries:condition3Queries}));
if (allPostgresQueries < 400) throw Error(\`M2 expected identity to reach PostgreSQL at least 400 times, got \${allPostgresQueries}\`);
if (condition3Queries !== 0) throw Error(\`M2 reached condition-3 PostgreSQL \${condition3Queries} times\`);

const edgeRequest = {url:edge + allowedPath, headers:{Cookie:cookie}};
const b38 = await run(Array.from({length:400}, () => edgeRequest), 200, 200);
const falseDenials = (b38.statuses['403'] ?? 0) + (b38.statuses['503'] ?? 0);
console.log('B38 ' + JSON.stringify({...b38, pacing_ms_between_batches:200}));
console.log('M3 ' + JSON.stringify({false_403_or_503:falseDenials, requests:400}));
if (falseDenials !== 0) throw Error(\`M3 saw \${falseDenials} false denials\`);

const invalidBefore = await authorizeCount();
for (const path of [
  \`/media/p/\${playbackId}/hls/not-a-uuid/x.m4s\`,
  \`/media/p/zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz/hls/\${videoId}/x.m4s\`,
]) {
  const response = await fetch(edge + path, {headers:{Cookie:cookie}});
  await response.arrayBuffer();
  if (response.status !== 403) throw Error(\`R20 \${path}: expected 403, got \${response.status}\`);
}
const invalidAfter = await authorizeCount();
if (invalidAfter !== invalidBefore) throw Error(\`R20 reached Go: counter moved \${invalidBefore} -> \${invalidAfter}\`);
console.log('R20 statuses=403,403 go_requests=0');

const m4Before = await authorizeCount();
const minutePaths = [
  \`/media/p/\${playbackId}/hls/\${videoId}/master.m3u8\`,
  \`/media/p/\${playbackId}/hls/\${videoId}/v0/index.m3u8\`,
  ...Array.from({length:10}, (_, index) => \`/media/p/\${playbackId}/hls/\${videoId}/v0/segment_\${String(index).padStart(3,'0')}.m4s\`),
];
for (const path of minutePaths) {
  const response = await fetch(edge + path, {headers:{Cookie:cookie}});
  await response.arrayBuffer();
  if (response.status !== 200) throw Error(\`M4 \${path}: expected 200, got \${response.status}\`);
}
const m4After = await authorizeCount();
console.log('M4 ' + JSON.stringify({segment_seconds:6, segments_per_minute:10, startup_playlists:2, authorize_per_first_minute:m4After-m4Before}));
if (m4After - m4Before !== 12) throw Error(\`M4 expected 12 authorize calls, got \${m4After-m4Before}\`);

for (const path of [
  \`/media/p/\${playbackId}/hls/\${videoId}/master.m3u8\`,
  \`/media/p/\${playbackId}/hls/\${videoId}/v0/segment_000.m4s\`,
]) {
  const response = await fetch(edge + path);
  await response.arrayBuffer();
  if (response.status !== 403) throw Error(\`R01 \${path}: expected 403, got \${response.status}\`);
}
console.log('R01 playlist=403 segment=403');

const internal = await fetch(edge + '/internal/playback/authorize', {headers:{Cookie:cookie}});
await internal.arrayBuffer();
if (internal.status !== 404) throw Error(\`R10 expected 404, got \${internal.status}\`);
console.log('R10 status=404');
const logout = await fetch(edge + '/api/v1/auth/logout', {method:'POST', headers:{Cookie:cookie, Origin:'http://localhost:8080'}});
if (logout.status !== 204) throw Error(\`benchmark logout: expected 204, got \${logout.status}\`);
JS

curl -fsS -o /dev/null -c "$cookie_jar" -X POST \
  -H 'Content-Type: application/json' -H 'Origin: http://localhost:8080' \
  --data "{\"email\":\"$admin_email\",\"password\":\"courseforge-admin-demo\"}" \
  "$origin/api/v1/auth/login"
sleep 3

media_path="$origin/media/p/$playback_id/hls/$video_id/v0/segment_000.m4s"

docker compose stop api >/dev/null
assert_status R11 503 -b "$cookie_jar" "$media_path"
docker compose start api >/dev/null
wait_for_exec api curl -fsS http://127.0.0.1:8080/readyz

docker compose stop redis >/dev/null
assert_status R12 503 -b "$cookie_jar" "$media_path"
docker compose start redis >/dev/null
wait_for_exec redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli ping'
wait_for_exec api curl -fsS http://127.0.0.1:8080/readyz
docker compose exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli SET "$1" "$2" EX 900 >/dev/null' sh "playback:owner:$admin_id" "$ownership"

docker compose stop postgres >/dev/null
assert_status R13 503 -b "$cookie_jar" "$media_path"
docker compose start postgres >/dev/null
wait_for_exec postgres sh -c 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
wait_for_exec api curl -fsS http://127.0.0.1:8080/readyz

printf '%s\n' 'PASS B37 B38 R01 R10 R11 R12 R13 R20 and measurements M1-M4'
