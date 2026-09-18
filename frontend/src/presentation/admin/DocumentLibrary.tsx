import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { TranscriptDocument, TranscriptErrorCode } from '../../domain/transcript/transcript';
import { Button, TextField } from '../ui';

export type DocumentLibraryActions = {
  add(document: TranscriptDocument): Promise<TranscriptErrorCode | null>;
  remove(documentId: string): Promise<TranscriptErrorCode | null>;
};

type DocumentLibraryProps = {
  readonly documents: readonly TranscriptDocument[];
  readonly actions: DocumentLibraryActions;
};

export function DocumentLibrary({ documents, actions }: DocumentLibraryProps) {
  const { t } = useTranslation();
  const [label, setLabel] = useState('');
  const [url, setUrl] = useState('');
  const [failure, setFailure] = useState<TranscriptErrorCode | null>(null);
  const [pending, setPending] = useState(false);

  const message = (code: TranscriptErrorCode) =>
    t(`administration.documentLibrary.errors.${code}`, {
      // A code outside the named set is contract drift; showing it verbatim would dress a drift
      // up as a state the author is expected to understand.
      defaultValue: t('administration.documentLibrary.errors.unavailable'),
    });

  return (
    <section className="document-library" aria-labelledby="document-library-heading">
      <h3 id="document-library-heading">{t('administration.documentLibrary.heading')}</h3>

      {documents.length === 0 ? <p>{t('administration.documentLibrary.empty')}</p> : null}

      <ul className="document-library-items">
        {documents.map((document) => (
          <li key={document.id}>
            <a href={document.url} target="_blank" rel="noreferrer noopener">
              {document.label}
            </a>
            <Button
              type="button"
              variant="secondary"
              isDisabled={pending}
              onPress={async () => {
                setPending(true);
                setFailure(await actions.remove(document.id));
                setPending(false);
              }}
            >
              {t('administration.documentLibrary.remove', { label: document.label })}
            </Button>
          </li>
        ))}
      </ul>

      <div className="document-library-add">
        <TextField
          label={t('administration.documentLibrary.label')}
          value={label}
          onChange={setLabel}
        />
        <TextField label={t('administration.documentLibrary.url')} value={url} onChange={setUrl} />
        <Button
          type="button"
          isDisabled={pending || label.trim() === '' || url.trim() === ''}
          onPress={async () => {
            setPending(true);
            const error = await actions.add({ id: crypto.randomUUID(), label, url });
            setFailure(error);
            setPending(false);
            if (error === null) {
              setLabel('');
              setUrl('');
            }
          }}
        >
          {t('administration.documentLibrary.add')}
        </Button>
      </div>

      {failure === null ? null : (
        <output className="document-library-error">{message(failure)}</output>
      )}
    </section>
  );
}
