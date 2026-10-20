// Type surface for the runtime SSR log builders in server-log.mjs (plain JS,
// copied verbatim beside server.mjs). Type-only; not shipped at runtime.
export function ssrRequestLine(line: { status: number; durationMs: number; route: string }): string;
export function ssrRenderFailedLine(line: { route: string }): string;
