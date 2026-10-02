# Runbook — O06c-live: real WayForPay callback through a tunnel (Pay25 production proof)

Who: the owner (holds the test-merchant secret key). This is the one step a stub cannot prove —
that WayForPay's **real** `serviceUrl` callback reaches the stack, verifies and applies once.
Everything runs in Docker; nothing to install on the host beyond what the repo already uses.
The secret key stays in your local `.env` only; it never goes into the repo or a commit.

**Heads-up:** a quick tunnel makes this stack public for as long as it runs, seeded demo accounts
included. Keep it short and run `make tunnel-down` when done.

## Known limits (first live run, 2026-10-02)
- **Tunnel provider.** `make tunnel-up` uses localhost.run by default because this network blocks
  Cloudflare's quick-tunnel API; `TUNNEL_PROVIDER=cloudflared make tunnel-up` selects cloudflared.
  The free localhost.run tunnel **rotates its hostname within minutes** — the api keeps the old URL,
  so a callback sent after a rotation is lost. Pay quickly after `tunnel-up`, or use a named tunnel.
- **Test card.** `4111 1111 1111 1111` (from WayForPay's API examples) is **declined** by
  `test_merch_n1` (reason 1105). WayForPay publishes no test-card list; an Approved payment needs a
  card WayForPay documents for test mode, or your own merchant in test mode. Never type a real card
  into the shared test merchant.
- **One checkout per order.** Opening «Перейти до оплати» twice for the same order is refused by
  WayForPay (`1112 Duplicate Order ID`) — start a new purchase for each attempt until B-25 is fixed.
- **Phone is required** on WayForPay's page.

## 1. Credentials — once
Add two lines to your local `.env` (gitignored):
```
WAYFORPAY_MERCHANT_ACCOUNT=<test account>
WAYFORPAY_SECRET_KEY=<test secret>
```
(The shared public test merchant `test_merch_n1` with its wiki key works, or your own account in
test mode.) Nothing else goes in `.env` — the tunnel URL and domain are set by the script.

## 2. Start
```
export DOCKER_CONFIG=/tmp/claude-501/dockercfg
make tunnel-up
```
What it does (`infrastructure/tunnel/up.sh`):
1. `make up` — the normal stack;
2. starts `cloudflared` (official `cloudflare/cloudflared` image, `docker-compose.cloudflared.yml`)
   pointing at `nginx:8080`, and reads the random `https://<random>.trycloudflare.com` from its log;
3. recreates the api with `docker-compose.tunnel.yml`, exporting `PUBLIC_ORIGIN`,
   `PUBLIC_BASE_URL` (= tunnel URL) and `WAYFORPAY_MERCHANT_DOMAIN` (= tunnel host);
4. reseeds identities and courses;
5. checks the callback route through the tunnel: **401** = configured (route present, unsigned body
   refused). It fails loudly on 404 or anything else.

It ends by printing the tunnel URL — call it `$PUB`.

## 3. Buy
Open **`$PUB`** in the browser (not localhost — WayForPay must return to the public host).
1. As admin (`admin@courseforge.local`): create a course with a lesson, save, publish, then set a
   price in the «Ціна курсу» panel.
2. Sign out; sign in as `student@courseforge.local`; open the locked lesson → «Купити курс» →
   «Перейти до оплати».
3. On `secure.wayforpay.com` pay with a WayForPay **test card** from their test-data wiki page.

## 4. Watch the real callback (second terminal)
```
export DOCKER_CONFIG=/tmp/claude-501/dockercfg
docker compose logs -f api | grep --line-buffered -E 'wayforpay/callback|/learn/orders'
```
Expect `"route":"POST /api/v1/payments/wayforpay/callback","status":200`. These lines are the
proof. (The callback body is never logged by design — no payload or key in logs.)
The order page shows «Оплату зараховано» and the lesson opens.

## 5. Confirm "applied once, a repeat is a no-op"
WayForPay may retry the callback; any repeat must change nothing:
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
Expected: exactly one active `purchase` grant; one `paid` event per WayForPay event id; the order
`paid`. If step 4 shows the callback more than once, the counts must still be one.

## 6. Stop
```
make tunnel-down
```
Removes the tunnel (the public URL dies with it) and recreates the api without WayForPay. `.env`
can keep the two lines: the base compose never forwards them, so the normal stack and smoke's
callback-404 guard are unaffected.

## What to send back
- The step-4 log lines and the step-5 query output (neither contains secrets).
- Optional: a screenshot of the opened lesson.
Courses 1 records them in `docs/payment/report.md` as the Pay25-live proof.
