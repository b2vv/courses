#!/bin/zsh
# M1 C4: render-error contract via nginx; usage: c4-throw.sh <label>. Diagnostic only.
set -u
. <kiri>/environment.sh >/dev/null 2>&1
L=$1; T=$(mktemp -d); O=http://localhost:8080
since=$(date -u +%Y-%m-%dT%H:%M:%S.000000000Z); sleep 1
for route in /m1-throw /courses; do
  curl -sS -D $T/h -o $T/b -H 'Accept: text/html' "$O$route"
  echo "$L $route: status=$(head -1 $T/h | cut -d' ' -f2) bytes=$(wc -c <$T/b | tr -d ' ') complete=$(tail -c 300 $T/b | grep -c '</html>') sentinelInHtml=$(grep -o 'm1-render-probe' $T/b | wc -l | tr -d ' ')"
done
sleep 1; until=$(date -u +%Y-%m-%dT%H:%M:%S.000000000Z)
docker logs --since $since --until $until courses-frontend-1 > $T/log 2>&1
echo "$L log interval [$since,$until]: lines=$(wc -l <$T/log | tr -d ' ')"
grep -E '"route":"/(m1-throw|courses)"' $T/log | sed 's/"release":"[^"]*",\{0,1\}//' 
echo "$L sentinel in log: $(grep -c 'm1-render-probe' $T/log)"; grep -m3 'm1-render-probe' $T/log | cut -c1-240
