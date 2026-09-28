import { randomUUID } from 'node:crypto';
import { expect, test } from '@playwright/test';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

const runNamespace = `e2e-learn-access-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
let course: { id: string; updatedAt: string } | undefined;

function token(payload: object) {
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string') {
    throw new Error('Admin response must carry id and updatedAt');
  }
  return { id, updatedAt };
}

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

test.use({ storageState: authStatePath('student') });

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined || course === undefined) return;
  const context = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    const response = await context.request.delete(`/api/v1/admin/courses/${course.id}`, {
      data: { expectedUpdatedAt: course.updatedAt },
      headers: origin(baseURL),
    });
    expect(response.status(), await response.text()).toBe(204);
  } finally {
    await context.close();
  }
});

test('R11 names access denial for a student visiting another author’s draft', async ({
  page,
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const adminContext = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    const slug = `${runNamespace}-draft`;
    const draft = {
      slug,
      title: `Недоступний курс ${runNamespace}`,
      summary: 'A run-scoped draft for access denial.',
      description: 'A draft owned by another account.',
      instructorId: adminID,
    };
    const created = await adminContext.request.post('/api/v1/admin/courses', {
      data: draft,
      headers: origin(baseURL),
    });
    expect(created.status(), await created.text()).toBe(201);
    const createdCourse = token((await created.json()) as object);
    const moduleID = randomUUID();
    const lessonID = randomUUID();
    const saved = await adminContext.request.put(`/api/v1/admin/courses/${createdCourse.id}`, {
      data: {
        expectedUpdatedAt: createdCourse.updatedAt,
        course: draft,
        modules: [
          {
            id: moduleID,
            title: `Модуль ${runNamespace}`,
            lessons: [{ id: lessonID, title: 'Урок доступу', description: '' }],
          },
        ],
      },
      headers: origin(baseURL),
    });
    expect(saved.status(), await saved.text()).toBe(200);
    course = token((await saved.json()) as object);

    await page.goto(`/learn/${slug}`);
    await expect(
      page.getByRole('heading', { name: /Немає доступу до курсу|Course access unavailable/ }),
    ).toBeVisible();
    await expect(page.getByText(/Цей курс недоступний|This course is not available/)).toBeVisible();
    await expect(
      page.getByRole('link', { name: /Перейти до каталогу|Go to course catalog/ }),
    ).toBeVisible();
    await expect(page.getByRole('link', { name: /Спробувати ще раз|Try again/ })).toHaveCount(0);
    await expect(
      page.getByRole('heading', {
        name: /Не вдалося перевірити сеанс|Could not check your session/,
      }),
    ).toHaveCount(0);
  } finally {
    await adminContext.close();
  }
});
