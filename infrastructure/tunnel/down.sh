#!/bin/sh
# `make tunnel-down` — removes the tunnel and recreates the api without WayForPay, so the normal
# stack is back to unconfigured (smoke's callback-404 guard expects that). .env is untouched: the
# base compose never forwards the WayForPay values, so leaving them there is harmless.
set -eu
cd "$(dirname "$0")/../.."
docker compose -f docker-compose.yml -f docker-compose.cloudflared.yml rm -sf cloudflared
make up
