NEXT: Read shared entry rules and the be-b mutation journal; close GAP rows with behavior tests, verify baseline and exact mutants, and restore production bytes after every row.
FACT: Verified authorized root; clean HEAD 4e569f6e919dc220ed3286aa6a2cb5a25b2ca767 (brief named 5a56965). Extra commit changes nine transcript test files only; production matches 5a56965. Preserve be-a tests. No AGENTS.local.md. Read briefing, adapter/standards and exact behavior-coverage contract.
DECISION: Tests-only Small task; owner authorizes mutants and Kiri ephemeral integration. Inline plan: existing repository harness for persistence/scoping/locking; subprocess startup config/wiring; fake-Docker reset. Independent review delivery remains with owner/existing PR reviewer; no agents/messages started (0/0). Author checks are not independent review PASS.
NEXT: Implement B25/B26/B27/B29/B34/B38/B39/B69 at existing PostgreSQL seam using SQL assertions and event/lock synchronization.
FACT: Initial repository baseline exit 1: gaps_integration_test.go:84: invalid_request. Test fixture used updated_at before the sibling created_at (mediaNow+1m). Corrected test timestamp to mediaNow+1h; this invalid fixture is not a product DEFECT. All other new repository cases passed. Raw: mut-be/.tmp/gap-db-baseline.log.
NEXT: Rerun corrected repository baseline before exact mutants.
NEXT: Apply exact B25 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B25 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B25.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -96,7 +96,7 @@
 		track.ID, track.VideoAssetID, string(track.Kind), track.Language, track.Label,
 		track.IsDefault, string(track.State), track.SourceLeaf,
 		track.DeclaredBytes, track.ReceivedBytes, track.OriginalName, expiresAt,
-		created, updated, track.Multilingual))
+		created, updated, false))
 	if err != nil {
 		if mapped, ok := audioTrackConflict(err); ok {
 			return domain.AudioTrack{}, mapped
```
FACT: B25 | targeted behavior test (see raw output) | baseline PASS |                 gaps_integration_test.go:46: multilingual returned=false SQL=false loaded=false want=true | CLOSED
FACT: B25 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Apply exact B26 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B26 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B26.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -273,7 +273,7 @@
 		return domain.AudioTrack{}, err
 	}
 	previous, err := scanAudioTrack(tx.QueryRow(ctx, `SELECT `+audioTrackColumns+` FROM audio_tracks
- WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid FOR UPDATE`, trackID, assetID))
+ WHERE id=$1::text::uuid AND $2::text IS NOT NULL FOR UPDATE`, trackID, assetID))
 	if errors.Is(err, pgx.ErrNoRows) {
 		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
```
FACT: B26 | targeted behavior test (see raw output) | baseline PASS |         gaps_integration_test.go:134: foreign locked track error=dependency_unavailable want video_not_found | CLOSED
FACT: B26 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Apply exact B27 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B27 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B27.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -273,7 +273,7 @@
 		return domain.AudioTrack{}, err
 	}
 	previous, err := scanAudioTrack(tx.QueryRow(ctx, `SELECT `+audioTrackColumns+` FROM audio_tracks
- WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid FOR UPDATE`, trackID, assetID))
+ WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid`, trackID, assetID))
 	if errors.Is(err, pgx.ErrNoRows) {
 		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
```
FACT: B27 | targeted behavior test (see raw output) | baseline PASS |         gaps_integration_test.go:176: prior-read lock was not reached: timeout: context deadline exceeded /         gaps_integration_test.go:238: sweeper entered before metadata completed: sweep=<nil> metadata=video_not_found; want contested row lock | CLOSED
FACT: B27 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Apply exact B29 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B29 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B29.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -278,7 +278,7 @@
 		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
 	if err != nil {
-		return domain.AudioTrack{}, persistenceError(err)
+		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
 	updated, err := scanAudioTrack(tx.QueryRow(ctx, `UPDATE audio_tracks
 		SET language=$3, label=$4, updated_at=$5, multilingual=COALESCE($6, multilingual)
```
FACT: B29 | targeted behavior test (see raw output) | baseline PASS |         gaps_integration_test.go:179: prior-read failure=video_not_found want dependency_unavailable | CLOSED
FACT: B29 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Apply exact B34 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B34 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B34.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -294,7 +294,7 @@
 		return domain.AudioTrack{}, persistenceError(err)
 	}
 	// NAME and LANGUAGE are rendered into the track's EXT-X-MEDIA line (D7).
-	if previous.Language != updated.Language || previous.Label != updated.Label {
+	if previous.Language != updated.Language {
 		if err := rewriteReadyTracks(ctx, tx, assetID, rewrite); err != nil {
 			return domain.AudioTrack{}, err
 		}
```
FACT: B34 | targeted behavior test (see raw output) | baseline PASS |         gaps_integration_test.go:98: label-only rewrite calls=0 err=<nil>; want 1 and nil | CLOSED
FACT: B34 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Apply exact B38 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B38 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B38.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -673,7 +673,7 @@
 			'audio-'||id::text,$4,0,$5,$6,$7,$7,$8
 		FROM new_id
 		RETURNING `+audioTrackColumns,
-		assetID, language, label, declaredBytes, originalName, expiresAt, now, multilingual))
+		assetID, language, label, declaredBytes, originalName, expiresAt, now, false))
 	if err != nil {
 		if mapped, ok := audioTrackConflict(err); ok {
 			return domain.AudioTrack{}, mapped
```
FACT: B38 | targeted behavior test (see raw output) | baseline PASS |                 gaps_integration_test.go:55: multilingual returned=false SQL=false loaded=false want=true | CLOSED
FACT: B38 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Apply exact B39 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B39 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B39.log
```diff
--- backend/internal/media/adapters/postgres/publish.go
+++ backend/internal/media/adapters/postgres/publish.go
@@ -88,7 +88,7 @@
 			id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at,multilingual)
 			VALUES($1::text::uuid,$2::text::uuid,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
 			embedded.ID, embedded.VideoAssetID, string(embedded.Kind), embedded.Language, embedded.Label,
-			embedded.IsDefault, string(embedded.State), embedded.SourceLeaf, storeTime(embedded.CreatedAt), storeTime(embedded.UpdatedAt), embedded.Multilingual); err != nil {
+			embedded.IsDefault, string(embedded.State), embedded.SourceLeaf, storeTime(embedded.CreatedAt), storeTime(embedded.UpdatedAt), false); err != nil {
 			if mapped, ok := audioTrackConflict(err); ok {
 				return domain.VideoAsset{}, nil, mapped
 			}
```
FACT: B39 | targeted behavior test (see raw output) | baseline PASS |                 gaps_integration_test.go:69: multilingual returned=false SQL=false loaded=false want=true | CLOSED
FACT: B39 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/publish.go` exit 0
NEXT: Apply exact B69 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP'`; baseline PASS from current unmutated targeted run.
FACT: B69 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B69.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -273,7 +273,7 @@
 		return domain.AudioTrack{}, err
 	}
 	previous, err := scanAudioTrack(tx.QueryRow(ctx, `SELECT `+audioTrackColumns+` FROM audio_tracks
- WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid FOR UPDATE`, trackID, assetID))
+ WHERE $1::text IS NOT NULL AND video_asset_id=$2::text::uuid FOR UPDATE`, trackID, assetID))
 	if errors.Is(err, pgx.ErrNoRows) {
 		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
```
FACT: B69 | targeted behavior test (see raw output) | baseline PASS |         gaps_integration_test.go:88: metadata no-ops rewrote master 1 times; want 0 | CLOSED
FACT: B69 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
FACT: Corrected PostgreSQL GAP baseline PASS exit 0 (.tmp/gap-db-baseline2.log); eight exact mutants B25/B26/B27/B29/B34/B38/B39/B69 each exit 1 with behavior assertions, all production bytes restored. B27 is killed by the sweeper-entered assertion; the separate lock-query test also times out under this mutant but is not counted as valid kill evidence.
FACT: Reset baseline PASS exit 0 (.tmp/gap-reset-baseline.log), seven scenarios run the real Makefile with PATH-bound fake Docker only.
NEXT: Baseline and exact mutants for four startup validation rows and seven reset rows; implement remaining executable default-wiring behaviors B47/B50 at ephemeral process seams.
NEXT: Apply exact B45 mutation; command `go test -count=1 ./cmd/api ./cmd/media-worker -run 'Test(API|Worker)RejectsInvalid'`; baseline PASS from current unmutated targeted run.
FACT: B45 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B45.log
```diff
--- backend/cmd/api/main.go
+++ backend/cmd/api/main.go
@@ -101,7 +101,7 @@
 		logger.Error("startup failed", "reason", err.Error())
 		os.Exit(1)
 	}
-	multilingualDefault, err := config.ASRMultilingualDefault(os.Getenv)
+	multilingualDefault, err := config.ASRMultilingualDefault(func(string) string { return "" })
 	if err != nil {
 		logger.Error("startup failed", "reason", err.Error())
 		os.Exit(1)
```
FACT: B45 | targeted behavior test (see raw output) | baseline PASS |         main_test.go:40: startup ASR=invalid error=exit status 1 output={"time":"2026-11-17T17:36:02.569855+02:00","level":"ERROR","msg":"startup failed","service":"api","reason":"POSTGRES_HOST is required"} | CLOSED
FACT: B45 restored original production bytes; `git diff --quiet -- backend/cmd/api/main.go` exit 0
NEXT: Apply exact B46 mutation; command `go test -count=1 ./cmd/api ./cmd/media-worker -run 'Test(API|Worker)RejectsInvalid'`; baseline PASS from current unmutated targeted run.
FACT: B46 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B46.log
```diff
--- backend/cmd/api/main.go
+++ backend/cmd/api/main.go
@@ -102,7 +102,7 @@
 		os.Exit(1)
 	}
 	multilingualDefault, err := config.ASRMultilingualDefault(os.Getenv)
-	if err != nil {
+	if false {
 		logger.Error("startup failed", "reason", err.Error())
 		os.Exit(1)
 	}
```
FACT: B46 | targeted behavior test (see raw output) | baseline PASS |         main_test.go:40: startup ASR=invalid error=exit status 1 output={"time":"2026-11-17T17:36:08.19286+02:00","level":"ERROR","msg":"startup failed","service":"api","reason":"POSTGRES_HOST is required"} | CLOSED
FACT: B46 restored original production bytes; `git diff --quiet -- backend/cmd/api/main.go` exit 0
NEXT: Apply exact B48 mutation; command `go test -count=1 ./cmd/api ./cmd/media-worker -run 'Test(API|Worker)RejectsInvalid'`; baseline PASS from current unmutated targeted run.
FACT: B48 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B48.log
```diff
--- backend/cmd/media-worker/main.go
+++ backend/cmd/media-worker/main.go
@@ -50,7 +50,7 @@
 }
 
 func run(ctx context.Context, logger *slog.Logger) error {
-	multilingualDefault, err := config.ASRMultilingualDefault(os.Getenv)
+	multilingualDefault, err := config.ASRMultilingualDefault(func(string) string { return "" })
 	if err != nil {
 		return err
 	}
```
FACT: B48 | targeted behavior test (see raw output) | baseline PASS |         main_test.go:22: worker startup ASR=invalid error=POSTGRES_HOST is required; want "ASR_MULTILINGUAL_DEFAULT must be true or false" | CLOSED
FACT: B48 restored original production bytes; `git diff --quiet -- backend/cmd/media-worker/main.go` exit 0
NEXT: Apply exact B49 mutation; command `go test -count=1 ./cmd/api ./cmd/media-worker -run 'Test(API|Worker)RejectsInvalid'`; baseline PASS from current unmutated targeted run.
FACT: B49 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B49.log
```diff
--- backend/cmd/media-worker/main.go
+++ backend/cmd/media-worker/main.go
@@ -51,7 +51,7 @@
 
 func run(ctx context.Context, logger *slog.Logger) error {
 	multilingualDefault, err := config.ASRMultilingualDefault(os.Getenv)
-	if err != nil {
+	if false {
 		return err
 	}
 	dependencies, err := runtime.Open(ctx, "media-worker", logger)
```
FACT: B49 | targeted behavior test (see raw output) | baseline PASS |         main_test.go:22: worker startup ASR=invalid error=POSTGRES_HOST is required; want "ASR_MULTILINGUAL_DEFAULT must be true or false" | CLOSED
FACT: B49 restored original production bytes; `git diff --quiet -- backend/cmd/media-worker/main.go` exit 0
NEXT: Apply exact B56 mutation; command `go test -count=1 ./platform/architecture -run TestE2EResetBehavior`; baseline PASS from current unmutated targeted run.
FACT: B56 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B56.log
```diff
--- Makefile
+++ Makefile
@@ -129,7 +129,7 @@
 	@config=$$($(E2E_RESET_COMPOSE) config --format json) \
 	  || { echo "e2e-reset: cannot resolve the compose configuration" >&2; exit 1; }; \
 	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + " " + v.name); });') \
-	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; exit 1; }; \
+	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; :; }; \
 	project=$$(echo "$$resolved" | sed -n 1p); \
 	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
 	for volume in $(E2E_RESET_VOLUMES); do \
```
FACT: B56 | targeted behavior test (see raw output) | baseline PASS |         e2e_reset_test.go:174: resolver failure must abort before listing: error=<nil> lists=6 | CLOSED
FACT: B56 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B62 mutation; command `go test -count=1 ./platform/architecture -run TestE2EResetBehavior`; baseline PASS from current unmutated targeted run.
FACT: B62 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B62.log
```diff
--- Makefile
+++ Makefile
@@ -131,7 +131,7 @@
 	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + " " + v.name); });') \
 	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; exit 1; }; \
 	project=$$(echo "$$resolved" | sed -n 1p); \
-	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
+	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | cat; }; \
 	for volume in $(E2E_RESET_VOLUMES); do \
 	  name=$$(echo "$$resolved" | awk -v v="$$volume" 'NR > 1 && $$1 == v { print $$2 }'); \
 	  if [ -z "$$name" ]; then echo "e2e-reset: $$volume is not a volume of project $$project" >&2; exit 1; fi; \
```
FACT: B62 | targeted behavior test (see raw output) | baseline PASS |         e2e_reset_test.go:203: removal order=[alias-2 alias-10 custom-postgres-data] want [alias-10 alias-2 custom-postgres-data] | CLOSED
FACT: B62 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B64 mutation; command `go test -count=1 ./platform/architecture -run TestE2EResetBehavior`; baseline PASS from current unmutated targeted run.
FACT: B64 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B64.log
```diff
--- Makefile
+++ Makefile
@@ -135,7 +135,7 @@
 	for volume in $(E2E_RESET_VOLUMES); do \
 	  name=$$(echo "$$resolved" | awk -v v="$$volume" 'NR > 1 && $$1 == v { print $$2 }'); \
 	  if [ -z "$$name" ]; then echo "e2e-reset: $$volume is not a volume of project $$project" >&2; exit 1; fi; \
-	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; exit 1; }; \
+	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; :; }; \
 	  ids=$$(echo "$$all" | pick "$$name" "$$volume"); \
 	  if [ -n "$$ids" ]; then docker volume rm $$ids > /dev/null || exit 1; fi; \
 	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; exit 1; }; \
```
FACT: B64 | targeted behavior test (see raw output) | baseline PASS |         e2e_reset_test.go:178: failed empty lookup must fail: error=<nil> | CLOSED
FACT: B64 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B78 mutation; command `go test -count=1 ./platform/architecture -run TestE2EResetBehavior`; baseline PASS from current unmutated targeted run.
FACT: B78 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B78.log
```diff
--- Makefile
+++ Makefile
@@ -143,7 +143,7 @@
 	  if [ -n "$$left" ]; then echo "e2e-reset: $$volume survived:" $$left >&2; exit 1; fi; \
 	  if [ -n "$$ids" ]; then echo "e2e-reset: $$volume: removed" $$ids; else echo "e2e-reset: $$volume: absent"; fi; \
 	done; \
-	echo "e2e-reset: project $$project done (prometheus-data kept)"
+	echo "e2e-reset: project unknown done (prometheus-data kept)"
 # Owner-run live WayForPay proof (Pay25-live), work/archive/runbook-o06c-tunnel.md. Needs only
 # WAYFORPAY_MERCHANT_ACCOUNT and WAYFORPAY_SECRET_KEY in .env; opens a public quick tunnel to this
 # stack (seeded demo accounts included), so keep it short-lived and run tunnel-down after.
```
FACT: B78 | targeted behavior test (see raw output) | baseline PASS |         e2e_reset_test.go:192: resolved project summary missing: /         e2e_reset_test.go:192: resolved project summary missing: /         e2e_reset_test.go:192: resolved project summary missing: /         e2e_reset_test.go:192: resolved project summary missing: | CLOSED
FACT: B78 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B79 mutation; command `go test -count=1 ./platform/architecture -run TestE2EResetBehavior`; baseline PASS from current unmutated targeted run.
FACT: B79 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B79.log
```diff
--- Makefile
+++ Makefile
@@ -133,7 +133,7 @@
 	project=$$(echo "$$resolved" | sed -n 1p); \
 	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
 	for volume in $(E2E_RESET_VOLUMES); do \
-	  name=$$(echo "$$resolved" | awk -v v="$$volume" 'NR > 1 && $$1 == v { print $$2 }'); \
+	  name=$$(echo "$$resolved" | awk -v v="$$volume" '$$1 == v { print $$2 }'); \
 	  if [ -z "$$name" ]; then echo "e2e-reset: $$volume is not a volume of project $$project" >&2; exit 1; fi; \
 	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; exit 1; }; \
 	  ids=$$(echo "$$all" | pick "$$name" "$$volume"); \
```
FACT: B79 | targeted behavior test (see raw output) | baseline PASS |         e2e_reset_test.go:189: remaining volumes=[{Name:custom-postgres-data Project: Volume:} {Name:custom-prometheus-data Project:postgres-data Volume:prometheus-data} {Name:foreign Project:another-project Volume:postgres-data}]; want only prometheus and foreign | CLOSED
FACT: B79 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B80 mutation; command `go test -count=1 ./platform/architecture -run TestE2EResetBehavior`; baseline PASS from current unmutated targeted run.
FACT: B80 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B80.log
```diff
--- Makefile
+++ Makefile
@@ -128,7 +128,7 @@
 	$(E2E_RESET_COMPOSE) down --remove-orphans
 	@config=$$($(E2E_RESET_COMPOSE) config --format json) \
 	  || { echo "e2e-reset: cannot resolve the compose configuration" >&2; exit 1; }; \
-	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + " " + v.name); });') \
+	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes)) console.log(k + " " + v.name); });') \
 	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; exit 1; }; \
 	project=$$(echo "$$resolved" | sed -n 1p); \
 	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
```
FACT: B80 | targeted behavior test (see raw output) | baseline PASS |         e2e_reset_test.go:182: missing volumes must identify missing logical volume: error=exit status 2 lists=0 | CLOSED
FACT: B80 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B81 mutation; command `go test -count=1 ./platform/architecture -run TestE2EResetBehavior`; baseline PASS from current unmutated targeted run.
FACT: B81 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B81.log
```diff
--- Makefile
+++ Makefile
@@ -123,7 +123,7 @@
 # rule would wipe it silently. Whoever adds a stateful volume that tests write to adds it here.
 E2E_RESET_VOLUMES := postgres-data redis-data media-data
 E2E_RESET_COMPOSE := docker compose -f docker-compose.yml -f docker-compose.e2e.yml
-E2E_RESET_LS := docker volume ls --format '{{.Name}} {{.Label "com.docker.compose.project"}} {{.Label "com.docker.compose.volume"}}'
+E2E_RESET_LS := docker volume ls --format '{{.Name}}'
 e2e-reset:
 	$(E2E_RESET_COMPOSE) down --remove-orphans
 	@config=$$($(E2E_RESET_COMPOSE) config --format json) \
```
FACT: B81 | targeted behavior test (see raw output) | baseline PASS |         e2e_reset_test.go:189: remaining volumes=[{Name:alias-2 Project:resolved-project Volume:postgres-data} {Name:alias-10 Project:resolved-project Volume:postgres-data} {Name:custom-prometheus-data Project:resolved-project Volume:prometheus-data} {Name:foreign Project:another-project Volume:postgres-data}]; want only prometheus and foreign /         e2e_reset_test.go:189: remaining volumes=[{Name:renamed-redis Project:resolved-project Volume:redis-data} {Name:custom-prometheus-data Project:resolved-project Volume:prometheus-data} {Name:foreign Project:another-project Volume:postgres-data}]; want only prometheus and foreign | CLOSED
FACT: B81 restored original production bytes; `git diff --quiet -- Makefile` exit 0
FACT: Startup validation baseline PASS exit 0 (.tmp/gap-startup-baseline.log); B45/B46/B48/B49 exact mutants fail expected ASR diagnostic assertions. Seven Go reset scenarios baseline PASS and all seven mutants failed; superseded their Go test location after discovering backend-only CI mount lacks root Makefile.
DECISION: Preserve reset behavior under frontend Node test project (src/adapters/api/e2e-reset.test.ts + test-only fake fixture); rerun baseline and exact mutants against final location. No CI production/config changes. Install locked frontend dependencies into disposable copy only with npm cache inside copy; first npm exec failed EPERM on default user cache, no source changed.
FACT: Initial executable fixture baseline failed setup (no migrations found): incorrect relative backend root; corrected fixture path (three parents) and ready endpoint (/readyz). Not a product defect, test had not reached startup.
NEXT: Verify final Node reset tests and actual binary true/false wiring; then exact B47/B50 mutants and final checks.
NEXT: Apply exact B56 mutation; command `cd ../frontend && npm test -- src/adapters/api/e2e-reset.test.ts`; baseline PASS from current unmutated targeted run.
FACT: B56 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B56.log
```diff
--- Makefile
+++ Makefile
@@ -129,7 +129,7 @@
 	@config=$$($(E2E_RESET_COMPOSE) config --format json) \
 	  || { echo "e2e-reset: cannot resolve the compose configuration" >&2; exit 1; }; \
 	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + " " + v.name); });') \
-	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; exit 1; }; \
+	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; :; }; \
 	project=$$(echo "$$resolved" | sed -n 1p); \
 	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
 	for volume in $(E2E_RESET_VOLUMES); do \
```
FACT: B56 | targeted behavior test (see raw output) | baseline PASS | NO ASSERTION FAILURE | REMAINING
FACT: B56 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B62 mutation; command `cd ../frontend && npm test -- src/adapters/api/e2e-reset.test.ts`; baseline PASS from current unmutated targeted run.
FACT: B62 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B62.log
```diff
--- Makefile
+++ Makefile
@@ -131,7 +131,7 @@
 	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + " " + v.name); });') \
 	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; exit 1; }; \
 	project=$$(echo "$$resolved" | sed -n 1p); \
-	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
+	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | cat; }; \
 	for volume in $(E2E_RESET_VOLUMES); do \
 	  name=$$(echo "$$resolved" | awk -v v="$$volume" 'NR > 1 && $$1 == v { print $$2 }'); \
 	  if [ -z "$$name" ]; then echo "e2e-reset: $$volume is not a volume of project $$project" >&2; exit 1; fi; \
```
FACT: B62 | targeted behavior test (see raw output) | baseline PASS | NO ASSERTION FAILURE | REMAINING
FACT: B62 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B64 mutation; command `cd ../frontend && npm test -- src/adapters/api/e2e-reset.test.ts`; baseline PASS from current unmutated targeted run.
FACT: B64 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B64.log
```diff
--- Makefile
+++ Makefile
@@ -135,7 +135,7 @@
 	for volume in $(E2E_RESET_VOLUMES); do \
 	  name=$$(echo "$$resolved" | awk -v v="$$volume" 'NR > 1 && $$1 == v { print $$2 }'); \
 	  if [ -z "$$name" ]; then echo "e2e-reset: $$volume is not a volume of project $$project" >&2; exit 1; fi; \
-	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; exit 1; }; \
+	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; :; }; \
 	  ids=$$(echo "$$all" | pick "$$name" "$$volume"); \
 	  if [ -n "$$ids" ]; then docker volume rm $$ids > /dev/null || exit 1; fi; \
 	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; exit 1; }; \
```
FACT: B64 | targeted behavior test (see raw output) | baseline PASS | NO ASSERTION FAILURE | REMAINING
FACT: B64 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B78 mutation; command `cd ../frontend && npm test -- src/adapters/api/e2e-reset.test.ts`; baseline PASS from current unmutated targeted run.
FACT: B78 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B78.log
```diff
--- Makefile
+++ Makefile
@@ -143,7 +143,7 @@
 	  if [ -n "$$left" ]; then echo "e2e-reset: $$volume survived:" $$left >&2; exit 1; fi; \
 	  if [ -n "$$ids" ]; then echo "e2e-reset: $$volume: removed" $$ids; else echo "e2e-reset: $$volume: absent"; fi; \
 	done; \
-	echo "e2e-reset: project $$project done (prometheus-data kept)"
+	echo "e2e-reset: project unknown done (prometheus-data kept)"
 # Owner-run live WayForPay proof (Pay25-live), work/archive/runbook-o06c-tunnel.md. Needs only
 # WAYFORPAY_MERCHANT_ACCOUNT and WAYFORPAY_SECRET_KEY in .env; opens a public quick tunnel to this
 # stack (seeded demo accounts included), so keep it short-lived and run tunnel-down after.
```
FACT: B78 | targeted behavior test (see raw output) | baseline PASS | NO ASSERTION FAILURE | REMAINING
FACT: B78 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B79 mutation; command `cd ../frontend && npm test -- src/adapters/api/e2e-reset.test.ts`; baseline PASS from current unmutated targeted run.
FACT: B79 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B79.log
```diff
--- Makefile
+++ Makefile
@@ -133,7 +133,7 @@
 	project=$$(echo "$$resolved" | sed -n 1p); \
 	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
 	for volume in $(E2E_RESET_VOLUMES); do \
-	  name=$$(echo "$$resolved" | awk -v v="$$volume" 'NR > 1 && $$1 == v { print $$2 }'); \
+	  name=$$(echo "$$resolved" | awk -v v="$$volume" '$$1 == v { print $$2 }'); \
 	  if [ -z "$$name" ]; then echo "e2e-reset: $$volume is not a volume of project $$project" >&2; exit 1; fi; \
 	  all=$$($(E2E_RESET_LS)) || { echo "e2e-reset: cannot list volumes" >&2; exit 1; }; \
 	  ids=$$(echo "$$all" | pick "$$name" "$$volume"); \
```
FACT: B79 | targeted behavior test (see raw output) | baseline PASS | NO ASSERTION FAILURE | REMAINING
FACT: B79 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B80 mutation; command `cd ../frontend && npm test -- src/adapters/api/e2e-reset.test.ts`; baseline PASS from current unmutated targeted run.
FACT: Final Node reset baseline PASS: seven tests, no skips (.tmp/gap-reset-ts-baseline2.log); npm ci --ignore-scripts completed using copy-local cache. Rerunning seven exact reset mutants against final TS test location.
DECISION: Tightened B29 synchronization: cancel the metadata context and consume its error while the target row remains locked; release only after the failure is received, ensuring the prior SELECT error branch is reached. No sleeps. Recheck B29 after this test delta. Worker true/false fixtures use different asset IDs so retained queue task IDs cannot mask the second run.
NEXT: Finish executable baseline; reprove changed concurrency test and all final reset tests.
FACT: B80 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B80.log
```diff
--- Makefile
+++ Makefile
@@ -128,7 +128,7 @@
 	$(E2E_RESET_COMPOSE) down --remove-orphans
 	@config=$$($(E2E_RESET_COMPOSE) config --format json) \
 	  || { echo "e2e-reset: cannot resolve the compose configuration" >&2; exit 1; }; \
-	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + " " + v.name); });') \
+	resolved=$$(printf '%s' "$$config" | node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes)) console.log(k + " " + v.name); });') \
 	  || { echo "e2e-reset: cannot resolve the compose project and volume names" >&2; exit 1; }; \
 	project=$$(echo "$$resolved" | sed -n 1p); \
 	pick() { awk -v n="$$1" -v p="$$project" -v v="$$2" '$$1 == n || ($$2 == p && $$3 == v) { print $$1 }' | sort -u; }; \
```
FACT: B80 | targeted behavior test (see raw output) | baseline PASS | NO ASSERTION FAILURE | REMAINING
FACT: B80 restored original production bytes; `git diff --quiet -- Makefile` exit 0
NEXT: Apply exact B81 mutation; command `cd ../frontend && npm test -- src/adapters/api/e2e-reset.test.ts`; baseline PASS from current unmutated targeted run.
FACT: B81 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B81.log
```diff
--- Makefile
+++ Makefile
@@ -123,7 +123,7 @@
 # rule would wipe it silently. Whoever adds a stateful volume that tests write to adds it here.
 E2E_RESET_VOLUMES := postgres-data redis-data media-data
 E2E_RESET_COMPOSE := docker compose -f docker-compose.yml -f docker-compose.e2e.yml
-E2E_RESET_LS := docker volume ls --format '{{.Name}} {{.Label "com.docker.compose.project"}} {{.Label "com.docker.compose.volume"}}'
+E2E_RESET_LS := docker volume ls --format '{{.Name}}'
 e2e-reset:
 	$(E2E_RESET_COMPOSE) down --remove-orphans
 	@config=$$($(E2E_RESET_COMPOSE) config --format json) \
```
FACT: B81 | targeted behavior test (see raw output) | baseline PASS | NO ASSERTION FAILURE | REMAINING
FACT: B81 restored original production bytes; `git diff --quiet -- Makefile` exit 0
FACT: FINAL B56 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset partial resolver failure | baseline PASS | mutated FAIL assertion `expect(run.status, output).not.toBe(0);`; verbatim failure "docker compose -f docker-compose.yml -f docker-compose.e2e.yml down --remove-orphans\ne2e-reset: postgres-data: removed custom-postgres-data\ne2e-reset: redis-data: removed custom-redis-data\ne2e-reset: media-data: removed custom-media-data\ne2e-reset: project resolved-project done (prometheus-data kept)\n[eval]:1\nlet s = \"\"; process.stdin.on(\"data\", (d) => (s += d)).on(\"end\", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + \" \" + v.name); });\n                                                                                                                                                                                                                                  ^\n\nTypeError: Cannot read properties of null (reading 'name')\n    at Socket.<anonymous> ([eval]:1:227)\n    at Socket.emit (node:events:520:35)\n    at endReadableNT (node:internal/streams/readable:1701:12)\n    at process.processTicksAndRejections (node:internal/process/task_queues:89:21)\n\nNode.js v24.11.1\ne2e-reset: cannot resolve the compose project and volume names\n: expected +0 not to be +0 // Object.is equality" | CLOSED (supersedes provisional parser label REMAINING; raw assertion report was not parsed by Go-only extractor)
FACT: FINAL B62 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset sorted matches | baseline PASS | mutated FAIL assertion `expect(removal).toEqual(['volume', 'rm', 'alias-10', 'alias-2', 'custom-postgres-data']);`; verbatim failure "expected [ 'volume', 'rm', 'alias-2', …(2) ] to deeply equal [ 'volume', 'rm', 'alias-10', …(2) ]" | CLOSED (supersedes provisional parser label REMAINING; raw assertion report was not parsed by Go-only extractor)
FACT: FINAL B64 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset absent lookup failure | baseline PASS | mutated FAIL assertion `expect(run.status, 'failed empty lookup must fail').not.toBe(0);`; verbatim failure "failed empty lookup must fail: expected +0 not to be +0 // Object.is equality" | CLOSED (supersedes provisional parser label REMAINING; raw assertion report was not parsed by Go-only extractor)
FACT: FINAL B78 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset sorted matches | baseline PASS | mutated FAIL assertion `expect(output).toContain(`e2e-reset: project ${project} done (prometheus-data kept)`);`; verbatim failure "expected 'docker compose -f docker-compose.yml …' to contain 'e2e-reset: project resolved-project d…'" | CLOSED (supersedes provisional parser label REMAINING; raw assertion report was not parsed by Go-only extractor)
FACT: FINAL B79 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset project key collision | baseline PASS | mutated FAIL assertion `expect(final.volumes, 'only prometheus and foreign volumes remain').toEqual([`; verbatim failure "only prometheus and foreign volumes remain: expected [ …(3) ] to deeply equal [ …(2) ]" | CLOSED (supersedes provisional parser label REMAINING; raw assertion report was not parsed by Go-only extractor)
FACT: FINAL B80 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset missing volumes diagnostic | baseline PASS | mutated FAIL assertion `expect(output).toContain('postgres-data is not a volume of project resolved-project');`; verbatim failure "expected 'docker compose -f docker-compose.yml …' to contain 'postgres-data is not a volume of proj…'" | CLOSED (supersedes provisional parser label REMAINING; raw assertion report was not parsed by Go-only extractor)
FACT: FINAL B81 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset sorted matches | baseline PASS | mutated FAIL assertion `expect(final.volumes, 'only prometheus and foreign volumes remain').toEqual([`; verbatim failure "only prometheus and foreign volumes remain: expected [ [ 'alias-2', …(2) ], …(3) ] to deeply equal [ …(2) ]" | CLOSED (supersedes provisional parser label REMAINING; raw assertion report was not parsed by Go-only extractor)
FACT: Wiring fixture baseline attempts 2–4 failed setup before assertions: users column is email_verified (boolean), ffmpeg source named original needs explicit -f mp4, and ContainerIP returns an empty string when attached to multiple networks. Corrected test schema/format and select exact isolated-network endpoint via Inspect; ambient services/volumes untouched. Raw .tmp/gap-wiring-baseline{2,3,4}.log. No correct unmutated behavior assertion has failed; DEFECT=0 so far.
FACT: Frontend scoped oxlint --deny-warnings and full tsc --noEmit both exit 0 (.tmp/gap-frontend-{lint,typecheck}.log). Production diff checks remain empty.
NEXT: Await corrected executable baseline, then run exact B47/B50 and final B29 proof.
NEXT: Apply exact B47 mutation; command `go test -count=1 -tags integration ./internal/testsupport/startup -run 'TestExecutableMultilingualWiring/api'`; baseline PASS from current unmutated targeted run.
FACT: B47 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B47.log
```diff
--- backend/cmd/api/main.go
+++ backend/cmd/api/main.go
@@ -171,7 +171,7 @@
 		mediaRepository := mediapostgres.NewRepository(dependencies.Database)
 		mediaQueue := mediaasynq.NewQueue(dependencies.Redis)
 		mediaTickets := mediaapplication.NewTickets(mediaRepository, mediaQueue, time.Now)
-		audioTracks := mediaapplication.NewAudioTracks(mediaRepository, mediaffmpeg.New("/media"), multilingualDefault, time.Now)
+		audioTracks := mediaapplication.NewAudioTracks(mediaRepository, mediaffmpeg.New("/media"), !multilingualDefault, time.Now)
 		mediaHandler := mediahttp.NewHandler(mediaTickets, audioTracks)
 		imageRepository := imagepostgres.NewRepository(dependencies.Database)
 		// The storage adapter opens an existing root (os.OpenRoot does not create it); the API owns
```
FACT: B47 | targeted behavior test (see raw output) | baseline PASS |             wiring_integration_test.go:226: API default=0x9cfa056c740 status=200 err=<nil> want=true /             wiring_integration_test.go:226: API default=0x9cfa0692528 status=200 err=<nil> want=false | CLOSED
FACT: B47 restored original production bytes; `git diff --quiet -- backend/cmd/api/main.go` exit 0
NEXT: Apply exact B50 mutation; command `go test -count=1 -tags integration ./internal/testsupport/startup -run 'TestExecutableMultilingualWiring/media-worker'`; baseline PASS from current unmutated targeted run.
FACT: Executable baseline PASS exit 0 (.tmp/gap-wiring-baseline5.log), actual API/worker binaries on ephemeral Kiri dependencies with true and false controls. Worker path uses a real generated two-second audio/video source and Postgres LISTEN/NOTIFY after the production update, no scheduling sleeps. Runtime image courses-integration-go:1.27.2-ffmpeg is already provisioned by infrastructure/tests/integration.sh.
NEXT: Exact B47/B50 mutation proof underway; final DB baseline plus B29/B27 after synchronization/cleanup delta, then gofmt/vet/unit/dependent integrations.
FACT: B50 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B50.log
```diff
--- backend/cmd/media-worker/main.go
+++ backend/cmd/media-worker/main.go
@@ -74,7 +74,7 @@
 	}()
 	repository := mediapostgres.NewRepository(dependencies.Database)
 	queue := mediaasynq.NewQueue(dependencies.Redis)
-	transcode := mediaapplication.NewTranscode(repository, storage, mediaffmpeg.New(mediaRoot), queue, multilingualDefault, time.Now)
+	transcode := mediaapplication.NewTranscode(repository, storage, mediaffmpeg.New(mediaRoot), queue, !multilingualDefault, time.Now)
 	if err := transcode.Recover(ctx); err != nil {
 		return errors.New("media recovery failed")
 	}
```
FACT: B50 | targeted behavior test (see raw output) | baseline PASS |             wiring_integration_test.go:260: worker embedded multilingual=false want=true /             wiring_integration_test.go:260: worker embedded multilingual=true want=false | CLOSED
FACT: B50 restored original production bytes; `git diff --quiet -- backend/cmd/media-worker/main.go` exit 0
NEXT: Apply exact B29 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP_metadata_lock_query_failure'`; baseline PASS from current unmutated targeted run.
FACT: B29 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B29.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -278,7 +278,7 @@
 		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
 	if err != nil {
-		return domain.AudioTrack{}, persistenceError(err)
+		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
 	updated, err := scanAudioTrack(tx.QueryRow(ctx, `UPDATE audio_tracks
 		SET language=$3, label=$4, updated_at=$5, multilingual=COALESCE($6, multilingual)
```
FACT: B29 | targeted behavior test (see raw output) | baseline PASS |         gaps_integration_test.go:181: prior-read failure=video_not_found want dependency_unavailable | CLOSED
FACT: B29 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Apply exact B27 mutation; command `go test -count=1 -tags integration ./internal/media/adapters/postgres -run 'TestMediaRepositoryAgainstPostgreSQL/GAP_metadata_lock_excludes_sweep'`; baseline PASS from current unmutated targeted run.
FACT: B27 command exit=1; raw output /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-B27.log
```diff
--- backend/internal/media/adapters/postgres/audiotrack.go
+++ backend/internal/media/adapters/postgres/audiotrack.go
@@ -273,7 +273,7 @@
 		return domain.AudioTrack{}, err
 	}
 	previous, err := scanAudioTrack(tx.QueryRow(ctx, `SELECT `+audioTrackColumns+` FROM audio_tracks
- WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid FOR UPDATE`, trackID, assetID))
+ WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid`, trackID, assetID))
 	if errors.Is(err, pgx.ErrNoRows) {
 		return domain.AudioTrack{}, domain.ErrVideoNotFound
 	}
```
FACT: B27 | targeted behavior test (see raw output) | baseline PASS |         gaps_integration_test.go:247: sweeper entered before metadata completed: sweep=<nil> metadata=video_not_found; want contested row lock | CLOSED
FACT: B27 restored original production bytes; `git diff --quiet -- backend/internal/media/adapters/postgres/audiotrack.go` exit 0
NEXT: Final gofmt check: `gofmt -l .` in backend; Kiri env, GOCACHE/GOTMPDIR/TMPDIR inside copy.
FACT: Final gofmt exit=0; seconds=0.48; test/subtest events={"pass": 0, "fail": 0, "skip": 0}; package events={"pass": 0, "fail": 0, "skip": 0}; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-final-gofmt.log; gofmt stdout empty=True
NEXT: Final vet check: `go vet ./...` in backend; Kiri env, GOCACHE/GOTMPDIR/TMPDIR inside copy.
FACT: Final corrected DB GAP baseline PASS exit 0 (.tmp/gap-db-final-baseline.log); final B29 cancellation proof and B27 contested-sweep proof each fail on their exact mutants with assertions, no timeout counted. All 21 IDs have baseline PASS and attributable mutated FAIL; no DEFECT.
FACT: FINAL B25 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_multilingual_persistence/true/create | baseline PASS | mutated FAIL `gaps_integration_test.go:46: multilingual returned=false SQL=false loaded=false want=true`; assertion `t.Fatalf("multilingual returned=%v SQL=%v loaded=%v want=%v", track.Multilingual, stored, loaded.Multilingual, want)` | CLOSED | DEFECT no
FACT: FINAL B26 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_metadata_foreign_lock_scope | baseline PASS | mutated FAIL `gaps_integration_test.go:134: foreign locked track error=dependency_unavailable want video_not_found`; assertion `t.Fatalf("foreign locked track error=%v want video_not_found", err)` | CLOSED | DEFECT no
FACT: FINAL B27 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_metadata_lock_excludes_sweep | baseline PASS | mutated FAIL `gaps_integration_test.go:247: sweeper entered before metadata completed: sweep=<nil> metadata=video_not_found; want contested row lock`; assertion `t.Fatalf("sweeper entered before metadata completed: sweep=%v metadata=%v; want contested row lock", sweepErr, metaErr)` | CLOSED | DEFECT no
FACT: FINAL B29 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_metadata_lock_query_failure | baseline PASS | mutated FAIL `gaps_integration_test.go:181: prior-read failure=video_not_found want dependency_unavailable`; assertion `t.Fatalf("prior-read failure=%v want dependency_unavailable", err)` | CLOSED | DEFECT no
FACT: FINAL B34 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_metadata_siblings_and_label_rollback | baseline PASS | mutated FAIL `gaps_integration_test.go:98: label-only rewrite calls=0 err=<nil>; want 1 and nil`; assertion `t.Fatalf("label-only rewrite calls=%d err=%v; want 1 and nil", calls, err)` | CLOSED | DEFECT no
FACT: FINAL B38 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_multilingual_persistence/true/mint | baseline PASS | mutated FAIL `gaps_integration_test.go:55: multilingual returned=false SQL=false loaded=false want=true`; assertion `t.Fatalf("multilingual returned=%v SQL=%v loaded=%v want=%v", track.Multilingual, stored, loaded.Multilingual, want)` | CLOSED | DEFECT no
FACT: FINAL B39 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_multilingual_persistence/true/publish | baseline PASS | mutated FAIL `gaps_integration_test.go:69: multilingual returned=false SQL=false loaded=false want=true`; assertion `t.Fatalf("multilingual returned=%v SQL=%v loaded=%v want=%v", track.Multilingual, stored, loaded.Multilingual, want)` | CLOSED | DEFECT no
FACT: FINAL B69 | backend/internal/media/adapters/postgres/gaps_integration_test.go::TestMediaRepositoryAgainstPostgreSQL/GAP_metadata_siblings_and_label_rollback | baseline PASS | mutated FAIL `gaps_integration_test.go:88: metadata no-ops rewrote master 1 times; want 0`; assertion `t.Fatalf("metadata no-ops rewrote master %d times; want 0", calls)` | CLOSED | DEFECT no
FACT: FINAL B45 | backend/cmd/api/main_test.go::TestAPIRejectsInvalidMultilingualBeforeDependencies/invalid | baseline PASS | mutated FAIL `main_test.go:40: startup ASR=invalid error=exit status 1 output={"time":"2026-11-17T17:36:02.569855+02:00","level":"ERROR","msg":"startup failed","service":"api","reason":"POSTGRES_HOST is required"}`; assertion `t.Fatalf("startup ASR=%s error=%v output=%s; want %q", value, err, output, want)` | CLOSED | DEFECT no
FACT: FINAL B46 | backend/cmd/api/main_test.go::TestAPIRejectsInvalidMultilingualBeforeDependencies/invalid | baseline PASS | mutated FAIL `main_test.go:40: startup ASR=invalid error=exit status 1 output={"time":"2026-11-17T17:36:08.19286+02:00","level":"ERROR","msg":"startup failed","service":"api","reason":"POSTGRES_HOST is required"}`; assertion `t.Fatalf("startup ASR=%s error=%v output=%s; want %q", value, err, output, want)` | CLOSED | DEFECT no
FACT: FINAL B48 | backend/cmd/media-worker/main_test.go::TestWorkerRejectsInvalidMultilingualBeforeDependencies/invalid | baseline PASS | mutated FAIL `main_test.go:22: worker startup ASR=invalid error=POSTGRES_HOST is required; want "ASR_MULTILINGUAL_DEFAULT must be true or false"`; assertion `t.Fatalf("worker startup ASR=%s error=%v; want %q", value, err, want)` | CLOSED | DEFECT no
FACT: FINAL B49 | backend/cmd/media-worker/main_test.go::TestWorkerRejectsInvalidMultilingualBeforeDependencies/invalid | baseline PASS | mutated FAIL `main_test.go:22: worker startup ASR=invalid error=POSTGRES_HOST is required; want "ASR_MULTILINGUAL_DEFAULT must be true or false"`; assertion `t.Fatalf("worker startup ASR=%s error=%v; want %q", value, err, want)` | CLOSED | DEFECT no
FACT: FINAL B47 | backend/internal/testsupport/startup/wiring_integration_test.go::TestExecutableMultilingualWiring/api/true | baseline PASS | mutated FAIL `wiring_integration_test.go:226: API default=0x9cfa056c740 status=200 err=<nil> want=true`; assertion `t.Fatalf("API default=%v status=%d err=%v want=%v", list.MultilingualDefault, response.StatusCode, err, want)` | CLOSED | DEFECT no
FACT: FINAL B50 | backend/internal/testsupport/startup/wiring_integration_test.go::TestExecutableMultilingualWiring/media-worker/true | baseline PASS | mutated FAIL `wiring_integration_test.go:260: worker embedded multilingual=false want=true`; assertion `t.Fatalf("worker embedded multilingual=%v want=%v", stored, want)` | CLOSED | DEFECT no
NEXT: Run final gofmt/vet/all-unit/dependent integration sequence, then final restored-production checks, diff and file fingerprints.
FACT: Final vet exit=0; seconds=3.37; test/subtest events={"pass": 0, "fail": 0, "skip": 0}; package events={"pass": 0, "fail": 0, "skip": 0}; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-final-vet.log
NEXT: Final unit check: `go test -json ./...` in backend; Kiri env, GOCACHE/GOTMPDIR/TMPDIR inside copy.
FACT: Final unit exit=0; seconds=58.63; test/subtest events={"pass": 2794, "fail": 0, "skip": 0}; package events={"pass": 61, "fail": 0, "skip": 26}; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-final-unit.log
NEXT: Final integration check: `go test -count=1 -json -tags integration ./cmd/api ./cmd/media-worker ./cmd/upload ./internal/media/adapters/postgres ./internal/media/adapters/tus ./internal/testsupport/startup` in backend; Kiri env, GOCACHE/GOTMPDIR/TMPDIR inside copy.
FACT: FINAL B56 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset partial resolver failure | baseline PASS (7/7; no skips) | mutated FAIL assertion `expect(run.status, output).not.toBe(0);`; verbatim failure "docker compose -f docker-compose.yml -f docker-compose.e2e.yml down --remove-orphans\ne2e-reset: postgres-data: removed custom-postgres-data\ne2e-reset: redis-data: removed custom-redis-data\ne2e-reset: media-data: removed custom-media-data\ne2e-reset: project resolved-project done (prometheus-data kept)\n[eval]:1\nlet s = \"\"; process.stdin.on(\"data\", (d) => (s += d)).on(\"end\", () => { const c = JSON.parse(s); if (!c.name) process.exit(1); console.log(c.name); for (const [k, v] of Object.entries(c.volumes ?? {})) console.log(k + \" \" + v.name); });\n                                                                                                                                                                                                                                  ^\n\nTypeError: Cannot read properties of null (reading 'name')\n    at Socket.<anonymous> ([eval]:1:227)\n    at Socket.emit (node:events:520:35)\n    at endReadableNT (node:internal/streams/readable:1701:12)\n    at process.processTicksAndRejections (node:internal/process/task_queues:89:21)\n\nNode.js v24.11.1\ne2e-reset: cannot resolve the compose project and volume names\n: expected +0 not to be +0 // Object.is equality" | CLOSED | DEFECT no
FACT: FINAL B62 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset sorted matches | baseline PASS (7/7; no skips) | mutated FAIL assertion `expect(removal).toEqual(['volume', 'rm', 'alias-10', 'alias-2', 'custom-postgres-data']);`; verbatim failure "expected [ 'volume', 'rm', 'alias-2', …(2) ] to deeply equal [ 'volume', 'rm', 'alias-10', …(2) ]" | CLOSED | DEFECT no
FACT: FINAL B64 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset absent lookup failure | baseline PASS (7/7; no skips) | mutated FAIL assertion `expect(run.status, 'failed empty lookup must fail').not.toBe(0);`; verbatim failure "failed empty lookup must fail: expected +0 not to be +0 // Object.is equality" | CLOSED | DEFECT no
FACT: FINAL B78 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset resolved summary | baseline PASS (7/7; no skips) | mutated FAIL assertion `expect(output).toContain(`e2e-reset: project ${project} done (prometheus-data kept)`);`; verbatim failure "expected 'docker compose -f docker-compose.yml …' to contain 'e2e-reset: project resolved-project d…'" | CLOSED | DEFECT no
FACT: FINAL B79 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset project key collision | baseline PASS (7/7; no skips) | mutated FAIL assertion `expect(final.volumes, 'only prometheus and foreign volumes remain').toEqual([`; verbatim failure "only prometheus and foreign volumes remain: expected [ …(3) ] to deeply equal [ …(2) ]" | CLOSED | DEFECT no
FACT: FINAL B80 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset missing volumes diagnostic | baseline PASS (7/7; no skips) | mutated FAIL assertion `expect(output).toContain('postgres-data is not a volume of project resolved-project');`; verbatim failure "expected 'docker compose -f docker-compose.yml …' to contain 'postgres-data is not a volume of proj…'" | CLOSED | DEFECT no
FACT: FINAL B81 | frontend/src/adapters/api/e2e-reset.test.ts::e2e-reset renamed label match | baseline PASS (7/7; no skips) | mutated FAIL assertion `expect(final.volumes, 'only prometheus and foreign volumes remain').toEqual([`; verbatim failure "only prometheus and foreign volumes remain: expected [ [ 'renamed-redis', …(2) ], …(2) ] to deeply equal [ …(2) ]" | CLOSED | DEFECT no
FACT: Row inventory reconciled against input journal: B25/B26/B27/B29/B34/B38/B39/B45/B46/B47/B48/B49/B50/B56/B62/B64/B69/B78/B79/B80/B81 = 21 CLOSED; 0 DEFECT; 0 remaining. Provisional parser-generated REMAINING labels for TS logs are superseded by final parsed assertion records. B51 EQUIVALENT is outside this GAP task and unchanged.
FACT: Final integration exit=0; seconds=92.17; test/subtest events={"pass": 113, "fail": 0, "skip": 0}; package events={"pass": 5, "fail": 0, "skip": 1}; raw /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/.tmp/gaps-final-integration.log
FACT: Final `git diff --name-only` (exit 0):
```text
backend/internal/media/adapters/postgres/repository_integration_test.go
```
FACT: Final `git diff --stat` (exit 0; excludes six unstaged new files by Git semantics):
```text
 .../internal/media/adapters/postgres/repository_integration_test.go  | 5 +++++
 1 file changed, 5 insertions(+)
```
FACT: Full test-only change inventory (six new untracked files intentionally not staged; existing registration file +5 lines):
FACT: SHA-256 82c443128b766d08abb552cb235b4f09e88d30ae63826ba3aa4f5924435963e5  backend/cmd/api/main_test.go (44 total lines; new untracked)
FACT: SHA-256 3bec440a97059a8fd3562358ac634836ca72f71387d3962e619675c63e9a467e  backend/cmd/media-worker/main_test.go (26 total lines; new untracked)
FACT: SHA-256 9f6001f80a24444150240039303ca8084500cd92355c6846a97e7449af50885b  backend/internal/media/adapters/postgres/gaps_integration_test.go (252 total lines; new untracked)
FACT: SHA-256 c2673d7b9435ffa9af30a5bddcbbcb33ae4e996b264423ada94bca7c7605dc1e  backend/internal/media/adapters/postgres/repository_integration_test.go (951 total lines; tracked modified)
FACT: SHA-256 a2a2ea8f2979920e07eea86cffb3040481130b6416f56ad594bfdb01a0b672e6  backend/internal/testsupport/startup/wiring_integration_test.go (276 total lines; new untracked)
FACT: SHA-256 b04c850eb5d90020aaa8fa61daa7da83782a388eaf8b6524df09f16a9fcd0174  frontend/src/adapters/api/e2e-reset.test.ts (113 total lines; new untracked)
FACT: SHA-256 11df7646959c5e5747dacb5f7411306735f9e0bb10e61a67de7ce818a42c01ec  frontend/src/adapters/api/test-fixtures/reset-docker.mjs (40 total lines; new untracked)
FACT: Including six new files: 7 test/fixture files, 756 added lines, 0 deletions. `git diff --check` exit 0. All tracked non-test bytes compare identically to original base 5a56965 (git diff --quiet excluding test suffixes exit 0); all mutated production files compare identically to current HEAD. No stage/commit/push, no main working tree edits, no shared Compose/volume/reset command; reset tests used fake Docker subprocess only.
FACT: FINAL CHECKS — gofmt -l . exit 0, stdout empty; go vet ./... exit 0; go test -json ./... exit 0 (2794 test/subtest passes, 0 failed, 0 skipped; 61 tested packages passed, 26 packages have no test files); dependent integration go test -count=1 -json -tags integration ./cmd/api ./cmd/media-worker ./cmd/upload ./internal/media/adapters/postgres ./internal/media/adapters/tus ./internal/testsupport/startup exit 0 (113 test/subtest passes, 0 failed, 0 skipped; 5 tested packages passed, cmd/upload has no test files). Kiri ephemeral dependencies, copy-local GOCACHE/GOTMPDIR/TMPDIR; raw commands/output in .tmp/gaps-final-{gofmt,vet,unit,integration}.log.
FACT: Frontend final reset baseline 7/7 PASS, no skips; scoped formatting and oxlint PASS; full tsc --noEmit PASS. No frontend product files changed. Actual remote CI execution and browser/full make-check gates were not run or claimed; this delivery follows the owner-scoped be-b checks. New reset tests are selected by existing adapter-node project; new executable/service tests selected by existing integration ./... gate (not skipped).
FACT: FINAL OUTCOME — CLOSED 21 / DEFECT 0 / remaining 0. Production bytes restored after every mutant and rechecked after all final gates. Preserved pre-existing be-a commit 4e569f6; all production still matches 5a56965. Seven test/fixture files (+756/-0), six new files untracked; not staged or committed. All seven SHA-256 values and exact git diff --stat are immediately above. Agents 0 launched / 0 returned. Author verification complete; no independent-review PASS claimed.
NEXT: Owner/existing PR #109 reviewer reviews the seven test/fixture files and this evidence journal; no be-b GAP row remains to resume.
