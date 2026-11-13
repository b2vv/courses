#!/bin/sh
set -eu
task_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
# R30 must run on the host before Docker: the Go test container mounts only
# backend/, so it cannot inspect the root-owned nginx configuration.
awk '
  function reset_statement(key) {
    for (key in tokens) delete tokens[key]
    token_count = 0
    token = ""
    token_started = 0
    statement = ""
    in_braced_variable = 0
  }
  function finish_token() {
    if (!token_started) return
    tokens[++token_count] = token
    token = ""
    token_started = 0
  }
  function validate_log_format(text, remaining, matched, variable) {
    log_format_safe_found = 1
    remaining = text
    while (match(remaining, /\$\{[[:alnum:]_]+\}|\$[[:alnum:]_]+/)) {
      matched = substr(remaining, RSTART, RLENGTH)
      variable = matched
      if (substr(matched, 2, 1) == "{") variable = "$" substr(matched, 3, length(matched) - 3)
      if (!(variable in allowed)) {
        print "nginx log_format safe contains unapproved variable " matched
        invalid = 1
      }
      remaining = substr(remaining, RSTART + RLENGTH)
    }
  }
  function validate_access_log(text) {
    if (token_count >= 3 && tokens[3] == "safe") {
      safe_access_log_found = 1
      return
    }
    if (token_count == 2 && tokens[2] == "off") return
    print "nginx access_log must use safe format or be off: " text
    invalid = 1
  }
  # A block is identified by its opening statement: "http", "server", "location ^~ /api/".
  function block_name(joined, slot) {
    joined = ""
    for (slot = 1; slot <= token_count; slot++) joined = joined (slot > 1 ? " " : "") tokens[slot]
    return joined
  }
  function inside(name, slot) {
    for (slot = 1; slot <= depth; slot++) if (block[slot] == name) return 1
    return 0
  }
  # nginx writes sizes with k/m/g suffixes and the domain writes bytes, so one side has to be
  # converted before they can be compared at all.
  function to_bytes(size, number, suffix) {
    number = size + 0
    suffix = tolower(substr(size, length(size), 1))
    if (suffix == "k") return number * 1024
    if (suffix == "m") return number * 1024 * 1024
    if (suffix == "g") return number * 1024 * 1024 * 1024
    return number
  }
  function validate_body_size(text) {
    if (token_count < 2) return
    if (block[depth] == "location ^~ /api/") {
      api_body_size = to_bytes(tokens[2])
      return
    }
    if (block[depth] == "http") http_body_size = to_bytes(tokens[2])
  }
  # gzip in http{} would reach the page routes, which stream with proxy_buffering off, and the
  # media locations, whose bytes are already compressed. It belongs to /api/ and nowhere else.
  function validate_gzip(text) {
    if (block[depth] == "location ^~ /api/") {
      gzip_in_api[tokens[1]] = 1
      return
    }
    print "nginx " tokens[1] " must live in location ^~ /api/, not in " block[depth] ": " text
    invalid = 1
  }
  function finish_statement() {
    finish_token()
    if (token_count >= 2 && tokens[1] == "log_format" && tokens[2] == "safe") {
      validate_log_format(statement)
    } else if (token_count >= 1 && tokens[1] == "access_log") {
      validate_access_log(statement)
    } else if (token_count >= 1 && tokens[1] == "client_max_body_size") {
      validate_body_size(statement)
    } else if (token_count >= 1 && tokens[1] ~ /^gzip/) {
      validate_gzip(statement)
    }
    reset_statement()
  }
  function scan_line(text, position, character) {
    for (position = 1; position <= length(text); position++) {
      character = substr(text, position, 1)
      if (escaped) {
        statement = statement character
        token = token character
        token_started = 1
        escaped = 0
        continue
      }
      if (character == "\\") {
        statement = statement character
        token_started = 1
        escaped = 1
        continue
      }
      if (!in_double && character == single_quote) {
        statement = statement character
        token_started = 1
        in_single = !in_single
        continue
      }
      if (!in_single && character == "\"") {
        statement = statement character
        token_started = 1
        in_double = !in_double
        continue
      }
      if (!in_single && !in_double && character == "#") {
        finish_token()
        return
      }
      if (!in_single && !in_double && character == ";") {
        finish_statement()
        continue
      }
      if (!in_single && !in_double && character == "{" &&
          substr(token, length(token), 1) == "$") {
        in_braced_variable = 1
      } else if (!in_single && !in_double && character == "}" && in_braced_variable) {
        in_braced_variable = 0
      } else if (!in_single && !in_double && !in_braced_variable && character == "{") {
        finish_token()
        # The statement before a brace names the block, so /api/ can be told from http{} and
        # from /uploads/. Without this the scanner saw three client_max_body_size directives and
        # could not say which belonged where - it reset on every brace and kept no context.
        depth++
        block[depth] = block_name()
        reset_statement()
        continue
      } else if (!in_single && !in_double && !in_braced_variable && character == "}") {
        finish_token()
        block[depth] = ""
        depth--
        reset_statement()
        continue
      } else if (!in_single && !in_double && character ~ /[[:space:]]/) {
        finish_token()
        statement = statement " "
        continue
      }
      statement = statement character
      token = token character
      token_started = 1
    }
  }
  BEGIN {
    single_quote = sprintf("%c", 39)
    allowed["$request_id"] = 1
    allowed["$request_method"] = 1
    allowed["$status"] = 1
    allowed["$request_time"] = 1
    gzip_required["gzip"] = 1
    gzip_required["gzip_proxied"] = 1
    gzip_required["gzip_types"] = 1
    gzip_required["gzip_vary"] = 1
  }
  {
    scan_line($0)
    if (!in_single && !in_double && !escaped) finish_token()
    statement = statement " "
  }
  END {
    if (token_count && tokens[1] == "access_log") {
      print "nginx access_log directive is missing a terminator: " statement
      invalid = 1
    }
    if (!log_format_safe_found) {
      print "nginx log_format safe not found or missing semicolon"
      invalid = 1
    }
    if (!safe_access_log_found) {
      print "nginx requires an access_log using safe; otherwise it defaults to combined and logs URI"
      invalid = 1
    }
    if (depth != 0) {
      print "nginx configuration blocks do not balance; the scanner cannot say which block a directive is in"
      invalid = 1
    }
    # Inherited is not declared. /api/ carrying no ceiling of its own means its limit is whatever
    # http{} happens to say, three screens away, and the transcript write path has to raise it
    # deliberately rather than discover it.
    if (api_body_size == 0) {
      print "nginx location ^~ /api/ must declare client_max_body_size explicitly rather than inherit it from http{}"
      invalid = 1
    }
    if (api_body_size > 0 && api_body_size < http_body_size) {
      print "nginx client_max_body_size for /api/ (" api_body_size " B) is below the http{} default (" http_body_size " B), so the narrower path is the API"
      invalid = 1
    }
    # gzip_proxied defaults to off and everything under /api/ is proxied; gzip_types defaults to
    # text/html alone; gzip_vary defaults to off. Each absent one is a compression that silently
    # does not happen, or a cache that serves gzip to a client that did not ask.
    for (required in gzip_required) {
      if (!(required in gzip_in_api)) {
        print "nginx location ^~ /api/ has gzip but not " required ", whose default undoes it"
        invalid = 1
      }
    }
    if (invalid) exit 1
  }
' "$task_root/infrastructure/nginx/nginx.conf" >&2
task_socket=$(docker context inspect --format '{{.Endpoints.docker.Host}}')
case "$task_socket" in unix://*) task_socket=${task_socket#unix://};; *) echo 'A local Docker Unix socket is required for integration tests.' >&2; exit 1;; esac
if [ "$(uname -s)" = Darwin ]; then
  set -- -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal
else
  set -- --network host -e TESTCONTAINERS_HOST_OVERRIDE=localhost
fi
# G37: the fixture credentials are documented as local-only and never reach
# shared configuration. The Go tests cannot assert this - their container
# mounts only backend/, so the repository root is invisible from inside them.
readme=$task_root/README.md
grep -q 'make seed-identities' "$readme" || { echo 'README must document make seed-identities' >&2; exit 1; }
grep -qi 'local development fixtures only' "$readme" || { echo 'README must state the fixtures are local-only' >&2; exit 1; }
for fixture in student instructor admin; do
  grep -q "$fixture@courseforge.local" "$readme" || { echo "README does not document $fixture@courseforge.local" >&2; exit 1; }
  secret=courseforge-$fixture-demo
  grep -q "$secret" "$readme" || { echo "README does not document the $fixture password" >&2; exit 1; }
  for shared in .env.example docker-compose.yml; do
    if grep -q "$secret" "$task_root/$shared"; then echo "$shared carries the $fixture fixture password" >&2; exit 1; fi
  done
done

# Ryuk is on; its create-to-start race is handled by internal/testsupport/containers.Start.
# TEMPORARY(testcontainers-go#3868): remove both when a release contains #3868 (backlog B-27).
# After a hard kill, list leftovers with `docker ps -a --filter label=org.testcontainers=true` and remove
# only this repository's.
# K5: keep the full `go test` output. A one-off failure (`TestAdministrationRepositoryAgainstPostgreSQL`
# during O05) was rerun green and its message was gone, because nothing persisted this stream. The
# log lives under backend/test-results/ (gitignored). The exit status is written out of band: under
# plain sh, `$?` after a pipe is tee's, not go test's, so a red gate would otherwise exit 0.
log_dir="$task_root/backend/test-results"
mkdir -p "$log_dir"
log="$log_dir/integration.log"
# The runner image adds ffmpeg/ffprobe to the Go toolchain so the real-adapter
# audio transcode tests run in the gate rather than skipping. The apt layer is
# cached, so only the first build downloads ffmpeg; later runs reuse the layer.
runner_image=courses-integration-go:1.27.2-ffmpeg
docker build -q -t "$runner_image" -f "$task_root/infrastructure/tests/Dockerfile.integration" "$task_root/infrastructure/tests" >&2
status_file=$(mktemp)
set +e
{
  docker run --rm "$@" \
    -v "$task_root/backend:/work" -w /work \
    -v "$task_socket:/var/run/docker.sock" \
    -e DOCKER_HOST=unix:///var/run/docker.sock \
    -e TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock \
    -v "${GO_MOD_CACHE:-courses-gomod}:/go/pkg/mod" -v "${GO_BUILD_CACHE:-courses-gobuild}:/root/.cache/go-build" \
    -e GOTOOLCHAIN=local \
    "$runner_image" go test -count=1 -tags=integration ./...
  echo $? > "$status_file"
} 2>&1 | tee "$log"
set -e
status=$(cat "$status_file")
rm -f "$status_file"
echo "integration log: $log"
exit "$status"
