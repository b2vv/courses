# Runbook — O06c-live: real WayForPay callback through a tunnel (Pay25 production proof)

Who: the owner (holds the test-merchant secret key). This is the one step a stub cannot prove —
that WayForPay's **real** `serviceUrl` callback reaches the stack, verifies and applies once.
The secret key stays in your local `.env` only; it never goes into the repo or a commit.

## Prerequisites
- WayForPay test merchant credentials: account, secret key, domain. (The shared public test
  merchant `test_merch_n1` with its wiki key works, or your own account in test mode.)
- Nothing to install: the tunnel runs from Cloudflare's official image `cloudflare/cloudflared`
  (pulled on first run). (`brew install cloudflared` also works if you prefer the host binary.)
- Docker as usual: `export DOCKER_CONFIG=/tmp/claude-501/dockercfg` in every terminal below.

## Steps
1. **Raise the base stack, then start the tunnel in Docker** on the same network, pointing at
   nginx. The tunnel URL is random, and the api must know it at start — so tunnel first:
   ```
   make up
   docker run -d --name courses-tunnel --network courses_default \
     cloudflare/cloudflared:latest tunnel --no-autoupdate --url http://nginx:8080
   sleep 10; docker logs courses-tunnel 2>&1 | grep -o 'https://[a-z0-9-]*\.trycloudflare\.com' | head -1
   ```
   The last command prints `https://<random>.trycloudflare.com` — call it `$PUB`.

2. **Add to your local `.env`** (gitignored; never commit it). Use `$PUB` literally, without a
   trailing slash:
   ```
   PUBLIC_ORIGIN=https://<random>.trycloudflare.com
   PUBLIC_BASE_URL=https://<random>.trycloudflare.com
   COOKIE_SECURE=true
   WAYFORPAY_MERCHANT_ACCOUNT=<test account>
   WAYFORPAY_SECRET_KEY=<test secret>
   WAYFORPAY_MERCHANT_DOMAIN=<random>.trycloudflare.com
   ```
   Why each matters: `PUBLIC_ORIGIN` — the same-origin check rejects purchases from any other
   origin (403 `origin_forbidden`); `PUBLIC_BASE_URL` — builds WayForPay's `returnUrl` and
   `serviceUrl`; without the four WayForPay values checkout answers 503. Leave
   `WAYFORPAY_PAY_URL` / `WAYFORPAY_STATUS_URL` unset → the real WayForPay hosts.

3. **Bring the stack up with the tunnel override** — the base compose does not pass these
   variables to the api; `docker-compose.tunnel.yml` does, and fails loudly if one is missing:
   ```
   docker compose -f docker-compose.yml -f docker-compose.tunnel.yml up -d --build
   make seed-identities seed
   ```
   This recreates the api with the new environment; the tunnel container keeps running.

   Quick check that the provider is configured:
   ```
   curl -s -o /dev/null -w '%{http_code}\n' -X POST "$PUB/api/v1/payments/wayforpay/callback" -d '{}'
   ```
   Expected `401` (route present, bad signature). `404` means the variables did not reach the api.

4. **Make a paid course and buy it** — open **`$PUB`** in the browser (not localhost). As admin
   (`admin@courseforge.local`): create/publish a course with a lesson, save, then set a price in the
   «Ціна курсу» panel. Sign out, sign in as `student@courseforge.local`, open the locked lesson →
   «Купити курс» → «Перейти до оплати». You land on `secure.wayforpay.com`; pay with a WayForPay
   **test card** from their test-data wiki page.

5. **Watch the real callback arrive** (second terminal):
   ```
   docker compose logs -f api | grep --line-buffered -E 'wayforpay/callback|/learn/orders'
   ```
   Expect `"route":"POST /api/v1/payments/wayforpay/callback","status":200`. These lines are the
   proof. (The callback body is never logged by design — no payload or key in logs.)

6. **Confirm the outcome:** the order page shows «Оплату зараховано», the lesson opens.

7. **Confirm "applied once, replay a no-op" with SQL** (WayForPay may retry the callback; any
   repeat must change nothing):
   ```
   docker compose exec -T postgres psql -U courses -d courses -c "
     SELECT g.source, g.state, count(*) AS grants
       FROM course_grants g JOIN users u ON u.id = g.user_id
      WHERE u.email = 'student@courseforge.local' AND g.source = 'purchase'
      GROUP BY 1, 2;
     SELECT e.event_id, e.type, e.amount_minor, o.status
       FROM payment_events e JOIN orders o ON o.id = e.order_id
      WHERE e.provider = 'wayforpay'
      ORDER BY e.received_at DESC LIMIT 5;"
   ```
   Expected: exactly one active `purchase` grant for the course; one `paid` event per WayForPay
   event id, and the order `paid`. If the log in step 5 shows the callback more than once, the
   counts must still be one.

## What to send back
- The step-5 log lines (they contain no secrets).
- The step-7 query output.
- Optional: a screenshot of the opened lesson.
Courses 1 records them in `docs/payment/report.md` as the Pay25-live proof.

## Teardown
- `docker rm -f courses-tunnel` (throwaway hostname; the public URL dies with it).
- Remove the six lines from `.env` (or set them back), then
  `docker compose up -d --build` without the override — otherwise smoke's callback-404 guard
  goes red on a configured stack.
