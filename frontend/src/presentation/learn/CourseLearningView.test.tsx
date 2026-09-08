import { expect, test } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import type { LearningCourse } from '../../domain/playback/playback';
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
        },
        {
          id: lessonIds[1],
          title: 'Другий крок',
          videoState: null,
          position: 0,
          completed: false,
          clientSeq: 0,
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
        },
        {
          id: lessonIds[3],
          title: 'Четвертий крок',
          videoState: 'failed',
          position: 0,
          completed: false,
          clientSeq: 0,
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
