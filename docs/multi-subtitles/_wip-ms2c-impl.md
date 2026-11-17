
NEXT: Read repository instructions, accepted MS2–3 plan and HTTP/OpenAPI precedents; implement only MS2c backend HTTP and overlay scope, preserving the shared working tree.
FACT: Read AGENTS, repository briefing/startup/adapter/standards, accepted plan MS-D5/MS-D6/MS-D10 and D3, and transcript precedents. Existing unrelated dirty/untracked files are preserved. This is implementation/resume of the accepted full-phase plan, not a new planning workflow.
DECISION: 2026-11-15, Codex: inventory uses the admin PUT course-editor rule (the cited admin GET does not exist). Verify via owner/foreign-course/anonymous HTTP cases and repository integration fixtures.
DECISION: Review/verification route = existing separate session Courses 1; owner/delivery owner Courses 1; exact session reference pending user handoff. Basis: explicit brief delegates generation, integration, gates, review and commits to Courses 1. No subagents or shared Docker runtime used here. Author checks only: requested Go build/vet/module tests.
NEXT: Implement per-track draft listing/regeneration and editor inventory, then media flag transport/write and OpenAPI overlay/registries; add row-named tests.
FACT: All GetDraft callers are confined to legacy draft HTTP/tests. Metadata changes currently always rewrite the master inside the repository's asset transaction. No migration is needed.
DECISION: 2026-11-15, Codex: compare language/label under that existing asset lock before deciding to rewrite; preserve optional multilingual atomically with COALESCE. This avoids a pre-read race; counting-fake and repository tests check the branch.
FACT: Pre-code READY for the explicitly scoped accepted implementation: backend HTTP + existing application/repository seams + overlay/registries; no frontend/generated-file writes. Requested author commands are authorized; shared-state/full gates remain Courses 1's responsibility. No new plan decisions outside the brief.
NEXT: Apply transcript backend slice and adapt obsolete single-draft tests.
FACT: Transcript slice now lists sorted per-track drafts, uses explicit track IDs, distinguishes track_not_found, rejects non-empty regeneration bodies (including unknown-length bodies), and adds editor-scoped published-language inventory. The temporary empty-ID/default convention is removed.
NEXT: Thread multilingual through media transport/application/repository, preserving strict optional-field decoding and transaction-owned master rewrite decisions.
FACT: Media mint accepts an explicit optional flag with env fallback; PATCH keeps language/label required and rejects null/duplicate/unknown flag input. GET now exposes per-track multilingual plus multilingualDefault. Repository preserves absent flag and compares actual metadata under lock before rewriting.
NEXT: Update OpenAPI registry/overlay and row-named contracts/integration fixtures; compile to locate any remaining signature callers.
FACT: OpenAPI registry/overlay now describe plural drafts, bodyless per-track regenerate, inventory and media flags. Generated files remain untouched. Initial compile-only check failed on sandbox denial writing the default Go cache (not an auto-review rejection); retry is running with GOCACHE=/private/tmp/courses-ms2c-go-cache. A journal append used backend cwd and failed without changing a file; this entry records that step correctly.
NEXT: Complete endpoint/error/fidelity and integration tests, then requested build/vet/module checks. Generated OpenAPI is tested in memory, not written.
FACT: Added M2-R10 HTTP contracts covering per-track 202, track/lesson 404 codes, 409 running, 403 ownership/origin, 401, strict body rejection, removed-route 404s, ordered list/snapshot staleness and empty arrays. Responses validate against the in-memory generated contract, not the intentionally stale generated YAML.
NEXT: Add inventory HTTP/DB cases and media M2-R12 flag/default/master-count contracts; finish Go compile feedback.
FACT: Added M2-R10 inventory HTTP cases plus private-lesson integration fixture (0/1/2 languages, primary/order/token fields, processing vs Ready staleness, foreign owner denied). Added M2-R12 HTTP default/explicit true/false matrix, PATCH flag-only/language/label/no-op master counts, GET flags and strict optional-bool cases. Unit fake checks callback wiring; the production skip decision is separately asserted against Postgres below.
NEXT: Add actual media repository flag-write/rewrite assertions; run checks and fix any fidelity failures.
FACT: First full module run: transcript tests PASS; media HTTP tests FAIL because checked-in generated YAML still has the old audio schemas and lacks track_not_found. This is generated-artifact drift, not a passed fidelity gate. The subsequent retry was launched from repository root and failed to locate the backend module; no tests ran in that retry.
NEXT: Use the explicitly permitted Docker-free cmd/openapi-gen (inspect flags first) to regenerate YAML from source, never hand-edit it; frontend generation remains Courses 1. Re-run all requested module tests after formatting.
FACT: Media production repository integration assertions now check flag-only=0 rewrites, language=1, no-op stays 1, explicit false persists, absent flag preserves true. Added a counting master fake over real media HTTP/application wiring. Scoped Go files formatted. A new test's unused net/http import was fixed before rerunning.
NEXT: Generate OpenAPI with the authorized Docker-free Go command; run module tests then build/vet/integration-tag vet. No make generate or frontend files.
FACT: Docker-free `GOCACHE=/private/tmp/courses-ms2c-go-cache go run ./cmd/openapi-gen` PASS (exit 0); backend/openapi/openapi.yaml was generated, not hand-edited. No frontend generation.
FACT: During fixture review, corrected the inventory staleness setup: Transcript.Stale intentionally treats authored generation 0 as fresh. The test now saves against Ready generation 1, proves processing generation 2 remains fresh, then promotes generation 2 and expects stale. No product-rule change.
NEXT: Run final module tests and requested build/vet checks, inspect scoped diff and record exact outcomes/remaining Courses 1 work.
FACT: Scoped diff whitespace check PASS. Added explicit repository M2-R10 lookup coverage for non-default/default, older Ready asset, foreign lesson, non-ready/missing track and no-Ready empty drafts; updated port comments to document optional flag/rewrite semantics. No migration or frontend changes.
NEXT: Finish active Go checks; assert new/removed OpenAPI routes and bodyless request shape explicitly, then record final evidence and handoff.
FACT: Requested author checks PASS (all exit 0): Go build ./..., go vet ./..., go vet -tags integration ./..., and go test ./internal/transcript/... ./internal/media/... using the temporary GOCACHE. Media HTTP included generated fidelity and new M2-R12 tests. These are author checks, not independent review or integration execution.
FACT: Added explicit M2-R10 OpenAPI route/schema assertions (legacy routes/request absent, bodyless per-track operation, declared status set and draft fields), plus PATCH response flag round-trip assertions. Final rerun follows these last tests.
NEXT: Final requested checks against current bytes; append changed-file manifest, evidence and Courses 1 next actions.
FACT: Final build → vet → integration-tag vet → transcript/media test chain PASS, exit 0 (GOCACHE=/private/tmp/courses-ms2c-go-cache). Last HTTP package runs: transcript 7.767s, media 12.014s. git diff --check PASS; branch/HEAD remain ms2-3-asr-per-track @ 6d11e76d00d508e376910a9498426cb2bbe302df.
FACT: Strengthened the no-Ready repository fixture to create/prove one actual draft before hiding all Ready assets, avoiding a vacuous empty-list assertion. This changes only the integration test; compile vet is repeated below.
NEXT: Compile the strengthened integration fixture and append final file manifest/handoff. Runtime integration/full gates remain NOT_RUN here by explicit task split.

FACT: Final strengthened integration-tag vet PASS (exit 0). Product-file manifest: 35 changed/new files; SHA-256 of this sorted hash manifest = 155b3645d63349ce3b345985fc930681c34067ed27d9752210d4bc969aeab59c. Journal itself is append-only and excluded from the fingerprint.

```text
3923f59bdcd63cda95e6a37ae172b6b58b0b4ebe06c3e67b78d439ba771aec24  backend/internal/media/adapters/http/handler.go
b5d937283878ba1aac8d7684bd9258342ef7c842052aece1e1e907b5fc5066ff  backend/internal/media/adapters/http/handler_test.go
d8598154a8b1d3c47cb4eef0535e42896bed4131c528fd2dccc494a4358cc7e3  backend/internal/media/adapters/http/multilingual_test.go
73d418b4c71c56528ff934ec3ba7e0ed64641fc94c727641a3155f6ec2b30092  backend/internal/media/adapters/http/openapi_contract.go
dda96b1b28d17d79d4ee070fc88cd2209afcfd8897631d85c1be0da5cf5d1f62  backend/internal/media/adapters/postgres/audiotrack.go
e79442c83ba0f99738f611b66c8d55ee74276c80cea56585d004f041f49f12fa  backend/internal/media/adapters/postgres/audiotrack_integration_test.go
923f1195ab350bca5d27c494343ec80c91c26e86649d8b4a6ace9b6cd5c3e9b2  backend/internal/media/adapters/postgres/language_shape_integration_test.go
90b9f4c5eb9e33188f4b8f81086e4c0fc718ea9bda123e1a5d4ab567c79026f2  backend/internal/media/adapters/postgres/repository_integration_test.go
07f19ee505a1a2f92aab09ddce37b43ecf6516829578b99bae41039bf3d1bb7c  backend/internal/media/application/audiotrack.go
7e48b386c7870612b31c4cc824bb6198c01f0bb46511058e255f493f44190434  backend/internal/media/application/regenerate_test.go
614b2d1087b19cd577f3171e536abcd93f28c91a53200d0d112f40f49e28aec0  backend/internal/media/application/upload_test.go
9fdb280914303b4f2c6808bcfa4366262b85cbc680fa586374da3786a16e13f2  backend/internal/media/ports/repository.go
b56db1c68da0da7e46c96efddb70ff23a371516b1237cfb2c5a26eaa29ec083e  backend/internal/transcript/adapters/http/admin_handler.go
d37393cf0404f674daa0585083b6011c6b7831f6f4a05c49d46684a3405281fd  backend/internal/transcript/adapters/http/draft_handler.go
3a1cac5c1673390dc41f6912caf9ecbfa39072d13fae0c5eaf89c9cf4d6e9327  backend/internal/transcript/adapters/http/draft_handler_test.go
246c072b36b65d8889768a57e4561b9af54544212bb2ffa3a1662d26d7ab5837  backend/internal/transcript/adapters/http/handler.go
2e4c54be3899b1c9d62af5b8e1ac18ad7f24075fe9ca2cb6da4bf9e6e4312d09  backend/internal/transcript/adapters/http/handler_test.go
0d52a7dd70a2809ad7044e0bd1c4937267e7749daa95bcee10d97b50ab7ed51c  backend/internal/transcript/adapters/http/inventory_test.go
5ba108985aa3b23b1556fa70f3c8b703ba48f367dd3deca9987902391569aa62  backend/internal/transcript/adapters/http/ms2_openapi_test.go
bc1ca6a45bfd147af4f5d66ade3b0d4807deb3b5a4dbb2364e396d51ce3eb5b0  backend/internal/transcript/adapters/http/openapi_contract.go
149a14423187bf47666d7c611438622ceeb5e5542efa425248c9e0f2eed2e5a2  backend/internal/transcript/adapters/postgres/draft.go
f95d994f7dd7eda0a5384a994b04cd1a6d16c4d4f0b6bf948af13b640ecadd8d  backend/internal/transcript/adapters/postgres/draft_integration_test.go
3bddb5f2c034953ed2e25f0a6af1b93767bcd3393b655e8d5d0cfd68d8d91691  backend/internal/transcript/adapters/postgres/inventory_integration_test.go
8b9dc03221e29deeaa64430a67bdf883cc89f7f3e42b3119997cc309528ed248  backend/internal/transcript/adapters/postgres/repository.go
67b54e9bd48978ab1420ca4109a542107b62819d908d769a3e59c8698ea9dcac  backend/internal/transcript/adapters/postgres/repository_integration_test.go
109c0882522b62df67130bb78f6d6e96139eec067aa02cea1554d20e3d5b1c78  backend/internal/transcript/adapters/postgres/s6_integration_test.go
fd5957ab1280daf1bdfe16d490c52b1c162aeee15aaf6eca9897d0e37e04b1a1  backend/internal/transcript/application/draft.go
bbd425c0eebeafd5d93045c03a9a732fa6d1f99e381841a66cb3222d9085df97  backend/internal/transcript/application/draft_test.go
b0bc07b56d935e5a0fe3a6da5cdc211c230fff191a6354982c79cce60dbf94fc  backend/internal/transcript/application/read_test.go
951d9a182da28d9a727ab435d26894bffc827f955d3fae42dedeaf907fcff91c  backend/internal/transcript/application/write.go
0c0a1a74bf2693599da10398e5c160e298fc127aca88f5a6744b0c01c124a8e0  backend/internal/transcript/domain/errors.go
fb4f8e1e37114003b7c16cd4b4d3576302211a88ea21b38661d19d0657019131  backend/internal/transcript/ports/draft.go
e32ea7e127d85b15acf241c5df0385cc2e1c5ed3bc23be96e6f658b9427d7154  backend/internal/transcript/ports/repository.go
7db54c2ee392c5cdf63a3c372bbb9be31fa373894a036ecd5e508bf80ab4a011  backend/openapi/openapi.overlay.yaml
bc602abb1d45da2d3208d5a70ea465b1301bea222ac59e6298dbca5aa20c0b4d  backend/openapi/openapi.yaml
```

FACT: Delivery summary: replaced legacy single-draft routes with sorted per-track drafts and bodyless explicit-track regeneration; added track_not_found; removed GetDraft/default-ID convention; added editor-scoped transcript inventory; threaded optional multilingual through media mint/PATCH and returned both GET fields; master rewrite only for changed language/label under the asset lock; added M2-R10/M2-R12 HTTP/OpenAPI and compiling integration coverage.
FACT: Verification (cwd backend, GOCACHE=/private/tmp/courses-ms2c-go-cache): `go build ./...` PASS exit 0; `go vet ./...` PASS exit 0; `go vet -tags integration ./...` PASS exit 0 (including final added fixture); `go test ./internal/transcript/... ./internal/media/...` PASS exit 0. Docker-free `go run ./cmd/openapi-gen` PASS exit 0; YAML changed only by that generator. `git diff --check` PASS.
FACT: NOT_RUN here: runtime Postgres integration, Docker/make generate, frontend generation/tests, full ordered gates/e2e, independent review. No commits, branch changes, checkout/reset/stash/clean/rm, migrations, frontend edits, or unrelated file edits performed. The frontend still needs the next run's migration to the new endpoints/generated client; this backend-only slice does not claim an end-to-end editor PASS.
QUESTION: None outstanding for this implementation scope; inventory access ambiguity was explicitly resolved by the brief and recorded above.
NEXT: Courses 1 owns `make generate`/frontend follow-up, actual integration tests (including inventory and multilingual rewrite cases), full ordered gates, independent review and commits. Review the exact current file manifest above; no independent PASS is claimed by this author session.
