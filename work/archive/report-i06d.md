# Report — I06d

Signed-in header and logout.

## What was implemented

- Added the prop-driven `IdentityHeader`. Anonymous visitors see the login link. Signed-in
  visitors see their complete display name as text and a React Aria `Button` for sign-out;
  instructors and administrators also see the administration `Link`.
- The root `Layout` observes `['session']` with the required `refetchOnMount: false` and leaves
  `refetchOnWindowFocus` at its default.
- The root owns the logout mutation and the failure `Toast`. Both 204 and failure outcomes navigate
  to `/courses` and replace the cached identity with `null`; a failure additionally leaves a
  localized uncertain-revocation message mounted after navigation.
- Catalog navigation is awaited before publishing `null`, so a protected route's mounted
  `useRedirectWhenSessionEnds` observer has unmounted and cannot race the intended destination.
- Added Ukrainian and English header, pending and logout-failure copy. The existing `Toast` now
  localizes its visible dismissal label and accessible name through its existing resource keys.
- Added wrapping constraints for a 120-code-point name and reused the UI kit's semantic colors,
  visible focus ring and 44px control minimum.

## Acceptance rows

| Row | Evidence added in I06d | Deferred browser half |
|---|---|---|
| G19 | `G19 renders the complete display name` covers 1 and 120 code points; header CSS permits wrapping without a fixed width or clipping | 200% layout observation in I06g |
| G20 | Parameterized `G20 renders the … header state` covers anonymous, student, instructor and admin, including the administration-link matrix | Live role/header navigation in I06g |
| G24 | `G24 signs out…` starts on `/learn`, observes only the guard's initial `/auth/me`, invokes logout, reaches `/courses`, clears the cache, renders the anonymous header, then proves `/learn` redirects again | Live cookie expiry and full Compose flow in I06g |
| G25 | `G25 ignores a second sign-out…` holds a request open, proves the pending button is disabled and observes one request after a second activation attempt | Repeated logout and two-tab browser contexts in I06g |
| G26 | Parameterized `G26/G31 ends the … local session…` covers 503 in uk/en, cache clearing, catalog navigation, anonymous header and the persistent localized toast | PostgreSQL-stop/recovery flow in I06g |
| G29 | `G29 reaches and activates sign-out…` uses `user-event` to prove the text name is skipped and the button is keyboard-operable | Visible-focus browser half in `accessibility.spec.ts` in I06g |
| G30 | Markup reuses UI-kit `Button`/`Link`; their `--control-min-size` is 44px, and the new CSS uses wrapping plus semantic tokens only | Browser measurement at 200% zoom in dark theme in I06g |
| G31 | Header tests cover Ukrainian states plus English anonymous/signed-in actions; the 503 route test is parameterized across uk/en, and every new string has both resources | Integrated browser locale coverage remains I06g |

## Verification

The required targets were attempted in order from `/Users/strelia/projects/courses`. A failed
target was not reported as passed, and no target was modified to bypass the environment.

| Step | Result |
|---|---|
| `make format-check` | pass |
| `make lint` | pass |
| `make typecheck` | pass |
| `make test` | pass — frontend 171 tests / 19 files / 0 failures; all Go packages ok |
| `make integration` | pass — 12 Go packages, 0 failures |

Supplemental evidence:

- Before the review correction, focused Rstest runs passed the four-state component test, the
  G24 route test, and then 33 tests across `IdentityHeader.test.tsx`, `route-guards.test.tsx` and
  `ui.test.tsx` with zero failures.
- After the final review correction, `npm run format:check`, `npm run lint`,
  `npm run typecheck` and `git diff --check` pass. The final focused Rstest rerun was attempted but
  hit the same `listen EPERM … ::1:3000` before collection, so the final test revision is not
  represented as executed.
- `make build` was not run; it is not requested by the I06d brief.
- `make e2e` was not run; the brief assigns browser evidence to I06g.

## Review

The required two-axis review found no documented-standard violation and no remaining spec issue.
The spec review initially found that G24 began on `/` and therefore did not exercise the mounted
protected-route subscription. The implementation and test were corrected as described above, and
the reviewer confirmed the finding resolved.

The standards review's only judgment call was duplicate authenticated-user fixture builders in
component and route tests. No shared helper was introduced: the same tiny local shape already
exists in a third domain test, and a new cross-suite fixture module would increase files and
coupling for no production behavior.

## Deviations and limitations

- No behavior deviates from the fourth-edition plan. The review clarified the ordering needed to
  fulfill its explicit no-race requirement.
- The browser halves named above remain deliberately with I06g. In particular, happy-dom does not
  measure 44px targets, text overflow, dark-theme color rendering, visible focus or 200% zoom.
- The header link is navigation convenience, not an authorization boundary. Server-side role
  enforcement remains Phase 4.
- A focused commit was attempted with message
  `feat(identity): add signed-in header and logout`, but Git could not create
  `/Users/strelia/projects/courses/.git/index.lock`: `Operation not permitted`. No commit was
  created. The supplied untracked `work/brief-i06d.md` remains unchanged.

## Next task

I06e adds the final `/learn` and `/admin` empty-state screens; I06g owns the Compose browser halves
listed above after I06e is present.


## Verified from the reviewing session

Re-run unsandboxed, since the implementer's environment denies Docker and a localhost port bind.
Frontend tests rose from 158 to 171.

Checked by reading rather than by trusting the report:

- `refetchOnMount: false` is on the header observer (`__root.tsx:86`), which is what G07 needs. It
  now appears exactly twice in production code, here and in `require-identity.ts:31`.
- Sign-out settles both outcomes in one `onSettled`: both navigate to `/courses` and both clear
  `['session']`, with the navigation in a `try/finally` so the cache is cleared even if navigation
  throws. The order is navigate-then-clear, which is what keeps `useRedirectWhenSessionEnds` from
  firing its own redirect on the way out — the race the plan warned about.
- The `Toast` renders from `Layout`, outside the content that navigates, so the failure message
  survives the move to a public page. That was the hole `plan-defense` found in edition 4.
- The administration link is gated on `canAccessAdministration` (`IdentityHeader.tsx:19`), the
  domain predicate, not a role literal.

## Known inconsistency, deliberately left

`Toast` was changed from hardcoded Ukrainian to i18n keys, because the header renders it and G31
requires both locales. That makes `overlays.tsx` the **only** file in `presentation/ui/` that uses
`useTranslation`; `data.tsx` still carries hardcoded Ukrainian in `Table` and `Pagination`.

The inconsistency is real, and it is left as it stands: the change is the correct direction and
the rest of the kit is what is wrong. Reverting an improvement to match a defect would be the
wrong trade. `Table` and `Pagination` carry the same latent G31-class defect for any future
feature that renders them in English — recorded here rather than fixed, because no row in this
slice touches them.
