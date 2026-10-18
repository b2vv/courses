import { useTranslation } from 'react-i18next';

/**
 * What learning here gives, shared by the home page and the catalog. Every
 * claim is a feature that exists and is measured (practice-first lessons,
 * searchable transcripts, lesson discussions) — the mockup's "mentor code
 * review" and "lifetime access" are not offered, so they are not claimed.
 */
export function ValueStrip() {
  const { t } = useTranslation();
  return (
    <section id="values" className="value-strip" aria-labelledby="values-heading">
      <h2 id="values-heading" className="ui-visually-hidden">
        {t('home.valuesTitle')}
      </h2>
      {(['practice', 'review', 'access'] as const).map((value) => (
        <article key={value} className="value-item">
          {/* The glyph is painted by CSS from data-glyph: as text in the DOM, axe
              reports symbol characters as unverifiable contrast ("nonBmp"). The
              tile is decorative and aria-hidden either way. */}
          <span
            className="ui-icon-tile"
            aria-hidden="true"
            data-glyph={value === 'practice' ? '⌘' : value === 'review' ? '✓' : '∞'}
          />
          <div>
            <h3>{t(`home.values.${value}.title`)}</h3>
            <p>{t(`home.values.${value}.description`)}</p>
          </div>
        </article>
      ))}
    </section>
  );
}
