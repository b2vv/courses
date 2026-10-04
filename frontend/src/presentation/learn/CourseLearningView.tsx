import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import type { LearningCourse } from '../../domain/playback/playback';
import type { PurchaseErrorCode } from '../../domain/purchase/purchase';
import { Button, Link } from '../ui';
import { purchaseErrorKey } from './LockedLessonPanel';

type CourseLearningViewProps = {
  readonly course: LearningCourse;
  readonly currentLessonId: string;
  readonly children: ReactNode;
  readonly onBuy?: (() => void) | undefined;
  readonly buying?: boolean | undefined;
  readonly buyError?: PurchaseErrorCode | undefined;
};

export function CourseLearningView({
  course,
  currentLessonId,
  children,
  onBuy,
  buying = false,
  buyError,
}: CourseLearningViewProps) {
  const { t } = useTranslation();
  // The preview badge only means something to a viewer who still has locked lessons.
  const hasLocked = course.modules.some((module) => module.lessons.some((lesson) => lesson.locked));
  return (
    <article className="course-learning">
      <div className="course-learning-player">{children}</div>
      <nav className="course-outline" aria-labelledby="course-outline-heading">
        <header className="course-outline-heading">
          <p className="eyebrow">{t('learn.program')}</p>
          <h2 id="course-outline-heading">{t('learn.outline')}</h2>
          {/* Only for a viewer with no locked lesson (an admin, G2): a student already gets the buy
              path on the locked lesson, and two would be a duplicate. */}
          {course.canPurchase && !hasLocked && onBuy ? (
            <>
              <Button onPress={onBuy} isPending={buying}>
                {buying ? t('learn.purchase.starting') : t('learn.purchase.buy')}
              </Button>
              {buyError === undefined ? null : (
                <p role="alert">{t(`learn.purchase.errors.${purchaseErrorKey(buyError)}`)}</p>
              )}
            </>
          ) : null}
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
                        {lesson.locked && (
                          <span className="course-outline-state" data-state="locked">
                            {t('learn.lockedLesson')}
                          </span>
                        )}
                        {hasLocked && lesson.isPreview && (
                          <span className="course-outline-state" data-state="preview">
                            {t('learn.preview')}
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
