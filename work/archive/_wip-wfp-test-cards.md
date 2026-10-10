# WayForPay test cards research — test_merch_n1

NEXT: search wiki.wayforpay.com for pages mentioning test cards/sandbox, and WebSearch for official sources

FACT: wiki.wayforpay.com/view/852472 (Тестові реквізити, official, uk) lists only merchant
  accounts: test_merch_n1 (merchantSecretKey flk3409refn54t54t*FNJRET), test_merch_p2p,
  p2p_credit. No card numbers, no mode-behavior text. Confirms prior finding.

FACT: wiki.wayforpay.com/view/852100 (Платежі overview, official) only links to the Тестові
  реквізити page; no card numbers inline.

FACT: wiki.wayforpay.com/ru/wiki/default/test (official, "Генерування тестової строки") is a
  signature-generator tool page (merchantSignature builder) — no card numbers, no test-mode
  behavior text.

FACT: github.com/wayforpay official org repos (via WebFetch of org page): Word-Press-Woocommerce,
  php-sdk, PHP, Opencart-3.0, Simpla, Magento-2.0 — all PHP, payment integrations.
  raw.githubusercontent.com/wayforpay/php-sdk/master/README.md (official): no test card numbers,
  no test-mode/sandbox section at all.

FACT (UNOFFICIAL, help.weblium.com/en-us/article/connecting-wayforpay-as-a-payment-method-68h3dc/,
  third-party platform help article, not WayForPay-owned):
  "you can use the following test cards: 4111 1111 1111 1111 and 5454 5454 5454 5454"
  immediately followed by: "Such payments will be processed as unsuccessful. For successful
  payments you need to use a real card using real funds."
  Also: "If the store is in test mode, all payments are refunded automatically within 15
  minutes." And: "If you want to test successful payments without the risk of losing funds,
  you can create a product for 1 USD and pay for it."
  This matches our live observation (4111... and 5454... both Declined) and explicitly states
  there is NO test card that yields Approved — a real card with real funds is required, with
  the store's test-mode safety net being an automatic refund within 15 minutes, not a fake
  approval.

FACT: help.wayforpay.com (official help centre) pages checked — /view/1737798 (registration):
  only "after shop creation, account moves to test mode, you can integrate and test payment
  instruments" — no card numbers or outcome rules. /view/647980774 (deleting a test shop): no
  test-card or outcome content.

DECISION: Treat "no official Approved test card exists" as the answer. No WayForPay-owned
  page (wiki.wayforpay.com, help.wayforpay.com, github.com/wayforpay) publishes a card number
  that yields Approved on test_merch_n1, nor any blanket "any card is auto-approved in test
  mode" rule. The one concrete behavioral claim (any card test-mode payment is processed as
  unsuccessful; real card + real funds needed for Approved; auto-refund within 15 min) comes
  only from a third-party (Weblium) article, not from WayForPay itself, though it is consistent
  with the live Declined results already observed on 4111..., 5454..., 4222222222222.

NEXT: none — research exhausted per scope (wiki all langs effectively checked via search +
  direct fetch, github org repos listed, help centre checked, web search queries run). Done.
