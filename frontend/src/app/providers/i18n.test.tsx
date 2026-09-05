import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nProvider, createI18n } from './i18n';
import { en } from './locales/en';
import { uk } from './locales/uk';
import { LanguageSwitcher } from '../../presentation/LanguageSwitcher';
import { HomeView } from '../../presentation/storefront/HomeView';

test('keeps shared UI and administration keys aligned across locale bundles', () => {
  expect(Object.keys(uk.translation.ui)).toEqual(Object.keys(en.translation.ui));
  expect(Object.keys(uk.translation.administration)).toEqual(
    Object.keys(en.translation.administration),
  );
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
  expect(screen.getByRole('heading').textContent).toContain('Навички, які працюють у production');
  await user.click(screen.getByRole('button', { name: 'EN' }));
  expect(screen.getByRole('heading').textContent).toContain('Skills that work in production');
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
