#!/bin/zsh
# C5 supplementary: the Dockerfile's stage commands, run by hand in a real linux/amd64 node:24 container.
. <kiri>/environment.sh >/dev/null 2>&1
echo "START $(date -u +%FT%TZ) df=$(docker exec courses-postgres-1 df -h / | tail -1)"
docker pull --platform linux/amd64 node@sha256:0e2646ab57e77b7a384a0cb59b6815b182c77e1f9c3ba36bd8435bd9c4e6680e >/dev/null; echo "PULL_EXIT=$?"
docker run --rm --platform linux/amd64 -v <scratch>/m1-amd/frontend:/src:ro -v <scratch>/m1-amd5-out:/out node@sha256:0e2646ab57e77b7a384a0cb59b6815b182c77e1f9c3ba36bd8435bd9c4e6680e bash -ec '
  uname -m; node -p "process.version+\" \"+process.platform+\" \"+process.arch"; npm --version
  mkdir /app && cd /app && cp /src/package.json /src/package-lock.json . && sha256sum package-lock.json
  npm ci >/tmp/ci.log 2>&1 || { tail -30 /tmp/ci.log; exit 11; }; echo "deps: npm ci ok"
  npm ls @tanstack/react-router --parseable
  (cd /src && tar --exclude=node_modules --exclude=dist -cf - .) | tar -xf -
  PUBLIC_WAYFORPAY_PAY_HOST=localhost APP_RELEASE=dev npm run build 2>&1 | grep -E "built in|error|Error" | head -8
  test -f dist/server/index.js || exit 12
  echo "rspack bindings: $(ls node_modules/@rsbuild/core/node_modules/@rspack/ | xargs)"; echo "x64 binding e_machine: $(od -An -tx1 -j18 -N2 node_modules/@rsbuild/core/node_modules/@rspack/binding-linux-x64-gnu/*.node)"
  npm ci --omit=dev >/tmp/ci2.log 2>&1 || { tail -30 /tmp/ci2.log; exit 13; }; echo "runtime deps: npm ci --omit=dev ok"
  rm -rf /out/app && mkdir -p /out/app && cp -r dist node_modules package.json server.mjs server-log.mjs /out/app/
'; echo "MANUAL_EXIT=$?"
echo "END $(date -u +%FT%TZ) df=$(docker exec courses-postgres-1 df -h / | tail -1)"
