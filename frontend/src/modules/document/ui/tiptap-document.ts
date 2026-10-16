import type { LessonBlock, LessonMark, LessonRichText, LessonSpan } from '../domain/document';

// Pure conversion between our validated rich-text model and the ProseMirror/Tiptap JSON the editor
// works in. It imports no Tiptap runtime, so it is testable without a browser. Serialising out of
// the editor is defensive: anything the model does not allow (unknown nodes or marks, a non-http
// link) is dropped here rather than stored, so the stored document always re-validates.

export type PmMark = { readonly type: string; readonly attrs?: Record<string, unknown> };
export type PmNode = {
  readonly type: string;
  readonly attrs?: Record<string, unknown>;
  readonly content?: readonly PmNode[];
  readonly text?: string;
  readonly marks?: readonly PmMark[];
};
export type PmDoc = { readonly type: 'doc'; readonly content?: readonly PmNode[] };

function httpUrl(value: unknown): value is string {
  if (typeof value !== 'string') return false;
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

// --- model → ProseMirror (to seed the editor) ----------------------------------------------------

function spanToPmNode(span: LessonSpan): PmNode | null {
  if (span.text === '') return null; // ProseMirror forbids an empty text node.
  const marks: PmMark[] = [];
  if (span.marks?.includes('bold')) marks.push({ type: 'bold' });
  if (span.marks?.includes('italic')) marks.push({ type: 'italic' });
  if (span.href !== undefined) marks.push({ type: 'link', attrs: { href: span.href } });
  return marks.length === 0
    ? { type: 'text', text: span.text }
    : { type: 'text', text: span.text, marks };
}

function spansToPmContent(spans: readonly LessonSpan[]): PmNode[] {
  return spans.map(spanToPmNode).filter((node): node is PmNode => node !== null);
}

function paragraphNode(spans: readonly LessonSpan[]): PmNode {
  const content = spansToPmContent(spans);
  return content.length === 0 ? { type: 'paragraph' } : { type: 'paragraph', content };
}

function blockToPmNode(block: LessonBlock): PmNode {
  switch (block.type) {
    case 'paragraph':
      return paragraphNode(block.spans);
    case 'heading': {
      const content = spansToPmContent(block.spans);
      return content.length === 0
        ? { type: 'heading', attrs: { level: block.level } }
        : { type: 'heading', attrs: { level: block.level }, content };
    }
    case 'list':
      return {
        type: block.ordered ? 'orderedList' : 'bulletList',
        content: block.items.map((item) => ({ type: 'listItem', content: [paragraphNode(item)] })),
      };
  }
}

export function richTextToPmDoc(content: LessonRichText): PmDoc {
  if (content.blocks.length === 0) return { type: 'doc', content: [{ type: 'paragraph' }] };
  return { type: 'doc', content: content.blocks.map(blockToPmNode) };
}

// --- ProseMirror → model (to serialise out of the editor) ----------------------------------------

function pmNodeToSpan(node: PmNode): LessonSpan | null {
  if (node.type !== 'text' || typeof node.text !== 'string' || node.text === '') return null;
  const marks: LessonMark[] = [];
  let href: string | undefined;
  for (const mark of node.marks ?? []) {
    if (mark.type === 'bold') marks.push('bold');
    else if (mark.type === 'italic') marks.push('italic');
    else if (mark.type === 'link' && httpUrl(mark.attrs?.href)) href = mark.attrs.href as string;
  }
  return {
    text: node.text,
    ...(marks.length === 0 ? {} : { marks }),
    ...(href === undefined ? {} : { href }),
  };
}

function pmInlineToSpans(content: readonly PmNode[] | undefined): LessonSpan[] {
  return (content ?? []).map(pmNodeToSpan).filter((span): span is LessonSpan => span !== null);
}

function pmListToBlock(node: PmNode, ordered: boolean): LessonBlock {
  const items = (node.content ?? [])
    .filter((item) => item.type === 'listItem')
    .map((item) => {
      const paragraph = (item.content ?? []).find((child) => child.type === 'paragraph');
      return pmInlineToSpans(paragraph?.content);
    });
  return { type: 'list', ordered, items };
}

function pmNodeToBlock(node: PmNode): LessonBlock | null {
  switch (node.type) {
    case 'paragraph':
      return { type: 'paragraph', spans: pmInlineToSpans(node.content) };
    case 'heading': {
      const level = node.attrs?.level;
      return { type: 'heading', level: level === 3 ? 3 : 2, spans: pmInlineToSpans(node.content) };
    }
    case 'bulletList':
      return pmListToBlock(node, false);
    case 'orderedList':
      return pmListToBlock(node, true);
    default:
      return null; // Drop anything the model does not model (blockquote, code block, …).
  }
}

export function pmDocToRichText(doc: PmDoc): LessonRichText {
  return {
    blocks: (doc.content ?? [])
      .map(pmNodeToBlock)
      .filter((block): block is LessonBlock => block !== null),
  };
}
