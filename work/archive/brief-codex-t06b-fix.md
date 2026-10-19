# T06b follow-up 1 of 2 — fixes only (tests come in the next run)

Keep logging to `work/codex-t06b.log`. Do not commit. Only these items; nothing else.

1. **i18n in the teach review row** (`routes/teach.courses.$courseId.tsx` `ReviewModerationRow`): no literal
   user-facing strings — «Відхилити скаргу», «Зняти», «Скарг: N», the former-author fallback — use `t(...)` keys
   in `locales/uk.ts` and `en.ts` (reuse `ratings.formerAuthor`; plural for the report count). Format the row
   like the comment `ModerationRow` (multi-line JSX, the file's style); stars as `aria-hidden` glyphs + sr-only
   «N з 5» (as `RatingSummary`/the review card do).
2. **The five items in `work/_t06b-carryover.md`** (from the T06a review), applied in the shared review list
   component you extracted:
   - section error only when there is no data; a failed next page keeps the items and shows an inline retry
     (TanStack v5 sets `isError` for a failed `fetchNextPage` too — use `isFetchNextPageError`);
   - keep the polite live region mounted; announce «Завантажено ще N відгуків» / «Усі відгуки показано»; when the
     last page arrives move focus to the first newly added review (or the list), never leave it on `<body>`;
   - own-review query key includes the signed-in user id;
   - clear `reportError` on a later success; pending label or `aria-busy` on the report button while sending;
   - «змінено» shows the edit date.
3. The dismiss/remove `reason` sent from the teach row: do not hard-code `'moderation_reviewed'` as if a human
   wrote it — check what the comment moderation row sends today and do the same.

Run `npm run typecheck && npm run lint && npm run format:check`. Final answer: what you changed per item.
