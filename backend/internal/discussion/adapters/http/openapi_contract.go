package http

import (
	"courses/internal/discussion/domain"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"reflect"
)

// OpenAPIContribution describes discussion transport; the existing decoders remain unchanged.
func OpenAPIContribution() openapigen.ModuleContribution {
	kinds := make([]string, len(domain.ReportKinds))
	for i, value := range domain.ReportKinds {
		kinds[i] = string(value)
	}
	states := make([]string, len(domain.ReportStates))
	for i, value := range domain.ReportStates {
		states[i] = string(value)
	}
	queueKinds := make([]string, len(moderationKinds))
	for i, value := range moderationKinds {
		queueKinds[i] = string(value)
	}
	stateFilters := append(append([]string{}, states...), participantReportAllStates)
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "GET", Path: "/api/v1/courses/{slug}/reviews",
				Params: []openapigen.Param{
					{Name: "slug", In: "path", Type: "string", Required: true},
					{Name: "cursor", In: "query", Type: "string", Required: false},
				}, Responses: map[int]openapigen.Response{
					200: {Body: publicReviewsDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/learn/courses/{courseId}/reviews",
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
					{Name: "cursor", In: "query", Type: "string", Required: false},
				}, Responses: map[int]openapigen.Response{
					200: {Body: publicReviewsDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/learn/courses/{courseId}/review",
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					200: {Body: ownReviewDTO{}},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "PUT", Path: "/api/v1/learn/courses/{courseId}/review",
				Request: putReviewRequest{},
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					200: {Body: reviewDTO{}},
					201: {Body: reviewDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "DELETE", Path: "/api/v1/learn/courses/{courseId}/review",
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					204: {Empty: true},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/learn/reviews/{reviewId}/reports",
				Params: []openapigen.Param{
					{Name: "reviewId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					204: {Empty: true},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/learn/lessons/{lessonId}/comments",
				Params: []openapigen.Param{
					{Name: "lessonId", In: "path", Type: "string", Required: true},
					{Name: "cursor", In: "query", Type: "string", Required: false},
				}, Responses: map[int]openapigen.Response{
					200: {Body: pageDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/learn/lessons/{lessonId}/comments",
				Request: publishRequest{},
				Params: []openapigen.Param{
					{Name: "lessonId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					201: {Body: commentDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "PlaybackForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "PATCH", Path: "/api/v1/learn/comments/{commentId}",
				Request: updateRequest{},
				Params: []openapigen.Param{
					{Name: "commentId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					200: {Body: commentDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "DELETE", Path: "/api/v1/learn/comments/{commentId}",
				Params: []openapigen.Param{
					{Name: "commentId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					204: {Empty: true},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/learn/comments/{commentId}/reports",
				Params: []openapigen.Param{
					{Name: "commentId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					201: {Empty: true},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/learn/courses/{courseId}/report-lecturer",
				Request: participantReportRequest{},
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					201: {Body: participantReportDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/teach/courses/{courseId}/comments",
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
					{Name: "cursor", In: "query", Type: "string", Required: false},
				}, Responses: map[int]openapigen.Response{
					200: {Body: courseCommentsDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/teach/courses/{courseId}/moderation-queue",
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
					{Name: "cursor", In: "query", Type: "string", Required: false},
				}, Responses: map[int]openapigen.Response{
					200: {Body: queueDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/teach/comments/{commentId}/dismiss-report",
				Request: dismissReportsRequest{},
				Params: []openapigen.Param{
					{Name: "commentId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					204: {Empty: true},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/teach/reviews/{reviewId}/remove",
				Request: dismissReportsRequest{},
				Params: []openapigen.Param{
					{Name: "reviewId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					204: {Empty: true},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/teach/reviews/{reviewId}/dismiss-report",
				Request: dismissReportsRequest{},
				Params: []openapigen.Param{
					{Name: "reviewId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					204: {Empty: true},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/teach/courses/{courseId}/participants/{userId}/report",
				Request: participantReportRequest{},
				Params: []openapigen.Param{
					{Name: "courseId", In: "path", Type: "string", Required: true},
					{Name: "userId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					201: {Body: participantReportDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/admin/participant-reports",
				Params: []openapigen.Param{
					{Name: "cursor", In: "query", Type: "string", Required: false},
					{Name: "state", In: "query", Type: "string", Required: false, Enum: stateFilters},
				}, Responses: map[int]openapigen.Response{
					200: {Body: participantReportQueueDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/admin/participant-reports/{reportId}/resolve",
				Params: []openapigen.Param{
					{Name: "reportId", In: "path", Type: "string", Required: true},
				}, Responses: map[int]openapigen.Response{
					204: {Empty: true},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[publicReviewsDTO]():              "PublicReviewPage",
			reflect.TypeFor[reviewDTO]():                     "Review",
			reflect.TypeFor[reviewRatingDTO]():               "Rating",
			reflect.TypeFor[ownReviewDTO]():                  "OwnReview",
			reflect.TypeFor[putReviewRequest]():              "PutReviewRequest",
			reflect.TypeFor[authorDTO]():                     "CommentAuthor",
			reflect.TypeFor[commentDTO]():                    "Comment",
			reflect.TypeFor[threadDTO]():                     "CommentThread",
			reflect.TypeFor[pageDTO]():                       "CommentPage",
			reflect.TypeFor[publishRequest]():                "PublishCommentRequest",
			reflect.TypeFor[updateRequest]():                 "UpdateCommentRequest",
			reflect.TypeFor[dismissReportsRequest]():         "DismissReportsRequest",
			reflect.TypeFor[participantReportRequest]():      "ParticipantReportRequest",
			reflect.TypeFor[participantReportDTO]():          "ParticipantReport",
			reflect.TypeFor[participantReportQueueItemDTO](): "ParticipantReportQueueItem",
			reflect.TypeFor[participantReportQueueDTO]():     "ParticipantReportQueuePage",
			reflect.TypeFor[moderationCommentDTO]():          "TeachComment",
			reflect.TypeFor[courseCommentsDTO]():             "TeachCourseCommentPage",
			reflect.TypeFor[queueItemDTO]():                  "ModerationQueueItem",
			reflect.TypeFor[queueDTO]():                      "ModerationQueuePage",
		},
		Enums: map[reflect.Type][]string{
			reflect.TypeFor[domain.ReportKind]():  kinds,
			reflect.TypeFor[domain.ReportState](): states,
			reflect.TypeFor[moderationKind]():     queueKinds,
		},
		NullableFields: map[openapigen.Field]bool{
			{Type: reflect.TypeFor[publicReviewsDTO](), Name: "Items"}:          false,
			{Type: reflect.TypeFor[threadDTO](), Name: "Replies"}:               false,
			{Type: reflect.TypeFor[pageDTO](), Name: "Threads"}:                 false,
			{Type: reflect.TypeFor[participantReportQueueDTO](), Name: "Items"}: false,
			{Type: reflect.TypeFor[courseCommentsDTO](), Name: "Comments"}:      false,
			{Type: reflect.TypeFor[queueDTO](), Name: "Items"}:                  false,
			{Type: reflect.TypeFor[publishRequest](), Name: "PositionMs"}:       true,
			{Type: reflect.TypeFor[publishRequest](), Name: "ParentID"}:         true,
		},
	}
}
