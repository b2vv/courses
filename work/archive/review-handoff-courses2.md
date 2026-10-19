# Хендоф на Courses 2 — код-рев'ю (+ pattern-architecture-review)

Від Courses 1. Рішення власника: код фаз має пройти **незалежне код-рев'ю Courses 2**, якого я пропустив
(я робив лише GATE 2 на планах + самоверифікацію гейтами). Застосуй до кожного обсягу **два скіли**:
`code-review` (проти spec/standards/дифу) **і** `pattern-architecture-review` (GoF/KISS/DRY/SOLID/Clean Arch).

Рев'ю лише читає; нічого не змінюй і не комість. Знахідки — APPROVE / CHANGES, blocking/non-blocking, з
доказом `file:line`. Пиши у вказані файли (append-only).

## Обсяг 1 — фаза `accounts` (ЗМЕРДЖЕНО, ретроактивно)
- Уже в `main`, PR #7, merge-коміт `d263045`. Код-коміти фази: `1e4cd49` (A01) … `05727ec` (перед merge).
  Диф: `git diff 14aeb3b...d263045 -- backend frontend infrastructure` (від закриття ratings до accounts);
  або `git log --oneline 14aeb3b..d263045`.
- Фокус (чутливе до безпеки): гейт верифікації email через `contracts/access.CanStream` + шов
  `httpserver.Actor.EmailVerified`; 2FA — TOTP (RFC6238, constant-time), pending-токен у Redis (одноразовість,
  прив'язка user+IP), резервні коди (атомарний consume, хеш); CSP (nonce, enforcing прод/report-only e2e);
  login rate-limit; Mailer (CRLF-guard, без секретів у логах); токени як digest.
- Артефакти: `docs/accounts/{spec,plan,report}.md` (обмеження §5). Пиши у `work/review-courses2-accounts.md`.

## Обсяг 2 — фаза `enrollment` (У РОБОТІ, гілка `feat/enrollment`)
- GATE 1/2 пройдено (`docs/enrollment/{spec,plan}.md`; робочі файли GATE 2 — `work/_wip-enrollment-gate2-*.md`).
  Код реалізується позрізно L1–L9 (`work/tasks.md` «Фаза enrollment»); L1 (домен membership + міграція 00015 +
  backfill + seed) щойно згенеровано Codex (диф у робочому дереві гілки, ще не закомічено).
- Фокус: модель «grants→membership, один рядок на (user,course), стан=влада доступу» (G1); один рядок vs
  обхід мюту/revoke (повна UNIQUE(user,course), UPSERT-переходи); media-bytes auth-шлях
  (`playback/adapters/postgres/access.go`); `CanConsume`/`CanCommunicate` (B3); backfill-union (нікого не
  замкнути); person-reports у discussion читає членство власним SQL (ADR-013, без імпорту orders).
- Пиши у `work/review-courses2-enrollment.md`. Рев'ю кожного зрізу — по мірі готовності; Courses 1 чекає
  перед PR.

## Що повернути Courses 1
Коли закінчиш обсяг — одне повідомлення: вердикт (APPROVE / CHANGES-requested), число blocking/non-blocking,
і шлях до файлу зі знахідками. Courses 1 виправляє по одному з тестом (receiving-code-review).
