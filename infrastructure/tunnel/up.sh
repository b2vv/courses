#!/bin/sh
# `make tunnel-up` — the owner-run live WayForPay proof (Pay25-live), work/runbook-o06c-tunnel.md.
# Starts a public tunnel to nginx (localhost.run by default, TUNNEL_PROVIDER=cloudflared optional), then recreates the api with WayForPay configured for
# that tunnel's random public URL. Only two values come from .env (the merchant account and key);
# the URL and domain are exported here, and shell variables take precedence over .env in Compose,
# so .env is never edited by this script.
set -eu
cd "$(dirname "$0")/../.."

for name in WAYFORPAY_MERCHANT_ACCOUNT WAYFORPAY_SECRET_KEY; do
  if ! grep -Eq "^${name}=.+" .env; then
    echo "Add ${name}=<value> to .env (WayForPay test merchant) and run again." >&2
    exit 1
  fi
done

# localhost.run by default: some networks block Cloudflare's quick-tunnel API outright.
provider=${TUNNEL_PROVIDER:-localhostrun}
case $provider in
  localhostrun) file=docker-compose.localhostrun.yml; pattern='https://[a-z0-9-]+\.lhr\.life' ;;
  cloudflared) file=docker-compose.cloudflared.yml; pattern='https://[a-z0-9-]+\.trycloudflare\.com' ;;
  *) echo "TUNNEL_PROVIDER must be localhostrun or cloudflared, not $provider" >&2; exit 1 ;;
esac
with_tunnel="docker compose -f docker-compose.yml -f $file"

make up
$with_tunnel up -d "$provider"

url=""
for _ in $(seq 1 30); do
  # api.trycloudflare.com is cloudflared's own API, never the tunnel.
  url=$($with_tunnel logs "$provider" 2>&1 \
    | grep -oE "$pattern" | grep -v '^https://api\.' | head -1 || true)
  [ -n "$url" ] && break
  sleep 2
done
if [ -z "$url" ]; then
  echo "No $provider tunnel URL after 60 s. Inspect: $with_tunnel logs $provider" >&2
  exit 1
fi

PUBLIC_ORIGIN=$url
PUBLIC_BASE_URL=$url
# Your own WayForPay shop is registered for a domain; if .env names it, keep it. Otherwise the
# tunnel host (the shared test merchant accepts any).
WAYFORPAY_MERCHANT_DOMAIN=$(sed -n 's/^WAYFORPAY_MERCHANT_DOMAIN=//p' .env | tail -1)
[ -n "$WAYFORPAY_MERCHANT_DOMAIN" ] || WAYFORPAY_MERCHANT_DOMAIN=${url#https://}
export PUBLIC_ORIGIN PUBLIC_BASE_URL WAYFORPAY_MERCHANT_DOMAIN

$with_tunnel -f docker-compose.tunnel.yml up -d --build --wait --wait-timeout 180 api
make seed-identities seed

# 401 = the callback route exists and refused an unsigned body: WayForPay is configured.
# 404 = the values did not reach the api. A new trycloudflare host can take a few seconds to resolve.
code=000
for _ in $(seq 1 15); do
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$url/api/v1/payments/wayforpay/callback" \
    -H 'Content-Type: application/json' -d '{}' || true)
  [ "$code" = 401 ] || [ "$code" = 404 ] && break
  sleep 2
done

echo
echo "Tunnel:   $url"
echo "Callback: HTTP $code (expected 401)"
if [ "$code" != 401 ]; then
  echo "WayForPay is not reachable through the tunnel as configured — see the runbook." >&2
  exit 1
fi
echo "Open $url in the browser (not localhost) and follow the runbook from step 4."
echo "Logs:     docker compose logs -f api | grep --line-buffered -E 'wayforpay/callback|/learn/orders'"
echo "Stop:     make tunnel-down"
