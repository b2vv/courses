import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nProvider, createI18n } from './i18n';
import { en } from './locales/en';
import { uk } from './locales/uk';
import { LanguageSwitcher } from '../../presentation/LanguageSwitcher';
import { HomeView } from '../../presentation/storefront/HomeView';

const pluralSuffix = /_(?:zero|one|two|few|many|other)$/;

const withoutPluralForms = (keys: string[]) => [
  ...new Set(keys.map((key) => key.replace(pluralSuffix, ''))),
];

function flattenLocale(value: unknown, prefix = ''): Map<string, string> {
  if (typeof value === 'string') return new Map([[prefix, value]]);
  if (typeof value !== 'object' || value === null || Array.isArray(value))
    throw new Error(`Locale value at ${prefix || '<root>'} must be an object or string`);
  const flattened = new Map<string, string>();
  for (const [key, child] of Object.entries(value)) {
    const childPrefix = prefix === '' ? key : `${prefix}.${key}`;
    for (const [path, text] of flattenLocale(child, childPrefix)) flattened.set(path, text);
  }
  return flattened;
}

function baseKeys(locale: Map<string, string>) {
  return [...new Set([...locale.keys()].map((key) => key.replace(pluralSuffix, '')))].sort();
}

function pluralBases(...locales: Array<Map<string, string>>) {
  return [
    ...new Set(
      locales.flatMap((locale) =>
        [...locale.keys()]
          .filter((key) => pluralSuffix.test(key))
          .map((key) => key.replace(pluralSuffix, '')),
      ),
    ),
  ];
}

test('B26 keeps complete non-empty locale trees aligned with language-specific plurals', () => {
  const locales = {
    en: flattenLocale(en.translation),
    uk: flattenLocale(uk.translation),
  };

  expect(baseKeys(locales.uk)).toEqual(baseKeys(locales.en));
  for (const [language, locale] of Object.entries(locales)) {
    expect([...locale.entries()].filter(([, value]) => value.trim() === '')).toEqual([]);
    const categories = new Intl.PluralRules(language).resolvedOptions().pluralCategories;
    for (const base of pluralBases(locales.en, locales.uk)) {
      expect(
        categories.filter((category) => !locale.has(`${base}_${category}`)),
        `${language} plural forms for ${base}`,
      ).toEqual([]);
    }
  }
});

test('keeps shared UI and administration keys aligned across locale bundles', () => {
  expect(withoutPluralForms(Object.keys(uk.translation.ui))).toEqual(
    withoutPluralForms(Object.keys(en.translation.ui)),
  );
  expect(withoutPluralForms(Object.keys(uk.translation.administration))).toEqual(
    withoutPluralForms(Object.keys(en.translation.administration)),
  );
  expect(withoutPluralForms(Object.keys(uk.translation.adminUsers))).toEqual(
    withoutPluralForms(Object.keys(en.translation.adminUsers)),
  );
  expect(Object.keys(uk.translation.adminUsers.error)).toEqual(
    Object.keys(en.translation.adminUsers.error),
  );
  const adminListKeys = [
    'title',
    'empty',
    'emptyHint',
    'courses',
    'courseCount',
    'courseTable',
    'courseTitle',
    'slug',
    'status',
    'owner',
    'updated',
    'statusTabs',
    'allStatuses',
    'draft',
    'published',
    'archived',
    'pages',
    'loading',
    'failure',
    'failureHint',
  ] as const;
  for (const key of adminListKeys) {
    expect(uk.translation.administration[key]).toBeTruthy();
    expect(en.translation.administration[key]).toBeTruthy();
  }
  const adminUserKeys = [
    'title',
    'empty',
    'emptyHint',
    'userCount',
    'userTable',
    'email',
    'name',
    'role',
    'status',
    'actions',
    'statusTabs',
    'allStatuses',
    'active',
    'blocked',
    'deleted',
    'pages',
    'loading',
    'failure',
    'failureHint',
    'degraded',
    'degradedHint',
  ] as const;
  for (const key of adminUserKeys) {
    expect(uk.translation.adminUsers[key]).toBeTruthy();
    expect(en.translation.adminUsers[key]).toBeTruthy();
  }
});

test('renders one request locale and changes visible copy without a reload', async () => {
  const i18n = createI18n('uk');
  const user = userEvent.setup();
  render(
    <I18nProvider instance={i18n}>
      <LanguageSwitcher />
      <HomeView />
    </I18nProvider>,
  );
  expect(screen.getByRole('heading', { level: 1 }).textContent).toContain('Навички, які працюють у production');
  await user.click(screen.getByRole('button', { name: 'EN' }));
  expect(screen.getByRole('heading', { level: 1 }).textContent).toContain('Skills that work in production');
  expect(document.documentElement.lang).toBe('en');
  expect(document.cookie).toContain('courseforge_locale=en');
});

test('creates independent instances for concurrent SSR requests', () => {
  const uk = createI18n('uk');
  const en = createI18n('en');
  expect(uk.t('catalog.title')).toBe('Каталог курсів');
  expect(en.t('catalog.title')).toBe('Course catalog');
  expect(uk.language).toBe('uk');
});
