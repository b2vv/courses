import { describe, expect, test } from '@rstest/core';
import { asBrandedId } from '../../domain/course/course';
import type { CourseTree, CourseTreeEdit } from '../../domain/course/course';
import type { AdminCourseRepository } from '../../ports/AdminCourseRepository';
import { createAuthoring } from './authoring';

const tree: CourseTree = {
  id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
  slug: 'go-basics',
  title: 'Go',
  summary: 'Практика',
  description: 'Опис',
  status: 'draft',
  instructorId: '40000000-0000-4000-8000-000000000001',
  createdAt: '2026-09-01T00:00:00.000000Z',
  updatedAt: '2026-09-01T00:00:00.000001Z',
  publishedAt: null,
  price: null,
  studentCount: 0,
  modules: [],
};
const edit: CourseTreeEdit = {
  slug: tree.slug,
  title: tree.title,
  summary: tree.summary,
  description: tree.description,
  instructorId: tree.instructorId,
  modules: tree.modules,
};

function repository(overrides: Partial<AdminCourseRepository> = {}): AdminCourseRepository {
  const base: AdminCourseRepository = {
    list: async (query) => ({
      ok: true,
      value: {
        items: [],
        pagination: { limit: query.limit, offset: query.offset, total: 0 },
        statusCounts: { total: 0, draft: 0, published: 0, archived: 0 },
      },
    }),
    load: async () => ({ ok: true, value: tree }),
    listInstructors: async () => ({ ok: true, value: [] }),
    create: async () => ({ ok: true, value: tree }),
    save: async () => ({ ok: true, value: tree }),
    transition: async () => ({ ok: true, value: tree }),
    delete: async () => ({ ok: true, value: null }),
  };
  return { ...base, ...overrides };
}

describe('admin authoring use cases', () => {
  test('validates writes before opening the repository seam', async () => {
    let calls = 0;
    const authoring = createAuthoring(
      repository({
        create: async () => {
          calls += 1;
          return { ok: true, value: tree };
        },
      }),
    );
    const result = await authoring.create({
      slug: tree.slug,
      title: ' ',
      summary: tree.summary,
      description: tree.description,
    });
    expect(result).toEqual({ ok: false, error: { code: 'invalid_request' } });
    expect(calls).toBe(0);
  });

  test('passes the loaded concurrency token with a valid full-tree save', async () => {
    let expectedUpdatedAt = '';
    const authoring = createAuthoring(
      repository({
        save: async (_id, expected) => {
          expectedUpdatedAt = expected;
          return { ok: true, value: tree };
        },
      }),
    );
    expect(await authoring.save(tree, edit)).toEqual({ ok: true, value: tree });
    expect(expectedUpdatedAt).toBe(tree.updatedAt);
  });

  test('refuses illegal lifecycle and permanent-slug changes locally', async () => {
    const published = {
      ...tree,
      status: 'published',
      publishedAt: '2026-09-01T00:00:00.000001Z',
      price: null,
    } satisfies CourseTree;
    const authoring = createAuthoring(repository());
    expect(await authoring.transition(published, 'draft')).toEqual({
      ok: false,
      error: { code: 'invalid_status_transition' },
    });
    expect(await authoring.save(published, { ...edit, slug: 'new-slug' })).toEqual({
      ok: false,
      error: { code: 'slug_immutable' },
    });
    expect(await authoring.delete(published)).toEqual({
      ok: false,
      error: { code: 'course_not_deletable' },
    });
  });

  test('deletes a draft list row by id and token, and refuses a non-draft one', async () => {
    const calls: Array<{ id: string; updatedAt: string }> = [];
    const authoring = createAuthoring(
      repository({
        delete: async (id, expectedUpdatedAt) => {
          calls.push({ id, updatedAt: expectedUpdatedAt });
          return { ok: true, value: null };
        },
      }),
    );
    const draftRow = {
      id: tree.id,
      slug: tree.slug,
      title: tree.title,
      status: 'draft',
      instructorId: tree.instructorId,
      updatedAt: tree.updatedAt,
      lessonCount: 0,
      studentCount: 0,
      price: null,
    } as const;
    expect(await authoring.deleteListed(draftRow)).toEqual({ ok: true, value: null });
    expect(calls).toEqual([{ id: tree.id, updatedAt: tree.updatedAt }]);
    // A non-draft row is refused locally — the seam is never opened.
    expect(await authoring.deleteListed({ ...draftRow, status: 'published' })).toEqual({
      ok: false,
      error: { code: 'course_not_deletable' },
    });
    expect(calls).toHaveLength(1);
  });
});

describe('course price', () => {
  test('saves only the price, over the stored tree, and refuses a malformed amount unsent', async () => {
    const sent: CourseTreeEdit[] = [];
    const authoring = createAuthoring(
      repository({
        save: async (_id, _expected, edit) => {
          sent.push(edit);
          return { ok: true, value: tree };
        },
      }),
    );

    expect((await authoring.setPrice(tree, { minor: 189000, currency: 'UAH' })).ok).toBe(true);
    expect(await authoring.setPrice(tree, { minor: 0, currency: 'UAH' })).toEqual({
      ok: false,
      error: { code: 'invalid_request' },
    });
    expect(
      (await authoring.setPrice({ ...tree, price: { minor: 5, currency: 'UAH' } }, null)).ok,
    ).toBe(true);

    // The stored tree, unchanged, plus the price — a first price is refused with anything else.
    expect(sent).toEqual([
      { ...edit, price: { minor: 189000, currency: 'UAH' } },
      { ...edit, price: null },
    ]);
  });

  test('the course form never carries a price', async () => {
    const sent: CourseTreeEdit[] = [];
    const authoring = createAuthoring(
      repository({
        save: async (_id, _expected, saved) => {
          sent.push(saved);
          return { ok: true, value: tree };
        },
      }),
    );
    await authoring.save(tree, edit);
    expect('price' in (sent[0] ?? {})).toBe(false);
  });
});
