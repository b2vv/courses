#!/bin/zsh
# M1 C5 (owner 2026-11-20: real amd64 Dockerfile build): task-owned buildx builder that advertises linux/amd64,
# whose RUN steps execute through the VM's Rosetta binfmt handler. kiri-build and the global context are untouched.
S=<scratch>
P=$S/m1-probe
R=$S/r2
. <kiri>/environment.sh >/dev/null 2>&1
BASECOMPOSE="$COMPOSE_FILE"
log() { print -r -- "$*" | tee -a $R/c5-rosetta.log; }
disk() { docker exec courses-postgres-1 df -BG / | tail -1 | awk '{print $4}' | tr -d G; }
ARGS=(--builder m1-rosetta --platform linux/amd64 --build-arg PUBLIC_WAYFORPAY_PAY_HOST=localhost --build-arg APP_RELEASE=dev --progress=plain --load)

log "START c5-rosetta $(date -u +%FT%TZ) free=$(disk)G"
log "BUILDER image=$(docker image inspect m1-buildkit-norqemu:local --format {{.Id}}) qemu_files=$(docker run --rm --entrypoint sh m1-buildkit-norqemu:local -c 'ls /usr/bin | grep -c qemu')"
docker buildx inspect m1-rosetta --bootstrap > $R/c5r-builder-inspect.log 2>&1; log "BUILDER bootstrap exit=$? $(grep -E 'Platforms' $R/c5r-builder-inspect.log | xargs)"

build() { # build <tag> <ctx> <target>
  local free=$(disk); log "BUILD $1 ctx=${2#$S/} target=$3 free_before=${free}G"
  (( free < 10 )) && { log "BLOCKED disk below 10G"; return 20; }
  docker buildx build "${ARGS[@]}" -f $2/frontend/Dockerfile --target $3 -t $1 $2 > $R/c5-build-${1//[:\/]/_}.log 2>&1; local rc=$?
  log "BUILD $1 exit=$rc id=$(docker image inspect $1 --format '{{.Id}} {{.Os}}/{{.Architecture}}' 2>&1) free_after=$(disk)G"
  log "  content: $(docker run --rm --platform linux/amd64 $1 sh -c 'uname -m; node -p "process.version+\" \"+process.platform+\" \"+process.arch"; od -An -tx1 -j18 -N2 /usr/local/bin/node' 2>&1 | xargs)"
  return $rc
}
rc=0
build m1-deps:amd64-rosetta $S/m1-amd dependencies || rc=1
if (( rc == 0 )); then
  log "DEPS stage: $(docker run --rm --platform linux/amd64 m1-deps:amd64-rosetta bash -c 'npm --version; sha256sum package-lock.json | cut -c1-16; npm ls @tanstack/react-router --parseable; od -An -tx1 -j18 -N2 node_modules/@rsbuild/core/node_modules/@rspack/binding-linux-x64-gnu/*.node' 2>&1 | xargs)"
  grep -E 'RUN npm ci|DONE|ERROR' $R/c5-build-m1-deps_amd64-rosetta.log | tail -3 | tee -a $R/c5-rosetta.log
fi
(( rc == 0 )) && { build m1-front:amd64-rosetta $S/m1-amd production || rc=1; }
(( rc == 0 )) && { grep -E 'built in|ERROR' $R/c5-build-m1-front_amd64-rosetta.log | tee -a $R/c5-rosetta.log; }
(( rc == 0 )) && { build m1-probe:amd64-rosetta $S/m1-cand production || rc=1; }

swap() {
  COMPOSE_FILE="$BASECOMPOSE:$P/m1-override.yml" M1_FRONT_IMAGE=$1 M1_FRONT_PLATFORM=linux/amd64 \
    docker compose up -d --no-build --no-deps m1-delay frontend > /dev/null 2>&1; local up=$? s=
  for i in {1..90}; do s=$(docker inspect courses-frontend-1 --format '{{.State.Health.Status}}'); [[ $s == healthy ]] && break; sleep 2; done
  local want=$(docker image inspect $1 --format '{{.Id}}') run=$(docker inspect courses-frontend-1 --format '{{.Image}}')
  log "SWAP $1 up_exit=$up health=$s image=$want running=$run match=$([[ $want == $run ]] && echo yes || echo NO) in_container='$(docker exec courses-frontend-1 sh -c 'uname -m; node -p process.arch' | xargs)' healthz=$(docker exec courses-frontend-1 node -e "fetch('http://127.0.0.1:3000/healthz').then(r=>console.log(r.status)).catch(()=>console.log('ERR'))") courses=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/courses)"
  [[ $want == $run && $s == healthy ]]
}
runp() { local l=$1; shift; "$@" > $R/$l.out 2>&1; log "RUN $l exit=$? :: $(grep -oE '"verdict":"[A-Z_]+"|C3 verdict: [A-Z]+' $R/$l.out | tail -1)"; }
c2() { docker run --rm --network courses_default -v $P:/p:ro node:24-bookworm-slim node /p/c2-probe.mjs $1 $2; }
if (( rc == 0 )); then
  swap m1-front:amd64-rosetta && { runp c1-admin-amd64rosetta node $P/c1-admin-spa.mjs; runp c1-catalog-amd64rosetta node $P/c1-spa.mjs; runp c3-amd64rosetta $P/c3-compress.sh; }
  swap m1-probe:amd64-rosetta && { runp c2-amd64rosetta-identity c2 r2-amdreal-identity-$(date +%s) identity; runp c2-amd64rosetta-gzip c2 r2-amdreal-gzip-$(date +%s) gzip; runp c4-amd64rosetta $P/c4-throw.sh amd64rosetta; runp c4-browser-amd64rosetta node $P/c4-browser.mjs; }
fi
docker compose up -d --no-build --no-deps frontend > /dev/null 2>&1; COMPOSE_FILE="$BASECOMPOSE:$P/m1-override.yml" docker compose rm -sf m1-delay > /dev/null 2>&1
for i in {1..90}; do [[ $(docker inspect courses-frontend-1 --format "{{.State.Health.Status}}") == healthy ]] && break; sleep 2; done
log "RESTORE frontend=$(docker inspect courses-frontend-1 --format "{{.Image}} {{.State.Health.Status}}")"
docker buildx rm m1-rosetta > /dev/null 2>&1; log "BUILDER removed exit=$? remaining: $(docker buildx ls 2>&1 | grep -c m1-rosetta)"
log "END c5-rosetta $(date -u +%FT%TZ) free=$(disk)G rc=$rc"
