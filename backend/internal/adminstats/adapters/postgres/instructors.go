package postgres

import (
	"context"

	"courses/internal/adminstats/domain"
)

// leaderboardLimit caps the per-instructor leaderboard to the top N rows.
const leaderboardLimit = 10

// loadInstructors builds the per-instructor leaderboard: one row per user who
// owns at least one course, ranked by paid revenue then active enrollments.
//
// Each metric is pre-aggregated in its own subquery joined on instructor_id, so
// the three independent one-to-many relations (courses→grants, courses→orders)
// never fan out and inflate one another. LEFT JOIN keeps an instructor with no
// published course, no active enrollment or no sale in the list at 0 rather than
// dropping them (9c), mirroring the snapshot's zero-fill discipline.
func (repository *Repository) loadInstructors(ctx context.Context) ([]domain.InstructorStat, error) {
	rows, err := repository.pool.Query(ctx, `
SELECT
  i.instructor_id::text AS instructor_id,
  u.display_name,
  COALESCE(pub.published_courses, 0) AS published_courses,
  COALESCE(enr.active_enrollments, 0) AS active_enrollments,
  COALESCE(rev.revenue, 0)::bigint AS revenue
FROM (SELECT DISTINCT instructor_id FROM courses) i
JOIN users u ON u.id = i.instructor_id
LEFT JOIN (
  SELECT instructor_id, COUNT(*) AS published_courses
  FROM courses WHERE status = 'published' GROUP BY instructor_id
) pub ON pub.instructor_id = i.instructor_id
LEFT JOIN (
  SELECT c.instructor_id, COUNT(*) AS active_enrollments
  FROM course_grants g JOIN courses c ON c.id = g.course_id
  WHERE g.state = 'active' GROUP BY c.instructor_id
) enr ON enr.instructor_id = i.instructor_id
LEFT JOIN (
  SELECT c.instructor_id, SUM(o.amount_minor) AS revenue
  FROM orders o JOIN courses c ON c.id = o.course_id
  WHERE o.status = 'paid' GROUP BY c.instructor_id
) rev ON rev.instructor_id = i.instructor_id
ORDER BY COALESCE(rev.revenue, 0) DESC, COALESCE(enr.active_enrollments, 0) DESC, i.instructor_id ASC
LIMIT $1`, leaderboardLimit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()

	var instructors []domain.InstructorStat
	for rows.Next() {
		var stat domain.InstructorStat
		if err := rows.Scan(&stat.InstructorID, &stat.DisplayName, &stat.PublishedCourses, &stat.ActiveEnrollments, &stat.RevenueMinorUAH); err != nil {
			return nil, persistenceError(err)
		}
		instructors = append(instructors, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return instructors, nil
}
