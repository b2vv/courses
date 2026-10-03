# Course ratings and reviews — spec

Status: **GATE 1 passed 2026-10-03** (owner, `grilling`; trace in the Obsidian session note
«2026-10-03 Phase ratings — GATE 1» and in `assumptions.md`). Next: `plan.md` → GATE 2.
The payment phase closed on 2026-10-03 (PR #3), so R5 is met. The project stays in test mode with no
production server (owner, 2026-10-03): everything here is built and proven locally.

## Problem

A visitor choosing a course has no signal of its quality. The storefront mockup shows
`★ 4.9 (1284)` on every card; we have no ratings, and showing invented numbers would destroy
the very trust the signal is meant to give. Ratings must come from real students, and the course's
own author must not be able to shape them.

## User scenarios

1. **A student rates a course they are taking.** A student who has started a free course, or who
   holds a grant for a paid course and has started it, opens the course and leaves 1–5 stars with a
   text review. The review appears on the course page at once, with «пройдено N% на момент відгуку».
2. **A student changes their mind.** The author edits the stars or the text; the percentage is
   recaptured at the edit. Or the author deletes the review and later writes a new one.
3. **A visitor chooses a course.** Anyone, signed in or not, sees on the catalog card and on the
   course page the average and the count once a course has 3 ratings, or «Новий курс» before that,
   and reads the reviews on the course page, newest first.
4. **Someone reports an abusive review.** A signed-in user, the course owner included, reports a
   review; it enters the existing moderation queue. An admin removes it: it disappears from the page
   and from the average, and the audit log records who removed it.
5. **An admin or an instructor takes someone else's course as a student.** An admin buys a paid
   course (the buy path is shown to an admin too) or an instructor buys a colleague's course; after
   starting it, they rate it like any student.
6. **Refused attempts.** The course owner tries to rate their own course; a user who only watched a
   free preview of a paid course tries to rate it; an admin with only role access tries to rate; a
   user with no progress tries to rate — each is refused with a business error code.

## Acceptance criteria

Each is observable; `plan.md` maps them to acceptance rows and tests.

- **AC1 — eligibility (R1, G1, G2).** A review is accepted only from a user who has progress on the
  course (at least one lesson with a `lesson_progress` row) **and**, for a paid course, an active
  grant. Role access (admin) and preview access do not count. The course owner is refused even with
  a grant. Each refusal returns a distinct business error code.
- **AC2 — no self-grant (G5).** An admin cannot issue a grant to themselves: `GrantCourse` with
  `userID == actor` is refused with a business error code and writes nothing.
- **AC3 — admin can buy (G2).** An admin opening a paid course they hold no grant for is offered the
  buy path and can complete a purchase; afterwards the course counts as theirs for AC1.
- **AC4 — format (R3, G7).** Stars are an integer 1–5; text is required, 20–2000 characters after
  trimming. One review per user per course; a second write is refused, an edit is allowed.
- **AC5 — progress context (R2, G6).** The stored percentage equals completed lessons
  (`completed_at` set) / published lessons at the moment of the write or edit, and does not change
  when the course later gains lessons or the author progresses further.
- **AC6 — edit and delete (R3, G4, G10).** The author can edit while holding the right of AC1, and
  can delete at any time; a deleted review leaves the page and the average, and the author may write
  again under AC1. An author who lost access keeps the review but cannot edit it.
- **AC7 — moderation (R3, G8, G9).** A new review is visible and counted at once. A report puts it in
  the existing queue. Only an admin can remove a review; the course owner cannot. A removed review
  leaves the page and the average, the audit log names the remover, and its author cannot replace
  it with a new one.
- **AC8 — display threshold (R4).** With fewer than 3 counted reviews the card and the course page
  show «Новий курс»; from 3 they show `★ avg (count)`. Counted = not removed and not deleted.
- **AC9 — public reading.** The course page lists reviews to anyone, signed in or not, newest first,
  each with the author's display name, stars, text, the captured percentage and the date.
- **AC10 — archive (G3).** Archiving changes nothing for reviews: they stay stored and visible to
  those who can see the course, writable under AC1; re-publishing shows the rating on the card again.
- **AC11 — catalog contract.** The catalog `Course` gains the rating summary and `lessonCount`
  (published lessons); total video duration only if the data gives a real value.
- **AC12 — no derived table (R6).** The summary is computed at read time; the migration adds no
  summary table, counter or trigger.

## Decisions taken (owner, 2026-09-30, in the design session)

Each closes an alternative; each has a way to check it once built.

| # | Decision | Alternative closed | How to check |
|---|---|---|---|
| R1 | **Who may rate:** a student who can access the course **and has started it** (at least one lesson with progress). Refined by G1 and G2. | Anyone signed in (easy to game); only after completion (too few, too late). | AC1. |
| R2 | **Steam-style context:** each review shows **how much of the course its author had completed when writing it** (e.g. «пройдено 35% на момент відгуку»). Captured at write time, recaptured on edit. | Hiding progress; showing current progress (changes meaning after the fact). | AC5. |
| R3 | **Format:** 1–5 stars **plus a text review**; one review per student per course, editable. Text reviews are **moderated**, reusing the instructor moderation workspace (queue, remove, audit). | Stars only. | AC4, AC7. |
| R4 | **Display threshold:** average and count only from **3 ratings**; below that «Новий курс». | Showing from the first rating (★ 5.0 (1) overstates). | AC8. |
| R5 | **Order:** after payment (O05–O07) closes. Met 2026-10-03. | Before O05. | — |

## Decisions taken (owner, 2026-10-03)

| # | Decision | Alternative closed | How to check |
|---|---|---|---|
| R6 | **No derived (summary) table for ratings.** The average, the count and `lessonCount` on the catalog card are computed at read time by an aggregate query (`GROUP BY course_id` over reviews that are not removed, backed by an index on `course_id`). | A stored summary table (`course_rating_summary` with avg/count), a trigger-maintained counter, or a materialized view. | AC12. |

**R6 — why (mechanism, not a guess):** under R3 the average must change both when a student edits a review
and when a review is removed. Removal happens in the moderation workspace of the `discussion` module, so a
stored summary would need a second writer from that module, or a cross-module call, which ADR 013
forbids: a module reads only its own facts. Any write path that forgets to update the summary leaves
a stale average in the catalog. A read-time aggregate cannot go stale. The volume is small: the
catalog publishes a handful of courses.

**R6 — trigger to revisit (magnitude), revised 2026-10-03 by the owner:** measure the production catalog
list query (`ListPublishedCourses`, default arguments: empty search, sort `recent`, page 20) with
`EXPLAIN (ANALYZE, BUFFERS)` on the data volume `plan.md` §9 fixes, after `VACUUM ANALYZE`, median of 5.
If the median execution time exceeds **5 ms**, take the first step up: a materialized view refreshed after
review writes and removals. A summary table comes only after that view proves insufficient. The first
version of this trigger — "more than 2× the same query without the aggregate" — was replaced because its
baseline is an index scan of ~16 µs, so any join exceeds it many times over and the ratio measures nothing
(measured 2026-10-03: 1.306 ms vs 16 µs, 81.6×; `assumptions.md`). "Derived table" in the SQL sense (a
subquery in `FROM`) is a query technique and stays allowed. This decision does not cover it.

## GATE 1 decisions (owner, 2026-10-03)

Facts behind each were read from the code during the gate; the references are where they live.

| # | Decision | Alternative closed | Fact it rests on |
|---|---|---|---|
| G1 | A **paid** course is rated only by a holder of an active grant (purchase, or admin-issued); a **free** course by anyone with progress. Watching a free preview gives no right to review. | Any progress counts, preview included (R1 read literally). | Preview lessons are open to anyone (`contracts/access/access.go:22`); progress is written for preview lessons too (`lesson_progress`, `migrations/00006_playback.sql`). |
| G2 | **Role does not decide — the student path does.** Role access (admin) does not count as a right; the course owner never rates their own course, even with a grant; an admin who bought a course and an instructor on a colleague's course rate as students. The buy path is shown to an admin. | Admin rates any course; owner rates with a grant; role exemption counts as a right. | One role per account (`users.role` CHECK, `migrations/00002_identity.sql:15`); owner = `course.instructor_id`; admin is exempt from purchase (`access.go:31`) so a locked lesson — the only buy entry (`LockedLessonPanel.tsx`) — is never shown to them; the backend does not forbid an admin's purchase (`orders/application/service.go:43`). |
| G3 | **Archive changes nothing** for reviews (kept, visible to those who see the course, writable under AC1; rating returns on re-publish). | Freeze reviews of an archived course. | archived↔published is reversible (`course/domain/authoring.go:231`); grant holders still see an archived course (`access.go:31`). |
| G4 | **Losing access later** (refund, revoked grant) keeps the review and its weight in the average; the author can no longer edit it. Only moderation removes it. | Hide reviews whose author has no active right. | `revoked` exists in the schema (`migrations/00011`) but nothing revokes today; refunds are ignored (`docs/payment/report.md` §6). |
| G5 | **No self-grant:** an admin cannot issue a grant to themselves (new business code in `GrantCourse`). | Self-grant allowed but not counted for reviews; self-grant counted. | `GrantCourse` (`orders/application/service.go:131`) has no such check today. |
| G6 | The R2 percentage is **frozen at write time** and recaptured only on the author's edit. | Showing current progress. | Completed = `completed_at` set at ≥95% of the duration (`playback/domain/progress.go:74`). |
| G7 | Stars and text both **required**; text **20–2000** characters. | 1–2000 like a comment; optional text. | Comment body is 1–2000 runes (`discussion/domain/comment.go:11`). |
| G8 | **Post-moderation**, as for comments: visible and counted at once; a report sends it to the existing queue. | Pre-moderation (pending until approved). | The queue is built on reports (`InsertReport`/`ListQueue`, tasks I02/I03). |
| G9 | **Only an admin removes** a review; the course owner can report it. Comments keep owner∨admin. | Owner∨admin as for comments. | The workspace guard is owner∨admin (`CanBeModeratedBy`, task I01). |
| G10 | The author can **delete** their review and write again under AC1; a review removed by an admin cannot be replaced. | Edit only, never delete. | — |

Taken as stated by the owner at the close of the gate: reviews are public (anonymous visitors
included), newest first; the 3-rating threshold applies on the course page as well as the card;
a review shows the author's display name, stars, text, captured percentage and date.

## Out of scope

Category and level on the card (no such fields on a course; would be a separate authoring change).
Instructor replies to reviews. Helpful/unhelpful votes. Review sorting beyond newest first.
Propagating refunds to access (a named payment gap, `docs/payment/report.md` §6). Production
deployment (deferred by the owner, 2026-10-03).

## Open questions

None left from the draft: questions 1–5 were closed by G3, G2, G6, G7 and G4. Questions that arise
while planning go to `plan.md` and, if they foreclose an alternative, to `assumptions.md`.
