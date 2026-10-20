#!/bin/sh
# Proves the monitoring stack end to end on the production path (it cannot run
# inside `make check`, whose `make up` has no --profile monitoring). It brings
# the profile up, generates one line per source, asserts each source is a
# queryable Loki label (AC4), asserts Grafana booted with both datasources and
# the overview dashboard from repo provisioning (AC5), and asserts every
# monitoring service is docker-network-only with no host port (AC6). Mirrors the
# exec-into-the-network technique of infrastructure/tests/image-authorize.sh.
set -eu

# Activate the monitoring profile via the env var (portable across compose
# versions, unlike the --profile flag's placement). Services without a profile
# still start; the monitoring services start too.
export COMPOSE_PROFILES=monitoring

http_port="${HTTP_PORT:-8080}"
origin="http://127.0.0.1:${http_port}"

# Reach loki/grafana (no host port) from inside the network through the nginx
# container's busybox wget.
innet() { docker compose exec -T nginx wget -q -O- "$@"; }

retry() { # retry <seconds> <desc> <cmd...>
  deadline=$(( $(date +%s) + $1 )); desc="$2"; shift 2
  until "$@" >/dev/null 2>&1; do
    if [ "$(date +%s)" -ge "$deadline" ]; then
      echo "timeout waiting for ${desc}" >&2
      return 1
    fi
    sleep 2
  done
}

echo '== bringing the monitoring profile up (base + loki/promtail/grafana) =='
docker compose up -d --build --wait --wait-timeout 180
# Migrate so the api read path answers 200; a 500 would still log an api line,
# but a clean run keeps the proof honest.
docker compose run --rm migrate >/dev/null

echo '== waiting for loki and grafana readiness =='
retry 90 'loki ready' innet 'http://loki:3100/ready'
retry 90 'grafana health' innet 'http://admin:admin@grafana:3000/api/health'

echo '== generating one line per source =='
curl -sS -o /dev/null "$origin/" || true                       # SSR -> source=ssr
curl -sS -o /dev/null "$origin/api/v1/courses" || true         # api -> source=api
curl -sS -o /dev/null -X POST -H 'Content-Type: application/json' \
  --data '{"level":"error","message":"monitoring-stack probe"}' \
  "$origin/api/v1/client-logs" || true                         # client -> source=client

echo '== asserting each source is a queryable Loki label (AC4) =='
source_values() {
  now=$(date +%s); start=$(( (now - 3600) * 1000000000 )); end=$(( now * 1000000000 ))
  innet "http://loki:3100/loki/api/v1/label/source/values?start=${start}&end=${end}"
}
has_source() { source_values | grep -q "\"$1\""; }
sources_ready() { has_source ssr && has_source api && has_source client && has_source media-worker; }
retry 90 'all sources indexed in loki' sources_ready
printf 'AC4 source labels: %s\n' "$(source_values)"

echo '== asserting Grafana datasources + dashboard from provisioning (AC5) =='
datasources="$(innet 'http://admin:admin@grafana:3000/api/datasources')"
echo "$datasources" | grep -q '"type":"prometheus"' || { echo 'AC5: prometheus datasource missing' >&2; exit 1; }
echo "$datasources" | grep -q '"type":"loki"' || { echo 'AC5: loki datasource missing' >&2; exit 1; }
dashboards="$(innet 'http://admin:admin@grafana:3000/api/search?type=dash-db')"
echo "$dashboards" | grep -q 'platform-overview' || { echo 'AC5: overview dashboard not provisioned' >&2; exit 1; }
echo 'AC5 datasources=prometheus,loki dashboard=platform-overview'

echo '== asserting monitoring services are network-only (AC6) =='
check_no_host_port() { # svc container-port
  # An unmapped port prints empty, ":0", or "invalid IP:0"; a real publish is an
  # IP:PORT with a non-zero port. Only the latter fails the network-only check.
  mapping="$(docker compose port "$1" "$2" 2>/dev/null || true)"
  case "$mapping" in
    *:[1-9]*)
      echo "AC6: $1 publishes a host port ($mapping)" >&2
      exit 1
      ;;
    *) : ;;
  esac
}
check_no_host_port grafana 3000
check_no_host_port loki 3100
check_no_host_port prometheus 9090
check_no_host_port promtail 9080
echo 'AC6 grafana/loki/prometheus/promtail: no host port'

echo 'PASS monitoring-stack AC4(source labels) AC5(datasources+dashboard) AC6(network-only)'
