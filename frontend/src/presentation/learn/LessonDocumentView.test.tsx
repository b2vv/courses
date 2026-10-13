import { expect, test } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import type { LessonDocument } from '../../domain/document/document';
import { LessonDocumentView } from './LessonDocumentView';

const imageId = '60000000-0000-4000-8000-000000000001';
const jpegId = '60000000-0000-4000-8000-000000000002';

function documentWith(...sections: LessonDocument['sections']): LessonDocument {
  return { schemaVersion: 1, locale: 'uk', sections };
}

test('a document with nothing but a video slot renders no body: the player shows the video', () => {
  const { container } = render(
    <LessonDocumentView document={documentWith({ type: 'video', id: 'v' })} />,
  );
  expect(container.firstChild).toBeNull();
});

test('an empty document renders nothing, the same as an absent publication', () => {
  const { container } = render(<LessonDocumentView document={documentWith()} />);
  expect(container.firstChild).toBeNull();
});

test('each section type renders as its own semantic element in order', () => {
  render(
    <LessonDocumentView
      document={documentWith(
        { type: 'video', id: 'v' },
        {
          type: 'text',
          id: 'text',
          content: {
            blocks: [
              { type: 'heading', level: 2, spans: [{ text: 'Замовляння' }] },
              { type: 'heading', level: 3, spans: [{ text: 'Крок' }] },
              {
                type: 'paragraph',
                spans: [
                  { text: 'жирне', marks: ['bold'] },
                  { text: 'курсивне', marks: ['italic'] },
                  { text: 'посилання', href: 'https://example.test/a' },
                ],
              },
              { type: 'list', ordered: false, items: [[{ text: 'пункт' }]] },
              { type: 'list', ordered: true, items: [[{ text: 'крок' }]] },
            ],
          },
        },
        {
          type: 'row',
          id: 'row',
          columns: [
            {
              type: 'image',
              id: 'c-img',
              imageId: jpegId,
              contentType: 'image/jpeg',
              alt: 'Практика',
            },
            {
              type: 'text',
              id: 'c-text',
              content: { blocks: [{ type: 'paragraph', spans: [{ text: 'Пояснення' }] }] },
            },
          ],
        },
        { type: 'image', id: 'img', imageId, contentType: 'image/png', alt: 'Схема' },
        {
          type: 'task',
          id: 'task',
          content: { blocks: [{ type: 'paragraph', spans: [{ text: 'Завдання' }] }] },
        },
      )}
    />,
  );

  const region = screen.getByRole('region', { name: 'Матеріали уроку' });

  // Rich text: headings at their level, marks, and a safe external link.
  expect(within(region).getByRole('heading', { level: 2, name: 'Замовляння' })).toBeTruthy();
  expect(within(region).getByRole('heading', { level: 3, name: 'Крок' })).toBeTruthy();
  expect(region.querySelector('strong')?.textContent).toBe('жирне');
  expect(region.querySelector('em')?.textContent).toBe('курсивне');
  const link = within(region).getByRole('link', { name: 'посилання' });
  expect(link.getAttribute('href')).toBe('https://example.test/a');
  expect(link.getAttribute('rel')).toContain('noopener');
  expect(within(region).getByText('пункт').closest('ul')).toBeTruthy();
  expect(within(region).getByText('крок').closest('ol')).toBeTruthy();

  // Image: served from /media/i/<id>.<ext>, extension derived from the content type.
  const png = within(region).getByRole('img', { name: 'Схема' });
  expect(png.getAttribute('src')).toBe(`/media/i/${imageId}.png`);
  const jpeg = within(region).getByRole('img', { name: 'Практика' });
  expect(jpeg.getAttribute('src')).toBe(`/media/i/${jpegId}.jpg`);

  // Row carries its column count for the responsive preset.
  const row = region.querySelector('.lesson-document-row');
  expect(row?.getAttribute('data-columns')).toBe('2');

  // Task is distinguishable from plain text.
  expect(region.querySelector('[data-section="task"]')?.textContent).toContain('Завдання');
});

test('a webp image uses the webp extension', () => {
  render(
    <LessonDocumentView
      document={documentWith({
        type: 'image',
        id: 'img',
        imageId,
        contentType: 'image/webp',
        alt: 'W',
      })}
    />,
  );
  expect(screen.getByRole('img', { name: 'W' }).getAttribute('src')).toBe(
    `/media/i/${imageId}.webp`,
  );
});

test('a migrated flat description renders as its single text section (scenario 7 parity)', () => {
  render(
    <LessonDocumentView
      document={documentWith({
        type: 'text',
        id: 'lesson-s1',
        content: {
          blocks: [{ type: 'paragraph', spans: [{ text: 'Старий плоский опис уроку' }] }],
        },
      })}
    />,
  );
  expect(screen.getByText('Старий плоский опис уроку')).toBeTruthy();
});
