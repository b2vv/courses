# Courses AI Factory — readiness record

## Scope and source

Authorized by the owner, 2026-10-17: “Давай тепер courses”. Additive documentation onboarding after the shared portable route was completed. No product/infrastructure/CI execution, permissions/credentials/model defaults, CRG DB/MCP registration, installs, live/reset/seed tests or commit/push/deploy. Small documentation scope; the local structural spec-flow threshold is not reached.

Inventory snapshot: `/Users/strelia/projects/courses`, branch `feat/editor-tabs-semantics`, HEAD `f18cf9673f267cdbc3d1464a23cb9a6b11c018f8`. Other sessions have been moving HEAD; this is a dated baseline, not a permanent branch requirement. Tracked files were clean at inventory, with many existing untracked phase/review artifacts retained. Scoped baseline hashes and protected source fingerprints guard the apply; never overwrite concurrent changes.

## Inventory and routing

| Area | Result |
|---|---|
| Standards | Existing `.claude/standards.md`; now discoverable from shared AGENTS entry |
| Briefing | Existing `work/CTO-RESEARCH.md`; a cache, not authority over accepted contracts |
| Contracts | Existing AGENTS and ADRs; Claude imports AGENTS |
| Persist/handoff | Existing phase docs, task table and append-only ledger; Claude Obsidian fallback reused, no external write |
| Artifacts | Existing `docs/<phase>/`; this small docs task uses this record |
| Tasks/archive | Existing rows stay in `work/tasks.md`, archive pointers and ledger preserved |
| Commands | Makefile + frontend package inventoried; effects in [testing](testing.md), no execution |
| Tracker | Existing repo tracker; no external issue tracker setup |
| Optional CRG | Registration file detected; current Courses tool availability/freshness/query NOT_TESTED; files/rg fallback |
| Optional Obsidian | Existing host-specific route; live availability NOT_TESTED, repo fallback available |
| Agenda | Existing `work/AGENDA.md`; not rebuilt or auto-executed |

Go modular monolith and TanStack Start/Rsbuild/React Aria stack are retained. [Adapter](ai-factory.md) owns routing only; [manifest](ai-factory-manifest.json) is explicit static input, not runtime certification. The adapter selects relevant shared skills only. Exact review fallback: `~/.agents/skills/arch/code-review/SKILL.md`, with local standards plus actual acceptance, independent Standards/Spec axes and pinned files. No Cargo review/FSD/mandatory mutation gate is copied.

| Host | File route | Discovery | Native tools | Invocation | Model roles |
|---|---|---|---|---|---|
| Codex CLI | Exact shared paths readable; review fallback available | Courses session NOT_TESTED | UNKNOWN | NOT_TESTED | UNBOUND pending same-host/session catalog, entitlement/budget |
| Claude Code | Same exact file route; existing hook instructions retained | NOT_TESTED | UNKNOWN | NOT_TESTED | UNBOUND pending its own inventory |

The supplied manifest session names identify this static inventory, not a running CLI session. Any previously reported global Codex discovery is dated evidence, not invalidated by absent list/load tools and not silently promoted to Courses invocation proof. Native agents used for this documentation review are separate from those CLI claims.

## Reconciled facts, without new policy

- Media: lesson-builder owner decisions authorize image `auth_request` with CanConsume + EmailVerified and no playback lease (spec AC9/Gate1; plan owner decision 2026-10-12). AGENTS now names both playback and image routes; nginx/product code unchanged. Earlier lesson-builder plan preview wording is historical and is overridden by that explicit later decision, not a newly invented vulnerability or access rule.
- ADR010's Phase6 access body is historical; ADR013/014 retain the pure-contract and CanView/payment decisions. Dated notes link enrollment CanConsume/CanCommunicate and image decisions without replacing ADRs wholesale.
- Standards distinguish descriptive code facts from normative acceptance/ADR/owner decisions. Task spec remains its scoped acceptance source; no blanket promotion of all reports to standards.
- Testing documents actual local reset/drain/generation/seed effects and current CI verify vs tag/manual browser, including ASR exclusion. Required local gates and CI implementation unchanged.
- README points to current measured phase outcomes instead of claiming missing playback/payment/transcription. It does not claim production deployment or a fresh product PASS.

## Checks and acceptance

Documentary readiness: **reviewed and configured when installed hashes match the reviewed receipt**. Independent Standards and Spec reviews returned PASS with zero findings; the fresh forward-review passed all nine routing scenarios. This is not native CLI or product-runtime acceptance.

- Static path check: 87 local Markdown references resolve; no fragments in this changed scope. Two pre-existing ledger navigation paths were corrected without changing historical entries.
- Installed shared preflight was run read-only against Courses. Both required skill files were detected; both hosts have exact-file review fallback, native inventory UNKNOWN and invocation NOT_TESTED. Before adaptation, new adapter was PROPOSE_ADD and differing existing entries BLOCKED automatic overwrite; the reviewed hash-guarded merge resolves that expected planning state.
- Disposable Courses fixture: all 12 files applied from backups, canonical adapter/entry/standards repeat returned three NOOPs, single entry/task row preserved, unrelated owner content retained; guarded restore and repeat restore passed. This is fixture proof only.
- All affected records retain existing gates and lifecycle. No product command, browser, hook, MCP query or model probe ran.

Applied file hashes, postwrite comparison and live repeat results are recorded in `applied-receipt.json` / `postflight.json` in the evidence workspace. They certify only this documentation batch; consult them before restoring. A concurrent content mismatch blocks the affected write/restore.

Evidence workspace on the onboarding host: `/Users/strelia/Documents/Codex/2026-10-16/task/courses-onboarding/` (baseline backups, manifest, raw preflight, review and receipt). These are historical evidence paths, not dependencies required to use the adapter on another machine. Portable rules and source links live in this repository.

## Readiness boundaries

No unresolved new access-policy decision was needed for the documented factual corrections: the owner decisions already exist. Historical ADR text remains preserved with explicit current-scope notes. No product gate was reduced.

Future live product checks remain outside this task: an authorized destructive-run scope and exclusive heavy-gate window are required before `make check`/`e2e-local`; missing entitlement/model budget/reviewer tools remain UNBOUND/BLOCKED only for the dependent runtime work. Unknown optional CRG/Obsidian does not block file-based documentation work. Production deployment remains deferred to the owner; no new deployment decision is implied.

Final pre-apply identity observed: `main` / `ddd8c000f1144f665bcb171359c617661d56aebb`; earlier snapshot retained above to make concurrent movement explicit. Scoped document baselines and all 11 protected normative/executable sources were unchanged at this reconciliation. Application checks them again; current HEAD is never reset.


## Pinned full collection rollout — 2026-10-18

Commit `626fd6852e7ca53a4022feff30b7bcd492768733`, package `0.3.1-candidate`; [pin](ai-factory-package.json). Повний plan/original skills через `collection/agents`, локальні contracts/QA/tasks/skills збережено. Baseline HEAD `deec70a64cb6ac78bb80275de5f9b104303b5f8c`, branch `feat/refunds-view-designpass`, status records `211`; dirty bytes збережені. Поточний статус **APPLIED_DOCS / INDEPENDENT_REVIEW_PENDING**. Попередні acceptance/PASS не поширюються на цю delta. Author validation: hashes, references, static routing, guarded repeat/rollback; product checks та native runtime не запускались. Докази й manual review у `~/project-data/ai-factory-rollout/runs/626fd68-20261018T051145Z/courses/`. Manifest hashes означають installed bytes, не незалежне прийняття.


## Canonical layout upgrade — 036ccad-20261018T061121Z

Pin `036ccad5b3f12de35fa9a0895d281a739a4e581b`, version `0.4.0-candidate`; previous `626fd6852e7ca53a4022feff30b7bcd492768733` remains in the immutable cache for rollback. Live plan/skills/rules resolve through `ai-factory/`, `skills/`, `rules/`; explicit runtime variants and non-native adapter profiles retain their behavior. Existing local contracts, architecture, task lifecycle, entry files and gates are unchanged.

Fresh baseline: `main` / `80a167dd4f328884b171d3222676da56b2d4dd2b`, 212 status records. Evidence, exact backups and guarded apply/rollback: `~/project-data/ai-factory-rollout/runs/036ccad-20261018T061121Z/courses/`. No product commit/push or gates; no live hooks, permissions, models, graph stores or services changed. Static hash/link/anchor/routing and exact-file loading checks only; native CLI discovery/invocation NOT_TESTED. Repeat matching target bytes is NOOP; rollback is tested in an isolated fixture, not performed on this live checkout.

Status: APPLIED_DOCS / INDEPENDENT_REVIEW_PENDING. Existing project Claude reviewer retained; no duplicate session or iTerm input. No response to the previous package ID was found in the inspected project-data/Obsidian handoff locations. Prior review proofs do not accept this new snapshot. Fresh immutable package is prepared for manual handoff; delivery and independent verdict are unconfirmed. Manifest hashes describe installed bytes, not independent acceptance.


## Pre-code checkpoint repair — 2026-10-18

Pin `2f2fa943a2a2cf411c656230ba94fefbac658567`, version `0.5.0-candidate`. Додано один checkpoint у чинний lifecycle: per-task READY, proportional grilling/рішення, dependencies, scoped impact/freshness або file fallback, planning review/defense. Сильніші локальні gates збережено; новий tracker не створено, історичні task verdicts не переписано.

Baseline `b2b91a0b33890addfebf4ff6330a04826dab28ed`, branch `feat/modules-catalog`, status records 214. Backups/hash guards, точний diff, ізольовані NOOP/rollback fixtures і postflight: `~/project-data/ai-factory-sprint-repair/courses/`. Author static checks не є product/CLI/UI PASS або independent acceptance. Відповідь чинної незалежної Claude-сесії PENDING; новий пакет для ручної передачі, delivery не підтверджено. Hooks/models/permissions/credentials/graph stores не змінюються; product commit/push/deploy не виконуються.


### Claude entrypoint follow-up і QnA comparison

CLAUDE.md веде прямо до checkpoint. У CLAUDE.local.md змінено лише несекретні workflow формулювання: per-task READY, прийнятий independent route, explicit handoff pointer замість latest-note authority. Український префікс і всі байти поза точними замінами збережено; private-file backup/full diff не створено. Порівняно read-only openai-qna: decision-surface, durable critique disposition і snapshot evidence корисні; branch-archive/стек/permission політики не перенесені. Нове незалежне review цієї delta PENDING.


## QnA concepts adoption — 2026-10-18

Shared `26c68f6a4c9e3e024e7c9f075b1365778087509b` / `0.6.0-candidate`. Авторський generic versioned plan-critic prompt, точна прив’язка verdict до plan version/hash і target snapshot, material-change критерії та command/exit/status/revision evidence доповнюють чинний checkpoint. Decision surface, per-task READY, impact fallback і small/resume вже були; нових tracker чи native registrations немає. Незалежна планова критика відрізняється від code review, чинний existing-reviewer маршрут збережено.

Baseline `8c2fe63f96bf2bf9087111ed042d7e6b28a892b4` / `feat/modules-teach`; 237 dirty status records. Backups/guards, isolated rollback, NOOP і review handoff: `~/project-data/ai-factory-qna-adoption/courses/`. Статус APPLIED_DOCS / INDEPENDENT_REVIEW_PENDING. Product gates, native discovery, live hooks/graphs/paid probes не запускалися. CLAUDE.local.md не читався і не змінювався; українські правила та чужа робота не є targets. QnA read-only, його код/тексти/стек/landing політики не скопійовані.
