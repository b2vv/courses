#!/bin/zsh
# M1 r2: full rerun after review r1, recording identity for every build and run. Diagnostic only.
S=<scratch>
P=$S/m1-probe
R=$S/r2
AMD=node@sha256:0e2646ab57e77b7a384a0cb59b6815b182c77e1f9c3ba36bd8435bd9c4e6680e
mkdir -p $R
. <kiri>/environment.sh >/dev/null 2>&1
BASECOMPOSE="$COMPOSE_FILE"
ARGS=(--build-arg PUBLIC_WAYFORPAY_PAY_HOST=localhost --build-arg APP_RELEASE=dev --progress=plain --load)
log() { print -r -- "$*" | tee -a $R/run.log; }
disk() { docker exec courses-postgres-1 df -BG / | tail -1 | awk '{print $4}' | tr -d G; }

build() { # build <tag> <context> <target> [platform]
  local tag=$1 ctx=$2 target=$3 plat=${4:-}
  local free=$(disk)
  log "BUILD $tag ctx=${ctx#$S/} target=$target platform=${plat:-native} free_before=${free}G"
  if (( free < 10 )); then log "BLOCKED disk below 10G before $tag"; exit 20; fi
  local p=(); [[ -n $plat ]] && p=(--platform $plat)
  docker build "${p[@]}" "${ARGS[@]}" -f $ctx/frontend/Dockerfile --target $target -t $tag $ctx > $R/build-${tag//[:\/]/_}.log 2>&1
  local rc=$?
  log "BUILD $tag exit=$rc id=$(docker image inspect $tag --format '{{.Id}} {{.Os}}/{{.Architecture}}' 2>&1) free_after=$(disk)G"
  return $rc
}

swap() { # swap <tag> [platform]
  local tag=$1 plat=${2:-linux/arm64}
  COMPOSE_FILE="$BASECOMPOSE:$P/m1-override.yml" M1_FRONT_IMAGE=$tag M1_FRONT_PLATFORM=$plat \
    docker compose up -d --no-build --no-deps m1-delay frontend > $R/swap.tmp 2>&1
  local rc=$? s=
  for i in {1..90}; do s=$(docker inspect courses-frontend-1 --format '{{.State.Health.Status}}'); [[ $s == healthy ]] && break; sleep 2; done
  local want=$(docker image inspect $tag --format '{{.Id}}') run=$(docker inspect courses-frontend-1 --format '{{.Image}}')
  local inarch=$(docker exec courses-frontend-1 sh -c 'uname -m; node -p process.arch' | xargs)
  local hz=$(docker exec courses-frontend-1 node -e "fetch('http://127.0.0.1:3000/healthz').then(r=>console.log(r.status)).catch(()=>console.log('ERR'))")
  local pg=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/courses)
  log "SWAP $tag up_exit=$rc health=$s image=$want running=$run match=$([[ $want == $run ]] && echo yes || echo NO) platform=$plat in_container='$inarch' healthz=$hz courses=$pg"
  [[ $want == $run && $s == healthy ]] || { log "SWAP FAILED $tag"; return 1; }
}

runp() { # runp <label> <cmd...>
  local label=$1; shift
  "$@" > $R/$label.out 2>&1; local rc=$?
  log "RUN $label exit=$rc :: $(grep -oE '"verdict":"[A-Z_]+"|C3 verdict: [A-Z]+' $R/$label.out | tail -1)"
}
c2() { docker run --rm --network courses_default -v $P:/p:ro node:24-bookworm-slim node /p/c2-probe.mjs $1 $2; }

log "START r2 $(date -u +%FT%TZ) free=$(disk)G orig_frontend=$(docker inspect courses-frontend-1 --format '{{.Image}}')"
log "BUILDX $(docker buildx ls 2>&1 | grep -E 'kiri-build0' | head -1 | xargs)"

# ---- arm64 images (plan instrument) ----
build m1-front:base $S/m1-base0 production || exit 21
build m1-front:cand $S/m1-amd production || exit 21
build m1-probe:base $S/m1-base production || exit 21
build m1-probe:cand $S/m1-cand production || exit 21
( cd $S/m1-base/frontend && sed -i '' 's|^      return response;$|      return new Response(await response.arrayBuffer(), response); // M1 control: materialize body|' server.mjs && git diff --stat server.mjs | tail -1 )
log "CONTROL delta sha256=$(cd $S/m1-base/frontend && git diff server.mjs | shasum -a 256 | cut -d' ' -f1)"
build m1-probe:control $S/m1-base production; crc=$?
( cd $S/m1-base/frontend && git checkout -- server.mjs )
(( crc == 0 )) || exit 21

for t in base cand; do
  swap m1-front:$t || exit 22
  runp c1-admin-$t node $P/c1-admin-spa.mjs
  runp c1-catalog-$t node $P/c1-spa.mjs
  runp c3-$t $P/c3-compress.sh
done
for t in base cand; do
  swap m1-probe:$t || exit 22
  runp c2-$t-identity c2 r2-$t-identity-$(date +%s) identity
  runp c2-$t-gzip c2 r2-$t-gzip-$(date +%s) gzip
  runp c4-$t $P/c4-throw.sh $t
  runp c4-browser-$t node $P/c4-browser.mjs
done
swap m1-probe:control || exit 22
runp c2-control-identity c2 r2-control-identity-$(date +%s) identity

# ---- C5 plan instrument: buildx --platform linux/amd64 (record what it produces) ----
build m1-front:cand-amd64-buildx $S/m1-amd production linux/amd64
log "C5-BUILDX image arch: $(docker image inspect m1-front:cand-amd64-buildx --format '{{.Os}}/{{.Architecture}}') content: $(docker run --rm --platform linux/amd64 m1-front:cand-amd64-buildx sh -c 'uname -m; node -p process.arch; od -An -tx1 -j18 -N2 /usr/local/bin/node' | xargs)"
grep -E 'FROM docker.io' $R/build-m1-front_cand-amd64-buildx.log | head -2 | tee -a $R/run.log

# ---- C5 supplementary: Dockerfile stage commands by hand on real x86_64 ----
manual() { # manual <src-frontend-dir> <outdir> <label>
  local src=$1 out=$2 label=$3
  log "MANUAL $label src=${src#$S/} free_before=$(disk)G"
  mkdir -p $out
  docker run --rm --platform linux/amd64 -v $src:/src:ro -v $out:/out $AMD bash -c '
    set -o pipefail
    echo "uname=$(uname -m) node=$(node -p "process.version+\" \"+process.platform+\" \"+process.arch") npm=$(npm --version)"
    mkdir /app && cd /app && cp /src/package.json /src/package-lock.json . && sha256sum package-lock.json
    npm ci > /tmp/ci.log 2>&1; echo "STEP npm_ci exit=$?"
    npm ls @tanstack/react-router --parseable; echo "STEP npm_ls_router exit=$?"
    (cd /src && tar --exclude=node_modules --exclude=dist -cf - .) | tar -xf -; echo "STEP copy_src exit=$?"
    echo "rspack x64 binding e_machine: $(od -An -tx1 -j18 -N2 node_modules/@rsbuild/core/node_modules/@rspack/binding-linux-x64-gnu/*.node)"
    PUBLIC_WAYFORPAY_PAY_HOST=localhost APP_RELEASE=dev npm run build > /tmp/build.log 2>&1; echo "STEP npm_run_build exit=$?"
    grep -E "built in" /tmp/build.log
    npm ci --omit=dev > /tmp/ci2.log 2>&1; echo "STEP npm_ci_omit_dev exit=$?"
    rm -rf /out/app && mkdir -p /out/app && cp -r dist node_modules package.json server.mjs server-log.mjs /out/app/; echo "STEP export exit=$?"
  ' > $R/manual-$label.log 2>&1
  log "MANUAL $label exit=$? steps: $(grep -E '^STEP' $R/manual-$label.log | xargs) free_after=$(disk)G"
  ! grep -qE '^STEP .* exit=[1-9]' $R/manual-$label.log
}
commitimg() { # commitimg <outdir> <tag>
  local c=$(docker create --platform linux/amd64 $AMD true)
  docker cp $1/app $c:/app >/dev/null && docker commit --change 'WORKDIR /app' --change 'USER node' --change 'ENV NODE_ENV=production PORT=3000' --change 'CMD ["node","server.mjs"]' $c $2 >/dev/null
  local rc=$?; docker rm $c >/dev/null
  log "COMMIT $2 exit=$rc id=$(docker image inspect $2 --format '{{.Id}} {{.Os}}/{{.Architecture}}') content: $(docker run --rm --platform linux/amd64 $2 sh -c 'uname -m; node -p process.arch; od -An -tx1 -j18 -N2 /usr/local/bin/node' | xargs)"
  return $rc
}
docker pull --platform linux/amd64 $AMD > /dev/null 2>&1; log "PULL amd64 node exit=$? local tag node:24-bookworm-slim=$(docker image inspect node:24-bookworm-slim --format '{{.Id}} {{.Architecture}}')"
manual $S/m1-amd/frontend $S/r2-amd-out clean || exit 23
commitimg $S/r2-amd-out m1-front:cand-amd64-manual || exit 23
manual $S/m1-cand/frontend $S/r2-amd-probe-out probe || exit 23
commitimg $S/r2-amd-probe-out m1-probe:cand-amd64-manual || exit 23
swap m1-front:cand-amd64-manual linux/amd64 || exit 22
runp c1-admin-amd64 node $P/c1-admin-spa.mjs
runp c1-catalog-amd64 node $P/c1-spa.mjs
runp c3-amd64 $P/c3-compress.sh
swap m1-probe:cand-amd64-manual linux/amd64 || exit 22
runp c2-amd64-identity c2 r2-amd64-identity-$(date +%s) identity
runp c2-amd64-gzip c2 r2-amd64-gzip-$(date +%s) gzip

log "END r2 $(date -u +%FT%TZ) free=$(disk)G"
