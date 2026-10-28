package http

import (
	"courses/internal/clientlog/domain"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"reflect"
)

// OpenAPIContribution describes the public client-log transport without exporting DTOs.
func OpenAPIContribution() openapigen.ModuleContribution {
	levels := make([]string, len(domain.Levels))
	for i, level := range domain.Levels {
		levels[i] = string(level)
	}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{{
			Method: "POST", Path: "/api/v1/client-logs",
			Request: clientLogRequest{},
			Responses: map[int]openapigen.Response{
				204: {Empty: true},
				400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
			},
		}},
		NameMap: map[reflect.Type]string{reflect.TypeFor[clientLogRequest](): "ClientLogRequest"},
		Enums:   map[reflect.Type][]string{reflect.TypeFor[domain.Level](): levels},
	}
}
