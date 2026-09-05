import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequestHeader } from '@tanstack/react-start/server';

export type Locale = 'uk' | 'en';

export function normalizeLocale(value: string | null | undefined): Locale | undefined {
  const language = value
    ?.trim()
    .toLowerCase()
    .split(/[-_,;]/)[0];
  return language === 'uk' || language === 'en' ? language : undefined;
}

export function cookieLocale(cookie: string | null | undefined): Locale | undefined {
  return normalizeLocale(
    cookie
      ?.split(';')
      .map((part) => part.trim())
      .find((part) => part.startsWith('courseforge_locale='))
      ?.split('=')[1],
  );
}

export const detectLocale = createIsomorphicFn()
  .server(
    () =>
      cookieLocale(getRequestHeader('cookie')) ??
      normalizeLocale(getRequestHeader('accept-language')) ??
      'uk',
  )
  .client(
    () =>
      cookieLocale(document.cookie) ??
      normalizeLocale(document.documentElement.lang) ??
      normalizeLocale(navigator.language) ??
      'uk',
  );
