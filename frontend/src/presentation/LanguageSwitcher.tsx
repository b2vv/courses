import { useTranslation } from 'react-i18next';
import type { Locale } from '../app/composition/locale';

export function LanguageSwitcher() {
  const { i18n } = useTranslation();
  const selected: Locale = i18n.resolvedLanguage === 'en' ? 'en' : 'uk';
  function select(locale: Locale) {
    document.cookie = `courseforge_locale=${locale}; Path=/; Max-Age=31536000; SameSite=Lax`;
    document.documentElement.lang = locale;
    void i18n.changeLanguage(locale);
  }
  return (
    <fieldset className="language-switcher">
      <legend className="ui-visually-hidden">Language / Мова</legend>
      <button type="button" lang="uk" aria-pressed={selected === 'uk'} onClick={() => select('uk')}>
        UA
      </button>
      <button type="button" lang="en" aria-pressed={selected === 'en'} onClick={() => select('en')}>
        EN
      </button>
    </fieldset>
  );
}
