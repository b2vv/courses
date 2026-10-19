ДАЛІ: запустити браузерний `accounts.spec.ts` у Compose (Courses 1), не змінюючи паралельність `Courses 1`.

2026-10-05 — Додано Acc26–Acc29 і Acc31 до `frontend/e2e/accounts.spec.ts`: RFC6238 TOTP обчислюється в тесті через Node stdlib, перевірено enroll (QR/secret/recovery codes один раз), другий крок входу та його axe-аудит, помилковий TOTP, одноразовість recovery code, регенерацію зі старим кодом, user-disable і admin-disable через захищений HTTP маршрут. Тест використовує наявний run-scoped користувач і наявний `afterAll` cleanup.
