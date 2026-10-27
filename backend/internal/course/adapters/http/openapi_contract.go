package http

import (
	"reflect"
	"time"

	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPIContribution owns the catalog's transport bindings without exporting its DTOs.
func OpenAPIContribution() openapigen.ModuleContribution {
	errors := func() map[int]openapigen.Response {
		return map[int]openapigen.Response{
			400: {Body: httpserver.ErrorResponse{}, Component: "BadRequest"},
			500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
			503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
		}
	}
	list := errors()
	list[200] = openapigen.Response{Body: pageDTO{}}
	detail := errors()
	detail[200] = openapigen.Response{Body: courseDetailDTO{}}
	detail[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "NotFound"}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "GET", Path: "/api/v1/courses", Responses: list},
			{Method: "GET", Path: "/api/v1/courses/{slug}", Responses: detail},
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[courseDTO](): "Course", reflect.TypeFor[courseDetailDTO](): "CourseDetail",
			reflect.TypeFor[priceDTO](): "Price", reflect.TypeFor[ratingDTO](): "Rating",
			reflect.TypeFor[programLessonDTO](): "ProgramLesson", reflect.TypeFor[programModuleDTO](): "ProgramModule",
			reflect.TypeFor[paginationDTO](): "Pagination", reflect.TypeFor[pageDTO](): "CoursePage",
		},
		Overrides: map[reflect.Type]*openapi3.Schema{
			reflect.TypeFor[time.Time](): openapi3.NewDateTimeSchema(),
		},
		NullableFields: map[openapigen.Field]bool{
			{Type: reflect.TypeFor[pageDTO](), Name: "Items"}:            false,
			{Type: reflect.TypeFor[courseDetailDTO](), Name: "Program"}:  false,
			{Type: reflect.TypeFor[programModuleDTO](), Name: "Lessons"}: false,
		},
		Components: openapi3.Components{
			Schemas: openapi3.Schemas{"NullablePrice": {Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{{Ref: "#/components/schemas/Price"}}, Nullable: true}}},
			Headers: openapi3.Headers{
				"RequestID": {Value: &openapi3.Header{Parameter: openapi3.Parameter{Required: true, Schema: &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}}}},
				"NoStore":   {Value: &openapi3.Header{Parameter: openapi3.Parameter{Required: true, Schema: &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}}}},
			},
		},
	}
}
