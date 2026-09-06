import { describe, expect, it } from '@rstest/core';
import {
  adminCourseQuery,
  adminListState,
  adminUserListState,
  adminUserQuery,
  validateAdminSearch,
  validateAdminUserSearch,
} from './admin-search';
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

describe('admin user list URL search contract', () => {
  it('emits no defaults and creates the default directory query', () => {
    expect(validateAdminUserSearch({})).toEqual({});
    expect(adminUserListState({})).toEqual({ q: '', page: 1 });
    expect(adminUserQuery(adminUserListState({}))).toEqual({ q: '', limit: 20, offset: 0 });
  });

  it('echoes the known user fields verbatim and drops unknown fields', () => {
    expect(validateAdminUserSearch({ status: 'blocked', q: 'Ada', page: '3', x: '1' })).toEqual({
      status: 'blocked',
      q: 'Ada',
      page: '3',
    });
  });

  it.each([
    ['?status=deleted&q=Ada&page=2', { status: 'deleted', q: 'Ada', page: 2 }],
    ['?status=nonsense&q=Ada&page=abc', { q: 'Ada', page: 1 }],
    ['?status=active&q=%00&page=3', { status: 'active', q: '', page: 3 }],
    [`?q=${'a'.repeat(101)}&page=2`, { q: '', page: 2 }],
    ['?status=blocked&q=Ada&q=Grace&page=2', { status: 'blocked', q: '', page: 2 }],
  ])('parses each user field independently for %s', (url, expected) => {
    expect(adminUserListState(validateAdminUserSearch(parseUrlSearch(url)))).toEqual(expected);
  });

  it('never throws and maps the last bounded page to the backend contract', () => {
    expect(() =>
      adminUserListState(validateAdminUserSearch({ status: Symbol('active'), q: {}, page: [] })),
    ).not.toThrow();
    expect(adminUserQuery({ status: 'blocked', q: 'Ada', page: 501 })).toEqual({
      status: 'blocked',
      q: 'Ada',
      limit: 20,
      offset: 10000,
    });
  });
});
