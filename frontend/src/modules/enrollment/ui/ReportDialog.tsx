import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { EnrollmentErrorCode } from '../domain/enrollment';
import { Button, Dialog, TextArea } from '../../../presentation/ui';

const maximumReasonLength = 500;

function reportErrorKey(
  code: EnrollmentErrorCode,
):
  | 'report_already_open'
  | 'report_own_person'
  | 'report_invalid'
  | 'report_forbidden'
  | 'email_not_verified'
  | 'course_not_found'
  | 'generic' {
  switch (code) {
    case 'report_already_open':
    case 'report_own_person':
    case 'report_invalid':
    case 'report_forbidden':
    case 'email_not_verified':
    case 'course_not_found':
      return code;
    default:
      return 'generic';
  }
}

function reportFailureCode(failure: unknown): EnrollmentErrorCode {
  if (!(failure instanceof Error)) return 'dependency_unavailable';
  switch (failure.message) {
    case 'report_already_open':
    case 'report_own_person':
    case 'report_invalid':
    case 'report_forbidden':
    case 'email_not_verified':
    case 'course_not_found':
      return failure.message;
    default:
      return 'dependency_unavailable';
  }
}

export function ReportDialog({
  triggerLabel,
  onReport,
  onSent,
}: {
  readonly triggerLabel: string;
  readonly onReport: (reason: string) => Promise<void>;
  readonly onSent: () => void;
}) {
  const { t } = useTranslation();
  const [reason, setReason] = useState('');
  const [touched, setTouched] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<EnrollmentErrorCode>();
  const valid = reason.trim().length >= 1 && reason.trim().length <= maximumReasonLength;

  return (
    <Dialog
      triggerLabel={triggerLabel}
      title={triggerLabel}
      closeLabel={t('enrollment.report.cancel')}
    >
      {(close) => (
        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault();
            setTouched(true);
            if (!valid) return;
            setPending(true);
            setError(undefined);
            void onReport(reason.trim())
              .then(() => {
                close();
                onSent();
              })
              .catch((failure: unknown) => setError(reportFailureCode(failure)))
              .finally(() => setPending(false));
          }}
        >
          <TextArea
            label={t('enrollment.report.reasonLabel')}
            value={reason}
            onChange={setReason}
            maxLength={maximumReasonLength}
            isRequired
            isDisabled={pending}
            isInvalid={touched && !valid}
            {...(touched && !valid
              ? { errorMessage: t('enrollment.report.errors.report_invalid') }
              : {})}
          />
          <Button type="submit" isPending={pending}>
            {t('enrollment.report.submit')}
          </Button>
          {error === undefined ? null : (
            <p role="alert">{t(`enrollment.report.errors.${reportErrorKey(error)}`)}</p>
          )}
        </form>
      )}
    </Dialog>
  );
}
