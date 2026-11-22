import type { i18n as I18n } from 'i18next';
import type { uk } from '../providers/locales/uk';

export type PageTitleKey = keyof (typeof uk)['translation']['pageTitle'];

const brand = 'CourseForge';

/** Document title for a page: "<page> — CourseForge", in the request's locale. */
export function pageTitle(i18n: I18n, key: PageTitleKey): string {
  return namedPageTitle(i18n.t(`pageTitle.${key}`));
}

/** Document title for a page named by data (a course or lesson title). */
export function namedPageTitle(name: string): string {
  return `${name} — ${brand}`;
}

/** `head` result that sets only the document title. */
export function pageTitleHead(i18n: I18n, key: PageTitleKey) {
  return { meta: [{ title: pageTitle(i18n, key) }] };
}
