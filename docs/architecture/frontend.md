# Frontend architecture

Dependency rules: domain → language only; ports → domain; application → domain/ports; adapters → ports/domain/external tools; presentation → use-case bindings/UI; app → composition of all layers. Route files translate URL/search state and delegate loaders/rendering. Generated Orval DTOs live in adapters/api/generated and are validated and mapped before entering domain.

`getRouter` creates a QueryClient, API adapter and use cases for every SSR request. TanStack Query hydration carries public catalog data; cookies remain in request-only adapter closures. A server-only helper reads inbound cookies and the internal API origin. Browser requests use same-origin `/api/v1` with credentials. Query state is never copied into Zustand.

SSR routes: `/`, `/courses`, `/courses/$slug`, `/login`. `/learn`, `/admin` and descendants render client-only locked placeholders until identity is implemented. No auth success is simulated. API calls have bounded deadlines and cancellation.

The UI is a Ukrainian technical course library: compact chapter-like labels, blue navigation and a calm white/slate surface. Repeated course metadata rows provide identity without decorative dashboard metrics. Design tokens define light/dark themes, focus and all colors. System body typography and a restrained Georgia display heading avoid external font requests. Minimum controls are 44px; layout wraps at zoom and small widths.

UI kit wraps React Aria primitives. Advanced React Hook Form, Zustand, hls.js, TanStack Table/Virtual, TipTap and ECharts dependencies arrive with their consuming phases. Foundation uses native semantic tables where virtualization/sorting is not needed. This follows the rule against unused libraries.
