package http

import (
	"courses/internal/platform/openapigen"
	"courses/internal/transcript/domain"
	"courses/platform/httpserver"
	"reflect"
)

// OpenAPIContribution describes the transcript transport; bespoke decoders remain unchanged.
func OpenAPIContribution() openapigen.ModuleContribution {
	states := make([]string, len(domain.DraftStates))
	for i, state := range domain.DraftStates {
		states[i] = string(state)
	}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "GET", Path: "/api/v1/admin/courses/{courseId}/documents",
				Params: []openapigen.Param{{Name: "courseId", In: "path", Type: "string", Required: true}},
				Responses: map[int]openapigen.Response{
					200: {Body: documentsResult{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/admin/courses/{courseId}/documents",
				Request: documentDTO{},
				Params:  []openapigen.Param{{Name: "courseId", In: "path", Type: "string", Required: true}},
				Responses: map[int]openapigen.Response{
					201: {Empty: true},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "DELETE", Path: "/api/v1/admin/courses/{courseId}/documents/{documentId}",
				Params: []openapigen.Param{{Name: "courseId", In: "path", Type: "string", Required: true}, {Name: "documentId", In: "path", Type: "string", Required: true}},
				Responses: map[int]openapigen.Response{
					204: {Empty: true},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/admin/lessons/{lessonId}/draft",
				Params: []openapigen.Param{{Name: "lessonId", In: "path", Type: "string", Required: true}},
				Responses: map[int]openapigen.Response{
					200: {Body: draftDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "InvalidRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "Forbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/admin/lessons/{lessonId}/draft/publish",
				Request: publishRequest{},
				Params:  []openapigen.Param{{Name: "lessonId", In: "path", Type: "string", Required: true}},
				Responses: map[int]openapigen.Response{
					200: {Body: publishResult{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "InvalidRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "AdminMutationForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "POST", Path: "/api/v1/admin/lessons/{lessonId}/draft/regenerate",
				Request: regenerateRequest{},
				Params:  []openapigen.Param{{Name: "lessonId", In: "path", Type: "string", Required: true}},
				Responses: map[int]openapigen.Response{
					202: {Empty: true},
					400: {Body: httpserver.ErrorResponse{}, Component: "InvalidRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "AdminMutationForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "DraftConflict"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "PUT", Path: "/api/v1/admin/lessons/{lessonId}/transcript",
				Request: saveTranscriptRequest{},
				Params:  []openapigen.Param{{Name: "lessonId", In: "path", Type: "string", Required: true}},
				Responses: map[int]openapigen.Response{
					200: {Body: saveResult{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					409: {Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
			{Method: "GET", Path: "/api/v1/learn/lessons/{lessonId}/transcript",
				Params: []openapigen.Param{{Name: "lessonId", In: "path", Type: "string", Required: true}, {Name: "language", In: "query", Type: "string", Required: false}},
				Responses: map[int]openapigen.Response{
					200: {Body: transcriptDTO{}},
					400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
					401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
					403: {Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"},
					404: {Body: httpserver.ErrorResponse{}, Component: "NotFound"},
					500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
					503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
				}},
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[transcriptDTO]():         "LessonTranscript",
			reflect.TypeFor[draftDTO]():              "LessonDraft",
			reflect.TypeFor[saveTranscriptRequest](): "SaveTranscriptRequest",
			reflect.TypeFor[publishRequest]():        "PublishLessonDraftRequest",
			reflect.TypeFor[regenerateRequest]():     "RegenerateLessonDraftRequest",
			reflect.TypeFor[documentDTO]():           "TranscriptDocumentLink",
			reflect.TypeFor[segmentDTO]():            "TranscriptSegment",
			reflect.TypeFor[draftSegmentDTO]():       "DraftSegment",
		},
		Enums: map[reflect.Type][]string{reflect.TypeFor[domain.DraftState](): states},
		NullableFields: map[openapigen.Field]bool{
			{Type: reflect.TypeFor[transcriptDTO](), Name: "Segments"}:                  false,
			{Type: reflect.TypeFor[transcriptDTO](), Name: "Documents"}:                 false,
			{Type: reflect.TypeFor[draftDTO](), Name: "Segments"}:                       false,
			{Type: reflect.TypeFor[segmentDTO](), Name: "DocumentIDs"}:                  false,
			{Type: reflect.TypeFor[documentsResult](), Name: "Documents"}:               false,
			{Type: reflect.TypeFor[saveTranscriptRequest](), Name: "ExpectedUpdatedAt"}: false,
			{Type: reflect.TypeFor[saveTranscriptRequest](), Name: "Segments"}:          false,
			{Type: reflect.TypeFor[saveSegmentRequest](), Name: "Speaker"}:              true,
			{Type: reflect.TypeFor[saveSegmentRequest](), Name: "Confidence"}:           true,
			{Type: reflect.TypeFor[saveSegmentRequest](), Name: "DocumentIDs"}:          false,
			{Type: reflect.TypeFor[regenerateRequest](), Name: "Language"}:              false,
		},
	}
}
