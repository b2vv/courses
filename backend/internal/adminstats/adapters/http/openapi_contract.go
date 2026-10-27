package http

import (
	"courses/internal/adminstats/domain"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"reflect"
)

// OpenAPIContribution describes the admin statistics transport without exporting DTOs.
func OpenAPIContribution() openapigen.ModuleContribution {
	errors := func() map[int]openapigen.Response {
		return map[int]openapigen.Response{
			401: {Body: httpserver.ErrorResponse{}, Component: "Unauthenticated"},
			403: {Body: httpserver.ErrorResponse{}, Component: "GrantForbidden"},
			500: {Body: httpserver.ErrorResponse{}, Component: "InternalError"},
			503: {Body: httpserver.ErrorResponse{}, Component: "Unavailable"},
		}
	}
	snapshot := errors()
	snapshot[200] = openapigen.Response{Body: snapshotDTO{}}
	series := errors()
	series[200] = openapigen.Response{Body: []timeseriesPointDTO{}}
	series[400] = openapigen.Response{Body: httpserver.ErrorResponse{}, Component: "InvalidRequest"}
	return openapigen.ModuleContribution{
		Routes: []openapigen.RouteBinding{
			{Method: "GET", Path: "/api/v1/admin/stats", Responses: snapshot},
			{Method: "GET", Path: "/api/v1/admin/stats/timeseries", Responses: series, Params: []openapigen.Param{
				{Name: "metric", In: "query", Type: "string", Required: true, Enum: enumStrings(domain.Metrics)},
				{Name: "bucket", In: "query", Type: "string", Required: true, Enum: enumStrings(domain.Buckets)},
				{Name: "from", In: "query", Type: "string"},
				{Name: "to", In: "query", Type: "string"},
			}},
		},
		NameMap: map[reflect.Type]string{
			reflect.TypeFor[snapshotDTO]():          "AdminStatsSnapshot",
			reflect.TypeFor[usersDTO]():             "AdminStatsUsers",
			reflect.TypeFor[roleDTO]():              "AdminStatsRole",
			reflect.TypeFor[coursesDTO]():           "AdminStatsCourses",
			reflect.TypeFor[grantsDTO]():            "AdminStatsGrants",
			reflect.TypeFor[ordersDTO]():            "AdminStatsOrders",
			reflect.TypeFor[reportsDTO]():           "AdminStatsReports",
			reflect.TypeFor[mediaDTO]():             "AdminStatsMedia",
			reflect.TypeFor[engagementDTO]():        "AdminStatsEngagement",
			reflect.TypeFor[instructorDTO]():        "AdminStatsInstructor",
			reflect.TypeFor[platformDTO]():          "AdminStatsPlatform",
			reflect.TypeFor[timeseriesPointDTO]():   "AdminStatsTimeseriesPoint",
			reflect.TypeFor[[]timeseriesPointDTO](): "AdminStatsTimeseries",
		},
		NullableFields: map[openapigen.Field]bool{
			{Type: reflect.TypeFor[snapshotDTO](), Name: "Instructors"}: false,
		},
	}
}

func enumStrings[T ~string](values []T) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = string(value)
	}
	return result
}
