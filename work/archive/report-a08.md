# A08 report — UI-kit i18n debt and locale split

## Files

- `frontend/src/presentation/ui/data.tsx` — made `Table.emptyMessage` required and made
  all five `Pagination` labels caller-supplied: navigation, previous, next, current-page and
  empty-page labels.
- `frontend/src/presentation/ui/states.tsx` — made `Skeleton.label` required and made
  `ErrorState.retryLabel` required whenever `retryHref` is present; removed both Ukrainian
  defaults.
- `frontend/src/presentation/ui/ui.test.tsx` — updated the existing pagination caller and
  added P17 evidence covering English pagination, empty-table, skeleton and error/retry
  output, including the pagination `nav` aria-label and a no-Cyrillic assertion over text
  and markup.
- `frontend/src/app/providers/locales/uk.ts`, `en.ts` — moved the two resource bundles out
  of `i18n.tsx`; preserved existing keys and added only the list vocabulary named by the
  plan: courses, title, slug, status, owner, updated time and the four status-tab labels.
- `frontend/src/app/providers/i18n.tsx` — now contains only per-request i18next creation and
  the React provider, importing the split bundles.
- `frontend/src/app/composition/locale.ts` — now owns the existing `Locale` type and locale
  normalization/cookie detection alongside `detectLocale`.
- `frontend/src/presentation/LanguageSwitcher.tsx` and
  `frontend/src/routes/route-guards.test.tsx` — followed the moved `Locale` type.
- `frontend/src/test-setup.ts` and `frontend/src/app/providers/i18n.test.tsx` — import the
  split resource bundles directly; the i18n test records uk/en key parity for shared UI and
  administration vocabulary.
- `work/report-a08.md` — this report.

## Decisions

- The kit remains i18n-free. Callers pass already-translated strings; `Pagination` does not
  receive a translator or construct language-specific page text.
- `Pagination.pageLabel` is a required, fully formatted caller value. Its empty alternative
  is the separately required `emptyLabel`, so the component owns only state selection.
- `ErrorState` uses a prop union: retry-less states remain valid, while a caller that supplies
  `retryHref` must also supply `retryLabel`. The component owns no fallback language.
- No existing translation key was renamed. Only short plan-named admin-list vocabulary was
  added; A10/A11 still own their screens' complete copy.
- The existing optional `lines` behavior on `Skeleton` is unchanged. There are no visual,
  layout or styling changes and no new component or dependency.

## Verified / how

- Read-through traced the split bundles into `createI18n` and global test setup, and traced
  the `Locale` type/helpers through locale detection, the language switcher and route tests.
- Repository-wide `rg` over `frontend/src` and `frontend/e2e` found every `Table`,
  `Pagination`, `Skeleton` and `ErrorState` reference. Both production skeleton callers and
  both production error-state callers already pass labels through `t(...)`; the UI tests now
  pass every required prop. There is no current production table or pagination caller.
- A read-only Python Unicode scan found no Cyrillic code point in `data.tsx`, `states.tsx`,
  `en.ts` or `i18n.tsx`. The same script inspected every JSX caller and found no missing
  required label prop.
- Manual P17 read-through confirms both pagination display states are represented, all five
  caller strings are asserted in rendered content or accessible names, table empty text,
  skeleton label and error retry label are English, and the markup assertion includes both
  the retry link and `nav` aria-label.
- The standard Git-based `code-review` skill workflow was denied because it requires Git
  fixed-point and diff commands. Brief A08 says never touch Git; no workaround was attempted.
- No file inside a `generated/` directory was written. No filesystem write was denied.

## Not run

Environment denial: **No Docker and no local test run.** No workaround was attempted, and
none of the following is represented as passing:

- Targeted Rstest for `presentation/ui/ui.test.tsx` and `app/providers/i18n.test.tsx` — not
  run; P17 is written evidence only.
- `make format-check` — not run.
- `make lint` — not run.
- `make typecheck` — not run.
- `make test` — not run.
- `make integration` — not run.
- `make check-generated` — not run; it also touches Git, which is separately forbidden.
- `make build` — not run.
- `make up`, `make seed-identities`, `make seed` and `make smoke` — not run.
- `make e2e` — not run.
- Aggregate `make check` — not run.
- No formatter, compiler, test runner, generator or Docker command was run.
- No Git command was run; Git state was not inspected or changed, and no commit was created.

## Reviewer verification (measured in the reviewing session)

One review return: the first pass left `ErrorState.retryLabel` defaulting to Ukrainian —
caught by an independent whole-kit Python audit (Codex's targeted audit covered only the
three named components). After the fix: kit-wide Cyrillic count 0 (python, not grep);
`make test` pass (0 failed frontend tests), `make typecheck`, `make format-check`,
`make lint` pass. Provenance: implemented with Codex (gpt-5.6-sol); reviewed and
verified here.
