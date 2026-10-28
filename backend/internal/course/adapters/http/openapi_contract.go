package http

import (
	"reflect"
	"time"

	"courses/internal/course/domain"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPIContribution owns catalog, learning and admin course contracts without exporting DTOs.
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
	documentResponses := func(success openapigen.Response, conflict bool) map[int]openapigen.Response {
		responses := errors()
		responses[200] = success
		responses[401] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"}
		responses[403] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "LearningForbidden"}
		responses[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "NotFound"}
		if conflict {
			responses[409] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "LessonDocumentConflict"}
		}
		return responses
	}
	adminResponses := func(status int, body openapigen.Response, mutation, detail bool) map[int]openapigen.Response {
		responses := errors()
		responses[status] = body
		responses[400] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "InvalidRequest"}
		responses[401] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"}
		responses[403] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "Forbidden"}
		if mutation {
			responses[400] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "AdminWriteBadRequest"}
			responses[403] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "AdminMutationForbidden"}
			responses[409] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "TranscriptConflict"}
		}
		if detail {
			responses[404] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "NotFound"}
		}
		return responses
	}
	idParam := openapigen.Param{Name: "id", In: "path", Type: "string", Required: true}
	// Frozen AdminLesson makes isPreview optional even though responses always emit it.
	// Request pointers and response values share this same non-nullable wire schema.
	closed := false
	adminLesson := openapi3.NewObjectSchema().WithProperty("id", openapi3.NewStringSchema()).
		WithProperty("title", openapi3.NewStringSchema()).WithProperty("isPreview", openapi3.NewBoolSchema())
	adminLesson.Required = []string{"id", "title"}
	adminLesson.AdditionalProperties.Has = &closed
	adminModule := openapi3.NewObjectSchema().WithProperty("id", openapi3.NewStringSchema()).WithProperty("title", openapi3.NewStringSchema())
	lessons := openapi3.NewArraySchema()
	lessons.Items = &openapi3.SchemaRef{Ref: "#/components/schemas/AdminLesson"}
	adminModule.Properties["lessons"] = &openapi3.SchemaRef{Value: lessons}
	adminModule.Required = []string{"id", "lessons", "title"}
	adminModule.AdditionalProperties.Has = &closed
	// optionalPriceDTO is decode-only; its Value/Set bookkeeping is not a wire object.
	nullablePrice := &openapi3.Schema{AllOf: openapi3.SchemaRefs{{Ref: "#/components/schemas/Price"}}, Nullable: true}
	lessonParam := openapigen.Param{Name: "lessonId", In: "path", Type: "string", Required: true}
	localeParam := openapigen.Param{Name: "locale", In: "query", Type: "string", Enum: []string{"uk", "en"}}
	// Block has a custom marshaler; preserve the frozen permissive union explicitly.
	blockType := openapi3.NewStringSchema()
	for _, value := range domain.BlockTypes {
		blockType.Enum = append(blockType.Enum, string(value))
	}
	spanArray := func(nullable bool) *openapi3.Schema {
		schema := openapi3.NewArraySchema()
		schema.Items = &openapi3.SchemaRef{Ref: "#/components/schemas/LessonSpan"}
		schema.Nullable = nullable
		return schema
	}
	block := openapi3.NewObjectSchema().WithProperty("type", blockType).
		WithProperty("level", openapi3.NewIntegerSchema()).
		WithProperty("ordered", openapi3.NewBoolSchema()).
		WithProperty("spans", spanArray(false)).
		WithProperty("items", openapi3.NewArraySchema().WithItems(spanArray(true)))
	block.Required = []string{"type"}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "GET", Path: "/api/v1/admin/courses", Responses: adminResponses(200, openapigen.Response{Body: adminCoursePageDTO{}}, false, false), Params: []openapigen.Param{
				{Name: "limit", In: "query", Type: "integer"},
				{Name: "offset", In: "query", Type: "integer"},
				{Name: "status", In: "query", Type: "string", Enum: enumStrings(domain.CourseStatuses)},
				{Name: "search", In: "query", Type: "string"},
				{Name: "sort", In: "query", Type: "string", Enum: enumStrings(domain.AdminCourseSorts)},
			}},
			{Method: "POST", Path: "/api/v1/admin/courses", Request: createAdminCourseRequest{}, Responses: adminResponses(201, openapigen.Response{Body: adminCourseTreeDTO{}}, true, false)},
			{Method: "GET", Path: "/api/v1/admin/courses/{id}", Responses: adminResponses(200, openapigen.Response{Body: adminCourseTreeDTO{}}, false, true), Params: []openapigen.Param{idParam}},
			{Method: "PUT", Path: "/api/v1/admin/courses/{id}", Request: saveAdminCourseRequest{}, Responses: adminResponses(200, openapigen.Response{Body: adminCourseTreeDTO{}}, true, true), Params: []openapigen.Param{idParam}},
			{Method: "DELETE", Path: "/api/v1/admin/courses/{id}", Request: deleteAdminCourseRequest{}, Responses: adminResponses(204, openapigen.Response{Empty: true}, true, true), Params: []openapigen.Param{idParam}},
			{Method: "GET", Path: "/api/v1/admin/lessons/{lessonId}/document", Responses: documentResponses(openapigen.Response{Body: lessonDocumentDraftResponse{}}, false), Params: []openapigen.Param{lessonParam, localeParam}},
			{Method: "PUT", Path: "/api/v1/admin/lessons/{lessonId}/document", Request: saveLessonDocumentRequest{}, Responses: documentResponses(openapigen.Response{Body: lessonDocumentUpdatedResponse{}}, true), Params: []openapigen.Param{lessonParam}},
			{Method: "POST", Path: "/api/v1/admin/lessons/{lessonId}/document/publish", Request: lessonDocumentUpdatedRequest{}, Responses: documentResponses(openapigen.Response{Empty: true}, true), Params: []openapigen.Param{lessonParam, localeParam}},
			{Method: "GET", Path: "/api/v1/learn/lessons/{lessonId}/document", Responses: documentResponses(openapigen.Response{Body: domain.LessonDocument{}}, false), Params: []openapigen.Param{lessonParam, localeParam}},
			{Method: "POST", Path: "/api/v1/admin/courses/{id}/status", Request: transitionAdminCourseRequest{}, Responses: adminResponses(200, openapigen.Response{Body: adminCourseTreeDTO{}}, true, true), Params: []openapigen.Param{
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
			reflect.TypeFor[domain.SectionType]():        enumStrings(domain.SectionTypes),
			reflect.TypeFor[domain.Mark]():               enumStrings(domain.Marks),
			reflect.TypeFor[domain.LearningVideoState](): enumStrings(domain.LearningVideoStates),
			reflect.TypeFor[domain.Status]():             enumStrings(domain.CourseStatuses),
			reflect.TypeFor[domain.BlockType]():          enumStrings(domain.BlockTypes),
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[adminCoursePageDTO]():         "AdminCoursePage",
			reflect.TypeFor[adminCourseDTO]():             "AdminCourse",
			reflect.TypeFor[adminCourseStatusCountsDTO](): "AdminCourseStatusCounts",
			reflect.TypeFor[adminCourseTreeDTO]():         "AdminCourseTree",
			reflect.TypeFor[adminCourseWriteRequest]():    "AdminCourseWrite",
			reflect.TypeFor[createAdminCourseRequest]():   "AdminCourseWrite",
			reflect.TypeFor[saveAdminCourseRequest]():     "AdminCourseSaveRequest",
			reflect.TypeFor[deleteAdminCourseRequest]():   "AdminCourseDeleteRequest",
			reflect.TypeFor[adminLessonDTO]():             "AdminLesson",
			reflect.TypeFor[saveAdminLessonRequest]():     "AdminLesson",
			reflect.TypeFor[adminModuleDTO]():             "AdminModule",
			reflect.TypeFor[saveAdminModuleRequest]():     "AdminModule",
			reflect.TypeFor[optionalPriceDTO]():           "NullablePrice",

			reflect.TypeFor[domain.LessonDocument]():         "LessonDocument",
			reflect.TypeFor[domain.Section]():                "LessonSection",
			reflect.TypeFor[domain.RichText]():               "LessonRichText",
			reflect.TypeFor[domain.Block]():                  "LessonBlock",
			reflect.TypeFor[domain.Span]():                   "LessonSpan",
			reflect.TypeFor[lessonDocumentDraftResponse]():   "LessonDocumentDraft",
			reflect.TypeFor[lessonDocumentUpdatedResponse](): "LessonDocumentUpdated",
			reflect.TypeFor[saveLessonDocumentRequest]():     "SaveLessonDocumentRequest",
			reflect.TypeFor[lessonDocumentUpdatedRequest]():  "PublishLessonDocumentRequest",

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
			reflect.TypeFor[optionalPriceDTO]():       nullablePrice,
			reflect.TypeFor[adminLessonDTO]():         adminLesson,
			reflect.TypeFor[saveAdminLessonRequest](): adminLesson,
			reflect.TypeFor[adminModuleDTO]():         adminModule,
			reflect.TypeFor[saveAdminModuleRequest](): adminModule,

			reflect.TypeFor[domain.Block](): block,
			reflect.TypeFor[time.Time]():    openapi3.NewDateTimeSchema(),
		},
		FieldRefs: map[openapigen.Field]string{
			{Type: reflect.TypeFor[adminCourseDTO](), Name: "UpdatedAt"}:                   "AdminTimestamp",
			{Type: reflect.TypeFor[adminCourseTreeDTO](), Name: "CreatedAt"}:               "AdminTimestamp",
			{Type: reflect.TypeFor[adminCourseTreeDTO](), Name: "UpdatedAt"}:               "AdminTimestamp",
			{Type: reflect.TypeFor[saveAdminCourseRequest](), Name: "ExpectedUpdatedAt"}:   "AdminTimestamp",
			{Type: reflect.TypeFor[deleteAdminCourseRequest](), Name: "ExpectedUpdatedAt"}: "AdminTimestamp",

			{Type: reflect.TypeFor[transitionAdminCourseRequest](), Name: "ExpectedUpdatedAt"}: "AdminTimestamp",
		},
		NullableFields: map[openapigen.Field]bool{
			{Type: reflect.TypeFor[adminCoursePageDTO](), Name: "Items"}:                   false,
			{Type: reflect.TypeFor[adminCourseTreeDTO](), Name: "Modules"}:                 false,
			{Type: reflect.TypeFor[adminModuleDTO](), Name: "Lessons"}:                     false,
			{Type: reflect.TypeFor[adminCourseWriteRequest](), Name: "Slug"}:               false,
			{Type: reflect.TypeFor[adminCourseWriteRequest](), Name: "Title"}:              false,
			{Type: reflect.TypeFor[adminCourseWriteRequest](), Name: "Summary"}:            false,
			{Type: reflect.TypeFor[adminCourseWriteRequest](), Name: "Description"}:        false,
			{Type: reflect.TypeFor[adminCourseWriteRequest](), Name: "InstructorID"}:       false,
			{Type: reflect.TypeFor[createAdminCourseRequest](), Name: "Slug"}:              false,
			{Type: reflect.TypeFor[createAdminCourseRequest](), Name: "Title"}:             false,
			{Type: reflect.TypeFor[createAdminCourseRequest](), Name: "Summary"}:           false,
			{Type: reflect.TypeFor[createAdminCourseRequest](), Name: "Description"}:       false,
			{Type: reflect.TypeFor[createAdminCourseRequest](), Name: "InstructorID"}:      false,
			{Type: reflect.TypeFor[saveAdminCourseRequest](), Name: "ExpectedUpdatedAt"}:   false,
			{Type: reflect.TypeFor[saveAdminCourseRequest](), Name: "Course"}:              false,
			{Type: reflect.TypeFor[saveAdminCourseRequest](), Name: "Modules"}:             false,
			{Type: reflect.TypeFor[deleteAdminCourseRequest](), Name: "ExpectedUpdatedAt"}: false,

			{Type: reflect.TypeFor[domain.LessonDocument](), Name: "Sections"}:              true,
			{Type: reflect.TypeFor[domain.RichText](), Name: "Blocks"}:                      true,
			{Type: reflect.TypeFor[domain.Section](), Name: "Columns"}:                      false,
			{Type: reflect.TypeFor[domain.Section](), Name: "Content"}:                      true,
			{Type: reflect.TypeFor[domain.Block](), Name: "Spans"}:                          false,
			{Type: reflect.TypeFor[domain.Block](), Name: "Items"}:                          false,
			{Type: reflect.TypeFor[domain.Span](), Name: "Marks"}:                           false,
			{Type: reflect.TypeFor[saveLessonDocumentRequest](), Name: "ExpectedUpdatedAt"}: true,

			{Type: reflect.TypeFor[learningCourseDTO](), Name: "Modules"}:                      false,
			{Type: reflect.TypeFor[learningModuleDTO](), Name: "Lessons"}:                      false,
			{Type: reflect.TypeFor[transitionAdminCourseRequest](), Name: "Status"}:            false,
			{Type: reflect.TypeFor[transitionAdminCourseRequest](), Name: "ExpectedUpdatedAt"}: false,
			{Type: reflect.TypeFor[pageDTO](), Name: "Items"}:                                  false,
			{Type: reflect.TypeFor[courseDetailDTO](), Name: "Program"}:                        false,
			{Type: reflect.TypeFor[programModuleDTO](), Name: "Lessons"}:                       false,
		},
		Components: openapi3.Components{
			Schemas: openapi3.Schemas{"NullableAdminTimestamp": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true}}, "AdminTimestamp": {Value: openapi3.NewStringSchema()}, "NullablePrice": {Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{{Ref: "#/components/schemas/Price"}}, Nullable: true}}},
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
