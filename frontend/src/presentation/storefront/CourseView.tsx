import type { CourseDetail } from '../../domain/course/course';
import { Link } from '../ui';
import { useTranslation } from 'react-i18next';

export function CourseView({ course }: { course: CourseDetail }) {
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
        <div>
          <section>
            <h2>{t('course.about')}</h2>
            <p className="description">{course.description}</p>
          </section>
          {course.program.length > 0 && (
            <section className="program">
              <h2>{t('course.program')}</h2>
              <ol className="course-program">
                {course.program.map((module, moduleIndex) => (
                  <li key={`${module.title}-${moduleIndex}`}>
                    <h3>{module.title}</h3>
                    {module.lessons.length > 0 && (
                      <ol>
                        {module.lessons.map((lesson, lessonIndex) => (
                          <li key={`${lesson.title}-${lessonIndex}`}>{lesson.title}</li>
                        ))}
                      </ol>
                    )}
                  </li>
                ))}
              </ol>
            </section>
          )}
        </div>
        <aside className="availability">
          <h2>{t('course.comingSoon')}</h2>
          <p>{t('course.unavailable')}</p>
          <Link href="/courses">{t('course.continue')}</Link>
        </aside>
      </div>
    </article>
  );
}
