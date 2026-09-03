/** Generated URL builders share this immutable, validated runtime origin. */
export function apiOrigin(): string {
  if (typeof window !== 'undefined') return window.location.origin;
  const configured = process.env.API_BASE_URL;
  if (!configured) throw new Error('API_BASE_URL is required for SSR');
  const url = new URL(configured);
  if (
    !['http:', 'https:'].includes(url.protocol) ||
    url.username ||
    url.password ||
    url.pathname !== '/' ||
    url.search ||
    url.hash
  )
    throw new Error('API_BASE_URL must be an HTTP origin');
  return url.origin;
}
