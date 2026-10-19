# T04 follow-up from Courses 1 review — LearningCourse carries rating/lessonCount without data

Keep logging to `work/codex-t04.log`. Do not commit.

Finding: `learningCourseDTO` embeds `courseDTO`, so the learning response (and the `LearningCourse`
OpenAPI schema you extended) now carries `rating` and `lessonCount`. But the learning query in
`backend/internal/course/adapters/postgres/learning.go` was not changed, so for every learning course
`lessonCount` is 0 and `rating` is null — a contract field that is always false.

Fix by giving them real values (do not drop them — the embedding makes that brittle):
1. `lessonCount` on the learning course = the number of lessons in the modules the learning query already
   loads (no new SQL needed); set it where the learning domain course is built.
2. `rating` on the learning course: add the same `LEFT JOIN LATERAL` over `discussion_reviews`
   (`removed_at IS NULL`, count + sum of stars) to the learning course query and map it through
   `rating.Summary`, exactly as the catalog does.
3. Test: an integration subtest in the course repository tests — one course with lessons in two modules
   and 3 counted reviews (4, 5, 5): the learning course and `GetPublishedCourse` report the same
   `lessonCount` and the same `{4.7, 3}`; a course with 2 reviews → `rating` null in both. Add a learning
   HTTP contract assertion that `lessonCount` equals the number of lessons in `modules`.

Also: the generated sqlc file was edited by hand because Docker is denied to you. Leave it as close to
sqlc's output as you can; Courses 1 will run `make generate` and `check-generated`.

Final answer: changed files, tests added, what you ran, what you could not.
