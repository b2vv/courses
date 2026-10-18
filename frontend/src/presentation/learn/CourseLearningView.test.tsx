import { expect, test } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import type { LearningCourse } from '../../modules/playback';
import { CourseLearningView } from './CourseLearningView';

const lessonIds = [
  '30000000-0000-4000-8000-000000000001',
  '30000000-0000-4000-8000-000000000002',
  '30000000-0000-4000-8000-000000000003',
  '30000000-0000-4000-8000-000000000004',
] as const;

const course: LearningCourse = {
  id: '10000000-0000-4000-8000-000000000001',
  slug: 'go-basics',
  title: 'Go на практиці',
  summary: 'Курс',
  description: 'Опис',
  status: 'published',
  price: null,
  continueLessonId: lessonIds[1],
  modules: [
    {
      id: '20000000-0000-4000-8000-000000000001',
      title: 'Основи',
      lessons: [
        {
          id: lessonIds[0],
          title: 'Перший крок',
          videoState: 'ready',
          position: 300,
          completed: true,
          clientSeq: 2,
          isPreview: false,
          locked: false,
        },
        {
          id: lessonIds[1],
          title: 'Другий крок',
          videoState: null,
          position: 0,
          completed: false,
          clientSeq: 0,
          isPreview: false,
          locked: false,
        },
      ],
    },
    {
      id: '20000000-0000-4000-8000-000000000002',
      title: 'Практика',
      lessons: [
        {
          id: lessonIds[2],
          title: 'Третій крок',
          videoState: 'processing',
          position: 0,
          completed: false,
          clientSeq: 0,
          isPreview: false,
          locked: false,
        },
        {
          id: lessonIds[3],
          title: 'Четвертий крок',
          videoState: 'failed',
          position: 0,
          completed: false,
          clientSeq: 0,
          isPreview: false,
          locked: false,
        },
      ],
    },
  ],
};

test('renders modules and lessons in curriculum order with completed and current states', () => {
  render(
    <CourseLearningView course={course} currentLessonId={lessonIds[1]}>
      <div>Плеєр уроку</div>
    </CourseLearningView>,
  );

  const outline = screen.getByRole('navigation', { name: 'Зміст курсу' });
  const lessonLinks = within(outline).getAllByRole('link');
  expect(lessonLinks).toHaveLength(4);
  expect(lessonLinks.map((link) => link.getAttribute('href'))).toEqual(
    lessonIds.map((id) => `/learn/go-basics/${id}`),
  );
  expect(lessonLinks.map((link) => link.textContent)).toEqual([
    '1.1Перший крокПереглянуто',
    '1.2Другий крокПоточний урок',
    '2.1Третій крок',
    '2.2Четвертий крок',
  ]);
  expect(lessonLinks[0]?.getAttribute('aria-current')).toBeNull();
  expect(lessonLinks[1]?.getAttribute('aria-current')).toBe('page');
  expect(screen.getByText('Плеєр уроку')).toBeTruthy();
});

test('B19 renders all 120 lesson rows from a renderer-owned boundary fixture', () => {
  const maximumCourse: LearningCourse = {
    ...course,
    modules: Array.from({ length: 40 }, (_, moduleIndex) => ({
      id: `module-${moduleIndex}`,
      title: `Модуль ${moduleIndex + 1}`,
      lessons: Array.from({ length: 3 }, (_, lessonIndex) => ({
        id: `lesson-${moduleIndex}-${lessonIndex}`,
        title: `Урок ${moduleIndex + 1}.${lessonIndex + 1}`,
        videoState: null,
        position: 0,
        completed: false,
        clientSeq: 0,
        isPreview: false,
        locked: false,
      })),
    })),
  };

  render(
    <CourseLearningView course={maximumCourse} currentLessonId="lesson-0-0">
      <div>Плеєр</div>
    </CourseLearningView>,
  );

  expect(
    within(screen.getByRole('navigation', { name: 'Зміст курсу' })).getAllByRole('link'),
  ).toHaveLength(120);
  expect(screen.getByText('Урок 40.3')).toBeTruthy();
});

test('marks locked lessons, and preview lessons only while something is still locked', () => {
  const lessons = course.modules[0]?.lessons ?? [];
  const [first, second] = lessons;
  if (first === undefined || second === undefined) throw new Error('fixture lessons missing');
  const withPreview = (locked: boolean): LearningCourse => ({
    ...course,
    modules: [
      {
        id: 'm',
        title: 'Основи',
        lessons: [
          { ...first, isPreview: true, locked: false },
          { ...second, isPreview: false, locked },
        ],
      },
    ],
  });

  const { unmount } = render(
    <CourseLearningView course={withPreview(true)} currentLessonId={first.id}>
      <p>player</p>
    </CourseLearningView>,
  );
  const outline = screen.getByRole('navigation');
  expect(within(outline).getByText('Після покупки')).toBeTruthy();
  expect(within(outline).getByText('Відкритий урок')).toBeTruthy();
  unmount();

  // An owner, an admin or a buyer has nothing locked, so "open lesson" would be noise.
  render(
    <CourseLearningView course={withPreview(false)} currentLessonId={first.id}>
      <p>player</p>
    </CourseLearningView>,
  );
  expect(screen.queryByText('Відкритий урок')).toBeNull();
  expect(screen.queryByText('Після покупки')).toBeNull();
});

test('shows the admin purchase entry only when the course can be purchased and starts it', async () => {
  const user = (await import('@testing-library/user-event')).default.setup();
  let started = 0;
  const { rerender } = render(
    <CourseLearningView
      course={{ ...course, canPurchase: true }}
      currentLessonId={lessonIds[0]}
      onBuy={() => {
        started += 1;
      }}
    >
      <p>player</p>
    </CourseLearningView>,
  );
  await user.click(screen.getByRole('button', { name: 'Купити курс' }));
  expect(started).toBe(1);
  rerender(
    <CourseLearningView course={{ ...course, canPurchase: false }} currentLessonId={lessonIds[0]}>
      <p>player</p>
    </CourseLearningView>,
  );
  expect(screen.queryByRole('button', { name: 'Купити курс' })).toBeNull();
});

test('hides the outline purchase entry when a lesson is locked and shows progress while starting', () => {
  // A student with a locked lesson already has the buy path on that lesson; a second one would be a duplicate.
  const lockedCourse = {
    ...course,
    canPurchase: true,
    modules: course.modules.map((module, moduleIndex) => ({
      ...module,
      lessons: module.lessons.map((lesson, lessonIndex) => ({
        ...lesson,
        locked: moduleIndex === 0 && lessonIndex === 0,
      })),
    })),
  };
  const { rerender } = render(
    <CourseLearningView course={lockedCourse} currentLessonId={lessonIds[0]} onBuy={() => {}}>
      <p>player</p>
    </CourseLearningView>,
  );
  expect(screen.queryByRole('button', { name: 'Купити курс' })).toBeNull();
  rerender(
    <CourseLearningView
      course={{ ...course, canPurchase: true }}
      currentLessonId={lessonIds[0]}
      onBuy={() => {}}
      buying
    >
      <p>player</p>
    </CourseLearningView>,
  );
  expect(screen.getByRole('button', { name: 'Створюємо замовлення…' })).toBeTruthy();
});
