import { createInstance } from 'i18next';
import { I18nextProvider } from 'react-i18next';
import type { ReactNode } from 'react';
import type { Locale } from '../composition/locale';
import { en } from './locales/en';
import { uk } from './locales/uk';

export function createI18n(locale: Locale = 'uk') {
  const instance = createInstance();
  void instance.init({
    lng: locale,
    fallbackLng: 'uk',
    supportedLngs: ['uk', 'en'],
    resources: { uk, en },
    initAsync: false,
    interpolation: { escapeValue: false },
  });
  return instance;
}

export function I18nProvider({
  instance,
  children,
}: {
  instance: ReturnType<typeof createI18n>;
  children: ReactNode;
}) {
  return <I18nextProvider i18n={instance}>{children}</I18nextProvider>;
}
