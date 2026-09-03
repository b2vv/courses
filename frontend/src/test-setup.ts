import { afterEach } from '@rstest/core';
import { cleanup } from '@testing-library/react';
import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import { resources } from './app/providers/i18n';

void i18next.use(initReactI18next).init({
  lng: 'uk',
  fallbackLng: 'uk',
  resources,
  initAsync: false,
  interpolation: { escapeValue: false },
});

afterEach(() => cleanup());
