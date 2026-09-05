import { afterEach } from '@rstest/core';
import { cleanup } from '@testing-library/react';
import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import { en } from './app/providers/locales/en';
import { uk } from './app/providers/locales/uk';

void i18next.use(initReactI18next).init({
  lng: 'uk',
  fallbackLng: 'uk',
  resources: { uk, en },
  initAsync: false,
  interpolation: { escapeValue: false },
});

afterEach(() => cleanup());
