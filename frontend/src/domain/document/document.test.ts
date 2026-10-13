import { expect, test } from '@rstest/core';
import {
  hasRenderableSections,
  imageExtensionFor,
  imageServeUrl,
  type LessonDocument,
} from './document';

test('image content types map to the extension the uploader stored', () => {
  expect(imageExtensionFor('image/jpeg')).toBe('jpg');
  expect(imageExtensionFor('image/png')).toBe('png');
  expect(imageExtensionFor('image/webp')).toBe('webp');
});

test('the serve url is /media/i/<id>.<ext>', () => {
  expect(imageServeUrl('abc', 'image/png')).toBe('/media/i/abc.png');
  expect(imageServeUrl('abc', 'image/jpeg')).toBe('/media/i/abc.jpg');
});

test('a document that is only a video slot has nothing for the reader to render', () => {
  const videoOnly: LessonDocument = {
    schemaVersion: 1,
    locale: 'uk',
    sections: [{ type: 'video', id: 'v' }],
  };
  expect(hasRenderableSections(videoOnly)).toBe(false);
  expect(hasRenderableSections({ schemaVersion: 1, locale: 'uk', sections: [] })).toBe(false);
  expect(
    hasRenderableSections({
      schemaVersion: 1,
      locale: 'uk',
      sections: [
        { type: 'video', id: 'v' },
        { type: 'text', id: 't', content: { blocks: [] } },
      ],
    }),
  ).toBe(true);
});
