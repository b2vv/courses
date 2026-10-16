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
    expect(adminListState({})).toEqual({ search: '', page: 1 });
    expect(adminCourseQuery(adminListState({}))).toEqual({ limit: 20, offset: 0, search: '' });
  });

  it('echoes known parameters verbatim and drops unknown parameters', () => {
    expect(validateAdminSearch({ status: 'draft', search: 'Go', page: '3', x: '1' })).toEqual({
      status: 'draft',
      search: 'Go',
      page: '3',
    });
  });

  it.each([
    ['?status=published&page=2', { search: '', status: 'published', page: 2 }],
    ['?status=nonsense&page=3', { search: '', page: 3 }],
    ['?status=archived&page=abc', { search: '', status: 'archived', page: 1 }],
    ['?status=draft&page=0', { search: '', status: 'draft', page: 1 }],
    ['?status=draft&page=502', { search: '', status: 'draft', page: 1 }],
    ['?status=draft&status=published&page=2&page=3', { search: '', page: 1 }],
    ['?search=%20Go%20&page=2', { search: 'Go', page: 2 }],
    [`?search=${'a'.repeat(101)}&page=2`, { search: '', page: 2 }],
    ['?search=%00&page=2', { search: '', page: 2 }],
    ['?search=Go&search=Rust&page=2', { search: '', page: 2 }],
  ])('parses each field independently for %s', (url, expected) => {
    expect(adminListState(validateAdminSearch(parseUrlSearch(url)))).toEqual(expected);
  });

  it('never throws for malformed or non-object input', () => {
    expect(adminListState(validateAdminSearch(null))).toEqual({ search: '', page: 1 });
    expect(adminListState(validateAdminSearch(['draft']))).toEqual({ search: '', page: 1 });
    expect(() =>
      adminListState(validateAdminSearch({ status: Symbol('draft'), search: {}, page: {} })),
    ).not.toThrow();
  });

  it('turns a usable page into the bounded backend pagination contract', () => {
    expect(adminCourseQuery({ status: 'archived', search: 'Go', page: 501 })).toEqual({
      status: 'archived',
      search: 'Go',
      limit: 20,
      offset: 10000,
    });
  });
});

describe('admin user list URL search contract', () => {
  it('emits no defaults and creates the default directory query', () => {
    expect(validateAdminUserSearch({})).toEqual({});
    expect(adminUserListState({})).toEqual({ q: '', page: 1, sort: 'email' });
    expect(adminUserQuery(adminUserListState({}))).toEqual({ q: '', limit: 20, offset: 0 });
  });

  it('echoes the known user fields verbatim and drops unknown fields', () => {
    expect(
      validateAdminUserSearch({
        status: 'blocked',
        role: 'instructor',
        sort: 'recent',
        q: 'Ada',
        page: '3',
        x: '1',
      }),
    ).toEqual({
      status: 'blocked',
      role: 'instructor',
      sort: 'recent',
      q: 'Ada',
      page: '3',
    });
  });

  it.each([
    ['?status=deleted&q=Ada&page=2', { status: 'deleted', q: 'Ada', page: 2, sort: 'email' }],
    ['?status=nonsense&q=Ada&page=abc', { q: 'Ada', page: 1, sort: 'email' }],
    ['?status=active&q=%00&page=3', { status: 'active', q: '', page: 3, sort: 'email' }],
    [`?q=${'a'.repeat(101)}&page=2`, { q: '', page: 2, sort: 'email' }],
    ['?status=blocked&q=Ada&q=Grace&page=2', { status: 'blocked', q: '', page: 2, sort: 'email' }],
    ['?role=instructor&sort=recent&page=2', { role: 'instructor', sort: 'recent', q: '', page: 2 }],
    ['?role=nonsense&sort=nonsense&page=1', { q: '', page: 1, sort: 'email' }],
  ])('parses each user field independently for %s', (url, expected) => {
    expect(adminUserListState(validateAdminUserSearch(parseUrlSearch(url)))).toEqual(expected);
  });

  it('carries role and non-default sort into the backend query, omitting the email default', () => {
    expect(adminUserQuery({ role: 'student', sort: 'name', q: '', page: 1 })).toEqual({
      q: '',
      limit: 20,
      offset: 0,
      role: 'student',
      sort: 'name',
    });
    expect(adminUserQuery({ sort: 'email', q: '', page: 1 })).toEqual({
      q: '',
      limit: 20,
      offset: 0,
    });
  });

  it('never throws and maps the last bounded page to the backend contract', () => {
    expect(() =>
      adminUserListState(
        validateAdminUserSearch({ status: Symbol('active'), role: {}, sort: [], q: {}, page: [] }),
      ),
    ).not.toThrow();
    expect(adminUserQuery({ status: 'blocked', sort: 'email', q: 'Ada', page: 501 })).toEqual({
      status: 'blocked',
      q: 'Ada',
      limit: 20,
      offset: 10000,
    });
  });
});
