#!/bin/zsh
# M1: reproduce r2's manual amd64 build (m1-r2.sh manual(), verbatim steps) keeping every build log; N runs.
S=<scratch>
R=$S/r3
AMD=node@sha256:0e2646ab57e77b7a384a0cb59b6815b182c77e1f9c3ba36bd8435bd9c4e6680e
. <kiri>/environment.sh >/dev/null 2>&1
log() { print -r -- "$*" | tee -a $R/manual-repro.log; }
disk() { docker exec courses-postgres-1 df -BG / | tail -1 | awk '{print $4}' | tr -d G; }
log "START $(date -u +%FT%TZ) free=$(disk)G"
docker pull --platform linux/amd64 $AMD > $R/pull.log 2>&1; log "PULL exit=$? arch=$(docker image inspect $AMD --format '{{.Architecture}}')"
for i in 1 2 3 4 5 6; do
  free=$(disk); (( free < 10 )) && { log "BLOCKED disk ${free}G"; break; }
  out=$R/run$i; mkdir -p $out
  docker run --rm --platform linux/amd64 -v $S/m1-amd/frontend:/src:ro -v $out:/out $AMD bash -c '
    set -o pipefail
    echo "uname=$(uname -m) node=$(node -p "process.version+\" \"+process.platform+\" \"+process.arch") npm=$(npm --version) nproc=$(nproc) mem=$(free -m | awk "/Mem/{print \$2\"M total \"\$7\"M avail\"}")"
    mkdir /app && cd /app && cp /src/package.json /src/package-lock.json . && sha256sum package-lock.json
    npm ci > /out/ci.log 2>&1; echo "STEP npm_ci exit=$?"
    (cd /src && tar --exclude=node_modules --exclude=dist -cf - .) | tar -xf -; echo "STEP copy_src exit=$?"
    PUBLIC_WAYFORPAY_PAY_HOST=localhost APP_RELEASE=dev npm run build > /out/build.log 2>&1; echo "STEP npm_run_build exit=$?"
    grep -E "built in" /out/build.log
  ' > $out/console.log 2>&1
  log "RUN $i docker_exit=$? $(grep -E '^STEP|built in|^uname' $out/console.log | xargs) free_before=${free}G free_after=$(disk)G"
done
log "END $(date -u +%FT%TZ) free=$(disk)G"
