package http

import (
	"courses/internal/media/domain"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"reflect"
)

// OpenAPIContribution describes the public media transport without exporting DTOs.
// Request validation remains single-layer metadata (§16.2); decodeMintVideo is bespoke.
func OpenAPIContribution() openapigen.ModuleContribution {
	errors := func() map[int]openapigen.Response {
		return map[int]openapigen.Response{
			400: {Body: httpserver.ErrorResponse{}, Component: "InvalidRequest"},
			401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
			403: {Body: httpserver.ErrorResponse{}, Component: "MediaForbidden"},
			500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
			503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
		}
	}
	mint, list, retry := errors(), errors(), errors()
	mint[201] = openapigen.Response{Body: uploadTicketDTO{}}
	mint[400] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "MediaBadRequest"}
	mint[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"}
	mint[409] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "UploadConflict"}
	list[200] = openapigen.Response{Body: courseVideosDTO{}}
	retry[200] = openapigen.Response{Body: videoDTO{}}
	retry[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "VideoNotFound"}
	retry[409] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "VideoNotRetryable"}
	states := make([]string, len(domain.States))
	for i, state := range domain.States {
		states[i] = string(state)
	}
	failureReasonStrings := make([]string, len(domain.FailureReasons))
	for i, reason := range domain.FailureReasons {
		failureReasonStrings[i] = string(reason)
	}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "POST", Path: "/api/v1/admin/lessons/{lessonId}/video", Request: mintVideoRequest{}, Responses: mint, Params: []openapigen.Param{{Name: "lessonId", In: "path", Type: "string", Required: true}}},
			{Method: "GET", Path: "/api/v1/admin/courses/{courseId}/videos", Responses: list, Params: []openapigen.Param{{Name: "courseId", In: "path", Type: "string", Required: true}}},
			{Method: "POST", Path: "/api/v1/admin/videos/{videoId}/retry", Responses: retry, Params: []openapigen.Param{{Name: "videoId", In: "path", Type: "string", Required: true}}},
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[mintVideoRequest](): "MediaTicketRequest",
			reflect.TypeFor[uploadTicketDTO]():  "MediaUploadTicket",
			reflect.TypeFor[videoDTO]():         "MediaVideo",
			reflect.TypeFor[courseVideosDTO]():  "CourseVideos",
		},
		Enums: map[reflect.Type][]string{
			reflect.TypeFor[domain.State]():         states,
			reflect.TypeFor[domain.FailureReason](): failureReasonStrings,
		},
		NullableFields: map[openapigen.Field]bool{{Type: reflect.TypeFor[courseVideosDTO](), Name: "Items"}: false},
		FieldRefs: map[openapigen.Field]string{
			{Type: reflect.TypeFor[uploadTicketDTO](), Name: "ExpiresAt"}: "AdminTimestamp",
			{Type: reflect.TypeFor[videoDTO](), Name: "UpdatedAt"}:        "AdminTimestamp",
		},
	}
}
