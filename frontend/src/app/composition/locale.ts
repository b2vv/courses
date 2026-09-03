import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequestHeader } from '@tanstack/react-start/server';
import { cookieLocale, normalizeLocale } from '../providers/i18n';

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
