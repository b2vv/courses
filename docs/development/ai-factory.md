# AI Factory in Courses — Codex and Claude

<a id="portable-package-pin"></a>

## Pinned full collection — 26c68f6

[Package metadata](ai-factory-package.json) фіксує private `b2vv/ai-factory`, commit `26c68f6a4c9e3e024e7c9f075b1365778087509b`, release `0.6.0-candidate`. Єдине джерело повного спільного плану: `<factory>/ai-factory-plan.md`; оригінальні скіли: `<skills>/`.

- `<release>` = `~/project-data/ai-factory-rollout/releases/26c68f6a4c9e3e024e7c9f075b1365778087509b` (розгорнути `~` поточного користувача).
- `<factory>` = `<release>/ai-factory`; `<skills>` = `<release>/skills`; спільні rules = `<release>/rules`.
- Перед використанням звірити commit metadata, SHA release manifest/checksums і потрібних файлів із pin; повний release перевіряється `python3 -B <release>/tools/verify_release.py --root <expanded-release>`. Відсутній або змінений cache → BLOCKED потрібної залежності, без auto-clone/install.

Цей adapter — локальний router; чинні acceptance/contracts/ADR, stack, task/phase/journal, thresholds і QA/gates мають пріоритет. Entry-посилання на shared правила читаються через відповідний pinned rule, якщо локальний контракт не задає інший exact source. Повний успадкований plan — defaults і довідник; його історичні machine/project PASS, backlog і global AGENTS не переносяться як нові дозволи. `docs/portable/` і чотири `adapters/legacy-planning/*-portable.md` — явно окремі profiles, не підміна originals; у `skills/` лише канонічні оригінали. Ці adapter documents не є native-discovered SKILL.md.

84 внутрішні визначення враховано: 73 originals доступні, 11 synced license-restricted — лише inventory. Зовнішні та `b2b*` не входять до пакета. Локальні/proprietary/shared capabilities, яких немає в ньому, **не видаляти, не relink і не публікувати**. За колізії обрати exact source явно; локальні Cass/Cargo/Courses contracts та скіли залишаються чинними. Для `<skills>` відносні шляхи таблиці нижче зберігають групи. References усередині bundle читаються поряд із ним; зовнішні посилання на project docs/tools — runtime dependencies, не дозвіл на встановлення або послаблення gate.

Codex/Claude native discovery/invocation **NOT_TESTED**. Exact-file fallback: прочитати обраний `SKILL.md` та references й назвати точний файл; читання не є native invocation. Історичні `~/.agents`/`~/.claude` згадки всередині source bundle не перемикають чинний pin: для включених shared skills/rules застосувати цей mapping, для зовнішніх — явний local fallback або BLOCKED. Small/full/resume визначають незмінені секції нижче; другий tracker/bootstrap не створюється.

Hooks/MCP/models/permissions/global symlinks та active CRG/Graphify цим оновленням не змінюються. Успадковані helpers тепер у `<release>/runtime/legacy/`; їхня поведінка збережена, але автоматично їх не запускати. Portable `<release>/runtime/session-size/session_guard.py` потребує explicit store; `<release>/runtime/review/review_package.py` має hardened Git reads. Shared preflight: `<release>/runtime/onboarding/preflight.py`. Жодний helper не активується цим оновленням. Наявні live callbacks продовжують свій чинний контракт; це не їхня міграція.

Статус: **APPLIED_DOCS / INDEPENDENT_REVIEW_PENDING**. Наявна незалежна Claude-сесія цього проєкту; exact session reference pending; свіжий файловий пакет підготовлено для ручної передачі, доставка/отримання не підтверджені, без дубля reviewer/повідомлень. Author checks не є product/CLI/UI/review PASS. Baseline, diff, receipt і immutable review package: `~/project-data/ai-factory-qna-adoption/courses/`.


This is an adapter to the existing workflow, not a new pipeline, Cargo FSD port or tracker. Both hosts enter through [AGENTS](../../AGENTS.md) → [standards](../../.claude/standards.md) → this file. Shared startup: `<factory>/onboarding.md`; resolve `~` for the current user. Missing shared files are reported precisely, not installed automatically.

## Contracts and scope

Preserve Go modular monolith/media worker, inward domain/application/ports/adapters boundaries, and frontend TanStack Start/Router/Query + Rsbuild/Rspack + React Aria/Tailwind. No Cargo-specific architecture or gates. Accepted ADRs/owner decisions and the task spec are normative; code/scripts are implementation evidence. A stale descriptive note may be corrected; a conflicting norm is recorded with its owner, not silently replaced by code. Existing user authorization to restore an unambiguous contract is sufficient.

## Startup review routing і pinned package

Перед роботою в чинному task/phase record запиши й покажи користувачу для critique/review/verification: `мета → новий субагент | наявна окрема сесія | НЕ ВИЗНАЧЕНО; owner; exact session/agent reference або pending; підстава поточного вибору; delivery owner; shared-resource slot/дозвіл або BLOCKED`. Явний вибір користувача зберігається. Якщо його немає й поточне доручення не визначає маршрут — запитай до spawn/надсилання; safe inventory/planning можуть тривати. Не створюй дубль reviewer й не надсилай іншій сесії без явної авторизації; за ручної передачі пакет віддається користувачу. Author checks окремі від незалежного review.

Reviewer отримує task/author/reviewer IDs, exact root/branch/base/HEAD, closed file scope, dirty та relevant untracked hashes і diff, immutable acceptance/contracts references з hashes/копіями, raw evidence із commands/exits/status/revision та constraints. Він сам читає джерела, не лише авторський summary; за замовчуванням read-only. Тести/спільний стенд — лише з узгодженим доступом, власником і side effects; чинні локальні gates не послаблюються. Findings мають file:line, impact, verification, blocking/confidence. Відсутня відповідь чи доказ → review NOT_TESTED / required gate BLOCKED.

PASS належить точному package/snapshot ID та scope. Після правок reviewer оцінює final delta і зачеплені criteria/checks; явно приймає лише незмінені докази. Без автоматичного stale PASS чи повтору всіх gates для нерелевантних змін. Один `package ID + reviewer reference` передається один раз у наявному record; Small не отримує другого tracker. Повний shared контракт і опційний stdlib helper: `<factory>/review-handoff.md`, `<release>/runtime/legacy/review/review_package.py`. За відсутнього helper — ті самі поля вручну; не install/bootstrap.

## One lifecycle

- Small: below the existing standards threshold, one inline plan and scoped evidence in the current phase/work record; required gates remain. No new task directory or parallel spec-flow.
- Full: new subsystem, public API change or cross-cutting change across more than one layer of both backend and frontend, or explicit owner opt-in. Use the existing `docs/<phase>/` spec/plan/tasks/gates/report; one chosen orchestrator. Existing phase-specific task files remain authoritative for their scope.
- Tasks remain in [work/tasks.md](../../work/tasks.md) with planned → in_progress → done/blocked and result/archive pointers; completed rows are not deleted. [Ledger](../../work/archive/ledger.md) remains append-only by meaning. This documentation onboarding's primary record is [readiness](ai-factory-readiness.md), not a new product phase.
- Resume: read the current phase record/ledger, verify root/branch/HEAD/relevant dirty and untracked inputs plus pinned evidence. An unchanged completed handoff does not reopen work or rerun accepted gates. Changed scope invalidates only affected proofs unless the local contract requires more.
- Handoff: goal/scope, sources and decisions, root/HEAD/diff, exact checks and verdicts, blockers, next owner/action. Keep detailed evidence in the repository. Claude's existing hook/Obsidian route is host-specific; its existing fallback is `docs/<phase>/_wip-<topic>.md`. Codex can use that same repo record. Do not make external note availability a prerequisite or write externally without authorization.

## Skills, hosts and review

Existence, per-host/session discovery, native tool availability and actual invocation are separate. Codex `$name` / `/skills`, Claude `/name` only when supported/discovered. No native list/load tool is not proof that a skill is undiscovered. This rollout reads files; it does not certify either CLI's discovery. Resolve exact selected files, including references, before fallback; do not invent slash success or install a global Courses/Cargo review skill.

| Scope | Exact shared source below `<skills>/` |
|---|---|
| Check-only inventory | `utils/repo-bootstrap/SKILL.md` |
| Independent two-axis review when selected | `arch/code-review/SKILL.md` |
| Explicit full orchestration | `deliver/spec-flow/SKILL.md` |
| Acceptance/plan critique/defense when required | `plan/acceptance-spec/SKILL.md`, `plan/plan-critique/SKILL.md`, `plan/plan-defense/SKILL.md` |
| UI/design with affected render/interaction | `design-pass/SKILL.md` |
| Structural change, including test helpers | `pattern-architecture-review/SKILL.md` |
| Entity/scoped-filter acceptance | `entity-e2e-coverage/SKILL.md` + `entity-e2e-coverage/references/testing-rules.md`; required fields × surfaces × states and matching/nonmatching filters only within acceptance |
| Briefing or delivery report when useful | `plan/cto-research/SKILL.md`, `ship/ship-report/SKILL.md` |

Do not invoke the entire table. Keep local task/phase artifact paths. Review's standards source is `.claude/standards.md` and its scoped referenced architecture/ADR; spec source is the actual task/phase acceptance, not a briefing. The existing repo tracker supplies task provenance; a generic skill's tracker candidate is not permission to set up an external tracker. If using shared code-review, retain independent Standards and Spec axes with the same pinned diff, separate findings and durable verdict; exact-file fallback retains these requirements. Any unmet required reviewer/tool is BLOCKED, not author self-review PASS. Human review chosen by the task remains valid. Specialty applicability is recorded once in the same review; no duplicate ceremony.

## Checks, models and optional integrations

[Testing](testing.md) is the gate/effect/CI matrix. `make check` and `e2e-local` reset dev volumes; drain/seed/generate/start also mutate state. Preserve the one-heavy-session rule. This docs rollout authorizes no product checks, installs, infrastructure/config edits or live service calls. Future product gates require their appropriate scope/authorization. UI copy still needs layout/accessibility assessment; build is not browser evidence. No screenshot-only or aggregate-gate shortcut for acceptance rows.

Helper/Coder/Reviewer are roles, not model IDs. Record host/session-specific catalog provenance, candidate, entitlement, budget and actual binding separately under `<factory>/models.md`; unknown stays UNBOUND with a next step. Do not copy an app catalog into CLI, change defaults or run paid probes.

CRG/Obsidian are optional. A config file proves neither tool listing nor a live query. CRG evidence needs this Courses root/branch/HEAD, file hashes and coverage including untracked, plus a relevant query corroborated on disk. Cargo's snapshot proves nothing about Courses freshness. Missing/stale/unavailable → files/rg; refresh only separately authorized with a quiet writer window. Do not change DB/registration/hooks/permissions during onboarding.

## Idempotence and completion

Canonical files: this adapter, [manifest](ai-factory-manifest.json), [readiness record](ai-factory-readiness.md), and the single AGENTS pointer. Update in place, never duplicate a block or tracker. Manifest host labels are static inventory identifiers; replace with actual session provenance before a runtime claim. Helper results remain declarations, not configured/runtime acceptance.

Before authorized changes capture scoped originals/backups/hashes and reviewed diff; recheck before each write. Concurrent difference → BLOCKED that file. Repeat exact adapter bytes → NOOP; changed owner content requires a reviewed merge. Restore only own applied bytes when current SHA equals applied SHA, from verified backup, with an explicit restore scope; no git reset/clean or recursive removal. Full protocol: `<factory>/onboarding/change-protocol.md`.

Completion requires independent review and exact checks/evidence in the readiness/current phase report. PASS/FAIL/BLOCKED/N/A carry reasons; docs/fixtures do not claim product, CLI, MCP or UI PASS. No commit/push/merge/deploy follows automatically. Remaining contract decisions must name their impact and owner.

<a id="pre-code-checkpoint"></a>

## Перед кодом: один checkpoint у чинному record

### Critique прив’язаний до плану й snapshot

Це уточнення **того самого checkpoint**, не додатковий tracker. До залежної реалізації окремий критик плану перевіряє decision surface; незалежний review коду лишається іншим gate. Використай вже обрану незалежну сесію/контекст, без дубля. Prompt — чинний локальний versioned prompt або `<release>/prompts/plan-critic.md`; запиши exact path/version/SHA-256 і release. Generic prompt не замінює чотири лінзи оригінального plan-critique skill.

Verdict має `stage: plan`, critic reference, plan path/version/SHA-256, target root/HEAD і relevant dirty/untracked hashes. Raw findings або durable pointer, disposition/evidence кожного finding та відкриті blockers живуть у чинному record. ACCEPTED із відкритим blocker не дає READY. Командний доказ містить command + exit + status + artifact + точний snapshot; запланована команда не є виконаною, unrun required check — BLOCKED/NOT_TESTED без вигаданого exit.

**Material change:** acceptance/scope, contracts/dependencies, архітектурне рішення, risk або test strategy. Потрібна нова критика зачепленого scope й звірка залежних задач; незалежні задачі sprint не інвалідуються автоматично. Для правопису/форматування — перевірений delta та явна прив’язка збереженого verdict до нової версії, без удаваного повторного review. За зміни target snapshot перевір зачеплені evidence; old HEAD/PASS не покриває нові dirty bytes. Small зберігає обґрунтований N/A лише за локальними правилами; resume не повторює незачеплені gates. Повний контракт і синтетичний batch-приклад — `<factory>/pre-code-checkpoint.md`, `<release>/examples/checkpoint/sprint-batch.md`.


До першої product mutation оператор заповнює, а виконавець звіряє цей блок у наявному task/plan/phase record; brief посилається на нього. **Схвалення sprint/batch не робить усі задачі READY.** Для кожної задачі є власний статус і залежності; окремого tracker або комплекту документів не додаємо.

| Поле | Мінімальний доказ |
|---|---|
| Scope / acceptance | Джерело й IDs, спостережуваний результат, non-goals, owner, small/full/resume з причиною |
| Grilling / рішення | Прогалини, припущення, альтернативи, потрібні питання та відповіді/рішення з джерелом; висновок автора окремо від рішення власника |
| Залежності / writers | Готовність потрібних попередників з доказом; спільні файли/ресурси, власник і порядок, якщо потрібні |
| Scoped impact | Перевірені entrypoints/callers/contracts/tests, конкретні file:line або search/query evidence, межі покриття; root/HEAD і fingerprint або явний inventory relevant dirty/untracked |
| Планова критика / defense | Для full/ризикового плану — незалежна критика й п’ять питань defense або явно погоджений еквівалент; applicability, маршрут, reference і доказ verdict; для small — конкретна причина N/A лише там, де локальна норма дозволяє |
| Допуск | READY / BLOCKED, owner, невирішене required питання/залежність/finding і зачеплений scope. Обов’язковий незакритий пункт не сумісний із READY |

**Impact обов’язковий, інструмент умовний.** Доступний CRG/ttsc можна використати, Graphify — необов’язковий; запуск усіх графів не потрібен. Graph evidence має root/revision, scope hashes, coverage та relevant untracked; stale/unknown/null не доводить безпечності. За недоступності, несвіжості, непропорційності чи відсутності дозволу — явний файловий fallback: причина, конкретні callers/contracts/checks і межі пошуку. Порожній результат потребує доказу охоплення. Не оновлювати graph DB автоматично.

Shared контракт: `<factory>/pre-code-checkpoint.md`; exact skills: `<skills>/plan/grilling/SKILL.md`, `<skills>/plan/plan-critique/SKILL.md`, `<skills>/plan/plan-defense/SKILL.md`. Якщо заявлено саме plan-critique, потрібні чотири незалежні лінзи його контракту; self-check не є PASS. Чинний вибір existing-session/subagent зберігається, без неявного spawn/дубля чи повідомлення; відсутній required reviewer блокує залежний gate. Post-code review не є ретроспективною плановою критикою.

**Small:** кілька речень у тому самому record, без вигаданих питань. **Resume:** послатися на checkpoint, звірити scope/HEAD/dirty delta, повторно перевірити лише зачеплені питання, залежності, impact і gates; незмінені докази зберігати явно. Історичні задачі не переписувати заднім числом, старий PASS не переносити на новий scope. READY не дозволяє commit/push/deploy чи інші ефекти поза чинним дорученням.

Опційний `<release>/runtime/checkpoint/precode.py` перевіряє декларації JSON і окремо наданий current snapshot; Markdown достатньо, дубль JSON не потрібен. Це не перевірка правдивості доказів або автоматичне CI/hook enforcement. Тести helper не доводять дотримання агентами. Сильніші локальні gates залишаються чинними.

Courses: checkpoint живе в чинному phase/task plan, на який веде [work/tasks.md](../../work/tasks.md). Фазові acceptance/gates, Go architecture і TanStack Start/Rsbuild зберігаються; історичні рядки tracker не стають доказом READY нової задачі. `make check` не запускати як безпечну перевірку документації: чинні обмеження side effects залишаються.
