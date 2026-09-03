import { createInstance } from 'i18next';
import { I18nextProvider } from 'react-i18next';
import type { ReactNode } from 'react';

export function createI18n() {
  const instance = createInstance();
  void instance.init({
    lng: 'uk',
    fallbackLng: 'uk',
    initAsync: false,
    resources: { uk: { translation: { catalog: 'Каталог курсів' } } },
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
