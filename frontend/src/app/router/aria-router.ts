/**
 * How a React Aria link press moves the app (design-pass C3b, plan §17 R1–R3).
 *
 * React Aria hands a press to the router only for a same-origin `_self` link without a
 * modifier key (`react-aria` `shouldClientNavigate`); everything else stays the browser's.
 */

/** The options a React Aria link may pass to the router (`routerOptions`). */
export type AriaRouterOptions = {
  /** A full document load instead of a move inside the app: retry after a failure (R3). */
  readonly reloadDocument?: boolean;
  readonly replace?: boolean;
};

/**
 * The admin and instructor workspaces are a separate app with their own shell; everything
 * else is the client. Matches the root layout's surface split (`__root.tsx`).
 */
export type Surface = 'workspace' | 'client';

export function surfaceOf(pathname: string): Surface {
  return /^\/(admin|teach)(\/|$)/.test(pathname) ? 'workspace' : 'client';
}

/**
 * The router call for a press on `href` from the page at `current` (a full URL).
 *
 * Crossing between the workspace and the client is a full document load, as before C3b (R2):
 * admin and instructor writes invalidate only their own queries (`admin-queries.ts`), and the
 * public catalog and course data live in the query cache for a minute, so a move without a
 * load would show a course as it was before the edit.
 */
export function ariaNavigation(
  current: string,
  href: string,
  options: AriaRouterOptions | undefined,
): { href: string; reloadDocument?: boolean; replace?: boolean } {
  const from = new URL(current);
  const to = new URL(href, from);
  const crossesSurface = surfaceOf(from.pathname) !== surfaceOf(to.pathname);
  const reloadDocument = options?.reloadDocument === true || crossesSurface;
  return {
    href: `${to.pathname}${to.search}${to.hash}`,
    ...(reloadDocument ? { reloadDocument: true } : {}),
    ...(options?.replace === undefined ? {} : { replace: options.replace }),
  };
}

declare module 'react-aria-components' {
  interface RouterConfig {
    routerOptions: AriaRouterOptions;
  }
}
