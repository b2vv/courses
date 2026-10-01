import { describe, expect, test } from '@rstest/core';
import {
  asBrandedId,
  makeAdminCourseQuery,
  makeCatalogQuery,
  makePublishedCourse,
  parseUahAmount,
  validCourseTree,
  validCourseTreeEdit,
  validSlugChange,
  validTransition,
} from './course';
import type { CourseTreeEdit } from './course';

describe('public catalog rules', () => {
  test('normalizes search and rejects unbounded pagination', () => {
    expect(makeCatalogQuery({ q: '  Go  ' })).toEqual({
      ok: true,
      value: { q: 'Go', sort: 'recent', limit: 20, offset: 0 },
    });
    expect(makeCatalogQuery({ limit: 51 }).ok).toBe(false);
    expect(makeCatalogQuery({ offset: -1 }).ok).toBe(false);
    expect(makeCatalogQuery({ q: 'я'.repeat(101) }).ok).toBe(false);
    expect(makeCatalogQuery({ q: '\u0000' }).ok).toBe(false);
    expect(makeCatalogQuery({ q: 'я'.repeat(100), offset: 10000, limit: 50 }).ok).toBe(true);
  });
  test('never turns an unpublished record into a published course', () => {
    const input = {
      id: 'id',
      slug: 'go-basics',
      title: 'Go',
      summary: 'Практика',
      description: 'Опис',
      status: 'draft',
      publishedAt: '2026-09-01T00:00:00Z',
      price: null,
      paid: false,
    };
    expect(makePublishedCourse(input).ok).toBe(false);
    expect(makePublishedCourse({ ...input, status: 'published', slug: '../secret' }).ok).toBe(
      false,
    );
    expect(makePublishedCourse({ ...input, status: 'published' }).ok).toBe(true);
    expect(
      makePublishedCourse({
        ...input,
        status: 'published',
        price: { minor: 149900, currency: 'UAH' },
        paid: true,
      }),
    ).toMatchObject({ ok: true, value: { price: { minor: 149900, currency: 'UAH' }, paid: true } });
    expect(
      makePublishedCourse({
        ...input,
        status: 'published',
        price: { minor: 149900, currency: 'UAH' },
        paid: false,
      }).ok,
    ).toBe(false);
  });
});

describe('admin authoring rules', () => {
  const moduleInput = {
    id: asBrandedId('50000000-0000-4000-8000-000000000001', 'module'),
    title: 'Модуль',
    lessons: [
      {
        id: asBrandedId('60000000-0000-4000-8000-000000000001', 'lesson'),
        title: 'Урок',
        description: '',
        isPreview: false,
      },
    ],
  };
  const edit: CourseTreeEdit = {
    slug: 'go-basics',
    title: 'Go',
    summary: 'Практика',
    description: 'Опис',
    instructorId: '40000000-0000-4000-8000-000000000001',
    modules: [moduleInput],
  };

  test('bounds and de-duplicates a full tree edit', () => {
    expect(validCourseTreeEdit(edit)).toBe(true);
    expect(validCourseTreeEdit({ ...edit, modules: Array(41).fill(moduleInput) })).toBe(false);
    expect(
      validCourseTreeEdit({
        ...edit,
        modules: [moduleInput, { ...moduleInput, title: 'Duplicate module identifier' }],
      }),
    ).toBe(false);
    expect(
      validCourseTreeEdit({
        ...edit,
        modules: [{ ...moduleInput, lessons: Array(121).fill(moduleInput.lessons[0]) }],
      }),
    ).toBe(false);
  });

  test('keeps transition and permanent-slug rules explicit', () => {
    expect(validTransition('draft', 'published')).toBe(true);
    expect(validTransition('published', 'draft')).toBe(false);
    expect(validSlugChange('old', null, 'new')).toBe(true);
    expect(validSlugChange('old', '2026-09-01T00:00:00.000000Z', 'new')).toBe(false);
  });

  test('refuses impossible lifecycle timestamps in a stored tree', () => {
    const stored = {
      ...edit,
      id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
      status: 'draft',
      createdAt: '2026-09-01T00:00:00.000000Z',
      updatedAt: '2026-09-01T00:00:00.000001Z',
      publishedAt: null,
      price: null,
    } as const;
    expect(validCourseTree(stored)).toBe(true);
    expect(validCourseTree({ ...stored, status: 'published' })).toBe(false);
    expect(validCourseTree({ ...stored, updatedAt: '2026-08-31T23:59:59.999999Z' })).toBe(false);
    expect(
      validCourseTree({
        ...stored,
        createdAt: '2026-09-01T00:00:00.000002Z',
        updatedAt: '2026-09-01T00:00:00.000001Z',
      }),
    ).toBe(false);
  });

  test('normalizes bounded admin pagination without inventing a status', () => {
    expect(makeAdminCourseQuery({})).toEqual({
      ok: true,
      value: { limit: 20, offset: 0 },
    });
    expect(makeAdminCourseQuery({ status: 'archived', limit: 50, offset: 10000 }).ok).toBe(true);
    expect(makeAdminCourseQuery({ limit: 51 }).ok).toBe(false);
  });
});

describe('course price input', () => {
  test('reads hryvnias as minor units and refuses anything it would have to round', () => {
    expect(parseUahAmount('1890')).toEqual({ minor: 189000, currency: 'UAH' });
    expect(parseUahAmount('1 890,5')).toEqual({ minor: 189050, currency: 'UAH' });
    expect(parseUahAmount('0.01')).toEqual({ minor: 1, currency: 'UAH' });
    for (const refused of ['', '0', '0.00', '-5', '1.005', '12a', '1e3', '99999999999999'])
      expect(parseUahAmount(refused)).toBeNull();
  });
});
