package domain

// RoleStat counts accounts by role (deleted excluded by the reader).
type RoleStat struct {
	Student    int
	Instructor int
	Admin      int
}

// UsersStat is the people domain of the snapshot.
type UsersStat struct {
	Active    int
	Blocked   int
	Deleted   int
	ByRole    RoleStat
	Verified  int
	TwoFactor int
}

// CoursesStat counts courses by lifecycle status.
type CoursesStat struct {
	Draft     int
	Published int
	Archived  int
}

// GrantsStat counts course memberships by state.
type GrantsStat struct {
	Active  int
	Blocked int
	Revoked int
}

// OrdersStat counts orders by status; RevenueMinorUAH sums paid orders.
type OrdersStat struct {
	Pending         int
	Paid            int
	Failed          int
	RevenueMinorUAH int64
}

// ReportsStat counts participant reports by state.
type ReportsStat struct {
	Open     int
	Resolved int
}

// MediaStat counts video assets by coarse state and aggregates duration.
// AvgDurationSeconds is nil when no asset carries a duration (AVG over zero
// rows), which the contract renders as nullable rather than a misleading 0.
type MediaStat struct {
	InProgress           int
	Ready                int
	Failed               int
	TotalDurationSeconds float64
	AvgDurationSeconds   *float64
	WithDurationCount    int
}

// EngagementStat aggregates learning activity across active memberships.
// AvgProgress is the mean completed-lesson fraction over active grants and is
// nil when there is no active grant (AVG over zero rows), never a misleading 0.
type EngagementStat struct {
	ActiveLearners30d int
	AvgProgress       *float64
	Completions       int
}

// InstructorStat is one row of the per-instructor leaderboard.
type InstructorStat struct {
	InstructorID      string
	DisplayName       string
	PublishedCourses  int
	ActiveEnrollments int
	RevenueMinorUAH   int64
}

// Platform carries platform-level flags surfaced to the admin UI.
type Platform struct {
	SingleInstructor bool
}

// Snapshot is the point-in-time platform aggregate.
type Snapshot struct {
	Users       UsersStat
	Courses     CoursesStat
	Grants      GrantsStat
	Orders      OrdersStat
	Reports     ReportsStat
	Media       MediaStat
	Engagement  EngagementStat
	Instructors []InstructorStat
	Platform    Platform
}
