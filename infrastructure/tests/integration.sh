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
docker run --rm "$@" \
  -v "$task_root/backend:/work" -w /work \
  -v "$task_socket:/var/run/docker.sock" \
  -e DOCKER_HOST=unix:///var/run/docker.sock \
  -e TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock \
  -v courses-gomod:/go/pkg/mod -v courses-gobuild:/root/.cache/go-build \
  golang:1.27.1-bookworm go test -count=1 -tags=integration ./...
