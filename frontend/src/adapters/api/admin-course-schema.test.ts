import { describe, expect, test } from '@rstest/core';
import { adminCoursePageSchema, adminCourseTreeSchema } from './admin-course-schema';

const tree = {
  id: '10000000-0000-4000-8000-000000000001',
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
  modules: [
    {
      id: '50000000-0000-4000-8000-000000000001',
      title: 'Модуль',
      lessons: [
        {
          id: '60000000-0000-4000-8000-000000000001',
          title: 'Урок',
        },
      ],
    },
  ],
};

describe('admin response schemas', () => {
  test('brands valid identifiers while parsing a strict tree', () => {
    const parsed = adminCourseTreeSchema.safeParse(tree);
    expect(parsed.success).toBe(true);
    if (parsed.success) expect(parsed.data.modules[0]?.lessons[0]?.title).toBe('Урок');
  });

  test('accepts the payment fields the API returns: a price and preview lessons', () => {
    const paid = {
      ...tree,
      price: { minor: 49900, currency: 'UAH' },
      modules: [
        {
          id: '50000000-0000-4000-8000-000000000001',
          title: 'Модуль',
          lessons: [
            {
              id: '60000000-0000-4000-8000-000000000001',
              title: 'Урок',
              isPreview: true,
            },
          ],
        },
      ],
    };
    expect(adminCourseTreeSchema.safeParse(paid).success).toBe(true);
    // A free course omits nothing: price is a required, nullable member.
    const { price: _price, ...withoutPrice } = tree;
    expect(adminCourseTreeSchema.safeParse(withoutPrice).success).toBe(false);
    expect(
      adminCourseTreeSchema.safeParse({ ...tree, price: { minor: 0, currency: 'UAH' } }).success,
    ).toBe(false);
    expect(
      adminCourseTreeSchema.safeParse({ ...tree, price: { minor: 100, currency: 'USD' } }).success,
    ).toBe(false);
  });

  test('refuses malformed timestamps, unknown members and duplicate node ids', () => {
    expect(
      adminCourseTreeSchema.safeParse({ ...tree, updatedAt: '2026-09-01T00:00:00Z' }).success,
    ).toBe(false);
    expect(adminCourseTreeSchema.safeParse({ ...tree, position: 0 }).success).toBe(false);
    expect(adminCourseTreeSchema.safeParse({ ...tree, status: 'published' }).success).toBe(false);
    expect(
      adminCourseTreeSchema.safeParse({
        ...tree,
        modules: [tree.modules[0], { ...tree.modules[0], title: 'Duplicate module identifier' }],
      }).success,
    ).toBe(false);
  });

  test('refuses a list page that exceeds the contract shape', () => {
    const item = {
      id: tree.id,
      slug: tree.slug,
      title: tree.title,
      status: tree.status,
      instructorId: tree.instructorId,
      updatedAt: tree.updatedAt,
    };
    expect(
      adminCoursePageSchema.safeParse({
        items: [item],
        pagination: { limit: 20, offset: 0, total: 1 },
      }).success,
    ).toBe(true);
    expect(
      adminCoursePageSchema.safeParse({
        items: [{ ...item, instructorId: 'not-a-uuid' }],
        pagination: { limit: 20, offset: 0, total: 1 },
      }).success,
    ).toBe(false);
  });
});
