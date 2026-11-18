import { useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { TranscriptDocument, TranscriptErrorCode } from '../../transcript';
import { Button, TextField } from '../../../presentation/ui';

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
  const formId = useId();
  const [isOpen, setIsOpen] = useState(false);
  const [submitted, setSubmitted] = useState(false);
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
    <section className="document-library ui-card" aria-labelledby="document-library-heading">
      <div className="card-heading">
        <h2 id="document-library-heading">{t('administration.documentLibrary.heading')}</h2>
      </div>

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

      <Button
        type="button"
        variant="secondary"
        aria-expanded={isOpen}
        aria-controls={formId}
        onPress={() => setIsOpen((open) => !open)}
      >
        {t('administration.documentLibrary.add')}
      </Button>
      {isOpen && (
        <div className="document-library-add" id={formId}>
          <TextField
            label={t('administration.documentLibrary.label')}
            value={label}
            onChange={setLabel}
            isRequired
            isInvalid={submitted && label.trim() === ''}
            errorMessage={t('administration.validation.required')}
          />
          <TextField
            label={t('administration.documentLibrary.url')}
            value={url}
            onChange={setUrl}
            isRequired
            isInvalid={submitted && url.trim() === ''}
            errorMessage={t('administration.validation.required')}
          />
          <Button
            type="button"
            isDisabled={pending}
            onPress={async () => {
              setSubmitted(true);
              if (label.trim() === '' || url.trim() === '') return;
              setPending(true);
              const error = await actions.add({ id: crypto.randomUUID(), label, url });
              setFailure(error);
              setPending(false);
              if (error === null) {
                setSubmitted(false);
                setLabel('');
                setUrl('');
              }
            }}
          >
            {t('administration.documentLibrary.save')}
          </Button>
        </div>
      )}

      {failure === null ? null : (
        <output className="document-library-error">{message(failure)}</output>
      )}
    </section>
  );
}
