# TE6 — e2e teach editor + full make check (WIP, Codex working file)

РОБОЧИЙ ФАЙЛ. Codex дописує ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite).
Перший запис — рядок `ДАЛІ:` до першого інструменту. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.

## Контекст
- main = `5b6b47b` (TE4+TE5 змерджено). Гілка `feat/teach-editor`.
- Це run — написати e2e spec `frontend/e2e/teach-editor.spec.ts`. e2e НЕ запускати (Docker-стек піднімає координатор).

## Журнал
ДАЛІ: перевірити гілку/WIP, lesson-builder/instructor/auth аналоги, schema та teach contracts.
ФАКТ: feat/teach-editor; frontend clean перед TE6. Прочитано lesson-builder.spec.ts, instructor.spec.ts, auth.ts. Lesson-builder не чистить fixtures; контексти закриває. Instructor fixture один; token у lesson-builder має unsafe cast, тому новий helper перевірятиме unknown runtime-полями.
ДАЛІ: звірити user schema/seed і document request/response, selector та acceptance.
ФАКТ: HEAD 5b6b47ba05ca18de9c6277a7ef5d73d823c23363; seed-identities має лише одного instructor. 00002_identity.sql вимагає display_name та argon2id hash 60–512 chars. Новий owner може отримати копію fixture hash SQL INSERT SELECT, без виведення hash/логіну. Publish handler вимагає валідний expectedUpdatedAt і повертає 200.
РІШЕННЯ: TE6 implementation READY у scope доручення: один spec, два scenarios. T9 чекає teach publish response + toast, перевіряє teach GET content і member published render; browser requests контролюються на відсутність /admin. T5 використовує admin-saved draft token для валідних denied mutations, усі 5 exact status/code; додатково student tree refusal. Fixtures namespace + UUID; lifecycle як lesson-builder (без data cleanup), contexts finally.close. Docker helper лише записується, не запускається тут.
ДАЛІ: написати spec, перевірити format/typecheck/lint; execution/AC6 лишаються coordinator.
ФАКТ: створено один teach-editor.spec.ts із двома scenarios, namespace, sql/origin/PNG/openAs helpers, runtime-checked token, 5 non-owner refusals + student tree refusal. Остаточна звірка wire schema показує paragraph.spans (не paragraph.content); assertion має використовувати саме spans. Multipart handler читає поле file.
ДАЛІ: завершити точне wire assertion та виконати лише format/typecheck/lint.
ФАКТ: frontend/ npm run format → PASS, exit 0 (379 files); npm run typecheck → PASS, exit 0; npm run lint → PASS, exit 0 (architecture 231 files).
ДАЛІ: фінально звірити diff/scope та зберегти checkpoint; e2e execution NOT_RUN, не переносити static PASS на T5/T9/AC6.
ФАКТ: фінальний scope — лише новий frontend/e2e/teach-editor.spec.ts та append цього журналу; existing specs/helpers/product files не змінено. git diff --check без повідомлень.
ФАКТ: SHA-256 frontend/e2e/teach-editor.spec.ts: 4c341b0a1ec727203cd2ec690da8a93fd4bd9e0b1b18c0879e705eec01b67cf6; base HEAD 5b6b47ba05ca18de9c6277a7ef5d73d823c23363.
ФАКТ: npm test/Playwright/e2e/Docker/make НЕ запускалися; T5/T9 execution та AC6 NOT_RUN, незалежне review pending. Commit/stage/push не виконувалися.
ДАЛІ: reviewer перевіряє spec, coordinator виконує повний ordered make check та фіксує actual acceptance results.
ФАКТ: native checkpoint → OK, exit 0, b997b10e98526e349cc48feaa37ac714-4-7a37597fa5b1. Spec готовий до review й coordinator gate.
