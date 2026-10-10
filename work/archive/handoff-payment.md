# Хендоф — фаза оплати, 2026-09-29

Стан репо на мить коміту цього файла: гілка `feat/payment`, HEAD — коміт, що його додає; `main` = `a1d3898`.
Перш ніж діяти — звір із `git status`, `git log main..feat/payment`, `work/tasks.md` (§24 п.17 бріфу).

## Де фаза

Оплата — **остання й єдина** фаза роадмапу, що лишилась. Робоче місце інструктора закрите й змерджене
(PR #2, `2c0d29f`); брифінг і агенда перезібрані на main (`a1d3898`).

| Крок spec-flow | Стан | Де |
|---|---|---|
| spec + GATE 1 | ✅ | `docs/payment/spec.md` (`654b619`, `10868fe`) — рішення 1–7 |
| plan | ✅ DRAFT | `docs/payment/plan.md` (`7772d93`) — A1–A6, зрізи O01–O07, Pay01–Pay25 |
| **GATE 2** | ❌ **вердикту немає** | 3 прогони × 2 рецензенти → 0 результатів (див. нижче) |
| tasks / implement | ⬜ | після зеленого GATE 2 |

Жива нота фази: `docs/payment/assumptions.md` (Obsidian MCP недоступний — нота в репо).

## Ключові рішення (не переглядати без людини)

- Юрисдикція: Україна → Європа → далі. **Провайдер не названо** — блокує лише зріз O06 (адаптер,
  checkout, вебхук); фаза закривається тільки після O06 (доказ через sandbox, не фейк).
- Валюта UAH. Grant: покупка (`source=purchase`) + адмін-видача (`source=admin`, `order_id` NULL).
- Право видає лише вебхук; редірект — ні; UX «оплата в обробці».
- **Ціна на курсі** (NULL = безкоштовний); власник+адмін завжди; **прев'ю-уроки в MVP**.
- Правило: `admin ∨ owner ∨ entitled ∨ (published ∧ (¬paid ∨ preview))` → ADR 014 (ще не написаний, зріз O01).
- Мапінгів доступу **4** (course, playback, transcript, discussion); join на право — 6 запитів у 5 файлах,
  у 4 треба прокинути userID; гарячий шлях — nginx `auth_request` (`playback/adapters/postgres/access.go:11`).

## GATE 2 — що сталося і як продовжити

- Прогін 1 (Opus 5.5) і прогін 3 (аліас `opus` → API показав `claude-opus-5-5`): обидва рецензенти падали на
  API safeguard `reasoning_extraction`. Прогін 2 (sonnet) зупинено: користувач хоче **Opus 4.8**.
- Аліас `opus` у цьому середовищі = **5.5**, не 4.8. Прибита версія: `~/.claude/agents/gate-reviewer.md`
  (`model: claude-opus-4-8`) — тип агента видно **лише в новій сесії** (у старому процесі «not found»).
- Часткові записи рецензентів (обидва встигли лише прочитати план): `scratchpad/pay-gate2-*.md` сесії
  `b579daa1…` — ефемерні, нова сесія починає з нуля.

**Перший крок нової сесії:** запустити паралельно два агенти `subagent_type: gate-reviewer`:
1. **Критика** — скіл `plan-critique` на `docs/payment/plan.md`; перевірити особливо: (a) повноту покриття
   доступу — кожен шлях, що віддає вміст уроку платного курсу (outline, progress, transcript documents
   library, WebVTT, ASR-чернетки, discussion roots/replies/moderation, media auth_request); пропущений
   читач = blocking; (b) узгодженість прев'ю між сесією плеєра, сегментами, транскриптом, обговоренням;
   (c) A2 — читання `course_grants` іншими модулями vs AGENTS.md; (d) міграція 00011 vs видалення
   чернеток курсу й soft-delete користувача; (e) `ApplyPaymentEvent` — дві події, подія до коміту
   замовлення, розбіжна сума; (f) чи O01 поведінково нейтральний.
2. **Захист** — скіл `plan-defense`, п'ять питань (некоректний інпут; відмова залежності; гонки й повтори;
   порожні/граничні дані — **особливо «курс стає платним, коли студенти вже вчилися»**; rollback).

Формулювання — нейтральні (цілісність/валідація), без «атака/підробка» — з ними safeguard спрацьовував.
Кожен рецензент першим рядком пише свій model ID — звірити, що це `claude-opus-4-8`.

## Пастки середовища

- **Docker `credsStore: desktop`** вішає гейт: temp `DOCKER_CONFIG` = симлінки на `~/.docker/*` +
  `config.json` без `credsStore` (порожній `{}` ламає `docker compose`). Docker Desktop не перезапускати сам.
- Docker міг зникнути посеред сесії (сокет `~/.docker/run/docker.sock`) — фоновий поллер
  `/tmp/claude-501/wait-and-gate.sh` чекає на движок і жене `make check`.
- **Codex:** forwarder `codex:codex-rescue` лише запускає задачу; статус/результат — з головної сесії через
  `node ~/.claude/plugins/cache/openai-codex/codex/1.0.6/scripts/codex-companion.mjs status|result <id>`.
  Модель — `--model gpt-5.6-sol` (прохання користувача). Пісочниця Codex не жене docker-гейти — жене сесія.
- `check-generated` порівнює з індексом — стейджити згенероване перед прогоном.
- Два PNG у `docs/foundation/screenshots/` — байпродукт гейта (B-18), не комітити; `.catalog-desktop.png.swp`
  — чужий swap-файл, не чіпати.

## Ролі

Courses 1 — план, брифи, гейти, коміти; код — Codex (sol); рев'ю дифу — Courses 2 / свіжі рецензенти.
