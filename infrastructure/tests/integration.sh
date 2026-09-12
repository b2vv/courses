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
  function finish_statement() {
    finish_token()
    if (token_count >= 2 && tokens[1] == "log_format" && tokens[2] == "safe") {
      validate_log_format(statement)
    } else if (token_count >= 1 && tokens[1] == "access_log") {
      validate_access_log(statement)
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
      } else if (!in_single && !in_double && !in_braced_variable &&
                 (character == "{" || character == "}")) {
        finish_token()
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

docker run --rm "$@" \
  -v "$task_root/backend:/work" -w /work \
  -v "$task_socket:/var/run/docker.sock" \
  -e DOCKER_HOST=unix:///var/run/docker.sock \
  -e TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock \
  -v courses-gomod:/go/pkg/mod -v courses-gobuild:/root/.cache/go-build \
  golang:1.27.1-bookworm go test -count=1 -tags=integration ./...
