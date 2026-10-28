// Package openapispec is the leaf composition root for the parallel OpenAPI artifact.
// Business modules and the generator must not import this collector.
package openapispec

import (
	adminstatshttp "courses/internal/adminstats/adapters/http"
	coursehttp "courses/internal/course/adapters/http"
	discussionhttp "courses/internal/discussion/adapters/http"
	identityhttp "courses/internal/identity/adapters/http"
	imagehttp "courses/internal/image/adapters/http"
	mediahttp "courses/internal/media/adapters/http"
	ordershttp "courses/internal/orders/adapters/http"
	"courses/internal/platform/openapigen"
	playbackhttp "courses/internal/playback/adapters/http"
	transcripthttp "courses/internal/transcript/adapters/http"
)

// AllOpenAPIContributions collects the modules migrated to Go-owned contracts.
func AllOpenAPIContributions() []openapigen.ModuleContribution {
	return []openapigen.ModuleContribution{coursehttp.OpenAPIContribution(), adminstatshttp.OpenAPIContribution(), imagehttp.OpenAPIContribution(), mediahttp.OpenAPIContribution(), playbackhttp.OpenAPIContribution(), transcripthttp.OpenAPIContribution(), discussionhttp.OpenAPIContribution(), identityhttp.OpenAPIContribution(), ordershttp.OpenAPIContribution()}
}
