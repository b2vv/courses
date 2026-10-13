import { useTranslation } from 'react-i18next';
import { Skeleton } from '../ui/states';

// Keep the route fallback independent of the lazy editor and its Tiptap imports.
export function LessonDocumentEditorLoading() {
  const { t } = useTranslation();
  return (
    <section className="lesson-editor" aria-labelledby="lesson-editor-heading">
      <header className="lesson-editor-header">
        <h1 id="lesson-editor-heading">{t('administration.document.heading')}</h1>
      </header>
      <Skeleton label={t('administration.document.loading')} lines={4} />
    </section>
  );
}
