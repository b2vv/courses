package http

import (
	"courses/internal/platform/openapigen"
	"courses/internal/playback/domain"
	"courses/platform/httpserver"
	"reflect"
)

// OpenAPIContribution describes only the public playback transport; nginx authorization is internal.
// Request validation remains single-layer metadata; the decoders are bespoke.
func OpenAPIContribution() openapigen.ModuleContribution {
	errors := func() map[int]openapigen.Response {
		return map[int]openapigen.Response{
			400: {Body: httpserver.ErrorResponse{}, Component: "InvalidRequest"},
			401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
			403: {Body: httpserver.ErrorResponse{}, Component: "PlaybackForbidden"},
			500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
			503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
		}
	}
	play, heartbeat := errors(), errors()
	play[201] = openapigen.Response{Body: playResponse{}}
	play[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "PlaybackNotFound"}
	heartbeat[204] = openapigen.Response{}
	heartbeat[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"}
	heartbeat[409] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "PlaybackTakenOver"}
	heartbeat[410] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "PlaybackExpired"}
	reasons := make([]string, len(domain.ProgressReasons))
	for i, reason := range domain.ProgressReasons {
		reasons[i] = string(reason)
	}
	av1Values := make([]string, len(av1SupportValues))
	for i, value := range av1SupportValues {
		av1Values[i] = string(value)
	}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "POST", Path: "/api/v1/playback/sessions", Request: playRequest{}, Responses: play},
			{Method: "POST", Path: "/api/v1/playback/heartbeat", Request: heartbeatRequest{}, Responses: heartbeat},
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[playRequest]():      "PlaybackSessionRequest",
			reflect.TypeFor[playResponse]():     "PlaybackSession",
			reflect.TypeFor[heartbeatRequest](): "PlaybackHeartbeatRequest",
		},
		Enums: map[reflect.Type][]string{
			reflect.TypeFor[domain.ProgressReason](): reasons,
			reflect.TypeFor[av1Support]():            av1Values,
		},
		NullableFields: map[openapigen.Field]bool{{Type: reflect.TypeFor[playRequest](), Name: "AV1Support"}: false},
	}
}
