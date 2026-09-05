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
});
