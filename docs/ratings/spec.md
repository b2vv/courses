# Course ratings and reviews — spec (draft)

Status: **draft, not started.** Scheduled **after the payment phase closes** (owner, 2026-09-30).
Next step when it starts: GATE 1 (`grilling`) on this file, then `plan.md` (spec-flow).

## Problem

A visitor choosing a course has no signal of its quality. The storefront mockup shows
`★ 4.9 (1284)` on every card; we have no ratings, and showing invented numbers would destroy
the very trust the signal is meant to give. Ratings must come from real students.

## Decisions taken (owner, 2026-09-30, in the design session)

Each closes an alternative; each has a way to check it once built.

| # | Decision | Alternative closed | How to check |
|---|---|---|---|
| R1 | **Who may rate:** a student who can access the course **and has started it** (at least one lesson with progress). | Anyone signed in (easy to game); only after completion (too few, too late). | A rating request from a user with no progress on that course is refused with a business error code. |
| R2 | **Steam-style context:** each review shows **how much of the course its author had completed when writing it** (e.g. «пройдено 35% на момент відгуку»). The percentage is captured at write time and updated if the review is edited. | Hiding progress; showing current progress (changes meaning after the fact). | The stored percentage equals completed lessons / published lessons at the moment of the write. |
| R3 | **Format:** 1–5 stars **plus a text review**; one review per student per course, editable. Text reviews are **moderated**, reusing the instructor moderation workspace (queue, remove, audit). | Stars only. | A removed review disappears from the course page and from the average; the audit log records who removed it. |
| R4 | **Display threshold:** the card shows average and count only from **3 ratings**; below that it shows «Новий курс». | Showing from the first rating (★ 5.0 (1) overstates). | A course with 2 ratings renders «Новий курс»; with 3 it renders `★ avg (3)`. |
| R5 | **Order:** after payment (O05–O07) closes. | Before O05. | — |

## Also in scope (same catalog-contract change)

The mockup card shows lesson count and duration. The catalog `Course` contract has neither
today (`backend/openapi/openapi.yaml` → `Course`). Add `lessonCount` (published lessons) and,
if the data allows, total video duration — real values only.

## Out of scope

Category and level on the card (no such fields on a course; would be a separate authoring change).
Instructor replies to reviews. Helpful/unhelpful votes. Review sorting beyond newest first.

## Open questions (for GATE 1)

1. Does an **archived** course keep and show its ratings?
2. Can the **course owner or an admin** rate their own course? (Proposed: no.)
3. When the course gains lessons later, the author's stored percentage stays as captured (R2) —
   confirm that is the intended reading.
4. Minimum/maximum review text length; is a star rating allowed without text?
5. Does a refund/revoked grant remove the right to keep a review?
