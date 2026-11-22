import { useRouter } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { Locale } from '../app/composition/locale';

export function LanguageSwitcher() {
  const { i18n } = useTranslation();
  const router = useRouter();
  const selected: Locale = i18n.resolvedLanguage === 'en' ? 'en' : 'uk';
  function select(locale: Locale) {
    document.cookie = `courseforge_locale=${locale}; Path=/; Max-Age=31536000; SameSite=Lax`;
    document.documentElement.lang = locale;
    // Route titles are computed in `head`, which runs on load; re-run it in the new language.
    void i18n.changeLanguage(locale).then(() => router.invalidate());
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
