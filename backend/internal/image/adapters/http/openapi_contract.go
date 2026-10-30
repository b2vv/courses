package http

import (
	"courses/internal/image/domain"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"reflect"
)

// OpenAPIContribution describes only the public image transport; multipart metadata lives in the overlay.
func OpenAPIContribution() openapigen.ModuleContribution {
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{{Method: "POST", Path: "/api/v1/admin/lessons/{id}/images",
			Params: []openapigen.Param{{Name: "id", In: "path", Type: "string", Required: true}},
			Responses: map[int]openapigen.Response{
				201: {Body: uploadResponse{}},
				400: {Body: httpserver.ErrorResponse{}, Component: "MediaBadRequest"},
				401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
				403: {Body: httpserver.ErrorResponse{}, Component: "MediaForbidden"},
				404: {Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"},
				413: {Body: httpserver.ErrorResponse{}},
				415: {Body: httpserver.ErrorResponse{}},
				503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
			}}, {Method: "POST", Path: "/api/v1/teach/lessons/{id}/images",
			Params: []openapigen.Param{{Name: "id", In: "path", Type: "string", Required: true}},
			Responses: map[int]openapigen.Response{
				201: {Body: uploadResponse{}},
				400: {Body: httpserver.ErrorResponse{}, Component: "MediaBadRequest"},
				401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
				403: {Body: httpserver.ErrorResponse{}, Component: "MediaForbidden"},
				404: {Body: httpserver.ErrorResponse{}, Component: "LessonNotFound"},
				413: {Body: httpserver.ErrorResponse{}},
				415: {Body: httpserver.ErrorResponse{}},
				503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
			}}},
		NameMap: map[reflect.Type]string{reflect.TypeFor[uploadResponse](): "ImageUploadResponse"},
		Enums:   map[reflect.Type][]string{reflect.TypeFor[uploadContentType](): domain.ContentTypes},
	}
}
