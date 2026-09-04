/**
 * SSR has no browser cookie jar: the request's own cookie is forwarded
 * explicitly, per request, and never held by a shared client. A request without
 * one must carry no Cookie header at all rather than an empty one.
 */
export function apiRequestHeaders(cookie: string | undefined): Headers {
  const headers = new Headers({ Accept: 'application/json' });
  if (cookie) headers.set('Cookie', cookie);
  return headers;
}
