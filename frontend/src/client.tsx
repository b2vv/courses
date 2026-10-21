import './app/zod-jitless';
import { StrictMode, startTransition } from 'react';
import { hydrateRoot } from 'react-dom/client';
import { StartClient } from '@tanstack/react-start/client';
import { createClientLogReporter } from './modules/observability';

// The client entry is the one truly client-only point that runs once per page
// load, so the global error bridge is installed here, not in a request-scoped
// composition factory (those also run under SSR). Channels are cleanly
// separated with no dedup: `onCaughtError` owns render errors an error boundary
// caught; `window.onerror` + `unhandledrejection` own everything else. React 19
// routes boundary-caught errors to onCaughtError (not window.onerror) and
// rethrows uncaught render errors to window.onerror, so the two channels do not
// overlap — one failure yields exactly one event without any dedup.
const reporter = createClientLogReporter();

window.addEventListener('error', (event) => {
  reporter.capture(event.error ?? event.message);
});
window.addEventListener('unhandledrejection', (event) => {
  reporter.capture(event.reason);
});

startTransition(() => {
  hydrateRoot(
    document,
    <StrictMode>
      <StartClient />
    </StrictMode>,
    {
      onCaughtError: (error) => {
        reporter.capture(error);
      },
    },
  );
});
