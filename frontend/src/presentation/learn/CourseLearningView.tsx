import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import type { LearningCourse } from '../../domain/playback/playback';
import { Link } from '../ui';

type CourseLearningViewProps = {
  readonly course: LearningCourse;
  readonly currentLessonId: string;
  readonly children: ReactNode;
};

export function CourseLearningView({ course, currentLessonId, children }: CourseLearningViewProps) {
  const { t } = useTranslation();
  return (
    <article className="course-learning">
      <div className="course-learning-player">{children}</div>
      <nav className="course-outline" aria-labelledby="course-outline-heading">
        <header className="course-outline-heading">
          <p className="eyebrow">{t('learn.program')}</p>
          <h2 id="course-outline-heading">{t('learn.outline')}</h2>
        </header>
        <ol className="course-outline-modules">
          {course.modules.map((module, moduleIndex) => (
            <li key={module.id} className="course-outline-module">
              <h3>
                <span aria-hidden="true">{String(moduleIndex + 1).padStart(2, '0')}</span>
                {module.title}
              </h3>
              <ol className="course-outline-lessons">
                {module.lessons.map((lesson, lessonIndex) => {
                  const isCurrent = lesson.id === currentLessonId;
                  return (
                    <li key={lesson.id}>
                      <Link
                        href={`/learn/${course.slug}/${lesson.id}`}
                        className="course-outline-lesson"
                        aria-current={isCurrent ? 'page' : undefined}
                        data-current={isCurrent || undefined}
                      >
                        <span className="course-outline-number">
                          {moduleIndex + 1}.{lessonIndex + 1}
                        </span>
                        <span className="course-outline-title">{lesson.title}</span>
                        {lesson.completed && (
                          <span className="course-outline-state" data-state="completed">
                            {t('learn.completed')}
                          </span>
                        )}
                        {isCurrent && (
                          <span className="course-outline-state" data-state="current">
                            {t('learn.current')}
                          </span>
                        )}
                      </Link>
                    </li>
                  );
                })}
              </ol>
            </li>
          ))}
        </ol>
      </nav>
    </article>
  );
}
