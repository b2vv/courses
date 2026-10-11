package postgres

import (
	"context"

	"courses/internal/adminstats/domain"
)

// loadEngagement aggregates learning activity across active memberships in one
// read. It is the heaviest of the snapshot's queries (it joins grants to every
// lesson of their course), so it is isolated here; at the platform's scale the
// intermediate (grant × lesson) set is small and the existing primary keys
// serve it, so no dedicated index is introduced (revisit if lesson_progress
// grows large).
//
//   - ActiveLearners30d: distinct students who hold an active grant and touched
//     a lesson of that course in the last 30 days (DB clock).
//   - AvgProgress: the mean completed-lesson fraction over active grants, where
//     a lesson counts as done by its completed_at; a course with no lessons
//     never enters the set (the join to totals requires at least one lesson), so
//     it is excluded rather than counted as 0, and no active grant at all leaves
//     the average nil (AVG over zero rows).
//   - Completions: active grants whose every course lesson is completed.
func (repository *Repository) loadEngagement(ctx context.Context) (domain.EngagementStat, error) {
	var stat domain.EngagementStat
	err := repository.pool.QueryRow(ctx, `
WITH totals AS (
  SELECT cm.course_id, COUNT(*) AS total_lessons
  FROM course_lessons cl JOIN course_modules cm ON cm.id = cl.module_id
  GROUP BY cm.course_id
),
grant_progress AS (
  SELECT g.user_id, g.course_id, t.total_lessons,
    COUNT(lp.lesson_id) FILTER (WHERE lp.completed_at IS NOT NULL) AS completed
  FROM course_grants g
  JOIN totals t ON t.course_id = g.course_id
  JOIN course_modules cm ON cm.course_id = g.course_id
  JOIN course_lessons cl ON cl.module_id = cm.id
  LEFT JOIN lesson_progress lp ON lp.lesson_id = cl.id AND lp.user_id = g.user_id
  WHERE g.state = 'active'
  GROUP BY g.user_id, g.course_id, t.total_lessons
)
SELECT
  (SELECT COUNT(DISTINCT lp.user_id)
     FROM lesson_progress lp
     JOIN course_lessons cl ON cl.id = lp.lesson_id
     JOIN course_modules cm ON cm.id = cl.module_id
     JOIN course_grants g ON g.user_id = lp.user_id AND g.course_id = cm.course_id AND g.state = 'active'
     WHERE lp.updated_at >= now() - interval '30 days') AS active_learners,
  AVG(gp.completed::double precision / gp.total_lessons) AS avg_progress,
  COUNT(*) FILTER (WHERE gp.total_lessons > 0 AND gp.completed = gp.total_lessons) AS completions
FROM grant_progress gp`).Scan(&stat.ActiveLearners30d, &stat.AvgProgress, &stat.Completions)
	return stat, persistenceError(err)
}
