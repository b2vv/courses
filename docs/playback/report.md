# Phase 6 — playback: delivery report

Delta to [plan.md](plan.md) and [spec.md](spec.md). Written 2026-09-12 at phase close from
the task evidence in `work/archive/report-t1.md` … `work/archive/report-t12.md` and the reviewer's ordered
gate recorded in T11 and T12. There is no production deployment to check.

## 1. What was done

A signed-in student can open a published course's learning surface, see its ordered outline,
start the server-selected lesson and play real HLS through nginx. The API mints a server-side
playback identity; nginx asks the playback module for authorization on every playlist and
segment request and serves the media bytes itself. Heartbeats renew the account-wide playback
slot and persist monotonic progress. Reloading or changing browser context restores an
unfinished position, reaching 95% marks the lesson watched, and reopening a completed lesson
starts it from zero. A second tab, device or course playback takes the same account-wide slot
and the displaced session loses media access. These claims are exercised by T4–T10 and the
60-test browser run in the ordered gate.

The same surface exposes explicit empty, no-video, processing and failed states, keyboard
controls, Ukrainian and English copy, light and dark themes, visible focus, 200% zoom support
and 44 px controls. T11 measured those browser properties with the qualifications in §6.
T12 added a bounded, once-per-page AV1 capability observation to the playback mint seam; its
frontend and backend halves are contract-bound but are not one wire-crossing test.

This phase did not add payments, orders or enrolment, comment threads, transcripts, an
instructor workspace or a production deployment.

## 2. Acceptance scenarios

**Amendment 2026-09-17 — «48 packages» is a double count; the runs are unchanged.** Every
place in this report that says Go `-race` **48 packages ok** (§2 and the P12 row) reports the
number of `ok` lines, not the number of packages. The head of the gate runs Go **twice** —
`make test` → `go test -race ./...`, and `infrastructure/tests/integration.sh` →
`go test -count=1 -tags=integration ./...` — and each prints its own `ok` lines, so packages
that run in both are counted twice. Measured on this phase's own gate logs
(`work/archive/make-check-final2.log`, `work/archive/make-check-final3.log`): **48 `ok` lines, 26 unique
packages** — 22 from the race run plus 4 that only `-tags=integration` adds. The runs
happened as recorded and no row's evidence weakens; only the word «packages» was wrong, and
it read as roughly twice the coverage actually exercised. Found by the parallel session while
auditing its own identical claim. Class: a number read off a log without checking what the
log counts.

**Current status: 102/102 verified**, as of the update after T16 below. The paragraphs that follow
are kept as the record of how the count moved there — 89 at T13, then 91, 92 and 95 — and are not
rewritten.

The header below §8 in the plan says **94**, but the three tables actually contain **102
unique rows**: P01–P14, B01–B47 plus B21a and B24a, and R01–R39. The evidence supports **89/102**.
Rows not verified at the planned outcome are **B07, B08, B20, B21, B21a, B22, B23, B28,
B31, B36, R23, R30 and R34**. A green aggregate gate is not substituted for a missing
assertion.

**Update 2026-09-12, after T14 — ordered `make check` EXIT=0.** B23 and R30 are now verified, so the evidence supports **91/102**. B36 is verified in part: 503 and `Retry-After` in smoke; its "no media bytes" clause moved to e2e together with six rows whose proof level changed (`docs/playback/plan.md`, amendment `8cc1d72`), and that clause stays open. Rows still not verified at the planned outcome: **B07, B08, B20, B21, B21a, B22, B28, B31, R23 and R34**, plus the open clause of **B36**. B07 turned out to be a product gap, not missing evidence: the learning read returns the stored position without bounding it by the current video's duration. The paragraph above is kept as the record at T13. The gate ran before the T14 commit, on the content that commit records.

**Update 2026-09-13, after B07 — the row was a defect, and it is fixed.** B07 is now verified, so the
evidence supports **92/102**. The write path bounded a saved position by the current asset's duration and
the read path did not, so replacing a 34:10 video with an 8:00 one left the lesson opening at its end and
unable to advance on its own; one manual seek backwards still recovered it. The bound now lives beside the
existing position validation in the course domain, duration is passed in and not stored — `LearningLesson`
keeps its shape, so `openapi.yaml` and the generated models are untouched — and both boundaries read only a
`ready` asset, matching `loadLessonDuration`, so the two deliberate copies of the policy cannot disagree
about which asset defines the duration. Fixed and gated in `819dcac`; tracked as its own row in
`work/tasks.md` rather than inside T15, which did not do it.

Measured on the tree of `819dcac`: the ordered `make check` was green through `integration`, including
`courses/internal/course/adapters/postgres` **ok 14.982s** — the level B07's row assigns — and then killed by
the host running out of memory during `e2e`. The tail was rerun as `seed-identities → seed → smoke → e2e`,
**60 passed**, because a killed browser run had already consumed a disposable identity fixture and rerunning
`e2e` alone reported that consumption as a failure. Stated that way rather than as a green `make check`: the
tree was unchanged between the two runs, the database was not.

**Update 2026-09-14, after T15 — B08, B31 and R34 are verified; B31 was a defect, and it is fixed.** The
evidence supports **95/102**. Rows still not verified at the planned outcome: **B20, B21, B21a, B22, B28 and
R23**, plus the open "no media bytes" clause of **B36** — the set T16 carries.

- **B08** — `learning_integration_test.go`, subtest `B08 completed mark survives video replacement`, against
  PostgreSQL; its precondition asserts the first asset evicted and the second active before the mark is read.
- **R34** — `playback.spec.ts`, `R34 a non-admin owner plays the archived course while another non-admin is
  refused`: the archived course is reassigned to a non-admin instructor, the owner receives segment 200, and a
  student minting the same lesson is refused. Unlike P11, the player here is not an administrator.
- **B31** — the extended `P13 B15 B31 B32 takeover …`: the displaced device retakes the slot, its segment
  returns 200, and the device it displaced is refused. The first gate with this assertion was **red on the
  product**: retry after any error state minted a new session — taking the slot and displacing the other
  device — while the media stayed bound to the `<video>` unmounted with the error state, so the player stayed
  dead until reload. `LessonPlayer` now attaches the element through a callback ref, and both the media effect
  and the retry effect follow the mounted element. A component test fails without the fix (`length 2 but got
  1`) and asserts media on the remounted element, exactly two mints, and the position after retry: 47 s against
  a fixture start of 12 s, so neither a reset nor a stale start point passes (a mutation restoring from 0 fails
  with `expected +0 to be 47`). A race between media creation and the new mint cannot arise: `onReady` follows
  only `load`, which runs inside the mint after `loadSession` has written the resume point.

The ordered `make check` ran on the T15 tree before one later test-only change: `GATE_EXIT=0`; frontend unit
tests **380 passed, 0 failed**; Go `-race` **48 packages ok**, including `courses/internal/course/adapters/postgres`
**ok 11.890s**; e2e **61 passed, 0 failed** (R34 is the one row added since T14). After the gate a guard was added
to the B31 component test, failing with its own message if the test's position ever equals the fixture start
point (set to 12 s, it fails with that message), and checked with `format:check`, `lint`, `typecheck` and the
full unit suite; the other targets cannot see
it, because a unit test file is not part of the Go, generated-code, production build or browser graphs.

**Update 2026-09-14, after T16 — the six rows moved to e2e and the open clause of B36 are verified.** The
evidence supports **102/102**; no row of `docs/playback/plan.md` §8 is open. All seven proofs live in
`frontend/e2e/playback.spec.ts`, in the serial `playback` project, placed before R04 in this order: B20/B21/B21a/B22,
R23, B36 (which restores Redis in `finally`), and B28 last, because it empties the directory of an asset that is
uploaded afresh by every run and that no later test reads.

- **R23 measures against what flowed, not what is stored.** The denominator is the sum of 200 `.m4s` bodies the
  browser received in the window, not the asset's files on disk: the disk holds every ABR rung and the player takes
  mostly one, and nginx authorizes every media request through the api, so the api's output grows with the number of
  segments fetched. Delivered in this run: `360p=1 of 17, audio=17 of 17, 720p=16 of 17`, 41 successful
  authorizations. The comparison is deliberately asymmetric — interface bytes with headers against bodies without
  headers or init segments — and both asymmetries make the bound stricter. "Whole lesson" means the session went from
  the start to `ended`; at 16× it does not mean every segment of every rung was fetched.
- **The two baselines are taken at different moments on purpose.** The api's precedes the page load, because the
  mint, heartbeats and per-segment authorizations belong to playback; the frontend's follows it. The first gate took
  both before `page.goto`, and the frontend then sent 458,999 bytes — the weight of the SSR page, constant against a
  denominator that scales with the asset, so the bound would have turned red on a shorter lesson or a heavier bundle
  with no media passing through SSR at all. With the frontend baseline after the page load, the frontend sent
  **132 bytes** during playback and the api **70,800**, against 17,208,719 bytes of delivered segments.
- **B28 guards its destructive path twice.** Preconditions prove the directory is an asset before anything is
  removed; a canonical UUID check at the call site proves which asset it is, because a path that names another real
  asset passes every precondition.

Ordered `make check` on the T16 tree, run in two parts in the same target order so that a memory kill in `e2e` would
not cost the earlier targets (no tail target calls `up`): `HEAD_EXIT=0` for `format-check` through `up`, `TAIL_EXIT=0`
for `seed-identities seed smoke e2e`, `GATE_EXIT=0`. Frontend unit tests **380 passed, 0 failed**; Go `-race`
**48 packages ok**; e2e **65 passed, 0 failed** — the four added tests carry the seven proofs. The frontend
baseline in R23 was then moved after the page load, and only the tail was rerun on that change — `TAIL_EXIT=0`, e2e
**65 passed, 0 failed** — after `format:check`, `lint`, `typecheck` and the unit suite; no head target reads an e2e
specification.

The reviewer ran the ordered `make check` on 2026-09-12 against the tree that became `0959fca`
(T12) and `efc9851` (T11) — the gate preceded those commits and saw that content, not the
committed state:
`GATE_EXIT=0`; format-check, lint and typecheck passed; frontend unit tests were **379 passed,
0 failed**; Go `-race ./...` reported **48 packages ok**; integration, generated-code check,
build, Compose startup, both seed steps and smoke passed; e2e was **60 passed, 0 failed,
0 did not run** in **7.7 minutes**.

In the table, `gate` means that ordered run. The named T-report is the primary row-specific
proof; `gate` only records the final execution of proofs that actually exist.

| ID | Planned observation | Actual proof and run | Verdict |
|---|---|---|---|
| P01 | Real HLS plays and segments return 200 | T7 real-playback e2e; gate 60/60 | matched |
| P02 | Outline marks the lesson completed through real playback | T8 playback→heartbeat→learning-model e2e; T8 47/47 and gate | matched |
| P03 | Start uses server-selected `continueLessonId` | T8 e2e; T8 47/47 and gate | matched |
| P04 | Position survives reload | T9 e2e; T9 52/52 and gate | matched |
| P05 | Position survives another browser context | T9 e2e; T9 52/52 and gate | matched |
| P06 | A 100 s asset completes at 95% | T9 dedicated-fixture e2e; T9 52/52 and gate | matched |
| P07 | A completed lesson reopens at zero | T9 e2e; T9 52/52 and gate | matched |
| P08 | Playback outlives ownership TTL through renewals | T7 e2e, at the 2/6 test cadence; gate | matched with §6 timing limit |
| P09 | hls.js manual quality selection requests that rung | T7 e2e; gate | matched |
| P10 | Seeking into unloaded media requests the covering segment | T7 e2e; gate | matched |
| P11 | Admin plays an archived course | T7 admin e2e; gate | matched |
| P12 | AV1 probe result reaches the mint seam and increments a series | T12 frontend/backend tests plus OpenAPI; frontend 379 and Go 48-package gate | matched with split-proof limit |
| P13 | A second device stops the first | T10 two-context e2e; T10 59/59 and gate | matched |
| P14 | The limit applies across courses | T10 two-course e2e; T10 59/59 and gate | matched |
| B01 | Empty course has an empty state and no player | T6 domain/HTTP/PostgreSQL evidence; Go/integration gate | matched |
| B02 | Lesson without video has an explicit state | T7 component test; frontend gate | matched |
| B03 | Processing video has an explicit state | T7 component test; frontend gate | matched |
| B04 | Failed video has an explicit state | T7 component test; frontend gate | matched |
| B05 | 95.0 completes and 94.9 does not | T3 domain boundary tests; Go gate | matched |
| B06 | NULL duration saves position without completion or clamp | T3 domain and PostgreSQL integration; Go/integration gate | matched |
| B07 | Shorter replacement clamps a saved position | `work/archive/report-b07.md`; the course postgres integration package green against real PostgreSQL, fix in `819dcac` | verified (gate) |
| B08 | Replacing video preserves the watched mark | T15 PostgreSQL integration: the completed mark survives eviction of the first asset (precondition asserted) | **verified** |
| B09 | 90 s of playing time produces exactly three writes | T9 frontend cadence test; frontend gate | matched |
| B10 | Pause, seek, hide and navigation each write | T9 frontend event tests; frontend gate | matched |
| B11 | Stale `clientSeq` cannot rewind progress | T9 PostgreSQL integration; integration gate | matched |
| B12 | Nginx is healthy and its cache stays empty | T4 smoke; smoke gate | matched |
| B13 | Three media requests cause three authorizations | T4 smoke counter; smoke gate | matched |
| B14 | Anonymous replay of an allowed URL returns 403 | T4 same-URL smoke transition; smoke gate | matched |
| B15 | The same segment changes 200→403 immediately after takeover | T10 e2e; T10 59/59 and gate | matched |
| B16 | An abandoned session does not delay a new play | T10 e2e; T10 59/59 and gate | matched |
| B17 | Repeated play gives the caller the fresh slot | T5 Redis/PostgreSQL integration; integration gate | matched |
| B18 | Network loss retains the last saved position | T9 e2e; T9 52/52 and gate | matched |
| B19 | The 120-lesson boundary renders all rows | T8 unit fixture; frontend gate | matched for rendering only |
| B20 | A 1,024-byte range returns 206 and exactly 1,024 bytes | T16 e2e: a served segment longer than 1,024 bytes answers `Range: bytes=0-1023` with 206, exactly 1,024 bytes and a matching `Content-Range` | **verified** |
| B21 | Three ranges cause three authorizations | T16 e2e: after pause and three quiet 500 ms counter intervals, three distinct ranges moved `courses_http_requests_total{route="GET /internal/playback/authorize"}` by +3 successful and 0 rejected (1400 → 1403, 7 → 7) | **verified** |
| B21a | Generated playlists contain no `EXT-X-BYTERANGE` | T16 e2e: both variant playlists of a published asset are non-empty, have no `#EXT-X-BYTERANGE`, and every `#EXTINF` is followed by its own distinct segment URI | **verified** |
| B22 | Media responses carry no-store and security headers | T16 e2e: a 200 on an allowed segment carries `Cache-Control: private, no-store`, `nosniff`, `strict-origin-when-cross-origin` and `DENY`, compared by value | **verified** |
| B23 | Media rate limiting returns 429 | T14 smoke: 720 concurrent requests under `/media/p/` must yield both 429 (the media zone sheds the excess) and 403 (admitted requests reach authorization, so the limit sits before it and does not reject everything); gate line `PASS media rate limit`. Diagnostic run on the same stack: 403 × 245, 429 × 475 | **verified** (T14) |
| B24 | Player is keyboard-operable and has no axe violations | T7 component/axe tests; frontend gate | matched |
| B24a | Focus is visible and controls are at least 44 px | T11 real-engine e2e in light/dark and before zoom; gate | matched |
| B25 | Contrast and 200% layout pass | T11 real-engine e2e; gate, with `readyState=0`, width/height 0/0 | matched only over empty video; see §6 |
| B26 | Locale base keys/forms agree and values are non-empty | T8 locale unit traversal; frontend gate | matched |
| B27 | Missing progress reads as concrete zero values | T6 domain/HTTP/PostgreSQL evidence; Go/integration gate | matched |
| B28 | DB `ready` with missing files gives edge 404 and visible error | T16 e2e: the asset directory is emptied from `media-worker` (existence and a `master.m3u8` proven before, an empty directory after); mint 201, edge 404 on master, and the named error state is visible | **verified** |
| B29 | A new session writes on its first heartbeat | T9 PostgreSQL/integration sequence seeding; integration gate | matched |
| B30 | Zero probe duration persists NULL and cannot complete | T2 worker/PostgreSQL integration; integration gate | matched |
| B31 | Returning caller retakes the slot and the displaced UI stops | T15 e2e extension of P13: retake mints, segment 200 for the returning device, 403 for the displaced one; the retry defect it exposed is fixed with a component regression | **verified** |
| B32 | Takeover removes byte access immediately | T10 same-segment e2e; T10 59/59 and gate | matched |
| B33 | Two tabs displace one another | T10 same-browser e2e; T10 59/59 and gate | matched |
| B34 | Restore alone cannot complete after a shorter replacement | Same scenario and proof as duplicate B41: T3 domain/T5 PostgreSQL; Go/integration gate | matched |
| B35 | `clientSeq=2^62` returns 400 without a write | T3/T5 validation and integration evidence; Go/integration gate | matched |
| B36 | Redis failure returns 503 with `Retry-After` and no bytes | T14 smoke: 503 with `Retry-After`. T16 e2e: a real served segment (299,957 bytes) is requested again with Redis stopped — 503 `text/html`, 592 bytes, containing neither of two 64-byte media windows proven present in the reference; recovery in `finally`, the same segment back to 200 | **verified** |
| B37 | Edge starts healthy with an empty cache | T1 smoke; final smoke gate | matched |
| B38 | Paced 400-request parallel run is 400/400 HTTP 200 | T1 measured 400/400, no false 403/429; final smoke gate | matched |
| B39 | Expiry remints silently rather than reporting takeover | T7/T9 e2e; gate | matched |
| B40 | Existing generation plays through replacement; reopen gets new | T7 e2e 200/no 403 or 404, then fresh generation; gate | matched |
| B41 | Position must grow before completion | T5 PostgreSQL integration; integration gate | matched |
| B42 | Growing to the end completes | T5 PostgreSQL integration; integration gate | matched |
| B43 | Invalid heartbeat/TTL ratio refuses API startup | T3 configuration unit test; Go gate | matched |
| B44 | Oversized position is accepted and clamped | T5 HTTP plus PostgreSQL integration; Go/integration gate | matched |
| B45 | Migration succeeds over populated data | T2 live PostgreSQL integration; integration gate | matched |
| B46 | Synthetic failing migration rolls back only itself | T2 live PostgreSQL integration; integration gate | matched |
| B47 | Lock timeout rolls back and retry succeeds | T2 live PostgreSQL integration; integration gate | matched |
| R01 | Anonymous playlists and segments return 403 | T1/T4 smoke; smoke gate | matched |
| R02 | Student cannot play a draft course | T5 real-mint smoke; smoke gate | matched |
| R03 | Student cannot play an archived course | T5 real-mint smoke; smoke gate | matched |
| R04 | Archiving stops a started student view | T7 e2e checks a previously served segment becomes 403; gate | matched |
| R05 | Expired ownership refuses the first segment | T5 real-mint, real-TTL smoke; smoke gate | matched |
| R06 | Another user cannot use the playback ID | T5 real-mint smoke; smoke gate | matched |
| R07 | Course visibility is rechecked per media request | T5 exact-ownership then inaccessible-course smoke; smoke gate | matched |
| R08 | Non-HLS media trees are closed | T5 smoke; smoke gate | matched |
| R09 | Traversal variants return 403/404 without bytes | T5 active-ID smoke; smoke gate | matched |
| R10 | Internal authorize is not externally routed | T1 smoke; smoke gate | matched |
| R11 | Unavailable Go returns 503 | T1 dependency-failure smoke; smoke gate | matched |
| R12 | Unavailable Redis returns 503 without bytes | T1 dependency-failure smoke; smoke gate | matched |
| R13 | Unavailable Postgres returns 503 | T1 dependency-failure smoke; smoke gate | matched |
| R14 | Displaced playback cannot bypass client stop | T10's stronger same-segment 200→403 proof; e2e gate | matched by stronger observation |
| R15 | Displaced heartbeat returns 409 and changes neither slot nor progress | T5 Redis/PostgreSQL integration; integration gate | matched |
| R16 | Non-owner heartbeat does not extend another owner's TTL | T5 Redis integration; integration gate | matched |
| R17 | Inaccessible-course heartbeat returns 403 and creates no progress | T6 distinct-student Redis/PostgreSQL integration; integration gate | matched |
| R18 | Unknown lesson returns 404 and the UI has a not-found state | T5 HTTP 404 plus T7 recovery/UI tests; Go/frontend gate | matched by split proof |
| R19 | Malformed lesson ID returns 400 before application lookup | T5 HTTP contract test; Go gate | matched |
| R20 | Malformed edge IDs return 403 without calling Go | T1 smoke counter; smoke gate | matched |
| R21 | Anonymous public course response exposes no lesson IDs | T6 HTTP contract test; Go gate | matched |
| R22 | Malformed heartbeat fields return 400 without changing progress | T5 HTTP/application/PostgreSQL evidence; Go/integration gate | matched |
| R23 | API and SSR network output stay far below video size | T16 e2e: a 100-second lesson played from 0.098 s to `ended` at 16× rate; raw `tx_bytes` of the same api and frontend containers against 17,208,719 bytes of delivered segment bodies — api +70,800 from before the page load, frontend +132 from after it | **verified** |
| R24 | The generated route set contains no “my courses” route | T8 exact route-set unit test plus generated check; frontend/generated gate | matched |
| R25 | Two accounts behind one IP play together | T10 e2e; T10 59/59 and gate | matched |
| R26 | One account from two network addresses still takes over | T10 mixed browser/container request using nginx source address; gate | matched |
| R27 | Logout changes the same segment from 200 to 403 | T5 real-mint smoke; smoke gate | matched |
| R28 | A displaced client stops sending progress | T9 frontend recovery test; frontend gate | matched |
| R29 | Anonymous and non-owner receive byte-identical 403 bodies | T4 smoke; smoke gate | matched |
| R30 | Nginx/Go safe logs cannot acquire request URI | T14: a host-side scanner in `infrastructure/tests/integration.sh` allowlists the variables of `log_format safe` and requires every `access_log` to name `safe` or be `off`, with at least one naming `safe`; the Go AST test `backend/platform/architecture/safe_request_logs_test.go` requires `"route", route`. Mutations went red in the conditions the gate runs them in. Limit: the scanner reads `nginx.conf` only, not `include`d files | **verified** (T14) |
| R31 | Unknown playback ID heartbeat returns 410 | T3 Redis/integration evidence; integration gate | matched |
| R32 | Deleted lesson heartbeat returns 404 without FK error | T5 application/HTTP/integration evidence; Go/integration gate | matched |
| R33 | Displaced heartbeat cannot write progress | T3/T5 Redis/PostgreSQL integration; integration gate | matched |
| R34 | Course owner plays their own archived course and receives segments | T15 e2e: a non-admin owner receives segment 200 on the archived course; another non-admin is refused on the same lesson | **verified** |
| R35 | hls.js ABR steps down under a narrowed channel | T11 e2e; two reviewer runs and gate | matched only as one step; see diagnostics and §6 |
| R36 | `/learn` leads to the catalog | T8 behavior e2e; T8 47/47 and gate | matched |
| R37 | Huge position clamps to the known 300 s duration | T8 dedicated e2e; T8 47/47 and gate | matched |
| R38 | Media 403 is classified by heartbeat before recovery/denial | T9 unit+e2e with O1 counters; T9 52/52 and gate | matched |
| R39 | A non-button pause advances persisted progress | T10 paired before/after learning-model e2e; T10 59/59 and gate | matched |

T11's mandatory diagnostics were:

- `B25 video at contrast audit`: `readyState=0`, `videoWidth=0`, `videoHeight=0`.
- `R35 master BANDWIDTH`: 360p **542300 bit/s**, 720p **3440800 bit/s**; derived
  throttle **170750 B/s** (**1365996 bit/s**).
- R35 run 1: baseline 360p=1, 720p=15; after throttle 360p=2, 720p=19.
- R35 run 2: baseline 360p=1, 720p=5; after throttle 360p=2, 720p=8.
- The decisive request sequence was `720p/segment_00024 → 720p/segment_00025 →
  360p/segment_00025`: abandon-and-retry of the same segment index at the lower rung.

T12's live metric scrape found exactly three pre-seeded zero series: `probe_failed`,
`supported` and `unsupported`. Its Help text counts browsers whose bounded probe reached the
mint endpoint, including denied mints. The live API run did not itself observe a 0→1 mint or
a repeated mint without the field; those transitions are the named Go and frontend tests in
the ordered unit/integration portions of the gate.

## 3. Non-obvious findings

- **The plan's own count drifted.** Its “94 rows” footer hides eight rows; direct enumeration
  yields 102. The delivery count therefore uses the IDs, not the stale footer.
- **A read-only proof hid a progress-write authorization bug.** T6's review replaced it with
  a real heartbeat attempt by a distinct student and moved the shared visibility check before
  Redis renewal and PostgreSQL progress write.
- **An edge denial needs a before-state.** T4 and T10 found that a bare 403 could pass because
  bytes were already unreadable for another reason. The lasting proofs use the same URL as
  200 before logout/takeover and 403 after it.
- **An ABR request is not a settled quality level.** The repeated segment index is strong
  evidence of a downward abandon-and-retry, while the `+1` lower-rung count is insufficient
  evidence that playback remained there.

## 4. Screenshots

No static screenshots were captured. Browser-visible behavior was exercised by the Compose
Playwright suite; T11 recorded computed theme, geometry, zoom, axe and video-state diagnostics
rather than treating a screenshot as an assertion.

## 5. User flow walk-through

Verified by the named browser scenarios, not by hand. A signed-in student starts in the public
catalog, opens a published course and enters its learning route at the server-selected lesson.
The outline shows modules, lessons and watched marks. Pressing play mints a session and real HLS
starts; quality can be left on Auto or chosen manually in hls.js. Progress is written while
playing and on exit events, restored after reload and in another browser context, and reflected
in the outline. At 95% the lesson becomes watched; reopening it starts at zero. If the same
account starts another playback, the former view stops and explains why. Archive and logout
remove byte access on the next checked request. Empty, unavailable and processing media states
remain readable UI states rather than blank pages.

## 6. Risks and limitations

- **Native HLS is unproven (D-9).** Production checks `Hls.isSupported()` first, and both
  Chromium and Playwright WebKit measured `{hlsJsSupported:true,mse:true,canPlayType:"maybe"}`;
  both therefore used hls.js. Playwright WebKit is not Safari. A real Safari run is required,
  and no claim is made that the native branch works.
- **A native fallback is dead and uncovered.** No test reaches the
  `return createNativeMedia(...)` after a failed `canPlayType` in `hls-media.ts`.
- **B25 audited an empty video element.** The course read model was supplied, but mint/media
  were not, so the measured element was `readyState=0`, `videoWidth=0`, `videoHeight=0`.
  Contrast over actual playing imagery was not tested. The earlier prediction that a video
  background would create an axe `incomplete` was not confirmed because the fixture never
  created that state.
- **One axe `incomplete` remains:** `color-contrast`/`nonBmp` for React Aria Select's
  decorative chevron. It is axe declining to decide a glyph-only node, not a measured contrast
  defect. Remove the pin when axe can decide non-BMP glyph contrast, or when an updated UI kit
  changes the chevron DOM so the current stable suffix no longer identifies it.
- **R35 proves a step, not settling.** The channel is restored immediately after the first
  360p request, and both runs add exactly one 360p request. The supported claim is “ABR steps
  down under a narrowed channel,” not “the player transitions to and stays at 360p.”
- **R35 uses a synthetic channel.** The **170750 B/s** throttle is the geometric mean of the
  two master-playlist BANDWIDTH values, not a real network condition. It proves reaction to a
  controlled capacity between rungs, not correctness on real networks.
- **Ownership timing was compressed.** Playback e2e recreates the API at heartbeat/TTL **2/6**,
  while the shipped Compose defaults are **30/180**. Both satisfy `heartbeat*3 <= ttl`, but
  renewal, expiry and takeover behavior was not timed at the shipped cadence.
- **Four start-position rows measured the host, not the player, until 2026-10-07.** P04
  (reload), P05 (another browser context), B18 (a new page after a network break) read
  `currentTime` after `openAndPlay` had already waited for the Pause label and polled
  `videoTime`, so their `[39, 41]` window was really "the read happens within ~1 s of playback
  starting"; P07 ("reopens from zero", `<= 1`) had the same shape. A local run beside a
  format/typecheck job measured **41.37 s** on P04 and failed; a 2 vCPU CI runner (K10) would fail
  these routinely. The four rows now capture `currentTime` at the `playing` event
  (`armStartPositionProbe`/`startPosition`: an init script armed on the page that will navigate,
  capture phase because `playing` does not bubble) and assert that value in the same windows —
  the start position itself, independent of how fast the test process reads afterwards. The
  product did not change; the proofs moved onto the observable the rows name. P04 fell first;
  the review (Courses 2) named P05/B18 as structural twins, and the scan added P07. A clean full
  local run afterwards is a regression check (all rows still pass), not proof that the flake is
  gone — the original needed concurrent load a quiet host does not reproduce; the proof of that is
  the mechanism (a capture at the event is load-independent by construction) and the 2 vCPU CI run.
  That run (37655371086, 2026-10-07) then showed a second shape in **P05**: the first page keeps
  playing and recording a tick every 2 s on the e2e profile while the second context is opened,
  so the "saved position" was a moving target — the second page started at **41.03 s** at
  `playing`, correctly restoring what the first page had saved meanwhile. P05 and B18 now pause
  the first page and anchor the expectation on the position the pause heartbeat recorded
  (`pauseAndRecord`, ±1 s), i.e. on the value a restore has to reproduce, not on the number the
  seek asked for. P04 reloads immediately after the seek (a `nav` heartbeat freezes the value) and
  P07 restarts from zero; neither has a page left playing.
- **AV1 telemetry begins only at the mint endpoint.** It includes denied mints, but a client
  consumes its page-scope probe before sending; if that request never reaches the server, the
  observation is lost. This narrow loss window avoids double counting when a response is lost.
- **P12 has no single wire-crossing test.** Frontend proves the field is sent on a denied mint;
  backend proves a received field increments the series; shared `openapi.yaml` binds the two.
- **All 102 planned observations are verified as of 2026-09-14** (T14–T16 and B07, recorded in the dated
  updates of §2). At T13 thirteen were not: B07, B08, B20, B21, B21a, B22, B23, B28, B31, B36, R23, R30 and R34.
  Two of those, B07 and B31, were product defects rather than missing tests, and six were moved to `e2e` by
  amendment `8cc1d72`. The limitations in this section are unaffected by that closure.
- **One audio rendition only; no AV1 rendition.** Multi-language audio delivery is not built,
  and the AV1 work is telemetry/probe only.
- **Still out of scope:** payments, orders and enrolment; transcripts; comments/discussions;
  instructor workspace; “my courses”; notes, bookmarks, search, offline mode and certificates;
  frontend observability; TLS/S01/S02; and production deployment. None is claimed here.

The phase is closed by its rows. The next task is to rebuild `work/AGENDA.md` and
`work/CTO-RESEARCH.md` — both last rebuilt on `b913e81`, before the close, and both still calling
the phase open — then take the selected Phase 7 slice through `spec-flow`. The roadmap currently names
transcripts and synchronized document links, with student video comment threads added by
`work/PRODUCT-BRIEF.md` §24.4; none of that roadmap work is delivered by this phase.
