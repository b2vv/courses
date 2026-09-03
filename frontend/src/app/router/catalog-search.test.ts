import { describe, expect, it } from '@rstest/core';
import { validateCatalogSearch } from './catalog-search';
import { parseUrlSearch, stringifyUrlSearch } from './url-search';

describe('catalog URL search contract', () => {
  it.each(['2026', 'true', 'false', 'null', '"Go"', '[1]', '{"topic":"Go"}', '001', 'Go & API'])(
    'preserves literal search text %s',
    (q) => {
      const url = `?${new URLSearchParams({ q })}`;
      expect(validateCatalogSearch(parseUrlSearch(url)).q).toBe(q);
      expect(validateCatalogSearch(parseUrlSearch(stringifyUrlSearch({ q }))).q).toBe(q);
    },
  );

  it.each([
    '?limit=1e1',
    '?limit=1.0',
    '?offset=1e2',
    '?offset=-0',
    '?limit=',
    '?limit=1&limit=2',
    '?q=Go&q=React',
  ])('rejects invalid or repeated parameters %s', (url) => {
    expect(() => validateCatalogSearch(parseUrlSearch(url))).toThrow();
  });

  it('round-trips normalized catalog state without JSON-quoting string values', () => {
    const state = { q: '2026', sort: 'title', limit: 10, offset: 20 };
    const search = stringifyUrlSearch(state);
    expect(new URLSearchParams(search).get('q')).toBe('2026');
    expect(validateCatalogSearch(parseUrlSearch(search))).toEqual(state);
  });
});
