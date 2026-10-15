import { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Link from '@tiptap/extension-link';
import type { LessonRichText } from '../../domain/document/document';
import { Button, TextField } from '../ui';
import { pmDocToRichText, richTextToPmDoc, type PmDoc } from './tiptap-document';

type LessonDocumentRichTextProps = {
  readonly value: LessonRichText;
  readonly label: string;
  readonly onChange: (value: LessonRichText) => void;
};

function httpUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

// One Tiptap editor for a rich-text block. Uncontrolled after mount: it is seeded once from `value`
// and reports every change back through `onChange`, so the parent never feeds content in and the
// caret never jumps. Only the marks and nodes the model allows are enabled; the converter drops
// anything else on the way out.
export function LessonDocumentRichText({ value, label, onChange }: LessonDocumentRichTextProps) {
  const { t } = useTranslation();
  const [linkUrl, setLinkUrl] = useState('');
  const [addingLink, setAddingLink] = useState(false);
  const [linkError, setLinkError] = useState(false);
  const linkInputRef = useRef<HTMLInputElement>(null);
  const editor = useEditor({
    immediatelyRender: false,
    // Re-render the toolbar on every transaction (incl. selection-only moves) so active-state
    // (aria-pressed) and the "remove link" control reflect the caret's current marks. Without this
    // @tiptap/react only re-renders on doc changes, so clicking into an existing link left the
    // remove-link button hidden. Lesson rich-text is short, so the per-transaction render is cheap.
    shouldRerenderOnTransaction: true,
    extensions: [
      StarterKit.configure({
        heading: { levels: [2, 3] },
        strike: false,
        code: false,
        codeBlock: false,
        blockquote: false,
        horizontalRule: false,
      }),
      Link.configure({ openOnClick: false, autolink: false }),
    ],
    // A mutable clone: Tiptap's Content type is not readonly, and the editor owns its copy.
    content: JSON.parse(JSON.stringify(richTextToPmDoc(value))),
    onUpdate: ({ editor: instance }) => onChange(pmDocToRichText(instance.getJSON() as PmDoc)),
    editorProps: {
      attributes: {
        'aria-label': label,
        role: 'textbox',
        'aria-multiline': 'true',
        class: 'lesson-prose',
      },
    },
  });
  if (editor === null) return null;
  const applyLink = () => {
    if (!httpUrl(linkUrl)) {
      setLinkError(true);
      linkInputRef.current?.focus();
      return;
    }
    editor.chain().focus().setLink({ href: linkUrl }).run();
    setLinkUrl('');
    setAddingLink(false);
  };
  return (
    <div className="lesson-editor-richtext">
      <div
        className="lesson-editor-toolbar"
        role="toolbar"
        aria-label={t('administration.document.toolbar')}
      >
        <div className="lesson-editor-toolbar-group">
          <Button
            variant="secondary"
            aria-pressed={editor.isActive('bold')}
            onPress={() => editor.chain().focus().toggleBold().run()}
          >
            {t('administration.document.bold')}
          </Button>
          <Button
            variant="secondary"
            aria-pressed={editor.isActive('italic')}
            onPress={() => editor.chain().focus().toggleItalic().run()}
          >
            {t('administration.document.italic')}
          </Button>
        </div>
        <div className="lesson-editor-toolbar-group">
          <Button
            variant="secondary"
            aria-pressed={editor.isActive('heading', { level: 2 })}
            onPress={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
          >
            {t('administration.document.heading2')}
          </Button>
          <Button
            variant="secondary"
            aria-pressed={editor.isActive('heading', { level: 3 })}
            onPress={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
          >
            {t('administration.document.heading3')}
          </Button>
        </div>
        <div className="lesson-editor-toolbar-group">
          <Button
            variant="secondary"
            aria-pressed={editor.isActive('bulletList')}
            onPress={() => editor.chain().focus().toggleBulletList().run()}
          >
            {t('administration.document.bulletList')}
          </Button>
          <Button
            variant="secondary"
            aria-pressed={editor.isActive('orderedList')}
            onPress={() => editor.chain().focus().toggleOrderedList().run()}
          >
            {t('administration.document.orderedList')}
          </Button>
        </div>
        <div className="lesson-editor-toolbar-group">
          <Button
            variant="secondary"
            aria-pressed={addingLink}
            onPress={() => setAddingLink((on) => !on)}
          >
            {t('administration.document.link')}
          </Button>
          {editor.isActive('link') && (
            <Button variant="secondary" onPress={() => editor.chain().focus().unsetLink().run()}>
              {t('administration.document.removeLink')}
            </Button>
          )}
        </div>
      </div>
      {addingLink && (
        <div className="lesson-editor-link">
          <TextField
            ref={linkInputRef}
            label={t('administration.document.linkUrl')}
            type="url"
            value={linkUrl}
            onChange={(value) => {
              setLinkUrl(value);
              setLinkError(false);
            }}
            isInvalid={linkError}
            {...(linkError ? { errorMessage: t('administration.document.linkInvalid') } : {})}
            placeholder="https://"
          />
          <Button variant="secondary" onPress={applyLink}>
            {t('administration.document.apply')}
          </Button>
        </div>
      )}
      <EditorContent editor={editor} className="lesson-editor-input" />
    </div>
  );
}
