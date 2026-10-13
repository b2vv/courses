import { expect, test } from '@rstest/core';
import { lessonDocumentSchema } from './document-schema';

const imageId = '60000000-0000-4000-8000-000000000001';

test('a well-formed document with every section shape parses', () => {
  const parsed = lessonDocumentSchema.safeParse({
    schemaVersion: 1,
    locale: 'uk',
    sections: [
      { id: 'v', type: 'video', content: null },
      {
        id: 'text',
        type: 'text',
        content: {
          blocks: [
            {
              type: 'paragraph',
              spans: [{ text: 'a', marks: ['bold'], href: 'https://ok.test/x' }],
            },
            { type: 'heading', level: 2, spans: [{ text: 'h' }] },
            { type: 'list', ordered: true, items: [[{ text: 'i' }], null] },
          ],
        },
      },
      { id: 'img', type: 'image', imageId, contentType: 'image/png', alt: 'Alt' },
      {
        id: 'row',
        type: 'row',
        columns: [{ id: 'c', type: 'image', imageId, contentType: 'image/webp', alt: 'A' }],
      },
    ],
  });
  expect(parsed.success).toBe(true);
});

test('null sections and null blocks parse as absent, not as a failure', () => {
  expect(
    lessonDocumentSchema.safeParse({ schemaVersion: 1, locale: 'uk', sections: null }).success,
  ).toBe(true);
  expect(
    lessonDocumentSchema.safeParse({
      schemaVersion: 1,
      locale: 'uk',
      sections: [{ id: 't', type: 'text', content: { blocks: null } }],
    }).success,
  ).toBe(true);
});

test('a non-http href is refused before it reaches an anchor', () => {
  const parsed = lessonDocumentSchema.safeParse({
    schemaVersion: 1,
    locale: 'uk',
    sections: [
      {
        id: 't',
        type: 'text',
        // eslint-disable-next-line no-script-url -- the point of the test is that this is refused
        content: {
          blocks: [{ type: 'paragraph', spans: [{ text: 'x', href: 'javascript:alert(1)' }] }],
        },
      },
    ],
  });
  expect(parsed.success).toBe(false);
});

test('an image id that is not a UUID is refused before it reaches a url path', () => {
  const parsed = lessonDocumentSchema.safeParse({
    schemaVersion: 1,
    locale: 'uk',
    sections: [
      { id: 'img', type: 'image', imageId: '../escape', contentType: 'image/png', alt: 'A' },
    ],
  });
  expect(parsed.success).toBe(false);
});

test('an unknown content type and an unknown section type are both refused', () => {
  expect(
    lessonDocumentSchema.safeParse({
      schemaVersion: 1,
      locale: 'uk',
      sections: [{ id: 'img', type: 'image', imageId, contentType: 'image/gif', alt: 'A' }],
    }).success,
  ).toBe(false);
  expect(
    lessonDocumentSchema.safeParse({
      schemaVersion: 1,
      locale: 'uk',
      sections: [{ id: 'x', type: 'iframe' }],
    }).success,
  ).toBe(false);
});

test('a row may not nest another row', () => {
  const parsed = lessonDocumentSchema.safeParse({
    schemaVersion: 1,
    locale: 'uk',
    sections: [{ id: 'row', type: 'row', columns: [{ id: 'nested', type: 'row', columns: [] }] }],
  });
  expect(parsed.success).toBe(false);
});
