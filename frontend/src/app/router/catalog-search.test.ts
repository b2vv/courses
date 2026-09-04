import { describe, expect, it } from '@rstest/core';
import { isNotFound } from '@tanstack/react-router';
import { catalogQuery, validateCatalogSearch } from './catalog-search';
import { parseUrlSearch, stringifyUrlSearch } from './url-search';

describe('catalog URL search contract', () => {
  it.each(['2026', 'true', 'false', 'null', '"Go"', '[1]', '{"topic":"Go"}', '001', 'Go & API'])(
    'preserves literal search text %s',
    (q) => {
      const url = `?${new URLSearchParams({ q })}`;
      expect(catalogQuery(validateCatalogSearch(parseUrlSearch(url))).q).toBe(q);
      expect(catalogQuery(validateCatalogSearch(parseUrlSearch(stringifyUrlSearch({ q })))).q).toBe(
        q,
      );
    },
  );

  it('emits no key for a plain /courses so the router does not rewrite the URL', () => {
    expect(validateCatalogSearch({})).toEqual({});
    expect(Object.keys(validateCatalogSearch({}))).toEqual([]);
    expect(catalogQuery({})).toEqual({ q: '', sort: 'recent', limit: 20, offset: 0 });
  });

  it('echoes only the parameters the address carried', () => {
    expect(validateCatalogSearch({ offset: '20' })).toEqual({ offset: '20' });
    expect(catalogQuery(validateCatalogSearch({ offset: '20' }))).toEqual({
      q: '',
      sort: 'recent',
      limit: 20,
      offset: 20,
    });
  });

  it.each(['?limit=51', '?offset=10001', '?limit=0'])(
    'refuses the out-of-range %s that the schema alone would accept',
    (url) => {
      expect(() => catalogQuery(validateCatalogSearch(parseUrlSearch(url)))).toThrow();
    },
  );

  it('keeps an unusable value in the address so the URL is not rewritten', () => {
    expect(validateCatalogSearch({ limit: 'abc' })).toEqual({ limit: 'abc' });
    expect(validateCatalogSearch({ role: 'admin' })).toEqual({});
  });

  it('refuses an unusable address as not found rather than as a server error', () => {
    try {
      catalogQuery(validateCatalogSearch(parseUrlSearch('?limit=abc')));
      throw new Error('expected a refusal');
    } catch (cause) {
      expect(isNotFound(cause)).toBe(true);
    }
  });

  it.each([
    '?limit=1e1',
    '?limit=1.0',
    '?offset=1e2',
    '?offset=-0',
    '?limit=',
    '?limit=1&limit=2',
    '?q=Go&q=React',
  ])('rejects invalid or repeated parameters %s', (url) => {
    expect(() => catalogQuery(validateCatalogSearch(parseUrlSearch(url)))).toThrow();
  });

  it('round-trips normalized catalog state without JSON-quoting string values', () => {
    const state = { q: '2026', sort: 'title', limit: 10, offset: 20 };
    const search = stringifyUrlSearch(state);
    expect(new URLSearchParams(search).get('q')).toBe('2026');
    expect(catalogQuery(validateCatalogSearch(parseUrlSearch(search)))).toEqual(state);
  });
});
