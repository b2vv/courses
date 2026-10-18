import { useState } from 'react';
import type { Key } from 'react-aria-components';
import { useTranslation } from 'react-i18next';
import { grantReason, grantReasonMaximum, type GrantErrorCode } from '../../modules/purchase';
import { Button, ComboBox, TextArea } from '../ui';
import type { ChoiceOption } from '../ui';

type CourseGrantPanelProps = {
  readonly users: readonly ChoiceOption[];
  readonly isUserSearchPending: boolean;
  readonly onSearch: (text: string) => void;
  /** Resolves to null on success, or the code the server named. */
  readonly onGrant: (userId: string, reason: string) => Promise<GrantErrorCode | null>;
};

function errorKey(code: GrantErrorCode) {
  switch (code) {
    case 'course_already_owned':
      return 'administration.grant.alreadyOwned';
    case 'grant_forbidden':
      return 'administration.grant.forbidden';
    case 'invalid_request':
      return 'administration.grant.invalid';
    default:
      return 'administration.error.unavailable';
  }
}

/** Admin-only: the server refuses anyone else with grant_forbidden, whatever this renders. */
export function CourseGrantPanel({
  users,
  isUserSearchPending,
  onSearch,
  onGrant,
}: CourseGrantPanelProps) {
  const { t } = useTranslation();
  // The chosen user is kept as chosen: once picked, the input shows the label, and the next
  // search (for that label) no longer has to contain the user for the grant to go out.
  const [chosen, setChosen] = useState<ChoiceOption | null>(null);
  const userId = chosen?.id ?? null;
  const [reason, setReason] = useState('');
  const [touched, setTouched] = useState(false);
  const [pending, setPending] = useState(false);
  const [outcome, setOutcome] = useState<
    { kind: 'granted'; user: string } | { kind: 'error'; code: GrantErrorCode }
  >();
  const trimmed = grantReason(reason);

  async function submit() {
    setTouched(true);
    if (chosen === null || trimmed === null) return;
    setPending(true);
    setOutcome(undefined);
    const code = await onGrant(chosen.id, trimmed);
    setPending(false);
    if (code === null) {
      setOutcome({ kind: 'granted', user: chosen.label });
      setReason('');
      setTouched(false);
    } else setOutcome({ kind: 'error', code });
  }

  return (
    <section className="admin-lifecycle admin-grant ui-card" aria-labelledby="course-grant-heading">
      <div>
        <p className="eyebrow">{t('administration.grant.eyebrow')}</p>
        <h2 id="course-grant-heading">{t('administration.grant.heading')}</h2>
        <p className="admin-price-note">{t('administration.grant.hint')}</p>
      </div>
      <form
        className="admin-grant-form"
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          void submit();
        }}
      >
        <ComboBox
          label={t('administration.grant.user')}
          options={users}
          filtering="server"
          isLoading={isUserSearchPending}
          loadingLabel={t('administration.grant.searching')}
          selectedKey={userId}
          onSelectionChange={(key: Key | null) =>
            setChosen(users.find((user) => user.id === key) ?? null)
          }
          onInputChange={(text) => {
            if (text !== chosen?.label) onSearch(text);
          }}
          isDisabled={pending}
          isInvalid={touched && userId === null}
          {...(touched && userId === null
            ? { errorMessage: t('administration.grant.userRequired') }
            : { description: t('administration.grant.userHint') })}
        />
        <TextArea
          label={t('administration.grant.reason')}
          value={reason}
          onChange={setReason}
          maxLength={grantReasonMaximum}
          isDisabled={pending}
          isInvalid={touched && trimmed === null}
          {...(touched && trimmed === null
            ? { errorMessage: t('administration.grant.reasonRequired') }
            : { description: t('administration.grant.reasonHint') })}
        />
        <Button type="submit" isDisabled={pending}>
          {pending ? t('administration.saving') : t('administration.grant.submit')}
        </Button>
      </form>
      {outcome?.kind === 'granted' && (
        <output className="admin-editor-message">
          {t('administration.grant.granted', { user: outcome.user })}
        </output>
      )}
      {outcome?.kind === 'error' && (
        <p className="admin-editor-message" role="alert">
          {t(errorKey(outcome.code))}
        </p>
      )}
    </section>
  );
}
