// Public surface of the `observability` module: the client-side failure sink.
// The global error bridge in the client entry reaches it only through this
// barrel (enforced by scripts/check-architecture.mjs — no deep cross-module
// import). No domain of its own; the event shape is the generated API model.

// Ports: the fire-and-forget sink contract.
export type { ClientLogReporter } from './ports/client-log';

// Adapters: the implementation that posts to /api/v1/client-logs.
export { createClientLogReporter } from './adapters/client-log-adapter';
