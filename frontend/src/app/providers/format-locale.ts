/**
 * Maps the active i18n language to the BCP-47 locale used by Intl formatters.
 * The app ships English and Ukrainian; anything that is not English formats as
 * Ukrainian. Kept in one place so the number/date/currency rule stays consistent
 * across admin views.
 */
export function formatLocale(resolvedLanguage: string | undefined): string {
  return resolvedLanguage === 'en' ? 'en' : 'uk-UA';
}
