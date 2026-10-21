# Backlog — named, not scheduled

Items that have been decided as *worth doing* but not placed in a phase. Not an agenda:
`work/AGENDA.md` ranks the next move and is rebuilt each time. An item leaves here by
becoming a phase in `work/PRODUCT-BRIEF.md` §21 or a task in `work/tasks.md`.

## B-27 — TEMPORARY retry for the testcontainers reaper race

**Raised:** 2026-10-03, integration-gate reliability follow-up.

**What:** `internal/testsupport/containers.Start` retries only testcontainers-go's transient
Ryuk create→start race, terminating the partially created target returned by the library before
each retry. Integration call sites use it and Ryuk is enabled again.

**Why:** testcontainers-go v0.44.0 can find a sibling process's Ryuk container in `created`
state and permanently fail rather than retry; upstream [PR #3868](https://github.com/testcontainers/testcontainers-go/pull/3868)
adds the library-level retry. See also the related [issue #3918](https://github.com/testcontainers/testcontainers-go/issues/3918).

**Trigger:** a testcontainers-go release containing #3868.

**Removal:** delete `internal/testsupport/containers` (it also holds the real-path race test and the two
guards below) and every call site, then run `make integration` three times.

**Proof (2026-10-03):** `race_integration_test.go` makes the race deterministic — a reaper for an isolated
`TESTCONTAINERS_SESSION_ID` is created but not started, then started 1 s later; the first postgres start fails
with the library's own `wait for reaper … unexpected container status "created"`, the target is terminated, the
retry reuses the running reaper. Passed 3 of 3 separate runs, 0 leftovers. Guards: `TestPinnedTo…` fails when
testcontainers-go leaves v0.44.0 (forces this decision on upgrade); `TestNoCallSiteUsesContainerReuse` fails if a
call site starts using container reuse (Start would terminate a shared container).

**Known bound:** a permanent #3918 orphan (a reaper that never starts) matches the same text, so each package
spends the full 30 s budget (~15 create/terminate cycles) and then fails loudly naming #3868. Bounded, not hidden.
Review: Courses 2, `work/archive/review-ryuk-retry.md` (approve, 0 blocking).

## B-23 — e2e B39 впав раз у повному гейті — ЗАКРИТО 2026-10-04: дефект тесту (застарілий знімок)

**Raised:** 2026-09-29, під час гейта Phase 7b D04 (Courses; розбір разом із Courses 2).
**Shape:** спостережене падіння, не діагноз. Слово «flaky» тут навмисно не вживається — це ярлик, а
не магнітуда.

### Що саме

`frontend/e2e/playback.spec.ts:964` «B39 O1 remints after 410 without takeover» упав у повному
`make check` на дереві з D04: `restoreWrite` = `undefined` (`:1139`) — запис відновлення позиції на
`<video>` після ремінту не спостережено; `maxFailures` зупинив ще 21 тест.

### Дані — з межею їхньої сили

| Контекст | Результат B39 |
|---|---|
| Повний гейт до D04 (P06, B-20, D01, D02, D03) | 5/5 ✓ |
| Повний гейт з D04, що дійшов до e2e | 1 ✘, 1 ✓ |
| Повний гейт D05 (D04 у дереві), 2026-09-29 | 1 ✓ |
| Повний гейт D06 (закриття 7b), 2026-09-29 | 1 ✓ |
| Повний гейт з D04, що НЕ дійшов до e2e (ryuk/testcontainers не стартував) | інфра-аборт — не дані по B39, у знаменник не входить |
| Ізоляція на стенді з D04, пересів перед кожним прогоном, за рядком самого B39 | 3/3 ✓ |

**Межа:** за такого N числа не атрибутують падіння ні до D04, ні від нього. Якщо базова частота в
повному гейті ~1/4, п'ять чистих прогонів поспіль мають імовірність (3/4)^5 ≈ 24%. Ізоляція не розрізняє
гіпотез узагалі: обидві («D04 винен» і «давня крихкість під навантаженням») передбачають там 0 падінь.

**Сильний сигнал — структурний:** шляху впливу D04 на плеєр у коді немає. `channel`/`positionMs`
існували до D04; `LessonPlayer` рендериться безумовно без `key`; `discussion` не торкається
playback-сесії; новий рантайм-ефект — один `GET …/comments` при відкритті уроку.

### Спосіб перевірки, коли повториться

Лише повним гейтом: A/B HEAD-до-D04 проти D04, по 2 повні прогони. Ізольований повтор інформації не
дає. Повторюючи окремий e2e, пересівати фікстури перед кожним прогоном і судити за рядком самого тесту:
`-g` тягне 46 тестів залежних проєктів, і `admin-users P08` витрачає `delete-me`.

### Повтор 2026-10-04 (фаза рейтингів, гілка feat/ratings)

Повний `make check` на T06b+T07: B39 упав ще раз — `playback.spec.ts:1139`
`expect(restoreWrite?.assigned).toBeGreaterThanOrEqual(preRemint - 1)`, отримано `undefined` (запис відновлення
позиції не зафіксовано), 76 passed / 1 failed / 21 did not run. Перед ним на тій самій гілці 5 повних прогонів
дійшли до кінця playback без падіння B39. Зміни фази рейтингів playback не торкаються (відгуки — окремий блок
сторінки уроку; `LessonPlayer` не змінювався). Діагнозу нема — лише ще одна точка спостереження; повний гейт
перезапущено.


### Закрито 2026-10-04 — механізм знайдено (Courses 1)

Не плеєр, а тест. `playback.spec.ts` B39 знімав `writes` (записи `currentTime`) **до** `expect.poll`, потім до
30 с чекав, поки потрібний запис з'явиться в **живому** масиві відео, і після успіху poll перевіряв
`restoreWrite`, обчислений зі **старого** знімка. Коли запис відновлення ставав між знімком і успіхом poll,
`restoreWrite` був `undefined` при щойно доведеному записі — рівно спостережене падіння (рядок перевірки після
poll, `Received: undefined`; обидва зафіксовані падіння — 2026-09-29 і 2026-10-04 — мали цю форму). Виправлення:
фінальні перевірки рахуються зі знімка, перечитаного після poll (`readWrites()` + `findRestore`), ранній вивід
діагностики лишився. Перевірка: повний гейт фази після виправлення (гілка `fix/b28-review-cursor-null`).

## B-22 — Duplicated current-video selectors — CLOSED 2026-10-20 (deliberate limitation)

**Raised:** 2026-09-28, during Phase 7b D02.
**Shape:** the filtered current-video selector is duplicated in several places: `transcript`
(`adapters/postgres/repository.go:25`, `adapters/postgres/draft.go:73,262`), `playback`
(`adapters/postgres/playback.go:18`, `progress.go:22`), `media`
(`adapters/postgres/generated/media.sql.go:48`, with `NULLS LAST`), and `discussion`
(`adapters/postgres/repository.go:29`). This list was verified by grep on 2026-09-28.
`course/adapters/postgres/learning.go:46` carries a deliberately different sorting form for
learning-state display.
**Механізм:** keep the filtered copies aligned by review comments and pin the `discussion` copy
with integration K09a; do not merge closed modules without a phase that owns their seams.
**Спосіб перевірки:** K09a proves `discussion` selects `state='ready' AND evicted_at IS NULL` at
the highest generation, including positive and negative ready/upload fixtures.
**Чого не стверджує:** this does not prove the filtered selectors are globally unified, nor does
it remove the duplication or claim that the `course` sorting form is equivalent.

### Outcome 2026-10-20 — закрито як свідома limitation (не код)

Звірено з кодом + ADR 013. Дубль — це **SQL-рядки**, по одному в кожному модулі (transcript/playback/
media/discussion), не Go-код. **Механізм, чому не прибирається:** ADR 013 забороняє модулю імпортувати
пакети іншого модуля, а контракт у `internal/contracts/…` — це **чиста Go-функція, не SQL**; винести
спільний SELECT нікуди, не перетнувши межу модуля або не завівши SQL у contracts (чого ADR не дозволяє).
media-копія ще й інша за формою (`NULLS LAST`), а `course/learning.go` — свідомо інша (сортує ready
першим для показу стану). Копії тримаються вирівняними рев'ю-коментарями, а discussion-копія пришпилена
integration-тестом K09a (`state='ready' AND evicted_at IS NULL` на найвищій генерації). Це свідомо
прийнятий дубль за архітектурною межею, не борг до виправлення — відкривати лише фазою, що володіє
seam'ами цих модулів (напр. спільний playback-read-contract), якщо така колись з'явиться.

---

## B-1 — Observability: frontend logging, log aggregation, dashboards

**Raised:** 2026-09-06 (user), during Phase 5.
**Shape:** end-of-roadmap hardening, after the product does something worth watching.
**CLOSED 2026-10-21** — spec-flow phase, PR #62 squash-merged `0b69d04`. Shipped: backend
`/api/v1/client-logs` ingest (§17 redaction) + `source=client`; frontend error bridge (onCaughtError
+ window.onerror + unhandledrejection, no dedup); SSR `server.mjs` lines tagged `source=ssr`+release+
route; Loki + promtail (docker_sd) + Grafana (provisioned datasources + overview dashboard) under the
opt-in `monitoring` profile, network-only; nginx `clientlog` zone. Ordered `make check` EXIT=0 (e2e
124); `infrastructure/tests/monitoring-stack.sh` PASS (AC4/AC5/AC6); independent 2-axis code-review
PASS. Limitations + 5 dated plan deviations in `docs/observability/report.md`. Not deployed to prod
(owner 2026-10-03): infra built, not wired; Grafana network-only until S01 (TLS/auth).

### What already exists (verified in code, 2026-09-06)

| Piece | State | Where |
|---|---|---|
| Backend structured JSON logs | **exists** — request id, trace id, no secrets, URLs without query strings | `backend/platform/observability/`, the rule in `AGENTS.md` |
| Prometheus metrics | **exists**, plus the two media metrics with pre-initialised labels | `platform/observability/metrics.go`, Phase 5 M12 |
| Prometheus server | **exists**, optional profile, metrics never leave the docker network | `docker-compose.yml` (`--profile monitoring`), `infrastructure/monitoring/prometheus.yml` |
| OpenTelemetry tracing | **partial** — the package exists; end-to-end spans across SSR → API → worker are not proven | `platform/observability/tracing.go` |

### What is missing — the actual work

1. **Frontend logging does not exist at all.** No logger, no error boundary reporting, no
   client-side error transport. Today a browser-side failure is invisible to the operator.
   The product brief §3 already lists a **Sentry adapter** as the intended shape, and
   `AGENTS.md` forbids `fetch` in presentation — so this is an adapter behind a port, not
   `console.error` sprinkled in components.
2. **No log aggregation.** Logs exist as container stdout and are lost on restart. Nothing
   ships or indexes them: no Loki, no Elasticsearch.
3. **No dashboards.** Prometheus stores series with no one reading them; there is no Grafana,
   and no Kibana because there is no Elasticsearch.
4. **SSR is a distinct log source** and is easy to forget: it is a Node process whose failures
   (a 500 in a route lane, a failed API call) are separate from both the browser and the Go API.

### The choice to make when this is scheduled

Two coherent stacks, and mixing them is the expensive mistake:

- **Grafana + Loki + Prometheus** — one UI for metrics and logs, Loki indexes labels rather
  than full text, cheap to run on a single host, and Prometheus is **already here**. Log
  search is weaker than Elasticsearch's.
- **Elasticsearch + Kibana** (+ Grafana for metrics, or Kibana for both) — much stronger
  full-text search and analytics; substantially heavier (JVM, memory, index lifecycle
  management) for a self-hosted single-host deployment.

**Leaning:** Grafana + Loki, because Prometheus is already running and the deployment target
is one host with 4 GB — but this is a decision to take with the phase, not now.

### Why it is not scheduled yet

The roadmap's remaining phases (6 playback, 7 transcript, 8 orders) each add things worth
watching — playback sessions, transcode failures under real load, payment webhooks. Building
dashboards before those exist means instrumenting an empty product and rebuilding them after.

**Trigger to schedule:** the first phase that runs against real users, or the first incident
where a browser-side or SSR failure cannot be explained from container stdout — whichever
comes first.

### Constraints it must respect

- `AGENTS.md`: no secrets in logs; URLs logged without query strings; no `fetch` in
  presentation or routes — the client transport is an adapter behind a port.
- The brief's §17 list of what must never be logged: passwords, session cookies, playback
  tokens, payment secrets, full private transcript documents.
- Metrics stay internal to the docker network unless TLS and auth land first (`S01`).

---

## B-2 — Adapter tests run under a DOM they do not need, which costs header coverage

**Raised:** 2026-09-06 (user pushed back on accepting it), during Phase 5 M14.
**Shape:** small test-infrastructure fix; likely a config change, not a rewrite.
**Status: DONE, 2026-09-07** — task `B02`, evidence `work/archive/report-b02.md`. What follows is
kept as the record of the diagnosis, with its one error corrected below.

### The symptom, corrected 2026-09-07

**This entry was wrong about half of the symptom, and the correction matters.** It claimed
custom headers die "in either direction". Measured with one probe on both environments
(reviewer, then confirmed independently by the implementer's `## Discussion`):

| Environment | request-side `Tus-Resumable` | response-side `Upload-Offset` |
|---|---|---|
| `happy-dom` | `"1.0.0"` — **arrives** | `null` — lost |
| `node` | `"1.0.0"` | `"42"` |

Only the **response** side breaks, which is consistent with happy-dom implementing the CORS
safelist for response headers; `Access-Control-Expose-Headers` did not change it. The
Phase 3 case (`docs/identity/report.md` §I05, cookie forwarding) is a **different
mechanism** — `Cookie` is a forbidden header name that script may not set — and a `node`
environment does not fix it. Filed as one symptom, they were two.

**Why the error was worth correcting rather than deleting:** cited as written, it would have
sent the next header-carrying adapter's request-side proof into slow e2e for no reason.

**What it costs today:** plan §24 routes the tus client's wire proof to e2e (M16) and unit-
tests the header logic as pure functions. That is honest and the coverage is real, but the
adapter layer permanently cannot assert what it actually sends and receives, and the next
header-carrying protocol will pay the same price.

### The leading hypothesis, and why it is cheap

`frontend/rstest.config.ts` sets `testEnvironment: 'happy-dom'` **globally**, while
**zero** tests under `frontend/src/adapters/api/` render anything — they are pure network
adapters that need no DOM. Under a `node` environment they would use Node's real `fetch`
(undici), where MSW round-trips headers normally.

Candidates, cheapest first:
1. **Per-file environment** for the adapter tests (an rstest docblock or equivalent), or a
   second test project scoped to `src/adapters/api/**` with `testEnvironment: 'node'`.
2. Upgrade happy-dom, if the header behaviour is a known fixed defect.
3. Leave as is and keep the e2e-only proof — the current state, and the one to abandon.

**Verify before adopting:** write one throwaway adapter test under `node` that asserts a
request header in a resolver and a response header at the client. If both pass, the fix is
option 1 and the §24 workaround can be narrowed to whatever genuinely needs a browser.

### Trigger

Take it as soon as Phase 5 closes — before Phase 6 adds a player, whose adapter will also
speak in headers (range requests, cache-control, signed URLs). Fixing it after that phase
means retrofitting two adapters instead of one.

### Constraint

Do not weaken what exists: the browser-dependent tests (components, routes) must stay on
happy-dom. This is about scoping the environment to the tests that need it, not replacing it.

### Outcome, 2026-09-07

Taken at its own trigger, before Phase 6's player adapter. `frontend/rstest.config.ts` now
declares two projects: `adapter-node` (`node`, `src/adapters/api/**`, with an
`@opentelemetry/api` alias to its CJS build and `API_BASE_URL` in the project's own `env`)
and `dom` (`happy-dom`, everything else). The constraint above held: nothing that renders
moved off happy-dom.

**Two obstacles this entry had not predicted, both real:**
1. `@opentelemetry/api` fails under `node` — its `module` condition points at `build/esm`,
   whose relative imports are extensionless and therefore bundler-only.
2. `apiOrigin()` takes its SSR branch under `node` and needs `API_BASE_URL`.

**The payoff was collected, not just unlocked:** the tus protocol is now proven on the wire
in the adapter layer — PATCH/HEAD request headers read from an MSW resolver, offset
progression driven only by the response's `Upload-Offset`, resume metadata from `Upload-Length`,
and refusal of a response whose `Tus-Resumable` is missing or foreign. Both new assertions
were shown red before green. 330 → 334 tests; `adapter-node` 11 files / 97, `dom` 27 / 237.

**Deliberately not done here:** removing the `typeof window` branch from `apiOrigin()`. It
was considered and measured — orval always emits `options?: Parameters<typeof mutator>[1]`,
so a mutator requiring `origin: string` fails to typecheck at every generated call site, and
only `origin?: string` with a runtime throw compiles. That trades a compile-time guarantee
for a runtime one. It is also unnecessary: after this split both branches of `apiOrigin()`
execute — the SSR branch under `adapter-node`, the browser branch under the `dom` tests that
still call it. Previously only the browser branch ran anywhere.

---

## B-3 — Контракт дозволяє вчетверо більше уроків, ніж приймає домен

**Знайдено:** 2026-09-07, лінзою `plan-critique` під час GATE 2 Phase 6; підтверджено
рецензентом власноруч. **Фаза 6 цей дефект не створила — лише виявила.**

**Status: DONE з датованою перекласифікацією, 2026-09-17** — доказ `work/archive/report-b3.md`; напрям
(«домен правий: 120 на курс») ухвалила людина. Запис нижче лишається як був, бо його §«Що
зробити» **виявилось нездійсненним у варіанті 1**, і це варте того, щоб пережити фазу:

- **Закрито:** дві конкуруючі істини. Число живе в домені (`MaximumLessons = 120`,
  кумулятивно), контракт його **не повторює** як самостійну істину.
- **Лишається, і причина не недбалість, а межа формату:** схема лишається **пермісивнішою за
  домен**. `maxItems` на модуль правдивий сам собою (один модуль справді може нести всі 120),
  але суму по сусідніх масивах JSON Schema виразити не може, тож курс на 121 урок «60+61»
  контрактно валідний і доменно відкинутий. Варіант 1 цього запису («додається явне обмеження
  курсу») в OpenAPI **не виражається** — це не вибір, який не зробили, а вимога, яку формат не
  приймає.
- **Компенсація, уся виміряна:** курсова стеля в описі схеми (і, через `make generate`, у
  doc-коменті генерованого `adminCourseSaveRequest.ts`); відмова дістала власний код
  `lesson_limit` замість голого `invalid_request`, тож клієнт називає причину; межа пінена в
  обидва боки — `assertAdminHTTPContract(..., validRequest: true)` доводить, що контракт
  «60+61» **приймає**, а `assertAdminRequestSchema(..., false)` доводить, що 121 одним модулем
  схема таки ловить.
- **Побічно відкрито доказ, а не лише прибрано ризик:** фікстура «межа схеми», яку Phase 6 не
  змогла побудувати (рядок `B19`/`B15` із неіснуючою стелею), тепер існує в обох напрямах.

Тобто дефект не «виправлений як записаний», а розділений на закриту частину й названий
залишок. Читати це як «контракт вирівняний» — помилка: він і далі оголошує досяжним те, що
домен відкидає, і саме тому відмова тепер має ім'я.

### Розбіжність, виміряна

| Джерело | Що дозволяє |
|---|---|
| `backend/openapi/openapi.yaml:743` | `maxItems: 120` **на модуль** (координати оновлено 2026-09-17: були `:633`) |
| `backend/openapi/openapi.yaml:761` | `maxItems: 40` модулів (були `:653`) |
| Та сама пара ще тричі | `:798`/`:818`, `:913`/`:932`, плюс `:953` — стеля продубльована по схемах, тож «одне число в одному місці» коштує більше, ніж обіцяв цей запис |
| **Разом за контрактом** | **4800 уроків на курс** |
| `backend/internal/course/domain/authoring.go:174-185` | `lessonCount` рахується **кумулятивно по всіх модулях**, `MaximumLessons = 120` |
| **Разом за доменом** | **120 уроків на курс** |

Тобто контракт обіцяє клієнтові стелю, якої домен ніколи не пропустить, а оголошений
`maxItems: 120` на модуль **не обмежує нічого понад те**, що вже обмежив домен: курс із
двох модулів по 100 уроків контрактно валідний і доменно відкинутий.

### Чому це той самий клас, що й «CHECK, який нічого не обмежував»

Phase 5 знайшла обмеження, яке було синтаксично правильним і семантично порожнім
(`array_length('{}',1)` — NULL, а NULL CHECK **виконано**; `docs/media/report.md` §3).
Тут те саме: обмеження існує, читається як захист, і не захищає. Різниця лише в тому, що
цього разу два джерела правди суперечать одне одному, а не одне джерело бреше саме собі.

### Що коштує сьогодні

Клієнт, згенерований Orval, типізує межу як 120 на модуль. Форма адмінки може дозволити
скласти курс, який сервер відкине з `invalid_request`, **не назвавши причини** — з погляду
клієнта це «валідні дані, які раптом відхилили». Тестова фікстура «межа схеми» побудувати
неможливо: Phase 6 спробувала й отримала рядок B15 із неіснуючою стелею.

### Що зробити

Вирівняти в один бік — і рішення тут продуктове, не технічне:

1. **Домен правий, контракт занижуємо**: `maxItems` на модуль лишається 120, але додається
   явне обмеження курсу; або модульну стелю знижують так, щоб добуток не перевищував 120.
2. **Контракт правий, домен піднімаємо**: `MaximumLessons` стає стелею **на модуль**, а не
   кумулятивною. Це змінює поведінку наявних курсів і потребує окремого зважування.

**Не робити «і те, і те».** Дві стелі, що не збігаються, — це і є теперішній стан.

### Тригер

Наступна робота, що торкається авторингу курсів або генерованого клієнта. Окремо: якщо
Phase 6 (`docs/playback/plan.md` B19) виявить, що фікстура межі потрібна раніше.

---

## B-4 — Дефолтний таймаут rstest тісний для 364 тестів на завантаженій машині

**Знайдено:** 2026-09-08, **двома незалежними сесіями** під час Phase 6 T7.

### Симптом, однаковий в обох

| Сесія | Що бачила |
|---|---|
| ведуча фазу | 1–5 падінь `CourseEditorView.test.tsx`, усі — `timed out in 5000ms`; **не відтворюється** ізольовано |
| рецензент T7 | 4 падіння того самого файлу, три з них — ті самі 5000 мс; повтор дав **364/364** |

Найпоказовіше: один із тестів устигав **162 асерти** до зупинки — тобто працював, а не завис.
Прицільний замір ведучої сесії: той самий файл дає **12/12 PASS** із `--testTimeout=30000`.

Найважчий фігурант — `B09 keeps a maximal module list keyboard reachable and free of axe
violations`: максимальний список модулів плюс повний прогін axe.

### Причина, виміряна, а не припущена

Рецензент заміряв хост під час своїх падінь: `load average 47`, і найбільші споживачі —
**сторонні** процеси (cassiopeia/ttscgraph 48%, elasticsearch 27%, WindowServer 64%, Chrome).
Контейнери самого проєкту в ту мить простоювали.

Отже це **не** конкуренція наших власних гейтів і **не** дефект тестів: 5 секунд на 364 тести
тісні, коли машина віддає частку CPU.

### Чому не виправлено одразу

Підняти дефолт означало б **послабити мірник** назавжди заради тимчасового зовнішнього
навантаження на ноуті розробника. На розвантаженій машині 5 с тримаються — обидві сесії
бачили 364/364.

### Варіанти, коли візьмемо

1. Підняти дефолтний `testTimeout` (наприклад, до 15 с) — просто, але знижує чутливість до
   справжніх зависань у всіх 364 тестах.
2. Зрізати паралелізм rstest — зберігає чутливість, подовжує прогін.
3. Підняти таймаут **точково** для важких a11y-тестів, лишивши дефолт як є — найвужча правка,
   але вимагає називати «важкі» вручну.

### Тригер

**Коли впаде в CI**, а не лише на завантаженому ноуті. Доти це шум середовища з відомою
причиною, і його не можна плутати з регресією: якщо падає «щоразу інший тест» і не
відтворюється в ізоляції — дивись на `load average`, перш ніж шукати дефект.

### Outcome 2026-10-20 — закрито (рішення власника: підняти дефолт)

Власник (2026-10-20) обрав варіант 1. `testTimeout: 30000` на верхньому рівні `frontend/rstest.config.ts`
(успадковують обидва проєкти). Значення — виміряний у цьому записі безпечний поріг (12/12 і 364/364).

**Знята пересторога «послаблює мірник» (механізм):** працюючий тест завершується далеко під 5 с навіть
без навантаження (один устигав 162 асерти до 5-с дедлайну), тож вищий дефолт не робить жоден робочий
тест повільнішим — він лише дає запас під зовнішнім навантаженням. Чутливість до **справжнього**
зависання не зникає: нескінченний await усе одно впирається в будь-який скінченний поріг, різниця —
30 с замість 5 с до червоного, і гейт його все одно ловить. Коштом є тільки повільніша діагностика
справжнього hang — прийнятно проти флаку «щоразу інший тест».

---

## B-5 — Черга asynq переживає прогони й отруює наступний

**Знайдено:** 2026-09-08 сесією-рецензентом, за симптомом ведучої сесії.

**Status: DONE, 2026-09-17** — доказ `work/archive/report-b5.md`. Причину прибрано з ланцюга гейта:
ціль `drain-queue` у `Makefile` стоїть між `up` і `seed-identities`, тобто **раніше за `smoke`** —
саме та дірка, яку цей запис називав («перед `smoke` дренаж не виконується ніколи»). Ціль друкує
`before=N after=M` і валить крок, якщо `after != 0`; червоне доведено відтворенням історичного
дефекту (префіксна форма замість `export` → `NOAUTH`, `before=3 after=3`, guard exit 1), тож
«дренаж, що мовчки не дренує» більше не може виглядати зеленим. Патерн лишився
`asynq:{media}:*`; реєстраційні ключі вціліли.

Дренажі в e2e лишаються — вони чистять те, що насмітив **сам** прогін; крок у `Makefile` прибирає
залишок **попереднього**. Варіант `tmpfs` для `redis-data` відкинуто з названою причиною: `make
up` перестворює лише змінені контейнери, тож на теплому стеку черга переживає `up` навіть на
`tmpfs`, а спостережений шлях відмови — саме `up` **без** `down`.

**Передумова перенесена в точку використання:** дренаж не розрізняє свою задачу від чужої, і
безпечний лише поки важкі гейти жене одна сесія за раз. Це тепер стоїть над цілю в `Makefile`
разом із вказівкою прибрати крок **першим**, якщо зʼявиться паралельний прогін — раніше вона
жила лише тут і в `media.setup.ts`.

### Механізм, а не збіг

Redis піднято з `appendonly yes` на **постійному** томі `redis-data`
(`docker-compose.yml:31`), а чергу asynq між прогонами **не чистить ніхто**: ані setup,
ані teardown, ані Makefile.

Отже прогін, обірваний посеред транскодування, лишає задачу в Redis. Наступний `make up`
піднімає воркера, той її **відновлює** (`Recover` перед `server.Start`, Phase 5) і зʼїдає
свій бюджет 2 CPU / 2 GB, поки транскод нового прогону стоїть за нею в черзі.

**Два спостережені наслідки:**
1. `smoke` падає на B19: `courses_transcoding_job_duration_seconds_count{outcome="ready"}`
   уже **не нульова**, бо воркер устиг обробити чужу задачу до перевірки.
2. Пʼятихвилинна фікстура не встигає у власний таймаут — червоне з причини, не повʼязаної
   з кодом.

### Що зроблено зараз

Дренаж `asynq:*` у `frontend/e2e/media.setup.ts` перед першою фікстурою.

**Передумова, за якої це безпечно, записана в коді:** важкі гейти запускає рівно **одна
сесія за раз** (домовленість між сесіями, 2026-09-08). Конкурентного прогону, чию легітимну
задачу можна знести, немає. **Якщо це зміниться — дренаж прибрати першим.**

### Що лишається відкритим

Дренаж лікує симптом у межах e2e, але **не** прибирає причину: постійний том черги плюс
відсутність будь-якого дренажу в `make up`/`down`. Варіанти на потім:
- `tmpfs` для `redis-data` у тестовому профілі, щоб черга не переживала стек;
- крок дренажу в самому Makefile перед `smoke`;
- лишити як є, якщо колись зʼявиться паралельний прогін — тоді дренаж стане небезпечним.

**Тригер:** поява другої сесії чи CI-джоба, що ганяє гейти одночасно.

**Друга половина закрита 2026-09-11.** Дренаж у setup не міг сягнути задач, які створює **сам
прогін**: блок `playback` вантажить відео вже після фікстур, і його задачі лишаються в черзі.
Через кілька хвилин проєкт `recovery` перезапускає стек, воркер одразу відновлює ті задачі,
ffmpeg забирає свій бюджет 2 CPU — і `/readyz` перестає відповідати в межах трисекундного
таймауту healthcheck (`interval 5s`, `retries 15`). Compose оголошує воркера `unhealthy`,
`docker compose up -d --wait` падає, і `recovery` червоніє **без жодного дефекту в продукті**.

Заміряно: після блоку playback у черзі лишалось **12** ключів.

Дренаж додано у `playback.teardown.ts` — чистить той блок, який насмітив. Це те саме правило, що
й з `lesson_progress`: **чистка на вході захищає того, хто чистить; від того, хто чистить,
захищає лише чистка на виході.**

**Межа початкового лікування, знайдена 2026-09-10.** Дренаж живе у `frontend/e2e/media.setup.ts`,
тобто **всередині e2e**, а ланцюг `make check` — `… up seed-identities seed smoke e2e`. Отже перед
`smoke` дренаж не виконується **ніколи**. Задача, що вціліла в Redis, буде оброблена воркером до
того, як `smoke` прочитає метрики. Це стосується саме черги; **лічильники в процесі воркера — це
окрема хвороба, див. B-10**.

### Поправка 2026-09-08 — дренаж порушував правило, яке репо саме й записало

Рецензент помітив, що перша редакція дренажу передавала пароль **аргументом**
(`redis-cli -a "$REDIS_PASSWORD"`), тоді як `docker-compose.yml:29-30` забороняє це прямим
текстом для самого сервера: пароль іде файлом, «never argv, where `ps` and `docker top` would
expose it». А `--no-auth-warning` глушив рівно те попередження, яким redis-cli про це каже.

Причина була технічна: префікс `VAR=cmd` ставить змінну лише для **першої** команди, тож другий
`redis-cli` у пайпі її не успадковував. Лікується `export` на початку sh-рядка.

**Заміряно, а не виведено** (Redis піднято, 2026-09-08):
- з `export` — ключ `asynq:{media}:__drain_probe__` створено, після пайпа лишилось **0**;
- з префіксною формою — друга половина пайпа дає `NOAUTH Authentication required.`, ключ
  **виживає**. Тобто `-a` у першій редакції був **несучим**: прибрати його без `export` означало
  б мовчазно неробочий дренаж, який виглядає зеленим.

Заодно звужено патерн до `asynq:{media}:*` — рівно черга воркера
(`Queues: {"media": 1}`, `cmd/media-worker/main.go:70`). **Замір показав, чому це не косметика:**
єдині asynq-ключі, наявні в Redis на момент перевірки, — `asynq:servers`,
`asynq:servers:{...}`, `asynq:workers`. Ширший `asynq:*` зніс би **все наявне і жодної задачі**.

### Поправка до самого доказу «черга порожня» — 2026-09-08

Факт, на якому стоїть уся ця історія («рецензент перевірив, черга порожня»), **тримався
випадково**, і рецензент сам це знайшов.

Перевірка була двома способами: пʼять лічильників `pending/active/retry/archived/scheduled` —
і скан. Лічильники міряли **`asynq:{default}:*`**. Черги з таким імʼям у цьому репо немає:
воркер слухає `media` (`cmd/media-worker/main.go:70`).

**Перевірено власним заміром (Redis піднято):**

```
asynq:{default}:pending    EXISTS=0  LLEN=0  ZCARD=0
asynq:{default}:active     EXISTS=0  LLEN=0  ZCARD=0   ← і так усі пʼять
```

`EXISTS=0` — ключів **не існує**, а `LLEN`/`ZCARD` неіснуючого ключа за визначенням нуль. Тобто
ті пʼять чисел дали б нуль **на вщент забитій черзі `media`** так само впевнено.

Висновок «порожня» вцілів лише тому, що в тому ж виклику робився ще й `--scan --pattern
"asynq*"`, і він повернув три ключі реєстрації без жодної задачі. **Правильним був скан, а не
лічильники.**

**Практичний наслідок:** тих пʼяти команд не переюзовувати — вони повертають нуль незалежно від
стану. Стан черги міряти сканом по `asynq:{media}:*`.

**Чому це варте запису.** Дві зелені брехні мало не наклались одна на одну: дренаж, який без
`export` мовчки не дренує, і спосіб перевірки, який показує «порожньо» незалежно від того, чи
дренаж спрацював. Кожна поодинці виглядає нешкідливою; разом вони дають повністю зелений гейт,
що не міряє нічого.

---

## B-6 — e2e не звірений із контрактом: кожен `as { … }` — незалежна гіпотеза

**Знайдено:** 2026-09-08, побічно з пастки імен у T8 (heartbeat приймає `positionSeconds`,
читальна модель віддає `position`).

**Status: DONE, 2026-09-17** — доказ `work/archive/report-b6.md`; узято перед фазою `7a` за правилом
§24 (`work/PRODUCT-BRIEF.md:1015`). `frontend/e2e/contract.ts` дає `contractField<Model>`, ключ
якого мусить бути полем генерованої моделі; анотовано 23 виклики у трьох spec-файлах, 18
унікальних пар. Червоний доказ — підміна пари з T8 в обидва боки дає два `TS2345`. Зонди
інструментації свідомо лишені на сирому `Reflect.get`. Далі — опис дефекту, як він був.

### Що не так

`frontend/e2e/` читає відповіді як `unknown` і звужує **рукописними** асершенами. Ім'я поля там
не звіряється ні з чим, тож
`(body as { positionSeconds?: number })` на `LearningLesson` **типо-легальне**: свита
збереться, побіжить і збреше — «прогрес не зберігся» замість «я взяв не те поле».

Це не гіпотетика: рівно ця описка була найімовірнішим сценарієм у T8, і врятував нас від неї
рецензент, а не інструмент.

### Чому це дешевше, ніж звучить

Мосту будувати не треба, все вже стоїть:

| Що | Де | Перевірено |
|---|---|---|
| Генеровані моделі | `frontend/src/adapters/api/generated/models/` | `learningLesson.ts` має `position: number` |
| e2e у тій самій програмі TS | `frontend/tsconfig.json:22` — `"include": ["src", "e2e", "*.ts"]` | так |
| Бар'єр архітектури на такий імпорт | `scripts/check-architecture.mjs` слова `e2e` не містить | перешкоди немає |

Робота — **анотувати ціль звуження генерованою моделлю**, а не будувати інфраструктуру. Одна
анотація перетворює хибне ім'я поля з мовчазного `undefined` на помилку `typecheck`.

### Що це купує — і чого НЕ купує

**Ловить:** описку в тесті — рівно наш клас дефекту.

**НЕ ловить:** розходження контракту з реальним сервером. Моделі походять з `openapi.yaml`, і
`make check-generated` стереже їхню відповідність **специфікації**. Якщо специфікація бреше про
сервер, застосунок і e2e брехатимуть **разом і однаково**. Це закривають контрактні тести на
HTTP, які `AGENTS.md` вимагає окремо.

Два різні дефекти, два різні докази — те саме правило, за яким ми щойно розділили P02 і R37.

**Не тягнути в T8.** Пункт ширший за задачу; розширювати її під нього означало б приліплювати
друге твердження до чужого доказу.

---

## B-7 — стан `playing` не стежить за відеоелементом, і каденція це переживає

**Status: DONE, 2026-09-09** — таск `T10`, доказ `work/archive/report-t10.md:240,261`. `onPause` і
`onPlaying` беруть стан із подій відеоелемента (`LessonPlayer.tsx:426,437`), і рядок **R39**
доводить, що пауза поза кнопкою зупиняє поновлення володіння **і** пише прогрес. Позначку
додано 2026-09-12: `work/tasks.md:163` заявляв закриття, а цей файл його не ніс, і два
реєстри розходились.

**Знайдено:** 2026-09-08, побічно з рецензії T8 (обговорення, як безпечно поставити паузу в e2e).

### Факт

`LessonPlayer.tsx` не має **жодного** слухача подій відео на паузу: ані `onPause`, ані
`addEventListener('pause')` — перевірено пошуком, збігів нуль. `playing` змінюється рівно в
чотирьох місцях: `useState(false)` (:65), скидання при зміні уроку (:74), `onEnded` (:331) і дві
гілки кнопки (:342, :351).

Отже **пауза поза кнопкою не гасить каденцію heartbeat**: ефект із гейтом `playing`
(`LessonPlayer.tsx:251`) далі вважає, що йде відтворення, і продовжує слати heartbeat-и.

### Чому це не лише про тести

У e2e наслідок обмежений — тому в T8 пауза йде **кнопкою**, і це записано як умова доказу.

Але heartbeat **тримає володіння живим**. Студент, чиє відео стало не через кнопку, продовжує
**утримувати слот акаунт-вайд**, не дивлячись нічого. За моделлю витіснення (ліміт 1 на акаунт)
це означає, що інший пристрій буде витіснений на користь сесії, яка стоїть на паузі.

### Наскільки це досяжно сьогодні

`<video>` **не має атрибута `controls`** (`:322-333`), тож рідного UI паузи в елемента немає, і
звичайним кліком користувач туди не потрапляє. Лишаються шляхи повз розмітку: медіаклавіші
клавіатури / Media Session, контекстне меню відео в Chrome, picture-in-picture.

**Не заміряно, які з них реально ставлять паузу в цьому плеєрі** — це висновок із коду.

### Що з цим робити

Лікується одним слухачем: `onPause={() => setPlaying(false)}` (і симетрично `onPlay`), після чого
стан стежить за елементом незалежно від того, хто його зупинив.

**Не в T8** — це поведінка володіння й витіснення, тобто **T10**. Тут лише зафіксовано, поки
видно.

---

## B-8 — стан переживає тест, який його створив: три прояви за годину

**Знайдено:** 2026-09-08, під час гейту T8. Рецензент назвав спільний корінь.

### Прояв 1 — збережена позиція робить урок спільним станом

Плеєр відновлює `lesson.position` на `onReady` (`LessonPlayer.tsx:186-192`): ставить
`video.currentTime = resumeAt` і **лише тоді** `play()`. Отже **кожен тест, який сікає вперед,
пересуває стартову точку всіх наступних на тому ж уроці**.

На уроці 1 сікають чотири тести: **80** (P01), **298** (B40), **290** (P02), **120** (R04).
Дві останні позиції летальні: урок, відновлений за дві секунди до кінця, не резолвить `play()`,
`setPlaying(true)` не виконується, і викликач вичікує весь свій таймаут (заміряно: B39, 420 с,
**жодного heartbeat** у трасі за сім хвилин).

**Порядок ліками бути не може — обидві сторони псують одна одну:**

| Порядок | Що ламається |
|---|---|
| Блок T8 перший | B39 успадковує **290** від P02 → повисає |
| Блок T8 останній | P02 успадковує **298** від B40 → повисає |

І гірше: у другому випадку, якби воно якось пройшло, асерт «урок 1 переглянуто» задовольнявся б
роботою **B40**, а не ланцюгом, який P02 існує довести. Головний доказ став би порожнім — те
саме, від чого ми страхувались, додаючи `not.toContainText` на решту рядків.

**Зроблено:** блок T8 має `beforeAll`, який чистить `lesson_progress` курсу через наявний
psql-механізм. Блок стартує з **відомого** стану, а не з очікуваного.

**Свита мала неявний контракт порядку через збережений прогрес, і його ніде не було записано.**

### Друга поправка — 2026-09-09: область хука ≠ його частота

Скидання лишалось **вхідним для одного блоку**, а чотири тести T7 (P01, P08, B39, B40) ділили
урок 1 **без жодного скидання**: 80 → ≈110 → ≈140, і B40 приходив із двома хвилинами залишку.
Коли транскод заміни, на який він чекає, став тривати довше за той залишок, відео добігало кінця
до `setPosition(298)`, і `oldResponses` лишався порожнім.

**Пастка, у яку я мало не вступив:** «скинути на початку всього блоку» через `beforeAll` було б
**порожньою операцією** — воно виконалось би один раз перед P01, на вже чистому стані.

> `beforeAll` прибирає вплив того, що було **до блоку**. `beforeEach` — вплив **сусідів
> усередині блоку**. Тут хвороба друга.

Це вчорашня помилка в новій обгортці: тоді сплутано **позицію** хука з його **областю**, тепер —
область із **частотою**. Обидві дає та сама модель «хук робить те, що я мав на увазі».

**Зроблено:** `test.beforeEach(resetLessonProgress)` на рівні `as a student`; хуки вкладеного
блоку прибрані як зайві. Перевірено, що це нікого не ламає — усі чотири тести від успадкованої
позиції лише **страждали**: P01 сікає на 80 і з нуля працює так само, P08 отримує 300 с замість
190, B39 байдужий, B40 отримує 300 замість 140.

**Точкове скидання перед B40 було б лікуванням симптому** — троє попередніх лишились би
зчепленими.

**Поправка того ж дня — чистки на вході НЕ досить.** Першою редакцією стояв лише `beforeAll`, і
рецензент спинив прогін до того, як він дійшов до міни: **R04 (`:650`) відкриває той самий урок 1
і успадкував би 290 від P02** — рівно стан, на якому повис B39. Правило, яке з цього виходить:

> **Чистка на вході захищає того, хто чистить; від того, хто чистить, захищає лише чистка на
> виході.**

Ми поставили перше й вирішили, що звʼязок розірвано, — а розірвали половину. Тепер скидання
стоїть і в `beforeAll`, і в `afterAll`, і порядок остаточно перестає бути несучим.

**P11 не зачеплений, і причина перевірена, а не припущена:** `lesson_progress` має первинний ключ
`(user_id, lesson_id)` (`00006_playback.sql:13`), тож прогрес привʼязаний до користувача. P11
ходить адміном, у якого свого запису на урок 1 немає. Зачеплений саме R04 — бо він студент, той
самий, що й P02.

### Прояв 2 — одноразові користувачі не відновлюються, і це блокує все нижче за течією

`delete-me@courseforge.local` споживається тестом `admin-users P08`; наступний прогін падає з
`is not an active disposable fixture; run make seed-identities`. Обірваний або повторний прогін
лишає фікстуру спожитою.

**Це не флейк, а детермінований блокатор — і ціна вища, ніж один червоний тест.** `chromium` є
**висхідною залежністю** ланцюга `media → playback-profile → playback` (`playwright.config.ts`),
тож спожита фікстура зупиняє **весь** блок плейбека: він навіть не запускається. Заміряно
2026-09-08: цілеспрямований `--project=playback` двічі не дійшов до жодного playback-тесту, бо
`admin-users P08` падав першим.

**Операційне правило, поки не полагоджено:** `make seed-identities` **перед кожним** окремим
прогоном e2e. У `make check` це зайве — ланцюг `… up seed-identities seed smoke e2e` (Makefile:54)
робить це сам; хвороба стосується лише окремого `make e2e` / `npx playwright test`.

**Справжнє лікування** (не в T8): свита має відновлювати одноразові фікстури сама — у власному
`setup`-проєкті, а не покладатись на те, що хтось згадає команду.

### Прояв 3 — черга asynq переживає прогони

Див. **B-5**: постійний том `redis-data`, `appendonly yes`, дренажу в `make up`/`down` немає.

### Спільний корінь

Три різні підсистеми — плеєр, ідентичності, черга — і **одна хвороба**: стан переживає тест,
який його створив, а наступний прогін мовчки успадковує чужі наслідки. Щоразу симптом читався
як дефект найсвіжішої правки.

**Що з цим робити системно** (не в T8): кожен блок, що пише стан, або чистить його за собою,
або стартує з явного скидання. Зараз це зроблено точково для `lesson_progress`; `delete-me` і
черга лишаються на ручному `make seed-identities` та дренажі в setup.

**Тригер перегляду:** T9/T10 додадуть нових писачів у ті самі уроки — там воно вибухне знову, і
тоді дешевше буде дати блоку T8 власний курс, ніж чистити прогрес.


---

## B-9 — три неоголошені контракти за один день

Кожен із трьох коштував червоного прогону, і кожен читався як дефект найсвіжішої правки.

| Що ми читали | Про що треба було спитати | Відповідь, яка чекала |
|---|---|---|
| «де в файлі стоїть тест» | **з якою позицією в уроці 1 фактично стартує B39?** | 290, лишені P02 |
| «скільки воркерів у конфізі» | **скільки браузерів працює одночасно?** | половина ядер машини |
| «де в файлі стоїть хук» | **у який момент фактично виконується чистка?** | один раз перед P01 |

Ліва колонка — питання про **позицію в коді**: на них відповідає читання, і читання успадковує
ту модель, яку читач уже має. Права — питання про **стан у момент X**: на них читанням не
відповісти, треба простежити виконання, і саме простежування модель перевіряє.

Рецензент побачив область хука не тому, що читав уважніше, а тому, що ставив праве питання: він
питав не «де стоїть хук», а «з якого стану P02 стартує» — і на це неможливо відповісти, не
зʼясувавши, коли чистка виконується.

**Третій — найповчальніший, бо його не помітили обидві сесії.** Я поставив
`beforeAll`/`afterAll` між B40 і P02 і прочитав це як «чистка між блоками»; рецензент,
дивлячись у діф, прочитав так само. Але Playwright прив'язує хуки до **describe, що їх
охоплює**, а не до позиції в тексті. Обидва хуки належали `as a student`, тож `beforeAll`
відпрацював один раз перед P01 на порожньому стані, а `afterAll` — після R04.

**Наслідок сходився з симптомом до дрібниць:** до P02 урок 1 приходив таким, яким його лишив
B40 — завершеним, на 298. `openAndPlay` **проходив** (мітка «Пауза» встигала з'явитись), але
через дві секунди `onEnded` гасив `playing`, cleanup знімав таймер раніше за перший heartbeat,
і `await firstHeartbeat` вичікував усі 420 с. А якби один запит і проскочив, він ніс би 298 —
проти предиката `< 285`. Порожня траса була правдою, а не неповнотою.

**Виправлено:** P02/P03 і R37 загорнуті у власний вкладений `describe`, хуки — всередині нього.
Serial успадковується від `configure`. Заодно закрито четверту міну, яку ми не бачили: у старій
розкладці `afterAll` стояв **після** R04, тож від 290, лишених P02, R04 не був захищений нічим.

### Межа, серйозніша за «неоголошений контракт»

Спокусливо сказати «інформації не було в тексті». **Це неправда, і поправка важлива.** Межі
`describe` були на рядках 360 і 570 — обидва видно, відстань між ними теж. Інформація була
присутня повністю.

Не спрацювало інше: обидві сесії приклали **одну й ту саму хибну модель** — «позиція = область
дії». Автор приклав її, пишучи; рецензент — читаючи той самий текст.

**Свіжий погляд рятує від помилок, специфічних для автора. Він не рятує від помилки, спільної
для обох**: однакова хибна модель дає однакове прочитання. Незалежність рецензента тут безсила —
допомагає лише звірка з семантикою фреймворку.

**Практичний детектор** — питати про **стан у момент X**, а не про позицію в коді. Праву колонку
таблиці вище не можна закрити читанням; вона примусово веде до простежування, а простежування
перевіряє модель замість того, щоб її успадкувати.


---

## B-10 — `smoke` вимагає нульових лічильників, а обіцяє перевірити реєстрацію серій

**Знайдено:** 2026-09-08, коли `make check` упав на `smoke` після дня прогонів.

**Status: DONE — виправлено 2026-09-09, позначено 2026-09-17.** Доказ `work/archive/report-b10.md`.
Відкритою була не робота, а **реєстр**: ліки з §Ліки уже стояли в
`infrastructure/tests/smoke.sh:61-68` (вісім серій у формі `…{reason="probe_failed"} ` з
пробілом, не `… 0`). Приїхали вони комітом **`8cec836`**, чия тема — *«the heredoc terminator
has to sit in column zero»*, тобто **інший** дефект. Правка йшла попутно, під чужою назвою, тож
ніщо не вказувало назад на цей запис. Механізм варто запам'ятати: коміт, що лікує два незалежні
дефекти під темою одного, робить другий невидимим для пошуку за назвою.

Поділ доказів лишається як є, бо він правильний: **реєстрацію на старті** доводить
`TestMetricsArePreinitialised` (`NewMetrics()` свіжий за побудовою, тож `0` там — факт, не
припущення про середовище), а **піднятий воркер** перевіряє `smoke` по **наявності**, бо на
теплому воркері значення ненульові.

Понад запис закрито дірку, якою можна було повернутись у ту саму сліпоту:
`TestBoundedFailureReasonsAreRegisteredEverywhere`
(`backend/platform/architecture/preinitialised_series_test.go`) пінить рівними три копії переліку
причин — домен (`media/domain/asset.go:34-38`), преторкання в `NewMetrics`
(`observability/metrics.go:43`) і асерти юніт-тесту. Червоне перевірено шостою причиною
`audio_missing`. Четверта копія — у `smoke.sh` — свідомо поза піном: Go-контейнер монтує лише
`backend/` (та сама межа, що розрізала `R30`), і наслідок названий у звіті.

**Це НЕ B-5, і сплутати їх легко.** Дренаж чистить **чергу в Redis**; лічильники живуть **у
процесі воркера**. Жоден дренаж не міг і не може їх обнулити — тільки перезапуск процесу. Дві
незалежні хвороби отримали одну мітку, і під нею друга загубилась би.

### Що ассерт обіцяє і що робить

`infrastructure/tests/smoke.sh:56-67` перебирає **вісім** серій — три `outcome` і пʼять
`reason` — і кожну шукає рядком, що закінчується на ` 0`. Вичерпність переліку плюс текст
помилки «Missing zero-value worker metric» видають намір: перевірити, що воркер **реєструє всі
комбінації міток на старті**, а не створює їх ліниво при першій події.

Інваріант цінний: лічильник, який зʼявляється лише після першої невдачі, робить дашборд і алерт
сліпими рівно до моменту, коли вони потрібні.

**Але записаний він через значення, а не через наявність.** `includes('… {outcome="ready"} 0')`
істинне лише поки лічильник нульовий — тобто **лише на щойно піднятому воркері**. Реєстрація
серії й нульовість — різні твердження, і друге паразитує на першому.

### Прихована передумова

`make up` перестворює лише **змінені** контейнери, тож локально `media-worker` живе годинами з
накопиченими лічильниками. На CI стек завжди свіжий — передумова виконується сама, і **дефект
невидимий там, де його вводять**.

### Ліки

Звіряти імʼя з мітками **без хвостового значення**: `'…{outcome="ready"} '` замість `… 0`.
Намір зберігається повністю, прихована передумова зникає, `make check` стає локально придатним.
Нічого не втрачається: «воркер уже щось обробив» не порушує жодного інваріанта, а `e2e` після
`smoke` однаково робить лічильники ненульовими.

**Обхід `make down` перед `make check` — не ліки:** він **відтворює** передумову замість того,
щоб її прибрати. Наступний локальний прогін без `down` упаде знову і знову виглядатиме дефектом
найсвіжішої правки.


---

## B-11 — вміст, чутливий до пробілів, усередині мови з вільним відступом

**Знайдено:** 2026-09-09, після того як власна правка зламала власний же хук.

### Що сталось

Термінатор shell-heredoc мусить стояти в **колонці нуль**. Але живе він усередині шаблонного
рядка TypeScript, де відступ вільний. Загортаючи блок у вкладений `describe`, я додав два пробіли
до кожного рядка — і термінатор зʼїхав. Оболонка документа не закрила, psql отримав слово `SQL`
як інструкцію.

**Доведено історією, а не памʼяттю:**

```
11a4370  562: SQL`,      ← колонка 0
07b2167  568:   SQL`,    ← два пробіли (коміт, що додав вкладеність)
```

**Моє перше пояснення було хибним.** Я написав, що правка «зробила видимою» стару помилку. Ні:
вона її **внесла**. Контраргумент, що це доводить, — арифметичний: падіння `beforeAll` завалило б
увесь serial-блок, а в тому прогоні 39 тестів пройшли. Тихих падінь `beforeAll` не буває.

### Чому це клас, а не випадок

Нічого не сторожує цей відступ: `lint`, `typecheck` і `format` бачать лише символи в рядку. Міна
**зводиться заново** щоразу, коли хтось додасть рівень `describe`, пересуне блок або
відрефакторить обгортку. Я рухав ці блоки двічі за день — і другого разу вона спрацювала.

### Ліки

Обидва heredoc винесені в константи **модульного рівня**
(`deleteCourseStatement`, `resetLessonProgressStatement`). Верхній рівень неможливо поглибити
вкладеністю тестів, тож майбутнє загортання термінатора не зрушить.

Друге місце — видалення курсу (`playback-cleanup`) — вціліло лише тому, що його ніхто не рухав
із часу рецензії T7. Тепер захищені обидва.

### Правило

**Якщо вміст чутливий до пробілів, він не повинен залежати від відступу коду, який його
оточує.** Або на модульному рівні, або в окремому файлі — але не всередині вкладеної структури,
глибину якої хтось колись змінить.


---

## B-12 — правило, що лишається правилом, порушується; вбудоване в шлях — ні

**Знайдено:** 2026-09-09, після того як автор двічі за годину обійшов власне правило з B-8.

### Спостереження

Правило «перед окремим e2e зроби `up` і `seed-identities`» було **записане** в B-8 — і двічі
порушене тим самим, хто його записав, у межах години. Не через недбалість: дешевий шлях
(`make e2e`) **існує й неправильний**, а правильний доводиться складати руками щоразу.

**Дисципліна тут програє не людині, а ергономіці: інструмент пропонує помилковий варіант
першим.**

### Зроблено

```make
e2e-local: up seed-identities seed e2e
```

`make e2e` лишається низькорівневою ціллю, яку компонує `check`; але жоден локальний прогін
більше не почнеться з непіднятого стеку чи спожитого `delete-me`. Дешевий шлях став правильним
**за побудовою**, а не за памʼяттю.

### Закономірність, а не збіг

Три випадки за один день, і всі три однакові за формою — контракт переведено з того, що треба
памʼятати, у те, що виконується саме:

| Було правилом | Стало структурою |
|---|---|
| «хуки мають охоплювати саме цей блок» | вкладений `describe`, який їх охоплює |
| «на цій машині запускай меншу паралельність» | `workers: 2` у конфізі |
| «перед окремим e2e підніми стек і засій» | ціль `e2e-local` |

**Фільтр для решти беклогу:** для кожного пункту спитати не «чи записано», а **«чи можна
зробити так, щоб забути було неможливо»**. Пункт, який лишається інструкцією, буде порушено —
питання лише коли й ким.

### Що цим НЕ лікується

Структура прибирає **забування**, а не **хибну модель**. Термінатор heredoc (B-11) забути не
можна — його можна не знати; там ліки інші: винести вміст туди, де відступ не залежить від
вкладеності. Це різні хвороби, і плутати їх означає ставити конструкцію там, де потрібне знання,
і навпаки.


---

## B-13 — публічний каталог є спільним станом між e2e-файлами

**Знайдено:** 2026-09-09, після того як `catalog.spec.ts:29` упав **удруге** й спокуса списати
його на навантаження була найбільшою.

**Status: DONE, 2026-09-17, із переглянутим діагнозом** — доказ `work/archive/report-b13.md`. Запис
нижче лікував **ізоляцію тестів**; справжній дефект був інший: тест вимагав стабільності
глобального offset-вікна, якої продукт ніколи не обіцяв
(`docs/foundation/assumptions.md:21` — «offset pagination може зміщуватися між окремими
запитами під час публікації»), і сильнішої за рядок, який він доводить
(`H04`, `docs/foundation/plan.md:49`, просить «той самий відфільтрований порядок» при
повторному відкритті URL, не позицію). `workers: 1` тримав вимогу, якої не давав продукт.

Вікно тепер скоуплене через `q`, якого не містить жодна run-scoped фікстура; три браузерні
твердження збережені, одне посилилось (`toHaveCount(1)` — ліміт **виконано**, а не прийнято).
Доведено прямо, не кореляцією логів: курс із `published_at = now()` зсуває старе вікно
(`offset=1` → «Go»), не зсуває нове (`q=React` → «React», `total=1`), валить версію файла з
`HEAD` (`1 failed, 8 passed`) і не валить нову (`9 passed`). Повний `e2e` — `65 passed (8.6m)`.

`workers: 1` **лишається** за другою, ресурсною причиною (замір 2026-09-08), і коментар у
`playwright.config.ts` тепер називає, яка причина жива, а яка закрита — інакше наступний читач
або не підніме число, маючи право, або підніме, перемірявши не те. Клас `B-8` цим **не**
закритий: будь-який новий тест, що опублікує курс і потім заасертить глобальне вікно, відтворить
дефект.

### Механізм, доведений, а не припущений

Каталог упорядкований **детерміновано**: `ORDER BY … published_at DESC, id ASC`
(`generated/courses.sql.go:357-361`), а id у `seed.sql` фіксовані. Нестабільності в запиті немає.

Але `admin.spec.ts` **P20 публікує курс** посеред свого сценарію, з `published_at = now` —
новішим за всі сіди (`2026-09-03` і раніше). Поки він опублікований, він стоїть **нульовим**, і
`/courses?limit=1&offset=1&sort=recent` віддає «Go» замість «React».

**Чому це трапляється не завжди:** `fullyParallel: false` упорядковує тести **всередині** файлу;
різні файли йдуть **паралельно** за кількістю воркерів. Збіг у логі прямий — тест 15
(`catalog.spec`) завершився між тестами 10 і 16 (`admin.spec`), тобто вікна перетинались:

```
✓ 10 [chromium] admin.spec P20 creates, authors, publishes, and archives (4.5s)
✘ 15 [chromium] catalog.spec search and pagination … (5.5s)
```

### Зроблено

`workers: 1`. Це третя редакція цього числа за добу, і кожна мала свою причину:

| Значення | Причина |
|---|---|
| не задано | паралелізм = половина ядер, тобто **властивість заліза** |
| `2` | пʼять воркерів на завантаженому хості давали 504 і таймаути **браузерів** |
| `1` | файли, що бігли поруч, ділили **глобальний стан** — публічний каталог |

Перші дві причини — про ресурси, третя — про **ізоляцію**, і саме вона не лікується жодним
числом, більшим за одиницю.

### Ціна

Серіалізується лише `chromium` (шість файлів). `media`, `playback` і `recovery` тримають по
одному файлу кожен — на них кількість воркерів не впливає взагалі, а саме вони займають більшу
частину годинника.

**Ціна переходу 2 → 1 названа окремо, щоб число не повернули «щоб швидше».** Оцінка ~1.7 хв на
фазу `chromium` робилась для **двох** воркерів; при одному шість файлів ідуть послідовно, і це
помітно більше. Детермінізм того вартий — але підстава мусить бути видимою, інакше наступний
побачить `1` без причини й підніме його.

### Ширше

Це той самий клас, що B-8: **стан переживає межі тесту**. Тільки там стан був у `lesson_progress`
і в черзі, а тут — у самому продукті: будь-який тест, що публікує курс, змінює те, що бачить
будь-який тест, який читає каталог. Тайбрейкер у SQL цього не рятує, бо проблема не в порядку, а
в **складі** вибірки.


---

## B-14 — тест сидів на межі власного бюджету, а я два дні звав це середовищем

**Знайдено:** 2026-09-10, коли `CourseEditorView.test.tsx` `B09` упав при `load 15` — після того,
як **проходив** при 19 і 20.

### Хибна класифікація, і чим вона трималась

Цей тест червонів чотири рази за два дні, і щоразу я списував його на навантаження хоста. Кореляція
справді була — на 50+ він падав завжди. Але вона **не була причиною**, і перший же прогін, що
суперечив їй (падіння при 15 після проходжень при 19–20), мав це показати.

### Замір, який треба було зробити першим

Ізольований прогін лише цього тесту на тихому хості, тричі:

```
real 9,20   passed
real 9,87   FAILED
real 9,63   passed
```

**Одна відмова з трьох без жодного навантаження.** Тест рендерить 40 модулів по 3 уроки (120
полів), обходить їх клавіатурою і проганяє `axe` — це близько пʼяти секунд роботи проти
**дефолтного** таймауту 5000 мс, який для нього ніхто не обирав.

### Зроблено

Явний бюджет `20_000` у самому тесті, з причиною в коментарі. Після цього — три прогони з трьох
зелені.

**Це не послаблення.** Жоден асерт не змінився: 40 модулів, 120 уроків, `tabIndex` кожного поля,
нуль порушень `axe`. Змінилось лише те, що зелений результат більше не залежить від того, чи
встиг планувальник.

### Урок, ширший за цей тест

Бриф T9 наказував Codex **не** піднімати таймаути — і це було правильно для середовища. Але я сам
застосував те саме правило там, де воно не діяло, і **правило захистило хибну класифікацію**.

> Кореляція з навантаженням не є причиною. Перевіряє її не ще один прогін гейту, а **ізольований
> замір того самого тесту**, повторений кілька разів.

Дешевий замір (три прогони по девʼять секунд) стояв доступним два дні й не був зроблений, бо
пояснення вже було.


---

## B-15 — бюджет старту контейнера в тестах був обраний, але не для цієї машини

**Знайдено:** 2026-09-10, після того як `integration` упав із прямою причиною в логу.

### Чим це відрізняється від B-14

Тут причина **записана**, а не виведена з кореляції:

```
container logs (wait until ready: "database system is ready to accept connections"
  matched 1 times, expected 2 — context deadline exceeded)
```

Postgres буквально ще стартував. Це не здогад про навантаження — це замір, який зробив сам
testcontainers.

### Що змінено

`WithStartupTimeout(time.Minute)` → `3*time.Minute` у **дванадцяти** місцях (одинадцять файлів).

Одна хвилина була обрана явно — на відміну від rstest-дефолту в B-14, — але обрана для машини,
яка не ділить десять ядер із двома десятками чужих контейнерів.

**Це чистий setup, не асерт.** Жодна перевірка не послаблена: те, що доводять репозиторні тести,
не змінилось ні на рядок. І це **таймаут, а не пауза** — на вільній машині Postgres стартує за
секунди, і бюджет не коштує нічого.

### Що це показало окремо

Той самий рядок повторено **дванадцять** разів у одинадцяти файлах — спільного помічника для
підняття Postgres у тестах немає. Тому будь-яка зміна політики очікування — це зміна в дванадцяти
місцях, і розійтися вони можуть непомітно. Кандидат на винесення в тестовий помічник, коли до
цього дійдуть руки.

**Тригер:** наступна зміна стратегії очікування. Якщо доведеться правити вдванадцяте — виносити.


---

## B-16 — правило, яке пояснює симптом, зупиняє пошук причини

**Виділено 2026-09-10 рецензентом як окремий механізм, бо решта пунктів його не описує.**

### Що сталось

Бриф T9 наказував: **не піднімай таймаути**. Правило правильне — воно захищає від того, щоб
тимчасова проблема середовища ставала постійною зміною свити.

Я застосував його до `CourseEditorView` B09 і **два дні** не шукав причину, бо пояснення вже
було: хост завантажений, таймаути очікувані, піднімати їх заборонено. Кореляція з `load`
підтверджувалась щоразу — на 50+ тест падав завжди.

Ізольований замір, який усе вирішив, коштував **девʼять секунд** і стояв доступним увесь час.

### Розрізнювач, якого бракувало

> Правило «не піднімай таймаути» стосується падінь, **спричинених середовищем**. Воно не
> стосується тесту, **чия власна робота впирається в бюджет**.

**Перевіряється одним числом:** скільки тест працює **на тихому хості** проти свого ліміту.
Пʼять секунд роботи проти пʼяти секунд бюджету — це не флейк за жодного навантаження; це
недобюджетований тест, а навантаження лише **робить видимим** те, що й так на межі.

### Механізм, ширший за цей випадок

**Правило, яке пояснює симптом, зупиняє пошук причини — і що правило краще, то надійніше
зупиняє.** Погане правило викликає сумнів і його перевіряють. Добре правило звучить як відповідь,
і перевіряти його не тягне.

**Практичний наслідок:** коли правило пояснює червоне, питати не «чи підходить правило», а
**«який замір відрізнив би випадок, до якого правило НЕ застосовне»** — і зробити саме його.

### Закрито того ж дня: 354 × 429 на краю

Порахував навмисними лише шістдесят і записав решту як непояснене. Насправді навмисні **всі**:
`smoke.sh` робить три сплески — 60 запитів проти `burst=20`, 30 логінів проти `burst=5`, 400
сторінок проти `burst=100`, тобто ≈39 + ≈24 + ≈299 ≈ **362** очікуваних 429 проти 354
спостережених. Третій сплеск я й не міг вважати навмисним, бо його асерт перевіряє **наявність**
429, а не кількість.

**Той самий клас:** я шукав пояснення для числа, замість того щоб порахувати, яким воно **має**
бути.

---

## B-18 — упорядкований гейт перезаписує чотири відслідковувані бінарники

**Знайдено:** 2026-09-18, під час `X14`, коли після **зеленого** `make check` дерево виявилось
брудним, а змін я не вносила.

### Механізм

`frontend/e2e/accessibility.spec.ts:164,395,400` і `recovery.spec.ts:47` пишуть скріншоти
**прямо в репозиторій**:

```
docs/foundation/screenshots/catalog-desktop.png
docs/foundation/screenshots/catalog-error.png
docs/foundation/screenshots/catalog-mobile-dark.png
docs/foundation/screenshots/course-desktop.png
```

Тобто **кожен** прогін `e2e` лишає чотири відслідковувані файли зміненими. Заміряно:
`catalog-desktop.png` 987 920 B → 201 820 B, `catalog-mobile-dark.png` 983 619 B → 186 772 B —
різниця **майже пʼятикратна**, і вона **не** регресія рендера: відкрила зображення, сторінка
правильна (три опубліковані курси, коректна українська копія, «1–3 із 3»). Причина
середовищна — інший device scale factor або кодувальник PNG на іншій машині.

### Чому це не косметика

1. **Після зеленого гейта `git status` ніколи не чистий.** А саме брудне дерево — передумова
   `B-17`: я сама за цю сесію ледь не закомітила `git add -A` чужі робочі файли, і тоді
   помітила їх лише тому, що читала вивід `git status`.
2. **Бінарний діф не рецензується.** Скріншот, що поїхав у коміт разом із кодом, не видно в
   рев'ю — його неможливо прочитати, як текст.
3. **Файли будуть перекидатись між машинами**: пʼятикратна різниця залежить від середовища, тож
   двоє контриб'юторів довічно перезаписуватимуть їх одне одному, і в історії лишиться шум без
   сигналу.

### Варіанти

- писати скріншоти в проігноровану теку (`frontend/playwright/.artifacts/`), а закомічені
  тримати як **свідомо** оновлювані артефакти документації;
- лишити шлях, але додати ці чотири файли в `.gitignore` — тоді вони зникнуть із документації,
  на яку посилається `docs/foundation/ui-evidence.md`, тож так робити не можна без рішення;
- лишити як є й покладатись на те, що кожен щоразу робить `git checkout -- docs/`.

**Рішення продуктове (де живе доказ UI фази 1), тож не ухвалюється тут.** Дерево цього разу
повернуто в чистий стан `git checkout -- docs/foundation/screenshots/`.

**Тригер:** будь-який наступний коміт після прогону гейта; або поява другого контриб'ютора, що
жене `e2e` на іншій машині.

### B-18 закрито 2026-10-20

Вже закрито раніше: PR #32 (`d6ec304`) перевів e2e-скріншоти на проігноровану теку
`frontend/playwright/.artifacts/`, тож прогін гейта більше не лишає відстежувані бінарники
зміненими. Запис AGENDA 13.10, що тримав B-18 відкритим, був застарілий. Дерево після гейта
цієї сесії чисте (лише vim-swap).

---

## B-17 — `git add -A` під час роботи відчепленого агента змішує коміти

**Сталось:** 2026-09-11. Коміт `6110a09` має повідомлення про правку документів, а містить
**усю реалізацію T10**: 381 рядок e2e, зміни `LessonPlayer`, звіт. Я зробив `git add -A`, поки
Codex іще писав, і замів його роботу у свій документаційний коміт.

**Чому це не дрібниця обліку.** Повідомлення коміта — єдине місце, де записано **навіщо** зроблено
зміну. Коміт, що описує одне, а несе інше, робить `git log` і `git blame` **тихо хибними**: той,
хто шукатиме, коли зʼявився `onPause`, знайде коміт про документи.

**Правило:** поки відчеплений агент працює, `git add -A` **заборонений**. Або перелічувати шляхи
явно, або чекати завершення. Історію не переписую — коміт уже запушено, і його міг забрати
рецензент; замість цього факт записано тут і у звіті T10.

**Це той самий клас, що B-12:** дешевий шлях (`add -A`) коротший за правильний (перелік шляхів) і
мовчки неправильний. Різниця в тому, що тут він не червоніє **ніколи** — жоден гейт не перевіряє,
чи коміт про те, що в ньому.

### Спроба полікувати це ігнором — і чому вона хибна (2026-09-12)

Я додав був `frontend/*.mjs` у `.gitignore`, щоб чужа тимчасова проба фізично не могла потрапити
в коміт. Рецензент спинив, і мав рацію двічі.

**Патерн накривав `frontend/server.mjs`** — відстежений вхідний файл SSR, що лежить точно в корені
пакета. Відстежене git не ігнорує, тож нічого не зламалось; але захист зникає, щойно файл
опиняється невідстеженим — свіжий worktree після видалення, гілка, що його прибирає й повертає.
Тоді він стає невидимим саме тоді, коли цього ніхто не перевіряє.

**Шрам уже є в цьому ж файлі.** `/media/` записано з провідною скісною, бо колись `media/` без
якоря збіглося з `docs/media/`, і затверджена специфікація фази 5 **ніколи не потрапила в
репозиторій**: коміт, що нібито її додавав, містив лише правку брифа. Знайшли за кілька днів.

**І глибше: ігнор лікував не те.** Ризик створила не проба, а `git add -A`. Правило, що ховає
пробні файли, **легітимізує** запис проб у репо — закріплює звичку, від якої треба відмовитись.

**Справжні ліки — явні шляхи в `git add`.** Вони фізично не можуть захопити те, про що я не знаю,
і при цьому нічого не ховають. Вужче за ігнор і лікує B-17 повністю, а не лише його пробний
випадок.

**Правило про ігнори, ширше за цей випадок:** ігнорувати **анкерований шлях**, а не **форму
імені**. `/frontend/.scratch/` не може проковтнути наявний файл; `*-probe.*` збігається на будь-якій
глибині з тим, чого ще немає.

---

## B-19 — ранг 360p упирається в `maxrate 365k` і дає VMAF 60 на рухливому відео — CLOSED 2026-10-20

**Raised:** 2026-09-20, під час власного виміру AV1 (`docs/media/research-codecs.md` §6).
**Shape:** відкрите продуктове питання про драбину, не дефект коду.

### Вимір

На натуральному відео з високим рухом (кіно-фрагменти, зерно, склейки), 30 с, ранг 360p
із чинними налаштуваннями `transcoder.go` (`crf 21`, `maxrate 365k`, `bufsize 730k`,
`preset veryfast`):

| | бітрейт | VMAF (нормалізований) |
|---|---|---|
| H.264, як зараз | 369 kbps — **101% cap**, тобто впирається | **60.23** |
| AV1 preset 8, той самий бюджет | 358-395 kbps | 71.30 |

Для порівняння, на тому ж джерелі ранг 720p дає VMAF 92.95 і **не** впирається в cap
(1833 kbps проти 3000k). Тобто просідає саме 360p, і саме через стелю бітрейту.

### Чому це питання, а не готовий баг

Найнижча сходинка **за задумом** обслуговує поганий канал, і низька якість там може бути
свідомою ціною. 365k обрано в Phase 5 разом із рештою драбини. Але VMAF 60 — це видимі
артефакти, а не «трохи гірше», і власник продукту має підтвердити, що це прийнятно, або
назвати нову цифру.

### Що робити, коли візьмуть

Рішення — одна константа в `settingsFor` (`backend/internal/media/adapters/ffmpeg/transcoder.go`),
але воно тягне: цільову смугу найнижчої сходинки, `AVERAGE-BANDWIDTH` у master-плейлисті
й ті e2e-рядки, що фіксують розміри рендишенів. Міряти тим самим способом, що й §6
дослідження: нормалізований VMAF, keyframe-паритет, реальне відео, не тест-патерн.

### Чого цей запис НЕ стверджує

Не стверджує, що треба переходити на AV1: §24.14 тримає AV1 за трьома умовами, і жодна
з них цим виміром не спрацювала. AV1 тут лише показує, що на тому самому бюджеті можливо
більше — тобто що впирається саме бюджет.

### Outcome 2026-10-20 — закрито (рішення власника: підняти бюджет)

Власник (2026-10-20) обрав «підняти бюджет через замір». Повторено метод §6.5 (keyframe-паритет,
обидва потоки нормалізовано до 1080, реальне відео) на демандинговому high-motion 1080p-семплі
(jellyfish, 10 с; §6-семпл у репо немає, тож абсолютні числа нижчі — jellyfish жорсткіший за film):

| maxrate | досягнуто | keyframes | VMAF(1080) |
|---|---|---:|---:|
| 365k (було) | 390k | 5 | 33.3 |
| 500k | 535k | 5 | 42.4 |
| 650k | 697k | 5 | 50.2 |
| 800k | 857k | 5 | 56.0 |
| 1000k | 1065k | 5 | 61.9 |

Крива **лінійна, без коліна** → це чистий компроміс смуга↔якість, не технічний поріг: число —
продуктова політика про те, який канал обслуговує найнижчий ранг. **Обрано 600k** (bufsize 1200k):
+64% до бюджету, лінійно підіймає якість (на цьому семплі ~33→~47; на §6-film-класі 60→~75+), але
лишає ранг справжнім low-bandwidth (досяжний на ~700 kbps) і далеко під 720p (3000k), тож ABR має
реальний нижній рівень (e2e R35 читає bandwidth динамічно й тримається). Коштом — найнижчий ранг
більше не цілить у ~400 kbps канал; це свідома межа, число легко скоригувати (одна константа).

**Зміни:** `transcoder.go` settingsFor 360p `365k→600k`/`730k→1200k`; `transcoder_test.go` 3 arg-рядки
оновлено. Master-плейлист BANDWIDTH генерує ffmpeg (Go-коду нема). Замір: `scratchpad/b19/` (sweep.sh,
results). Метод відтворюваний; семпл — CC test-media (jellyfish), стабільний URL.

## B-20 — draft-ендпоінти відсутні в `openapi.yaml`; фронт на рукописних zod-схемах

**ЗАКРИТО 2026-09-28 комітом `e8c238d`** («fix(transcript): B-20 — the three draft endpoints enter
the OpenAPI contract, and the frontend reads them through Orval»). Три ендпоінти описано в
`openapi.yaml` (схеми `LessonDraft`/`DraftSegment`/`RegenerateLessonDraftRequest`/`PublishLessonDraftRequest`),
`draft-schema.ts` видалено, `draft-repository.ts` ходить Orval-генерацією, бекенд має контракт-тести
(`draft_handler_test.go`). Доведено машинно 2026-10-13: `npm run generate` → git diff порожній (клієнт
збігається з openapi.yaml). Запис лишається як слід; воскрешати в AGENDA/брифінгу **не можна**.
⚠️ Урок: fix-коміт назвав «B-20», але цей запис не позначили закритим — тому AGENDA й брифінг два тижні
несли його відкритим. Закривати запис боргу в тому ж коміті/PR, що й фікс.

**Raised:** 2026-09-24, під час рев'ю кроку 2 A13 (дефекти C10/C14).
**Shape:** розрив контракту, не дефект коду сам по собі.

### Що саме

Три ендпоінти чернетки ASR зареєстровані в `draft_handler.go:25-29`:
`GET /api/v1/admin/lessons/{lessonId}/draft`, `POST …/draft/regenerate`,
`POST …/draft/publish`. У `backend/openapi/openapi.yaml` їх **немає** (описані `/transcript`,
`/video`, learn `/transcript` — жодного `/draft`). Зайшло з A09, коли ендпоінти додали без
поповнення контракту.

Тому фронт **не має Orval-моделей** для чернетки й ходить через рукописний адаптер із власними
zod-схемами: `frontend/src/adapters/api/draft-repository.ts` →
`frontend/src/adapters/api/draft-schema.ts` (`draftErrorSchema`, `lessonDraftSchema`). Це прямо
суперечить AGENTS.md §Architecture («Generated sqlc/Orval models stay in adapters and map
explicitly to domain entities»).

### Чому це не косметика — механізм

Контракт (коди помилок, форма чернетки) живе у **двох рукописних місцях** — Go-handler і
фронтовий zod — **нічим одне з одним не звірених**. Дрейф між ними не червоніє на жодному гейті.

Доказ, що це не гіпотеза, а вже спрацьоване:
- Плутанину кодів у самій A13 (`asr_recognition_in_progress` проти реального контракту
  `already_running`, який чекає `draft-repository.ts:112`) зловило **рев'ю брифу**, а не гейт.
- Реальний дефект C10 — бекенд **ніколи не віддавав** `already_running`, хоча фронт його чекав —
  теж пройшов повз усі рівні приймання (16→14/22), бо звіряти рукописну пару нічому.

Дрейф контракту тут — наступний C10-подібний баг, який ніхто не побачить, поки не впаде в UI.

### Спосіб перевірки, коли візьмуть

Коли draft-ендпоінти опишуть в `openapi.yaml`, draft-адаптер має перейти на **Orval**, а
`check-generated` почне ловити дрейф автоматично — тобто контракт стане машинно-перевірюваним, а
не тримається рев'ю. Обсяг: три path-описи (з кодами помилок і схемою чернетки) в `openapi.yaml`,
регенерація Orval, заміна `draft-repository.ts`/`draft-schema.ts` на згенероване з явним
мапінгом у домен.

### Чого цей запис НЕ стверджує

Не стверджує, що A13 має це закрити: A13 закриває **приймання** ASR, не контракт транспорту.
Це окремий пункт (рішення власника 2026-09-24). Не стверджує, що поточний фронт зламаний зараз —
рукописні схеми наразі збігаються з бекендом; ризик — у майбутньому дрейфі без машинного гейта.

## B-21 — читання без мови бере перший транскрипт за алфавітом, а не мову уроку — CLOSED 2026-10-20

**Raised:** 2026-09-24, під час діагнозу C18 у A13 (Courses 2 знайшла, Courses 1 звірила в коді).
**Shape:** латентна підміна змісту для студентів; зараз не проявляється, бо у фазі одна мова на урок.

### Що саме

Коли мову не передано, читання транскрипту обирає `ORDER BY language LIMIT 1`
(`backend/internal/transcript/adapters/postgres/repository.go:40-44`, `getTranscript`; коментар
над ним прямо каже «the first by normalised code when several exist»). Без мови читають обидва
споживачі: студентський плеєр (`frontend/src/routes/learn.$.tsx:84`, `learn.lesson(lessonId)`) і
редактор адмінки (`frontend/src/routes/admin.courses.$courseId.tsx:204`, `transcripts.lesson(lessonId)`).

### Механізм, як це зламається

Щойно в уроку з'явиться другий транскрипт іншою мовою, будь-яким шляхом, усі — зокрема
студенти — побачать той, чий код раніше за алфавітом, а не той, що автор вважає «транскриптом
уроку». Приклад: урок має `uk`, додається `en` → усі бачать `en`, бо `en` < `uk`. Мовчки, без
жодного сигналу авторові.

Один такий шлях уже існував: правило кроку 4a A13 («завантаження ASR-чернетки перемикає редактор на
її мову») давало зберегти `en`-чернетку поверх `uk`-транскрипту. Крок 4g його закрив: завантаження
чернетки іншою мовою поверх наявного транскрипту тепер відмовляється (рішення користувача
2026-09-24, `work/archive/brief-a13-language.md`). Але крихкість самого читання лишилась: алфавіт замість
мови уроку.

### Спосіб перевірки, коли візьмуть

Канонічне читання без мови має брати мову уроку (або явно обрану автором основну мову), а не
перший код за алфавітом. Integration-тест: урок з `uk` і `en`, основна мова `uk` → читання без мови
повертає `uk`. Наявний тест `language selection is deterministic`
(`repository_integration_test.go:234`) закріплює саме алфавітний вибір — його доведеться переписати.

### Чого цей запис НЕ стверджує

Не стверджує, що зараз студенти бачать не той транскрипт: у фазі одна мова на урок, а 4g закриває
єдиний відомий шлях до другої. Це читальний шлях 7a, не приймання A13 — окремий пункт за рішенням
користувача Courses 2 (2026-09-24).

### Fix 2026-10-20 — закрито (рішення власника: реалізувати)

Власник (2026-10-20) вирішив узяти B-21 зараз, **Варіант A + мінімальний зріз**: маркер первинності
живе на транскрипті, без редактор-UI (додати 2-гу мову через інтерфейс сьогодні все одно не можна).

- **Схема** (`00021_transcript_primary_language.sql`): `transcripts.is_primary boolean NOT NULL DEFAULT false`
  + partial unique `transcripts_one_primary ON (lesson_id) WHERE is_primary` (рівно 1 primary/урок).
  Backfill: перший за алфавітом транскрипт уроку → primary (той самий рядок, що й повертало старе
  читання, тож при міграції жодне читання не змінюється; у проді 1 мова/урок — маркується вона).
- **Читання** (`repository.go` getTranscript): `ORDER BY is_primary DESC, language` — без мови повертає
  primary; `language` лишається tie-break для детермінізму, якщо primary якось не позначено.
- **Запис** (`write.go`): перший транскрипт уроку стає primary (`is_primary = NOT EXISTS(... AND is_primary)`);
  додана пізніше мова приходить non-primary й не перехоплює канонічне читання.
- **Тести**: переписано `language selection is deterministic` → `the primary language is read when none is
  asked (B-21)` (twoLang uk-primary+en → читання без мови = `uk`, не `en`); додано production-path proof у
  write-тесті (через `authoring.Save` дописати `en` до uk-уроку → читання без мови лишається `uk`).
- **Гейт**: упорядкований `make check` GATE_EXIT=0 (frontend 773, усі Go-пакети ok, e2e 122 passed, 15.6m).
  Повний гейт зловив дефект поза дифом (колізія segment-id у новому write-subtest'і зламала сусідній
  library-тест) — рівно той клас, що вимагає повного e2e на зміні персистованого стану.

**Named limitation (механізм):** дві конкурентні вставки *перших* транскриптів різними мовами в один
порожній урок — другий програє з `transcript_exists` через one-primary індекс (не `transcript_modified`).
Недосяжно сьогодні: редактор single-locale, другу мову через UI додати не можна. Переглянути з форком
«друга мова», коли з'явиться вибір primary в редакторі.

**Не зроблено (поза зрізом):** редактор/reader locale-перемикач і вибір primary автором — це форк
«друга мова» (`work/AGENDA.md` §розвилки 2), відкладений до реального попиту.

## B-24 — payment PR #3 review findings (Courses 2, 2026-10-02, non-blocking)

From `work/archive/review-pr3-patterns.md`. 0 blockers; `orders` module judged exemplary. Four polish items:

1. **Duplicate-JSON-key check is asymmetric and copy-pasted.** `validateUniqueJSONObject` (reject
   duplicate keys) lives in `decodeGrant` (orders), and independently in `admin_handler.go` and
   `discussion/handler.go` — three copies. `decodeCheckout` (same orders module) does NOT have it
   (only `DisallowUnknownFields`), and has no dup-key refusal test. Resolve by deciding the rule
   once: if duplicate-key rejection is the contract (it is enforced on 3 write surfaces), lift the
   helper to a shared platform place and apply it to `decodeCheckout` too, with a test; the checkout
   impact today is nil (body is only an optional cosmetic slug).
2. **In-memory status throttle** (`service.go`) → N× CHECK_STATUS with N instances. Recorded in
   `docs/payment/report.md` §6. A shared store before horizontal scaling; Proxy pattern not worth it now.
3. **Variadic optional dependency** in `service.New` and `http.NewHandler` (two handler
   constructors). Prefer one explicit nilable parameter over the variadic + second constructor.
4. **Date-format literal repeated 3×** in `handler.go` — extract a const.

Trigger: fold 1/3/4 into a follow-up polish commit on `feat/payment` before or just after the PR
merges; 2 stays a named limitation.

### Outcome 2026-10-20 — закрито

Звірка з кодом (рев'ю було 2026-10-02; код відтоді еволюціонував):
- **п.4 вже зроблено** — `dateTimeLayout` const існує (`orders/adapters/http/handler.go:26`), літерал не повторюється.
- **п.1 (gap) вже зроблено** — `decodeCheckout` уже викликає перевірку унікальних ключів і мапить у `ErrOrderInvalid`.
- **п.3 вже зроблено** — `service.New` і `NewHandler` без variadic/другого конструктора (grep `...` порожній).
- **п.1 (residual) — зроблено цією сесією:** три копії `validateUniqueJSONObject` (course/orders — рекурсивні, discussion — лише верхній рівень) зведено в один `platform/jsonbody.ValidateUniqueKeys` з нейтральним sentinel; кожен handler мапить non-nil у свій domain-error на місці виклику. Логіку винесено в один unit-тест (`platform/jsonbody/unique_test.go`, 12 кейсів). **Рішення (2026-10-20, Courses 1):** єдина версія рекурсивна — для discussion це глибша перевірка (вкладений дубль-ключ тепер відхиляється), але без зміни для валідних (пласких) discussion-запитів; механізм: discussion-схеми не мають вкладених об'єктних полів.
- **п.2 лишається named limitation** (in-memory status throttle, `docs/payment/report.md` §6): спільний store — перед горизонтальним масштабуванням.

## B-25 — WayForPay refuses a second checkout of the same pending order (live run, 2026-10-02) — CLOSED `f306165`

**Closed** by option (b) plus a refund trace; see `docs/payment/assumptions.md` (2026-10-02) and the
phase report §6. Kept below as the record of the finding.


**Found by:** the Pay25-live run against `test_merch_n1`. **Mechanism:** our checkout sends the order
id as WayForPay's `orderReference`; WayForPay treats `orderReference` as unique, so a second hosted
form for the same order (the buyer opened «Перейти до оплати», left, came back and pressed it again)
answers `(1112) Duplicate Order ID`. The status page keeps offering «Перейти до оплати» for a pending
order, so the buyer hits a dead end; «Спробувати ще раз» appears only for `failed`. **Repro:** start a
purchase, open the checkout form twice for the same order (two POSTs to
`/api/v1/learn/orders/{id}/checkout` submitted to WayForPay). **Fix options:** (a) a per-attempt
`orderReference` (order id + attempt suffix) mapped back to the order in the callback and in
CHECK_STATUS — the money path must still match one order; (b) on a pending order whose checkout was
already opened, start a new order (new idempotency key) instead. Decide before launch; (a) keeps one
order per purchase intent.

## B-26 — the live proof needs a stable public URL and a working test card

Cloudflare's quick-tunnel API (`api.trycloudflare.com`) is unreachable from this network (TCP to its
IPs times out; cloudflare.com itself works); the localhost.run free tunnel works but rotates its
hostname within minutes, which breaks `serviceUrl`/`returnUrl` mid-run. The test card from WayForPay's
API examples (`4111…`) is declined (1105). To close Pay25-live: a named tunnel (Cloudflare account or
ngrok authtoken) or the deployment, plus a documented test card or the owner's own test-mode merchant.

### B-26 addendum — what WayForPay's own docs say (researched 2026-10-02)
Sources: Verify API `wiki.wayforpay.com/view/852189`, Purchase `/view/852102`, response codes `/view/852131`,
test credentials `/view/852472`.
- **No test-card list is published anywhere in the wiki.** Test credentials page lists only merchants
  (`test_merch_n1` + key, `test_merch_p2p`, `p2p_credit`). The "4111 = success, 5454 = failure" claim is
  not in the wiki; live: 5454 → 1101 (issuer declined) as claimed, 4111 → **1105** ("invalid card number
  **or card in a prohibited state**"). Ask WayForPay support which card is Approved on `test_merch_n1`.
- **Codes we saw:** 1100 ok; 1101 issuer declined; 1105 invalid/prohibited card; 1112 duplicate order id
  (= B-25); 1131 still processing; also 1113 = wrong merchant signature (never seen — our signatures
  are accepted live).
- **transactionStatus set:** InProcessing, WaitingAuthComplete (awaiting settlement), Approved, Pending
  (antifraud review), Expired, Declined, Refunded/Voided, RefundInProcessing. Our mapping
  (Approved → paid; Declined/Expired → failed; the rest acknowledged, no change) holds for MVP.
  **Gap to name:** `Refunded`/`Voided` after an `Approved` are ignored today — a refund done in the
  WayForPay cabinet would leave the buyer's grant active. Belongs with the refund flow (out of MVP).
- **`orderTimeout` / `orderLifetime`** (Purchase params we do not send): how long the hosted page allows
  re-trying payment, and how long the order can be paid at all. Setting `orderLifetime` would let
  WayForPay itself expire abandoned orders (→ `Expired` → our order `failed`), tidying the extra pending
  rows B-25 option (b) creates.
- **serviceUrl** response format and the 4-day retry are confirmed — matches our acknowledgement and the
  Pay21b "unmatched → 503 so WayForPay retries" behaviour.

## B-28 — Review list cursor: the API sends "" for "no next page" — CLOSED 2026-10-04

**Raised:** 2026-10-04, ratings phase (found by e2e; Courses 2 T06b-fixes review №3).
**What:** `GET /courses/{slug}/reviews` and `GET /learn/courses/{id}/reviews` return `nextCursor: ""` on the last
page (the discussion repository's cursor encoding). The frontend maps `""` → `null` at the adapter boundary
(`course-repository.ts`, test «treats the API's empty last-page cursor as no next page»), so behaviour is right.
**Why still open:** the contract says "string", so every future client has to know the `""` convention. Fix:
the HTTP layer returns `null` and OpenAPI marks `nextCursor` nullable; then the frontend mapping can go.
**Trigger:** the next change to the review list contract, or a second client.

### B-28 closed 2026-10-04
The review list handlers return `nextCursor: null` on the last page (`nextCursorOrNull`); OpenAPI marks it nullable;
Orval regenerated; the frontend `""` mapping is gone. Contract test «B-28 the last page says null, not an empty
cursor»; repository test passes the null through. Comment and queue lists keep their older `""` (out of scope).

### B-28 residual closed 2026-10-20 (PR #58 `a5f15d2`)

The lists left out of scope on 2026-10-04 — discussion threads (`pageDTO`), course comments
(`courseCommentsDTO`) and the moderation queue (`queueDTO`) — still sent `""` on the last page, which the
frontend turned into a phantom next page via `"" ?? undefined`. All three now return `null` through
`nextCursorOrNull`; the three OpenAPI schemas are `nullable`; Orval regenerated (`commentPage`,
`teachCourseCommentPage`, `moderationQueuePage` → `string | null`). New backend test
`TestB28ResidualListCursorsAreNull`. The `""` convention is gone from every list cursor.
