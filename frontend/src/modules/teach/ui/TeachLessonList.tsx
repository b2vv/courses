import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { CourseTree } from '../../../contracts/course';
import { EmptyState } from '../../../presentation/ui/states';

type TeachLessonListProps = {
  readonly courseId: string;
  readonly title: CourseTree['title'];
  readonly modules: CourseTree['modules'];
};

export function TeachLessonList({ courseId, title, modules }: TeachLessonListProps) {
  const { t } = useTranslation();
  if (modules.every((module) => module.lessons.length === 0))
    return <EmptyState title={t('teach.lessons.empty')} />;
  return (
    <>
      <p>{title}</p>
      {modules
        .filter((module) => module.lessons.length > 0)
        .map((module) => (
          <section key={module.id} aria-labelledby={`teach-module-${module.id}`}>
            <h3 id={`teach-module-${module.id}`}>{module.title}</h3>
            <ul>
              {module.lessons.map((lesson) => (
                <li key={lesson.id}>
                  <Link
                    className="ui-button ui-button--secondary"
                    to="/teach/courses/$courseId/lessons/$lessonId"
                    params={{ courseId, lessonId: lesson.id }}
                  >
                    {lesson.title}
                  </Link>
                </li>
              ))}
            </ul>
          </section>
        ))}
    </>
  );
}
