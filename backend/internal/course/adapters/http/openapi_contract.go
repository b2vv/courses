package http

import (
	"reflect"
	"time"

	"courses/internal/course/domain"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPIContribution owns catalog and learning responses and the admin status request without exporting DTOs.
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
	learn := errors()
	learn[200] = openapigen.Response{Body: learningCourseDTO{}}
	learn[401] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"}
	learn[403] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"}
	learn[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "NotFound"}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "POST", Path: "/api/v1/admin/courses/{id}/status", Request: transitionAdminCourseRequest{}, Params: []openapigen.Param{
				{Name: "id", In: "path", Type: "string", Required: true},
			}},
			{Method: "GET", Path: "/api/v1/courses", Responses: list, Params: []openapigen.Param{
				{Name: "limit", In: "query", Type: "integer"},
				{Name: "offset", In: "query", Type: "integer"},
				{Name: "q", In: "query", Type: "string"},
				{Name: "sort", In: "query", Type: "string", Enum: enumStrings(domain.PublicCatalogSorts)},
			}},
			{Method: "GET", Path: "/api/v1/courses/{slug}", Responses: detail, Params: []openapigen.Param{
				{Name: "slug", In: "path", Type: "string", Required: true},
			}},
			{Method: "GET", Path: "/api/v1/learn/courses/{slug}", Responses: learn, Params: []openapigen.Param{
				{Name: "slug", In: "path", Type: "string", Required: true},
			}},
		},
		Enums: map[reflect.Type][]string{
			reflect.TypeFor[domain.LearningVideoState](): enumStrings(domain.LearningVideoStates),
			reflect.TypeFor[domain.Status]():             enumStrings(domain.CourseStatuses),
			reflect.TypeFor[domain.BlockType]():          enumStrings(domain.BlockTypes),
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[learningCourseDTO]():            "LearningCourse",
			reflect.TypeFor[learningModuleDTO]():            "LearningModule",
			reflect.TypeFor[learningLessonDTO]():            "LearningLesson",
			reflect.TypeFor[transitionAdminCourseRequest](): "AdminCourseStatusRequest",
			reflect.TypeFor[courseDTO]():                    "Course", reflect.TypeFor[courseDetailDTO](): "CourseDetail",
			reflect.TypeFor[priceDTO](): "Price", reflect.TypeFor[ratingDTO](): "Rating",
			reflect.TypeFor[programLessonDTO](): "ProgramLesson", reflect.TypeFor[programModuleDTO](): "ProgramModule",
			reflect.TypeFor[paginationDTO](): "Pagination", reflect.TypeFor[pageDTO](): "CoursePage",
		},
		Overrides: map[reflect.Type]*openapi3.Schema{
			reflect.TypeFor[time.Time](): openapi3.NewDateTimeSchema(),
		},
		FieldRefs: map[openapigen.Field]string{
			{Type: reflect.TypeFor[transitionAdminCourseRequest](), Name: "ExpectedUpdatedAt"}: "AdminTimestamp",
		},
		NullableFields: map[openapigen.Field]bool{
			{Type: reflect.TypeFor[learningCourseDTO](), Name: "Modules"}:                      false,
			{Type: reflect.TypeFor[learningModuleDTO](), Name: "Lessons"}:                      false,
			{Type: reflect.TypeFor[transitionAdminCourseRequest](), Name: "Status"}:            false,
			{Type: reflect.TypeFor[transitionAdminCourseRequest](), Name: "ExpectedUpdatedAt"}: false,
			{Type: reflect.TypeFor[pageDTO](), Name: "Items"}:                                  false,
			{Type: reflect.TypeFor[courseDetailDTO](), Name: "Program"}:                        false,
			{Type: reflect.TypeFor[programModuleDTO](), Name: "Lessons"}:                       false,
		},
		Components: openapi3.Components{
			Schemas: openapi3.Schemas{"AdminTimestamp": {Value: openapi3.NewStringSchema()}, "NullablePrice": {Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{{Ref: "#/components/schemas/Price"}}, Nullable: true}}},
			Headers: openapi3.Headers{
				"RequestID": {Value: &openapi3.Header{Parameter: openapi3.Parameter{Required: true, Schema: &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}}}},
				"NoStore":   {Value: &openapi3.Header{Parameter: openapi3.Parameter{Required: true, Schema: &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}}}},
			},
		},
	}
}

// enumStrings maps the domain-owned set without maintaining another allowlist.
func enumStrings[T ~string](values []T) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = string(value)
	}
	return result
}
