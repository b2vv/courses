import { describe, expect, it } from '@rstest/core';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router';
import { render, screen, within } from '@testing-library/react';
import type { ComponentProps } from 'react';
import { asBrandedId } from '../../../contracts/course';
import { TeachLessonList } from './TeachLessonList';

const courseId = '10000000-0000-4000-8000-000000000001';
const modules = [
  {
    id: asBrandedId('20000000-0000-4000-8000-000000000001', 'module'),
    title: 'Основи',
    lessons: [
      {
        id: asBrandedId('30000000-0000-4000-8000-000000000001', 'lesson'),
        title: 'Вступ',
        isPreview: false,
      },
      {
        id: asBrandedId('30000000-0000-4000-8000-000000000002', 'lesson'),
        title: 'Перші кроки',
        isPreview: true,
      },
    ],
  },
  {
    id: asBrandedId('20000000-0000-4000-8000-000000000002', 'module'),
    title: 'Практика',
    lessons: [
      {
        id: asBrandedId('30000000-0000-4000-8000-000000000003', 'lesson'),
        title: 'Вправа',
        isPreview: false,
      },
    ],
  },
  {
    id: asBrandedId('20000000-0000-4000-8000-000000000003', 'module'),
    title: 'Порожній модуль',
    lessons: [],
  },
] satisfies ComponentProps<typeof TeachLessonList>['modules'];

async function renderList(overrides: Partial<ComponentProps<typeof TeachLessonList>> = {}) {
  const props = { courseId, title: 'Навчальний курс', modules, ...overrides };
  const rootRoute = createRootRoute({ component: Outlet });
  const index = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => <TeachLessonList {...props} />,
  });
  const editor = createRoute({
    getParentRoute: () => rootRoute,
    path: '/teach/courses/$courseId/lessons/$lessonId',
    component: () => <p>Editor</p>,
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([index, editor]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  await router.load();
  render(<RouterProvider router={router} />);
}

describe('TeachLessonList', () => {
  it('groups lessons by module and links each lesson to its teach editor', async () => {
    await renderList();
    expect(screen.getByText('Навчальний курс')).toBeTruthy();
    expect(screen.getAllByRole('link')).toHaveLength(3);
    for (const module of modules.filter((item) => item.lessons.length > 0)) {
      const group = within(screen.getByRole('region', { name: module.title }));
      expect(group.getByRole('heading', { level: 3, name: module.title })).toBeTruthy();
      expect(group.getAllByRole('link')).toHaveLength(module.lessons.length);
      for (const lesson of module.lessons) {
        expect(group.getByRole('link', { name: lesson.title }).getAttribute('href')).toBe(
          `/teach/courses/${courseId}/lessons/${lesson.id}`,
        );
      }
    }
    expect(screen.queryByText('Уроків ще немає.')).toBeNull();
  });

  it('does not render modules without lessons as empty groups', async () => {
    await renderList();
    expect(screen.queryByRole('region', { name: 'Порожній модуль' })).toBeNull();
    expect(screen.queryByRole('heading', { name: 'Порожній модуль' })).toBeNull();
    expect(screen.getAllByRole('list')).toHaveLength(2);
  });

  it('T13 renders the empty state when all modules have zero lessons', async () => {
    await renderList({ modules: modules.map((module) => ({ ...module, lessons: [] })) });
    expect(screen.getByRole('heading', { name: 'Уроків ще немає.' })).toBeTruthy();
    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(screen.queryAllByRole('region')).toHaveLength(0);
    expect(screen.queryAllByRole('list')).toHaveLength(0);
  });

  it('T13 renders the empty state when the course has zero modules', async () => {
    await renderList({ modules: [] });
    expect(screen.getByRole('heading', { name: 'Уроків ще немає.' })).toBeTruthy();
    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(screen.queryAllByRole('list')).toHaveLength(0);
  });
});
