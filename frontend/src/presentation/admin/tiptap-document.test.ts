import { expect, test } from '@rstest/core';
import type { LessonRichText } from '../../domain/document/document';
import { pmDocToRichText, richTextToPmDoc, type PmDoc } from './tiptap-document';

test('a rich text survives model → ProseMirror → model unchanged', () => {
  const content: LessonRichText = {
    blocks: [
      { type: 'heading', level: 2, spans: [{ text: 'Заголовок', marks: ['bold'] }] },
      {
        type: 'paragraph',
        spans: [
          { text: 'жирне', marks: ['bold'] },
          { text: 'курсив', marks: ['italic'] },
          { text: 'лінк', href: 'https://example.test/a' },
        ],
      },
      { type: 'list', ordered: true, items: [[{ text: 'крок' }]] },
      { type: 'list', ordered: false, items: [[{ text: 'пункт' }]] },
    ],
  };
  expect(pmDocToRichText(richTextToPmDoc(content))).toEqual(content);
});

test('an empty rich text seeds one empty paragraph and round-trips to no blocks', () => {
  const pm = richTextToPmDoc({ blocks: [] });
  expect(pm.content).toEqual([{ type: 'paragraph' }]);
  // An empty paragraph carries no spans.
  expect(pmDocToRichText(pm)).toEqual({ blocks: [{ type: 'paragraph', spans: [] }] });
});

test('unknown nodes and marks the model cannot hold are dropped on serialize', () => {
  const doc: PmDoc = {
    type: 'doc',
    content: [
      {
        type: 'blockquote',
        content: [{ type: 'paragraph', content: [{ type: 'text', text: 'q' }] }],
      },
      {
        type: 'paragraph',
        content: [{ type: 'text', text: 'x', marks: [{ type: 'code' }, { type: 'bold' }] }],
      },
    ],
  };
  expect(pmDocToRichText(doc)).toEqual({
    blocks: [{ type: 'paragraph', spans: [{ text: 'x', marks: ['bold'] }] }],
  });
});

test('a non-http link href is dropped, keeping the text', () => {
  const doc: PmDoc = {
    type: 'doc',
    content: [
      {
        type: 'paragraph',
        // eslint-disable-next-line no-script-url -- the point is that this href is dropped
        content: [
          {
            type: 'text',
            text: 'click',
            marks: [{ type: 'link', attrs: { href: 'javascript:alert(1)' } }],
          },
        ],
      },
    ],
  };
  expect(pmDocToRichText(doc)).toEqual({
    blocks: [{ type: 'paragraph', spans: [{ text: 'click' }] }],
  });
});

test('a heading level outside 2 and 3 is clamped to a model-legal level', () => {
  const doc: PmDoc = {
    type: 'doc',
    content: [
      { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: 'a' }] },
      { type: 'heading', attrs: { level: 3 }, content: [{ type: 'text', text: 'b' }] },
    ],
  };
  expect(pmDocToRichText(doc)).toEqual({
    blocks: [
      { type: 'heading', level: 2, spans: [{ text: 'a' }] },
      { type: 'heading', level: 3, spans: [{ text: 'b' }] },
    ],
  });
});

test('an empty text span is not emitted as an invalid ProseMirror text node', () => {
  const pm = richTextToPmDoc({
    blocks: [{ type: 'paragraph', spans: [{ text: '' }, { text: 'ok' }] }],
  });
  expect(pm.content).toEqual([{ type: 'paragraph', content: [{ type: 'text', text: 'ok' }] }]);
});
