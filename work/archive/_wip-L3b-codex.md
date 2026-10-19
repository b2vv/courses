# _wip L3b (Codex) — plumb membership state into 7 readers

(Codex дописує тут після кожного кроку.)

ДАЛІ: завершити локальну перевірку й передати Courses 1; інтеграційний Docker-гейт потребує середовища з доступом до Docker socket.

2026-10-06 — 7/7 читачів конвертовано: course learning; playback session; playback media-bytes; transcript; discussion lessonFacts; discussion commentQuery; discussion reviewFacts. Кожен читає nullable `grants.state` без `state='active'` у JOIN, мапить `MembershipNone` для NULL і `access.Membership(state)` для рядка, а `Entitled` явно як `state.Valid && state.String == "active"`.

2026-10-06 — Membership прокинуто в LearningCourse, VideoCourse, transcript Course, discussion Course і ReviewFacts, а далі в усі наявні `access.Course`-літерали. `CanBeViewedBy` не змінено семантично: рівно 4, усі повертають `access.CanView(...)`; L4-гейтинг не додано.

2026-10-06 — інтеграційні assert-и додано: course learning (active/blocked/revoked); playback media-bytes і session (кожен active/blocked/revoked); transcript (active/blocked/revoked); discussion lessonFacts і commentQuery (кожен active/blocked/revoked); reviewFacts (active/blocked/revoked). Для active: Entitled=true, Membership=active; для blocked/revoked: Entitled=false і відповідний Membership.

2026-10-06 — PASS: `GOCACHE=/private/tmp/courses-l3b-gocache go test ./internal/course/... ./internal/playback/... ./internal/transcript/... ./internal/discussion/... ./platform/architecture/...`; PASS: той самий GOCACHE `go vet` на цих чотирьох модулях. `go test -tags=integration` для п’яти читачів BLOCKED: sandbox не має доступу до `/var/run/docker.sock` або `~/.docker/run/docker.sock` (permission denied), тож не позначено як PASS. generated/sqlc не змінювався; коміту немає.

2026-10-06 — фінальна локальна перевірка: PASS `go build` і повторний `go test` на course/playback/transcript/discussion + `platform/architecture`; `git diff --check` PASS. Перевірено статично: 4/4 `CanBeViewedBy`, кожен `return access.CanView(...)`; generated/sqlc у diff відсутні.

2026-10-06 — L3b-fix: ізольовано два membership-підтісти. Playback створює власні `membershipViewerID`, `membershipCourseID`, module/lesson/video через `gen_random_uuid()`; Discussion створює власні `membershipStudentID`, `membershipCourseID`, module/lesson/comment. Відповідно спільні `viewerID`/`courseID`/`lessonID`/`videoID` і `discussionStudent`/`discussionLesson`/course d1 не мутуються; assert-и active/blocked/revoked лишилися незмінними за змістом.

2026-10-06 — PASS: `gofmt` для обох integration-тестів; `git diff --check`; `GOCACHE=/private/tmp/courses-l3b-fix-gocache go build ./...`; той самий GOCACHE `go test ./internal/playback/... ./internal/discussion/...`. Docker integration не запускався: для нього потрібен Docker socket, недоступний у поточному sandbox. generated/sqlc і читачі не змінювалися; коміту немає.
