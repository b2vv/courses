# Ratings phase — assumptions and decisions

Append-only by meaning: a correction is a new entry referencing the earlier one.

## 2026-10-03 — GATE 1 passed (owner, `grilling`)

**Decision (owner, 2026-10-03):** the ratings spec is agreed. Ten decisions G1–G10 were taken in the
gate, each with the code fact it rests on; they live in `spec.md` §«GATE 1 decisions», and the
acceptance criteria AC1–AC12 follow from them. **Alternatives closed:** listed per decision in that
table (e.g. preview counts as a right, owner moderates reviews, pre-moderation, optional text).
**Scope it adds beyond reviews:** G5 changes `GrantCourse` in the closed payment module (no
self-grant); G2 shows the buy path to an admin. **How to check:** each AC maps to an acceptance row
in `plan.md` and to a test; GATE 2 refuses a plan that leaves an AC unmapped.
