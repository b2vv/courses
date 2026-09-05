#!/bin/sh
set -eu
task_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
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
