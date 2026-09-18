import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type {
  TranscriptDocument,
  TranscriptErrorCode,
  TranscriptSegment,
} from '../../domain/transcript/transcript';
import { Button, TextArea, TextField } from '../ui';

type DraftLine = {
  id: string;
  startMs: string;
  endMs: string;
  text: string;
  documentIds: readonly string[];
};

export type TranscriptEditorActions = {
  save(
    language: string,
    expectedUpdatedAt: string | undefined,
    segments: readonly TranscriptSegment[],
  ): Promise<TranscriptErrorCode | null>;
};

type TranscriptEditorProps = {
  readonly lessonId: string;
  readonly language: string;
  readonly expectedUpdatedAt?: string;
  readonly segments: readonly TranscriptSegment[];
  readonly documents: readonly TranscriptDocument[];
  readonly actions: TranscriptEditorActions;
};

function toDraft(segment: TranscriptSegment): DraftLine {
  return {
    id: segment.id,
    startMs: String(segment.startMs),
    endMs: String(segment.endMs),
    text: segment.text,
    documentIds: segment.documentIds,
  };
}

export function TranscriptEditor({
  lessonId,
  language,
  expectedUpdatedAt,
  segments,
  documents,
  actions,
}: TranscriptEditorProps) {
  const { t } = useTranslation();
  const [lines, setLines] = useState<readonly DraftLine[]>(() => segments.map(toDraft));
  const [status, setStatus] = useState<'idle' | 'saving' | 'saved'>('idle');
  const [failure, setFailure] = useState<TranscriptErrorCode | null>(null);

  // Times are held as TEXT while editing and parsed only on save. Parsing on every keystroke turns
  // a half-typed "1" into a real value and reorders the lines under the author's cursor.
  const update = (index: number, change: Partial<DraftLine>) =>
    setLines((current) =>
      current.map((line, position) => (position === index ? { ...line, ...change } : line)),
    );

  const parsed = (): readonly TranscriptSegment[] | null => {
    const result: TranscriptSegment[] = [];
    for (const line of lines) {
      const startMs = Number(line.startMs);
      const endMs = Number(line.endMs);
      if (!Number.isInteger(startMs) || !Number.isInteger(endMs)) return null;
      result.push({
        id: line.id,
        startMs,
        endMs,
        text: line.text,
        documentIds: line.documentIds,
      });
    }
    return result;
  };

  return (
    <section className="transcript-editor" aria-labelledby={`transcript-editor-${lessonId}`}>
      <h4 id={`transcript-editor-${lessonId}`}>{t('administration.transcriptEditor.heading')}</h4>

      {lines.length === 0 ? <p>{t('administration.transcriptEditor.empty')}</p> : null}

      <ol className="transcript-editor-lines">
        {lines.map((line, index) => (
          <li key={line.id}>
            <TextField
              label={t('administration.transcriptEditor.start', { line: index + 1 })}
              value={line.startMs}
              onChange={(value) => update(index, { startMs: value })}
            />
            <TextField
              label={t('administration.transcriptEditor.end', { line: index + 1 })}
              value={line.endMs}
              onChange={(value) => update(index, { endMs: value })}
            />
            <TextArea
              label={t('administration.transcriptEditor.text', { line: index + 1 })}
              value={line.text}
              onChange={(value) => update(index, { text: value })}
            />
            <fieldset>
              <legend>{t('administration.transcriptEditor.documents', { line: index + 1 })}</legend>
              {documents.map((document) => (
                <label key={document.id}>
                  <input
                    type="checkbox"
                    checked={line.documentIds.includes(document.id)}
                    onChange={(event) =>
                      update(index, {
                        documentIds: event.target.checked
                          ? [...line.documentIds, document.id]
                          : line.documentIds.filter((id) => id !== document.id),
                      })
                    }
                  />
                  {document.label}
                </label>
              ))}
            </fieldset>
            <Button
              type="button"
              variant="secondary"
              onPress={() => setLines((current) => current.filter((_, at) => at !== index))}
            >
              {t('administration.transcriptEditor.remove', { line: index + 1 })}
            </Button>
          </li>
        ))}
      </ol>

      <Button
        type="button"
        variant="secondary"
        onPress={() =>
          setLines((current) => [
            ...current,
            {
              id: crypto.randomUUID(),
              startMs: String(current.at(-1)?.endMs ?? '0'),
              endMs: String(Number(current.at(-1)?.endMs ?? '0') + 1000),
              text: '',
              documentIds: [],
            },
          ])
        }
      >
        {t('administration.transcriptEditor.add')}
      </Button>

      <Button
        type="button"
        isDisabled={status === 'saving'}
        onPress={async () => {
          const segmentsToSave = parsed();
          if (segmentsToSave === null) {
            setFailure('invalid_request');
            return;
          }
          setStatus('saving');
          setFailure(null);
          const error = await actions.save(language, expectedUpdatedAt, segmentsToSave);
          setStatus(error === null ? 'saved' : 'idle');
          setFailure(error);
        }}
      >
        {t('administration.transcriptEditor.save')}
      </Button>

      {failure === null ? null : (
        <output className="transcript-editor-error">
          {t(`administration.transcriptEditor.errors.${failure}`, {
            defaultValue: t('administration.transcriptEditor.errors.unavailable'),
          })}
        </output>
      )}
      {status === 'saved' && failure === null ? (
        <output>{t('administration.transcriptEditor.saved')}</output>
      ) : null}
    </section>
  );
}
