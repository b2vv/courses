import { CatalogQueryError } from '../composition/catalog-queries';

/** URL values are text. Keep repeats visible so route validation can reject them. */
export function parseUrlSearch(search: string): Record<string, string | string[]> {
  const parameters = new URLSearchParams(search);
  const entries = new Map<string, string | string[]>();
  for (const key of parameters.keys()) {
    const values = parameters.getAll(key);
    const first = values[0];
    if (first !== undefined) entries.set(key, values.length === 1 ? first : values);
  }
  return Object.fromEntries(entries);
}

/** Pair the parser with plain URL serialization, never JSON-quote literal search text. */
export function stringifyUrlSearch(search: Record<string, unknown>): string {
  const parameters = new URLSearchParams();
  for (const [key, value] of Object.entries(search)) {
    if (value === undefined) continue;
    const values: readonly unknown[] = Array.isArray(value) ? value : [value];
    for (const item of values) {
      if (typeof item !== 'string' && typeof item !== 'number' && typeof item !== 'boolean') {
        throw new CatalogQueryError({ code: 'invalid_query' });
      }
      parameters.append(key, String(item));
    }
  }
  const encoded = parameters.toString();
  return encoded ? `?${encoded}` : '';
}
