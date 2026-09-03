# ADR 002 — TanStack Start with Rsbuild
Status: accepted, 2026-09-03.

Use the official `@tanstack/react-start/plugin/rsbuild` with Rsbuild/Rspack, React 19 and Node 24. Build emits client and fetch-style server bundles. A Node host serves assets, health and dynamic rendering with graceful shutdown. Query hydration is per request. Vite and Next are excluded by the brief.

Source: [official hosting guide](https://tanstack.com/start/latest/docs/framework/react/guide/hosting), [official Rsbuild example](https://github.com/rstackjs/rstack-examples/tree/main/rsbuild/tanstack-start). Pin exact installed versions and retain npm lockfile; verify emitted bundle names, do not assume a Vite deployment layout.
