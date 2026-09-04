import { describe, expect, it } from '@rstest/core';
import { validateLoginSearch } from './login-search';

describe('login URL search contract', () => {
  it('emits no key for a plain /login so the router does not rewrite the URL', () => {
    expect(validateLoginSearch({})).toEqual({});
    expect(Object.hasOwn(validateLoginSearch({}), 'returnTo')).toBe(false);
  });

  it.each(['/learn', '/courses/go-basics', '/courses?q=Go'])(
    'keeps the internal path %s',
    (returnTo) => {
      expect(validateLoginSearch({ returnTo })).toEqual({ returnTo });
    },
  );

  it.each(['https://evil.example', '//evil.example', '/\\evil.example', 'evil.example', ''])(
    'strips the unusable target %p from the address instead of honouring it',
    (returnTo) => {
      expect(validateLoginSearch({ returnTo })).toEqual({});
    },
  );

  it('ignores a malformed search object entirely', () => {
    expect(validateLoginSearch({ returnTo: 7 })).toEqual({});
  });
});
