import { type ReactNode, type RefObject, useRef, useState } from 'react';
import * as Aria from 'react-aria-components';
import { useTranslation } from 'react-i18next';
import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Link from '@tiptap/extension-link';
import type { LessonRichText } from '../../domain/document/document';
import { Button, TextField, Tooltip } from '../ui';
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

const toolbarIcons = {
  bulletList: 'M8 6h12M8 12h12M8 18h12M4 6h.01M4 12h.01M4 18h.01',
  orderedList: 'M9 6h11M9 12h11M9 18h11M4 4v5M3 12h3l-3 6h3',
  link: 'm10 14 4-4M8 16l-2 2a4 4 0 0 1-5-5l4-4a4 4 0 0 1 5 0m4-1 2-2a4 4 0 0 1 5 5l-4 4a4 4 0 0 1-5 0',
} as const;

function ToolbarIcon({ path }: { readonly path: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="18"
      height="18"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path d={path} />
    </svg>
  );
}

// A single toolbar command: an icon/glyph button with its label surfaced through a Tooltip and kept
// as the accessible name. `groupStart` draws the thin rule that separates formatting | headings | lists.
function ToolbarButton({
  label,
  isActive,
  groupStart = false,
  onPress,
  children,
}: {
  readonly label: string;
  readonly isActive: boolean;
  readonly groupStart?: boolean;
  readonly onPress: () => void;
  readonly children: ReactNode;
}) {
  return (
    <Tooltip content={label}>
      <Button
        variant="secondary"
        aria-label={label}
        aria-pressed={isActive}
        onPress={onPress}
        className={groupStart ? 'toolbar-button toolbar-group-start' : 'toolbar-button'}
      >
        {children}
      </Button>
    </Tooltip>
  );
}

// One Tiptap editor for a rich-text block. Uncontrolled after mount: it is seeded once from `value`
// and reports every change back through `onChange`, so the parent never feeds content in and the
// caret never jumps. Only the marks and nodes the model allows are enabled; the converter drops
// anything else on the way out.
export function LessonDocumentRichText({ value, label, onChange }: LessonDocumentRichTextProps) {
  const { t } = useTranslation();
  const [linkUrl, setLinkUrl] = useState('');
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
  // Apply keeps the dialog open and re-focuses the field on an invalid address, so a typo never
  // closes the box or silently drops the link. `editor.chain().focus()` restores the selection the
  // modal took DOM focus away from, then sets the link over it.
  const applyLink = (close: () => void) => {
    if (!httpUrl(linkUrl)) {
      setLinkError(true);
      linkInputRef.current?.focus();
      return;
    }
    editor.chain().focus().setLink({ href: linkUrl }).run();
    close();
  };
  const linkLabel = t('administration.document.link');
  return (
    <div className="lesson-editor-richtext">
      <div
        className="lesson-editor-toolbar"
        role="toolbar"
        aria-label={t('administration.document.toolbar')}
      >
        <ToolbarButton
          label={t('administration.document.bold')}
          isActive={editor.isActive('bold')}
          onPress={() => editor.chain().focus().toggleBold().run()}
        >
          <span aria-hidden="true">B</span>
        </ToolbarButton>
        <ToolbarButton
          label={t('administration.document.italic')}
          isActive={editor.isActive('italic')}
          onPress={() => editor.chain().focus().toggleItalic().run()}
        >
          <span className="italic" aria-hidden="true">
            I
          </span>
        </ToolbarButton>
        <ToolbarButton
          label={t('administration.document.heading2')}
          isActive={editor.isActive('heading', { level: 2 })}
          groupStart
          onPress={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
        >
          <span aria-hidden="true">H2</span>
        </ToolbarButton>
        <ToolbarButton
          label={t('administration.document.heading3')}
          isActive={editor.isActive('heading', { level: 3 })}
          onPress={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
        >
          <span aria-hidden="true">H3</span>
        </ToolbarButton>
        <ToolbarButton
          label={t('administration.document.bulletList')}
          isActive={editor.isActive('bulletList')}
          groupStart
          onPress={() => editor.chain().focus().toggleBulletList().run()}
        >
          <ToolbarIcon path={toolbarIcons.bulletList} />
        </ToolbarButton>
        <ToolbarButton
          label={t('administration.document.orderedList')}
          isActive={editor.isActive('orderedList')}
          onPress={() => editor.chain().focus().toggleOrderedList().run()}
        >
          <ToolbarIcon path={toolbarIcons.orderedList} />
        </ToolbarButton>
        <LinkDialog
          label={linkLabel}
          isActive={editor.isActive('link')}
          linkUrl={linkUrl}
          linkError={linkError}
          linkInputRef={linkInputRef}
          onUrlChange={(next) => {
            setLinkUrl(next);
            setLinkError(false);
          }}
          onReset={() => {
            setLinkUrl('');
            setLinkError(false);
          }}
          onApply={applyLink}
        />
        {editor.isActive('link') && (
          <Button
            variant="secondary"
            className="toolbar-button"
            onPress={() => editor.chain().focus().unsetLink().run()}
          >
            {t('administration.document.removeLink')}
          </Button>
        )}
        <span className="toolbar-caption" aria-hidden="true">
          {label}
        </span>
      </div>
      <EditorContent editor={editor} className="lesson-editor-input" />
    </div>
  );
}

// The link control: an icon button that opens a modal to type the address, rather than an inline
// field. Opening resets the field; dismissing (Escape, scrim, Cancel) returns focus to the button.
function LinkDialog({
  label,
  isActive,
  linkUrl,
  linkError,
  linkInputRef,
  onUrlChange,
  onReset,
  onApply,
}: {
  readonly label: string;
  readonly isActive: boolean;
  readonly linkUrl: string;
  readonly linkError: boolean;
  readonly linkInputRef: RefObject<HTMLInputElement | null>;
  readonly onUrlChange: (value: string) => void;
  readonly onReset: () => void;
  readonly onApply: (close: () => void) => void;
}) {
  const { t } = useTranslation();
  return (
    <Aria.DialogTrigger
      onOpenChange={(open) => {
        if (open) onReset();
      }}
    >
      <Tooltip content={label}>
        <Button
          variant="secondary"
          aria-label={label}
          aria-pressed={isActive}
          className="toolbar-button"
        >
          <ToolbarIcon path={toolbarIcons.link} />
        </Button>
      </Tooltip>
      <Aria.ModalOverlay className="ui-modal-overlay" isDismissable>
        <Aria.Modal className="ui-modal">
          <Aria.Dialog className="ui-dialog" aria-label={t('administration.document.linkHeading')}>
            {({ close }) => (
              <>
                <Aria.Heading slot="title">{t('administration.document.linkHeading')}</Aria.Heading>
                <TextField
                  ref={linkInputRef}
                  label={t('administration.document.linkUrl')}
                  type="url"
                  value={linkUrl}
                  onChange={onUrlChange}
                  isInvalid={linkError}
                  {...(linkError ? { errorMessage: t('administration.document.linkInvalid') } : {})}
                  placeholder="https://"
                />
                <div className="ui-dialog-actions">
                  <Button variant="secondary" onPress={close}>
                    {t('administration.document.linkCancel')}
                  </Button>
                  <Button onPress={() => onApply(close)}>
                    {t('administration.document.linkConfirm')}
                  </Button>
                </div>
              </>
            )}
          </Aria.Dialog>
        </Aria.Modal>
      </Aria.ModalOverlay>
    </Aria.DialogTrigger>
  );
}
