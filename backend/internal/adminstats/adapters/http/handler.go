package http

import (
	"errors"
	"net/http"
	"time"

	"courses/internal/adminstats/application"
	"courses/internal/adminstats/domain"
	"courses/platform/httpserver"
)

// Handler serves the admin-only platform statistics endpoints.
type Handler struct {
	service *application.Service
}

// NewHandler builds the admin-stats HTTP handler.
func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

// RegisterAdmin wires the two read-only endpoints behind authenticate +
// administrator (no same-origin: both are GETs). The administrator middleware
// is the admin gate — 401 for a guest, 403 grant_forbidden for a non-admin.
func (handler *Handler) RegisterAdmin(mux *http.ServeMux, authenticate, administrator func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/admin/stats", authenticate(administrator(http.HandlerFunc(handler.snapshot))))
	mux.Handle("GET /api/v1/admin/stats/timeseries", authenticate(administrator(http.HandlerFunc(handler.timeseries))))
}

func (handler *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	snapshot, err := handler.service.Snapshot(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, snapshotToDTO(snapshot))
}

func (handler *Handler) timeseries(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	query := r.URL.Query()
	for key, values := range query {
		switch key {
		case "metric", "bucket", "from", "to":
			if len(values) != 1 {
				writeError(w, r, domain.ErrInvalidRange)
				return
			}
		default:
			writeError(w, r, domain.ErrInvalidRange)
			return
		}
	}
	points, err := handler.service.Timeseries(r.Context(), query.Get("metric"), query.Get("bucket"), query.Get("from"), query.Get("to"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, timeseriesToDTO(points))
}

func noStore(w http.ResponseWriter) { w.Header().Set("Cache-Control", "no-store") }

// writeError maps the adminstats business errors to their HTTP status. The
// four range/metric codes are 4xx (never a 500 or a silent 0); a read-path
// failure is 503; anything else is an internal error.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRange):
		httpserver.WriteError(w, r, http.StatusBadRequest, string(domain.ErrInvalidRange), "The requested range is invalid")
	case errors.Is(err, domain.ErrRangeTooWide):
		httpserver.WriteError(w, r, http.StatusBadRequest, string(domain.ErrRangeTooWide), "The requested range is too wide")
	case errors.Is(err, domain.ErrUnknownMetric):
		httpserver.WriteError(w, r, http.StatusBadRequest, string(domain.ErrUnknownMetric), "The requested metric is unknown")
	case errors.Is(err, domain.ErrUnknownBucket):
		httpserver.WriteError(w, r, http.StatusBadRequest, string(domain.ErrUnknownBucket), "The requested bucket is unknown")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, string(domain.ErrDependencyUnavailable), "A dependency is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}

type roleDTO struct {
	Student    int `json:"student"`
	Instructor int `json:"instructor"`
	Admin      int `json:"admin"`
}

type usersDTO struct {
	Active    int     `json:"active"`
	Blocked   int     `json:"blocked"`
	Deleted   int     `json:"deleted"`
	ByRole    roleDTO `json:"byRole"`
	Verified  int     `json:"verified"`
	TwoFactor int     `json:"twoFactor"`
}

type coursesDTO struct {
	Draft     int `json:"draft"`
	Published int `json:"published"`
	Archived  int `json:"archived"`
}

type grantsDTO struct {
	Active  int `json:"active"`
	Blocked int `json:"blocked"`
	Revoked int `json:"revoked"`
}

type ordersDTO struct {
	Pending         int   `json:"pending"`
	Paid            int   `json:"paid"`
	Failed          int   `json:"failed"`
	RevenueMinorUAH int64 `json:"revenueMinorUAH"`
}

type reportsDTO struct {
	Open     int `json:"open"`
	Resolved int `json:"resolved"`
}

type mediaDTO struct {
	InProgress           int      `json:"inProgress"`
	Ready                int      `json:"ready"`
	Failed               int      `json:"failed"`
	TotalDurationSeconds float64  `json:"totalDurationSeconds"`
	AvgDurationSeconds   *float64 `json:"avgDurationSeconds"`
	WithDurationCount    int      `json:"withDurationCount"`
}

type engagementDTO struct {
	ActiveLearners30d int      `json:"activeLearners30d"`
	AvgProgress       *float64 `json:"avgProgress"`
	Completions       int      `json:"completions"`
}

type instructorDTO struct {
	InstructorID      string `json:"instructorId"`
	DisplayName       string `json:"displayName"`
	PublishedCourses  int    `json:"publishedCourses"`
	ActiveEnrollments int    `json:"activeEnrollments"`
	RevenueMinorUAH   int64  `json:"revenueMinorUAH"`
}

type platformDTO struct {
	SingleInstructor bool `json:"singleInstructor"`
}

type snapshotDTO struct {
	Users       usersDTO        `json:"users"`
	Courses     coursesDTO      `json:"courses"`
	Grants      grantsDTO       `json:"grants"`
	Orders      ordersDTO       `json:"orders"`
	Reports     reportsDTO      `json:"reports"`
	Media       mediaDTO        `json:"media"`
	Engagement  engagementDTO   `json:"engagement"`
	Instructors []instructorDTO `json:"instructors"`
	Platform    platformDTO     `json:"platform"`
}

type timeseriesPointDTO struct {
	BucketStart string  `json:"bucketStart"`
	Value       float64 `json:"value"`
}

func snapshotToDTO(snapshot domain.Snapshot) snapshotDTO {
	// A non-nil slice so an empty leaderboard renders as [] rather than null,
	// which the required array field demands.
	instructors := make([]instructorDTO, 0, len(snapshot.Instructors))
	for _, instructor := range snapshot.Instructors {
		instructors = append(instructors, instructorDTO{
			InstructorID:      instructor.InstructorID,
			DisplayName:       instructor.DisplayName,
			PublishedCourses:  instructor.PublishedCourses,
			ActiveEnrollments: instructor.ActiveEnrollments,
			RevenueMinorUAH:   instructor.RevenueMinorUAH,
		})
	}
	return snapshotDTO{
		Users: usersDTO{
			Active:    snapshot.Users.Active,
			Blocked:   snapshot.Users.Blocked,
			Deleted:   snapshot.Users.Deleted,
			ByRole:    roleDTO{Student: snapshot.Users.ByRole.Student, Instructor: snapshot.Users.ByRole.Instructor, Admin: snapshot.Users.ByRole.Admin},
			Verified:  snapshot.Users.Verified,
			TwoFactor: snapshot.Users.TwoFactor,
		},
		Courses: coursesDTO{Draft: snapshot.Courses.Draft, Published: snapshot.Courses.Published, Archived: snapshot.Courses.Archived},
		Grants:  grantsDTO{Active: snapshot.Grants.Active, Blocked: snapshot.Grants.Blocked, Revoked: snapshot.Grants.Revoked},
		Orders: ordersDTO{
			Pending:         snapshot.Orders.Pending,
			Paid:            snapshot.Orders.Paid,
			Failed:          snapshot.Orders.Failed,
			RevenueMinorUAH: snapshot.Orders.RevenueMinorUAH,
		},
		Reports: reportsDTO{Open: snapshot.Reports.Open, Resolved: snapshot.Reports.Resolved},
		Engagement: engagementDTO{
			ActiveLearners30d: snapshot.Engagement.ActiveLearners30d,
			AvgProgress:       snapshot.Engagement.AvgProgress,
			Completions:       snapshot.Engagement.Completions,
		},
		Media: mediaDTO{
			InProgress:           snapshot.Media.InProgress,
			Ready:                snapshot.Media.Ready,
			Failed:               snapshot.Media.Failed,
			TotalDurationSeconds: snapshot.Media.TotalDurationSeconds,
			AvgDurationSeconds:   snapshot.Media.AvgDurationSeconds,
			WithDurationCount:    snapshot.Media.WithDurationCount,
		},
		Instructors: instructors,
		Platform:    platformDTO{SingleInstructor: snapshot.Platform.SingleInstructor},
	}
}

func timeseriesToDTO(points []domain.TimeseriesPoint) []timeseriesPointDTO {
	out := make([]timeseriesPointDTO, 0, len(points))
	for _, point := range points {
		out = append(out, timeseriesPointDTO{BucketStart: point.BucketStart.UTC().Format(time.RFC3339), Value: point.Value})
	}
	return out
}
