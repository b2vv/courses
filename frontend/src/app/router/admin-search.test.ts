import { describe, expect, it } from '@rstest/core';
import { adminCourseQuery, adminListState, validateAdminSearch } from './admin-search';
import { parseUrlSearch } from './url-search';

describe('admin list URL search contract', () => {
  it('emits no defaults for a plain /admin address', () => {
    expect(validateAdminSearch({})).toEqual({});
    expect(adminListState({})).toEqual({ page: 1 });
    expect(adminCourseQuery(adminListState({}))).toEqual({ limit: 20, offset: 0 });
  });

  it('echoes known parameters verbatim and drops unknown parameters', () => {
    expect(validateAdminSearch({ status: 'draft', page: '3', x: '1' })).toEqual({
      status: 'draft',
      page: '3',
    });
  });

  it.each([
    ['?status=published&page=2', { status: 'published', page: 2 }],
    ['?status=nonsense&page=3', { page: 3 }],
    ['?status=archived&page=abc', { status: 'archived', page: 1 }],
    ['?status=draft&page=0', { status: 'draft', page: 1 }],
    ['?status=draft&page=502', { status: 'draft', page: 1 }],
    ['?status=draft&status=published&page=2&page=3', { page: 1 }],
  ])('parses each field independently for %s', (url, expected) => {
    expect(adminListState(validateAdminSearch(parseUrlSearch(url)))).toEqual(expected);
  });

  it('never throws for malformed or non-object input', () => {
    expect(adminListState(validateAdminSearch(null))).toEqual({ page: 1 });
    expect(adminListState(validateAdminSearch(['draft']))).toEqual({ page: 1 });
    expect(() =>
      adminListState(validateAdminSearch({ status: Symbol('draft'), page: {} })),
    ).not.toThrow();
  });

  it('turns a usable page into the bounded backend pagination contract', () => {
    expect(adminCourseQuery({ status: 'archived', page: 501 })).toEqual({
      status: 'archived',
      limit: 20,
      offset: 10000,
    });
  });
});
