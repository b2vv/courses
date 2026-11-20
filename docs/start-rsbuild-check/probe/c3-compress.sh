#!/bin/zsh
# M1 C3: compression scope through nginx :8080 (raw bytes via curl, no auto-decoding). Diagnostic only.
set -u
O=http://localhost:8080; T=$(mktemp -d); fail=0
hdr() { grep -i "^$2:" "$1" | head -1 | cut -d' ' -f2- | tr -d '\r'; }
curl -sS -D $T/gz.h -o $T/gz.b -H 'Accept-Encoding: gzip' "$O/api/v1/courses?limit=20"
curl -sS -D $T/id.h -o $T/id.b -H 'Accept-Encoding: identity' "$O/api/v1/courses?limit=20"
st=$(head -1 $T/gz.h | cut -d' ' -f2); ce=$(hdr $T/gz.h content-encoding); vary=$(hdr $T/gz.h vary)
echo "api gzip: status=$st ce=$ce vary=$vary bytes=$(wc -c <$T/gz.b)"
[[ $st == 200 && $ce == gzip && ${vary:l} == *accept-encoding* ]] || { echo FAIL api-gzip; fail=1; }
st=$(head -1 $T/id.h | cut -d' ' -f2); ce=$(hdr $T/id.h content-encoding)
echo "api identity: status=$st ce=${ce:-none} bytes=$(wc -c <$T/id.b)"
[[ $st == 200 && -z $ce ]] || { echo FAIL api-identity; fail=1; }
python3 -I -c "import gzip,json,sys;a=json.loads(gzip.decompress(open(sys.argv[1],'rb').read()));b=json.load(open(sys.argv[2]));print('api decode-equal', a==b);sys.exit(0 if a==b else 1)" $T/gz.b $T/id.b || { echo FAIL api-equal; fail=1; }
for route in /courses /courses/go-backend-foundations; do
  curl -sS -D $T/p.h -o $T/p.b -H 'Accept-Encoding: gzip' -H 'Accept: text/html' "$O$route"
  st=$(head -1 $T/p.h | cut -d' ' -f2); ce=$(hdr $T/p.h content-encoding); ct=$(hdr $T/p.h content-type)
  complete=$(tail -c 200 $T/p.b | grep -c '</html>')
  echo "page $route: status=$st ct=$ct ce=${ce:-none} bytes=$(wc -c <$T/p.b) complete=$complete"
  [[ $st == 200 && -z $ce && $ct == text/html* && $complete -ge 1 ]] || { echo FAIL page $route; fail=1; }
  if [[ $route == /courses ]]; then want=(h1=Каталог\ курсів 'h2=Go: фундамент надійного бекенду'); else want=('h1=Go: фундамент надійного бекенду'); fi
  python3 -I $(dirname $0)/c3-headings.py $T/p.b "${want[@]}" || { echo FAIL headings $route; fail=1; }
done
# Baseline observation (plan C3): srvx static-asset behaviour for one client script named in the catalog HTML.
curl -sS -o $T/c.b -H 'Accept: text/html' "$O/courses"
asset=$(grep -oE 'src="/[^"]+\.js"' $T/c.b | head -1 | cut -d'"' -f2)
curl -sS -D $T/a.h -o $T/a.b -H 'Accept-Encoding: gzip' "$O$asset"
echo "static $asset: status=$(head -1 $T/a.h | cut -d' ' -f2) ct=$(hdr $T/a.h content-type) ce=$(hdr $T/a.h content-encoding) cache=$(hdr $T/a.h cache-control) bytes=$(wc -c <$T/a.b | tr -d ' ') (observation only)"
echo "C3 verdict: $([[ $fail == 0 ]] && echo PASS || echo FAIL)"
