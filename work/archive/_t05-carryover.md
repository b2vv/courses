# T05 carry-over (from reviews)
- T03 review №1: integration test that a moderation-log row with BOTH or NEITHER target is rejected by the one-target CHECK.
- T03 review №2: RemoveReview on an already-removed review returns ErrNotFound (indistinguishable from missing) — T05 picks the HTTP status for re-remove (check what comment re-remove does and match it).
- T01/T02 review: rating.average 5.0 marshals as 5 — display formatting is the client's (T06), but add a contract test that average is a number.
- T04 review №1/№4: average formatted to one decimal on the client (T06); avoid float == in new average tests (compare tenths or with tolerance).
