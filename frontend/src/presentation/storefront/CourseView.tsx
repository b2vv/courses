import type { PublishedCourse } from '../../domain/course/course';
import { Link } from '../ui';
import { useTranslation } from 'react-i18next';

export function CourseView({ course }: { course: PublishedCourse }) {
  const { t } = useTranslation();
  return (
    <article className="course-detail">
      <div className="detail-heading">
        <Link href="/courses">{t('course.all')}</Link>
        <p className="eyebrow">{t('catalog.kind')}</p>
        <h1>{course.title}</h1>
        <p className="lede">{course.summary}</p>
      </div>
      <div className="detail-body">
        <section>
          <h2>{t('course.about')}</h2>
          <p className="description">{course.description}</p>
        </section>
        <aside className="availability">
          <h2>{t('course.comingSoon')}</h2>
          <p>{t('course.unavailable')}</p>
          <Link href="/courses">{t('course.continue')}</Link>
        </aside>
      </div>
    </article>
  );
}
