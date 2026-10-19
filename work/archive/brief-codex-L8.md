# Brief — L8: e2e enrollment.spec.ts + axe, і оновлення зламаних існуючих specs

РОБОЧИЙ ФАЙЛ: `work/_wip-L8-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. Стандартні правила — `work/codex-working-model.md`.
**Ти НЕ можеш запускати Docker/e2e** — пишеш spec, reviewing session (Courses 1) ганяє `make e2e` і тріажить.
Пиши spec акуратно за реальними селекторами UI (L7a-c2) і реальним бекендом. Можеш дробити роботу.

## Контекст
Фаза enrollment, гілка `feat/enrollment`. Бекенд (L5–L6c) і фронт (L7a enroll/«Моє навчання», L7b ростер+block,
L7c1 report-форми, L7c2 admin-черга) закомічені й gate-green (unit/integration). L8 — наскрізний e2e.
**Бекенд-гейтинг free-контенту вже перевірявся e2e на L4a** (seed self-active членства сідованим студентам),
тож існуючі specs мали лишитись зеленими на рівні бекенду. **Нове — фронт L7**, тож:

### A. Оновити ЗЛАМАНІ існуючі specs (поведінка змінилась свідомо)
- `frontend/e2e/route-guards.spec.ts`: R36 (рядок ~30-33) і ~63-64 асертять `/learn`→`/courses` редірект.
  L7a ЗНЯВ редірект: `/learn` тепер рендерить «Моє навчання». Онови ці асерти на нову поведінку
  (signed-in на `/learn` лишається на `/learn`, видно заголовок/порожній стан «Моє навчання»; НЕ 404).
  `/learn/go/intro` splat-поведінку (невалідний урок) не чіпай, якщо вона не змінилась.
- Пройдись по решті e2e specs і онови ЛИШЕ ті асерти, що прямо конфліктують із L7-фронтом (напр. сторінка
  курсу тепер має aside з CTA «Записатися»/«Ви учасник»). НЕ переписуй логіку інших фіч. Якщо не певен —
  занотуй у WIP як «кандидат на перевірку Courses 1 у make e2e», не вгадуй.

### B. Новий `frontend/e2e/enrollment.spec.ts`
Шаблони: `purchase.spec.ts` (admin створює/структурує/цінує/публікує курс через API з runNamespace +
afterAll archive; `wcagViolations`, `overflowAtDoubleText`, 44px boundingBox), `learn-access.spec.ts`
(content-gate), `instructor.spec.ts` (teach-екран), `discussion.spec.ts` (коментарі/відгуки). Фікстури:
`auth.ts` `authStatePath('student'|'instructor'|'admin')`; seeded IDs student `…0001`, instructor `…0002`,
admin `…0003`. Контракт-поля — `contract.ts` `contractField`.

Сценарії (кожен назви кодом En; створюй власні run-namespaced курси, прибирай у afterAll):
- **En01 free enroll**: admin публікує БЕЗКОШТОВНИЙ курс; student на сторінці курсу бачить «Записатися»
  (не член) → контент закрито, опис+аутлайн видно; тисне «Записатися» → стає «Ви учасник» + «Почати
  навчання»; на `/learn` курс з'являється у «Моє навчання».
- **En03 content-gate**: не-член не має доступу до контенту (плеєр/урок) платного й безкоштовного; опис+аутлайн видно.
- **En05 blocked**: instructor блокує student свого курсу → контент лишається (urok відкривається), але
  комунікація (коментар/відгук) → помилка `enrollment_blocked`; інші курси студента цілі.
- **En04 revoke**: admin відкликає доступ (ростер/адмін) → контент відрізано (сторінка навчання закрита);
  членський рядок лишається (не дубль).
- **En08 roster**: instructor на teach-екрані свого курсу бачить ростер (ім'я/стан/%/дата), 0%-прогрес,
  порожній ростер → порожній стан; блок/розблок перемикає.
- **En07 reports**: student (член) скаржиться на викладача (діалог) → admin у `/admin/participant-reports`
  бачить скаргу, «Вирішити» прибирає; instructor/admin скаржиться на учня з ростера.
- **En14 G10**: instructor АБО admin «Записується» учасником ЧУЖОГО курсу (членство active, роль не змінюється).
- **En15 single-row**: blocked/revoked студент тисне «Записатися» → НЕ створює 2-го active (перехід/відмова), без дублів.
- **a11y**: `wcagViolations` (0) на нових екранах (сторінка курсу з CTA, «Моє навчання», ростер, адмін-черга,
  діалог скарги), 44px на нових контролах, `overflowAtDoubleText` на «Моє навчання»/ростері.

- **En04/En10 media-bytes + preview (ТРЕБА e2e — security-крух, виправлення 2026-10-07):** nginx
  `auth_request` на `/media/p/<playbackId>/` ганяється ТІЛЬКИ в Compose-e2e (integration перевіряє лише
  Go-рішення CanConsume, не реальний nginx). Доведи на продакшн-шляху: seeded **член** → байти 200;
  seeded **revoked/не-член** → **403**; **не-член на PREVIEW-уроці published-курсу** → 200 (CanConsume
  Published&&Preview — єдиний випадок, коли контент відкрито не-члену); не-preview урок не-члену → 403.
  Дзеркаль `playback.spec.ts`/`playback.setup.ts`/`media.setup.ts` для того, як дістається playbackId і
  робиться запит на media-URL + перевірка статусу auth_request. Один сценарій member(200)+revoked|non-member(403)
  +preview-non-member(200) достатньо.
- Пропусти (integration-достатньо — познач у WIP, НЕ вигадуй e2e): **backfill** (En13 — одноразова
  data-міграція над заповненою pre-enrollment БД; свіжий e2e-seed не має історичного стану; доводиться
  integration-прогоном 00015 на seeded БД); **grandfather** (En10 price-set — backend-only orders-логіка,
  нема нового enforcement-шва; integration-достатньо).

## Межі
НЕ чіпати бекенд/generated/застосунковий код фіч (лише e2e specs). Якщо e2e виявить потребу в зміні
продукту — занотуй у WIP як знахідку для Courses 1, НЕ міняй продукт у цьому брифі. НЕ комітити.

## Перевірка (у WIP)
`cd frontend && npm run typecheck` (specs типобезпечні проти generated); `npm run lint`.
`make e2e` ТИ не запускаєш (Docker) — Courses 1 ганяє й тріажить; чесно познач e2e як not-run.
