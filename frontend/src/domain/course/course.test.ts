import { describe, expect, test } from '@rstest/core';
import { makeCatalogQuery, makePublishedCourse } from './course';

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
    };
    expect(makePublishedCourse(input).ok).toBe(false);
    expect(makePublishedCourse({ ...input, status: 'published', slug: '../secret' }).ok).toBe(
      false,
    );
    expect(makePublishedCourse({ ...input, status: 'published' }).ok).toBe(true);
  });
});
