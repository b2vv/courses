# Payment provider research — Ukraine-first, EU later

| | |
|---|---|
| Purpose | Inform the deferred provider decision that blocks the `orders` adapter (slice O06). |
| Scope | One-time course purchase; grant issued **only** by provider webhook; UAH now, EUR later; no refunds in MVP; card data must not touch our servers; signed webhooks with event-id idempotency. |
| Author | Courses 1 (research delegated) |
| Date | 2026-09-30 |
| Status | **Research only — this informs a human decision. It does not decide.** |

> Every claim carries a source URL. Anything not confirmed against a primary source is marked **⚠️ unverified**. Fees and onboarding terms change and are often individually negotiated — treat all percentages as *rough, verify in the merchant cabinet before committing code*.

---

## TL;DR

- **Ukraine-first MVP:** the realistic candidates are **LiqPay**, **WayForPay**, and **Portmone** — all Ukrainian PSPs that (a) onboard a UA business/ФОП, (b) settle in UAH, (c) use a hosted redirect page so card data never reaches us, and (d) POST a signed server-to-server callback on success. They are **PSPs, not Merchants of Record**: *we* remain responsible for tax. This fits the MVP because our webhook-issues-grant model maps cleanly onto their callback + signature scheme.
- **Fondy** technically fits the same slot but carries a **material Ukraine-serving risk**: in 2024 the NBU cancelled the registration of its Ukrainian operating entity over a stated national-security threat, and the service changed ownership (TBC, Georgia). Flagged below — verify current UA merchant status before choosing it.
- **Stripe** does **not** onboard Ukraine-based sellers directly. Only via a foreign entity (e.g. Stripe Atlas US company). Not a Ukraine-first option.
- **EU expansion:** two different shapes. Keep a UA PSP and *add* EU VAT handling yourself, **or** adopt a **Merchant-of-Record** (Paddle) that collects and remits EU VAT for you. Paddle onboards Ukrainian sellers (except occupied regions) but **⚠️ does not appear to price in UAH**, so it does not serve the UAH-first MVP — it is an *EU-phase* adapter, not the first one.

The `orders` port is provider-agnostic by design (AGENTS.md: no port without a consumer). A second jurisdiction is meant to be a *new adapter*, not a domain rewrite — so picking a UA PSP now does not foreclose Paddle later.

---

## Comparison table

| | LiqPay (PrivatBank) | WayForPay | Fondy | Portmone | Paddle |
|---|---|---|---|---|---|
| **Type** | PSP (acquirer: PrivatBank) | PSP | PSP / EMI (UK Fondy Ltd, FCA) | PSP | **Merchant of Record** |
| **UA seller / ФОП onboarding** | Yes — PrivatBank merchant; sandbox keys issued on signup [1][12] | Yes — account starts in test mode on registration; complete KYC to go live [4][13] | Yes historically, but **⚠️ see UA-serving risk** [14][15] | Yes — UA merchant, unique payee ID + private key issued [7] | Yes for Ukraine **except occupied regions** (Crimea, Donetsk, Kherson, Luhansk, Zaporizhzhia); account needs approval [16][17] |
| **Role for tax (MoR vs PSP)** | PSP — seller handles tax | PSP — seller handles tax | PSP — seller handles tax | PSP — seller handles tax | **MoR — Paddle calculates, collects & remits VAT** [18][19] |
| **UAH (now)** | Yes [11] | Yes (default) [3] | Yes [8] | Yes (default currency) [7] | **⚠️ UAH not seen in pricing currencies** (EUR/USD listed) [16] |
| **EUR (later)** | Yes (referenced) ⚠️ confirm in cabinet [6] | Yes [3] | Yes [8] | ⚠️ confirm | Yes [16] |
| **EU expansion on same account** | Ukrainian acquiring; EU selling not its purpose | UA-focused; EU selling limited | UK/EU entity + PL/CZ/SK/RO/HU coverage [8] | UA-focused | **Yes — sells in 200+ countries, no per-country setup** [17] |
| **Checkout model** | Hosted (LiqPay domain, 3DS redirect) — card data off our servers [6] | Hosted redirect to `secure.wayforpay.com/pay` — card data off our servers [2] | Hosted checkout **and** embedded/custom option [9][10] | Gateway + CSE (client-side encryption); hosted option [7] | Hosted overlay/checkout — card data off our servers [17] |
| **Server-to-server webhook on success** | Yes — POST `{data, signature}` to `server_url` after final status [6] | Yes — POST to `serviceUrl`; **retries up to 4 days** until a valid response [2] | Yes — callback with `order_status` [9] | Yes — callback to configured URL [7] | Yes — webhooks (`transaction.completed`, etc.) [18] |
| **Signature scheme** | `base64(sha1(private_key + data + private_key))` — **⚠️ docs overview says sha3-256 but all code examples use SHA1** [6] | `HMAC_MD5` over `;`-joined fields; callback sig over `orderReference;status;time` [2] | SHA1 order signature (40 chars) [9] | `HMAC-SHA256` over payeeId+timestamp+order#+amount+private key [7] | `Paddle-Signature` header, **HMAC-SHA256** over `timestamp:body` [18] |
| **Idempotency / event id** | `payment_id` (LiqPay txn id) + `order_id` in payload — dedupe on these [6] | `orderReference` unique in merchant system; **⚠️ no explicit dedup guarantee — enforce yourself** [2] | `order_id` in payload — dedupe yourself | order#/txn id in payload — dedupe yourself | Event has unique id + type — dedupe on event id [18] |
| **Fees (rough — verify)** | ⚠️ sources vary: ~0.6–1.5% domestic to ~2.75% intl [20] | ~2–2.5% [20] | ~2.2% (≤500k UAH/mo) / 2% (>500k), +0.2% intl [3] | ~2.3–2.5% [7] | **5% + $0.50** per txn (MoR premium; ⚠️ verify current) [21] |
| **Refunds (future seam)** | API supports refunds ⚠️ confirm | API supports reverse/refund [2] | API supports refunds ⚠️ confirm | API supports refunds ⚠️ confirm | Supported via dashboard/API [18] |
| **Sandbox for e2e** | Yes — `sandbox_`-prefixed keys after signup [12] | Yes — account starts in test mode [13] | ⚠️ sandbox exists, terms unverified | ⚠️ test env exists, verify | Yes — sandbox environment [18] |
| **API/docs maturity** | Mature, multi-language SDKs; docs bilingual [1][6] | Mature wiki, many recipe pages [2] | Mature docs (docs.fondy.eu / .io) [9][10] | Documented, PCI-DSS L1 [7] | Very mature developer docs [16][18] |

---

## Per-provider notes

### LiqPay (PrivatBank)
- **Onboarding:** tied to PrivatBank acquiring; a UA business/ФОП can register and immediately receives `sandbox_`-prefixed test keys, then production keys after activation [12]. This is the lowest-friction path for a Ukrainian seller who already banks with PrivatBank.
- **Webhook fit:** matches our model well — LiqPay POSTs `data` (base64 JSON) + `signature` to `server_url`; decoded payload has `order_id`, `payment_id`, `status`, `currency` [6]. Idempotency: dedupe on `payment_id`/`order_id`.
- **⚠️ Signature discrepancy:** the callback doc's prose mentions `sha3-256` while every code example uses `sha1` — `base64(sha1(private_key + data + private_key))` [6]. Pin the exact algorithm against a live sandbox callback before trusting either; do not hardcode from the prose alone.
- **Fees:** reported figures diverge widely across secondary sources (0.6%–2.75%) [20] — **⚠️ unverified**, confirm in the merchant agreement.

### WayForPay
- **Onboarding:** registration puts the account in test mode automatically; you finish KYC to reach the live cabinet [13]. Serves UA business/ФОП.
- **Checkout:** POST to `https://secure.wayforpay.com/pay` redirects the buyer to WayForPay's hosted page — card data stays off our servers [2].
- **Webhook fit:** POSTs to `serviceUrl`; notably **retries for up to 4 days** until it gets a valid response [2] — our inbox must be idempotent by design, which aligns with Pay04. Callback signature is `HMAC_MD5` over `orderReference;status;time` [2].
- **⚠️ Idempotency:** `orderReference` is described as merchant-unique but the docs do not promise dedup — we must enforce event-id uniqueness ourselves (which the spec already requires) [2].
- **Currencies:** UAH default, EUR and USD supported [3].

### Fondy
- **Technical fit is good:** hosted *and* embedded checkout [9][10], SHA1 order signature, `order_status` callback, turnover-tiered fees ~2.2%/2% [3], EU coverage (PL, CZ, SK, RO, HU) [8].
- **⚠️ Material Ukraine-serving risk (flag for the human):** in 2024 the NBU **cancelled the registration** of Fondy's Ukrainian operating entity ("FC Alliance" / "FC Elaens", TM Fondy) in the payment systems, citing a **threat to national security** after law-enforcement information [14][15]. Around the same time the Antimonopoly Committee allowed Georgia's TBC Bank Group to acquire the service [15]. The decision is disputed by Fondy and criticized by some experts as non-transparent [15], but the practical effect is that **Fondy's ability to onboard/serve *new Ukrainian* merchants is uncertain**. The UK entity (Fondy Ltd, FCA-regulated EMI) is a separate, EU/UK-facing vehicle [14]. **Do not choose Fondy for the UAH-first MVP without confirming current UA merchant onboarding status directly with Fondy.**

### Portmone
- **Onboarding:** UA merchant; each merchant gets a unique payee ID + private key + credentials [7]. PCI-DSS Level 1 gateway with client-side encryption (CSE) [7].
- **Webhook/signature:** `HMAC-SHA256` over payee ID, timestamp, shop order number and amount with a per-partner private key [7] — the cleanest signature scheme of the UA PSPs (SHA-256 vs LiqPay's SHA1 / WayForPay's MD5).
- **Fees:** ~2.3–2.5% [7] **⚠️ verify**.
- **Note:** CSE means a "transparent" integration is possible where card fields render on our page — we would deliberately choose the **hosted** flow to keep card data off our servers per the spec.

### Paddle (Merchant of Record) — the EU-phase candidate
- **What MoR buys us:** Paddle becomes the seller of record and **calculates, collects and remits VAT/sales tax** across 200+ jurisdictions [17][18][19]. For EU expansion this removes the entire EU-VAT-registration/OSS burden that a PSP would leave on us.
- **UA seller eligibility:** Ukraine-registered sellers can sign up (account approval takes a few days), **except sellers operating from occupied regions** (Crimea, Donetsk, Kherson, Luhansk, Zaporizhzhia) [16][17].
- **⚠️ Blocker for the MVP:** the pricing/currency documentation lists EUR/USD and others but **UAH was not found** [16]. If Paddle cannot charge Ukrainian buyers in UAH, it does not serve the Ukraine-first MVP. **Verify UAH support directly before considering Paddle as the *first* adapter.**
- **Webhook:** `Paddle-Signature` header, HMAC-SHA256 over `timestamp:body`; events carry a unique id — idempotency-friendly [18].
- **Fee premium:** ~**5% + $0.50** per transaction (MoR handles tax + compliance in exchange) [21] **⚠️ verify current rate**.

### Stripe — status for Ukrainian sellers
- **Not available** for sellers located in Ukraine; Stripe's own guidance and the NBU both confirm Stripe has not entered the Ukrainian market [22][23]. The only path is a **foreign entity** (e.g. a US company via Stripe Atlas) [24]. This means legal/banking overhead abroad, not a Ukraine-first option. Listed here only to close the question.

---

## Recommendation (for a human to decide)

**MVP (Ukraine-first, UAH):** a Ukrainian **PSP** is the fit — the webhook-issues-grant model maps directly onto their signed server-to-server callback, all settle in UAH, and all keep card data off our servers via a hosted page. Ranked by fit to *our* spec:

1. **LiqPay** — lowest onboarding friction (especially if the seller banks with PrivatBank), clear sandbox with `sandbox_` keys, well-documented callback. *Caveat:* resolve the SHA1-vs-SHA3 signature ambiguity against a live sandbox callback before writing the parser [6][12].
2. **WayForPay** — strong docs, hosted redirect, UAH/EUR/USD; its 4-day callback retry actually reinforces our idempotency requirement. Enforce event-id dedup ourselves [2][13].
3. **Portmone** — cleanest signature scheme (HMAC-SHA256), but choose the hosted flow deliberately to keep card data off our servers [7].

**Avoid for the MVP unless status is re-confirmed:** **Fondy** (UA entity's NBU registration was cancelled in 2024 over a national-security threat; ownership changed) [14][15]. **Stripe** (no direct UA onboarding) [22][23].

**EU-expansion path:** two options, decided later as a *new adapter* behind the same `orders` port —
- **Stay PSP + do VAT yourself:** cheapest fees, but we take on EU VAT registration/OSS filing and compliance.
- **Add Paddle (MoR):** higher fee (~5% + $0.50), but Paddle handles EU VAT end-to-end [17][18][19]. Ukraine sellers can onboard [16]. **⚠️ Confirm whether Paddle can charge in UAH** — if not, Paddle only makes sense once selling to EU/global buyers in EUR/USD, i.e. strictly the EU phase, never the UAH MVP.

**Before writing O06, verify against a live sandbox (not the prose docs):** (a) exact signature algorithm and field order, (b) the decoded success payload's field names for order id / payment id / currency, (c) that a duplicate callback is observable so our event-id inbox can be tested, (d) current fee and settlement/payout terms in the merchant agreement.

---

## Sources

1. LiqPay — payments overview: https://privatbank.ua/en/liqpay-oplaty
2. WayForPay — Accept payment (Purchase), serviceUrl & HMAC_MD5: https://wiki.wayforpay.com/en/view/852102
3. Fondy fees & turnover tiers (secondary summary): https://iwis.io/en/blog/integration-of-payment-systems/
4. WayForPay — Add Merchant: https://wiki.wayforpay.com/en/view/853322
6. LiqPay — Callback / signature / server_url: https://www.liqpay.ua/doc/api/callback
7. Portmone — Payment Gateway API (HMAC-SHA256, CSE, fees, currency): https://docs.portmone.com.ua/docs/en/PaymentGatewayEng/
8. Fondy — Supported countries: https://docs.fondy.eu/en/docs/page/supported-countries/
9. Fondy — Accept purchase / callback: https://docs.fondy.eu/en/docs/page/3/
10. Fondy — Embedded/custom checkout: https://docs.fondy.eu/en/docs/page/custom-checkout-v2-0/
11. LiqPay — documentation home: https://www.liqpay.ua/documentation/en
12. LiqPay — test data / sandbox keys: https://www.liqpay.ua/en/doc/api/POS/testingData
13. WayForPay — Merchant Edit / test mode: https://wiki.wayforpay.com/en/view/853327
14. Fondy Ltd (UK) company profile, FCA EMI: https://find-and-update.company-information.service.gov.uk/company/10063984
15. NBU cancels UA Fondy entity registration (national security) + TBC acquisition: https://newsukraine.rbc.ua/news/experts-assess-the-situation-surrounding-1717659893.html ; https://english.nv.ua/nation/nbu-suspends-fc-elaens-activities-due-to-security-reasons-50414362.html
16. Paddle — Supported countries & locales (currencies): https://developer.paddle.com/concepts/sell/supported-countries-locales/
17. Paddle — Which countries are supported / sanctions (Ukraine regions): https://www.paddle.com/help/start/intro-to-paddle/which-countries-are-supported-by-paddle ; https://www.paddle.com/help/legal/sanctions
18. Paddle — how Paddle works / webhooks / MoR: https://developer.paddle.com/get-started/how-paddle-works/digital-products/
19. Paddle — intro / merchant of record: https://www.paddle.com/help/start/intro-to-paddle
20. LiqPay & WayForPay fee summaries (secondary; ranges only): https://olimpdigital.com/blog/ecommerce/top-ukraine-payment-providers ; https://cityhost.ua/en/blog/obzor-platezhnyh-sistem-interkassa-liqpay-easypay-portmone-fondy-wayforpay.html
21. Paddle pricing / fee (secondary review, verify): https://dev.to/onsen/paddle-review-2026-pros-cons-pricing-explained-4cgk
22. Stripe — global availability (Ukraine not supported): https://stripe.com/global
23. NBU — Stripe/PayPal never requested a license: https://dev.ua/en/news/nbu-zaiavyv-shcho-stripe-i-paypal-zhodnoho-razu-ne-prosyly-litsenziiu-dlia-roboty-v-ukraini-1764321864
24. Stripe Atlas workaround for Ukrainian founders: https://www.linkedin.com/pulse/how-open-stripe-account-ukraine-2024-mazino-oyolo-fobtf
