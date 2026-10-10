# Report M03 — media domain and ports

## Changed

- `backend/internal/media/domain/asset.go` — adds `VideoAsset`, the six-state vocabulary,
  typed claim tokens and generations, the five closed failure reasons, the explicit T1–T9
  transition matrix and writers, the 1-byte through 2-GiB ticket bound, the 255-character
  original-name bound, `NewAsset` invariants, and opaque storage keys with four kinds.
- `backend/internal/media/domain/actor.go` — adds the capability-only
  `Actor{UserID, Administrator}` and the admin-or-owner `CanEditLesson` policy.
- `backend/internal/media/domain/source.go` — adds source probes, the Q2 H.264/HEVC and
  AAC/MP3 codec set, source acceptance, and the 360p/720p/1080p ladder selected by source
  height with a mandatory 360p floor rung.
- `backend/internal/media/domain/errors.go` — adds the §8 media error vocabulary plus the
  internal transition and unsupported-source refusals used by the pure domain.
- `backend/internal/media/domain/*_test.go` — table coverage for every legal and refused
  state pair under T1–T9, transition writers, asset construction, ticket bounds, actor policy,
  codec acceptance/refusal, and landscape, portrait, sub-360p, exact-360p and 4K ladders.
- `backend/internal/media/ports/{repository,storage,transcoder,queue}.go` — adds exactly the
  four planned dependency ports. `AssetRepository` contains ownership-sensitive `Mint`,
  `Retry`, and `ListCourseVideos`; no ownership port exists. Storage and transcoder methods
  exchange `domain.StorageKey`, never filesystem paths.
- `work/report-m03.md` — this report.

No application use case, adapter, SQL, HTTP, frontend, dependency, or generated file was
added or changed. `work/tasks.md` was not edited.

## Decisions within M03

- Transitions are operation-aware instead of only a `from`/`to` adjacency check. This keeps
  T6, T7, and T9 distinct even though each can write `processing → queued`, and records both
  valid T8 writers without inventing another state.
- `NewAsset` accepts the complete domain value, matching the sibling `NewCourse` style, and
  enforces the mint-only shape: valid asset and lesson ids, generation at least one,
  `created`, no claim/bytes/attempt/failure/renditions/eviction, one clock value for creation
  and update, and an expiry after creation.
- Ticket validation is a separate exported seam so M04 can refuse invalid input before the
  repository call; `NewAsset` reuses it. Only a length above 2 GiB maps to
  `upload_too_large`; zero length and invalid filename shape map to `invalid_request`.
- Ladder selection uses height exactly as D-11 specifies. A source below 360 pixels still
  receives the `360p` label; M08's fixed `min(source,target)` scale expression is what keeps
  its actual dimensions from being enlarged.
- `FailureReason` is separate from the wire `Error` type. The five values are the bounded
  D-14 set and can be stored or used as metric labels without accepting arbitrary error text.

## Verification performed

- Ran local `gofmt -w` over all new domain, port, and test files: exit 0.
- Ran `gofmt -d` over the same files: exit 0 with no output.
- Direct standards read-through: production domain imports only `encoding/hex`, `strings`,
  `time`, and `unicode/utf8`; ports import only standard-library packages and the media
  domain. No course, identity, framework, database, transport, or external-package import is
  present.
- Direct specification read-through matched the six states and every T1–T9 edge/writer to
  plan §1, the storage kinds to D-8/ADR 004, the capability to D-3, the ladder to D-11/B17,
  the codec matrix to Q2/R27, and the error strings to §8/D-14.
- Static inventory found only the eleven requested files under `backend/internal/media`, no
  `generated/` directory, and no separate ownership interface.
- The formal `code-review` skill could not run: its first mandatory steps are Git fixed-point
  resolution, commit-log inspection, and `git diff`, while this brief denies all Git access.
  No Git command or workaround was attempted. A direct standards/specification review was
  performed instead; it found one indirect ticket-bound test, which was changed to exercise
  `ValidateTicket` directly and share that implementation with `NewAsset`.

## Checks not run — denied by the brief

The brief states: “No Docker: no Go compile or test.” None of these checks is represented as
passed:

- `make format-check` — not run; its Go half uses Docker. Local `gofmt` is the only formatting
  evidence.
- `make lint` — not run; its Go-vet half compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; the new unit tests require Go compilation/execution.
- `make integration` — not run; Docker/Testcontainers are denied and M03 adds no integration
  implementation.
- `make check-generated` — not run; it invokes Git and writes generated output, both denied.
- `make build` — not run; it requires Docker and compilation.
- `make up` — not run; it requires Docker Compose.
- `make seed-identities` and `make seed` — not run; they require the Compose database.
- `make smoke` — not run; it requires Compose.
- `make e2e` — not run; it requires Compose.
- `make migrate` and aggregate `make check` — not run because they include or depend on the
  denied environments above.

No filesystem write was denied, so the brief's exact-text stop condition did not fire. No Git
command was run and no commit was created.

## Limitation and next task

The new Go surface and unit tables are written evidence awaiting the Docker-enabled reviewing
session; compilation and test success are deliberately not claimed. The next planned slice is
M04, the media application use cases over these four ports.

## Reviewer verification (measured in the reviewing session)

`make test` pass — `courses/internal/media/domain` green (7 test functions across asset,
actor and source, with table cases covering codec acceptance/refusal and the ladder in
landscape, portrait and sub-360p). `make integration` pass including
`courses/platform/architecture`, the machine-checked boundary test — so the domain's
stdlib-only rule is proven, not asserted; an inspection of the imports confirms the domain
files pull only `encoding/hex`, `strings`, `time` and `unicode/utf8`. format-check and lint
pass. The four ports exist with ownership folded into `AssetRepository` (no separate
ownership port, per the GATE 2 deletion). Provenance: implemented with Codex (gpt-5.6-sol);
reviewed and verified here.
