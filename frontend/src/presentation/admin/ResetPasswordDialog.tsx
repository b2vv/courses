import { useState } from 'react';
import { useController, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { AdminUserQueryError } from '../../app/composition/admin-user-queries';
import { validAdministrationPassword } from '../../domain/identity/administration';
import type { AdminUser } from '../../domain/identity/administration';
import { Button, Dialog, TextField } from '../ui';

type ResetPasswordDialogProps = {
  user: AdminUser;
  isSelf: boolean;
  onReset: (password: string) => Promise<void>;
  onAuthenticationRequired: () => void;
};

type PasswordForm = { password: string };

export function ResetPasswordDialog({
  user,
  isSelf,
  onReset,
  onAuthenticationRequired,
}: ResetPasswordDialogProps) {
  const { t } = useTranslation();
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<'status' | 'alert'>('status');
  const form = useForm<PasswordForm>({ defaultValues: { password: '' } });
  const { field, fieldState } = useController({
    control: form.control,
    name: 'password',
    rules: {
      validate: (password) => validAdministrationPassword(password) || t('adminUsers.passwordHint'),
    },
  });

  async function reset({ password }: PasswordForm) {
    setMessage(undefined);
    try {
      await onReset(password);
      form.reset();
      setMessageKind('status');
      setMessage(t(isSelf ? 'adminUsers.selfReset' : 'adminUsers.passwordReset'));
      if (isSelf) onAuthenticationRequired();
    } catch (error) {
      const code = error instanceof AdminUserQueryError ? error.code : undefined;
      setMessageKind('alert');
      setMessage(administrationErrorMessage(code, t));
      if (code === 'authentication_required') onAuthenticationRequired();
    }
  }

  return (
    <Dialog
      triggerLabel={t('adminUsers.resetPasswordShort')}
      triggerAriaLabel={t('adminUsers.resetPassword', { name: user.displayName })}
      title={t('adminUsers.resetTitle', { name: user.displayName })}
      closeLabel={t('common.close')}
    >
      {message && (
        <div className="admin-editor-message" role={messageKind}>
          {message}
        </div>
      )}
      <form
        className="admin-create-form"
        onSubmit={form.handleSubmit(reset)}
        noValidate
        aria-busy={form.formState.isSubmitting}
      >
        <TextField
          type="password"
          label={t('adminUsers.newPassword')}
          description={t('adminUsers.passwordHint')}
          value={field.value}
          name={field.name}
          onChange={field.onChange}
          onBlur={field.onBlur}
          isRequired
          isInvalid={fieldState.invalid}
          {...(fieldState.error?.message === undefined
            ? {}
            : { errorMessage: fieldState.error.message })}
        />
        <Button type="submit" isDisabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting
            ? t('adminUsers.savingPassword')
            : t('adminUsers.savePassword')}
        </Button>
      </form>
    </Dialog>
  );
}

export function administrationErrorMessage(
  code: AdminUserQueryError['code'] | undefined,
  t: ReturnType<typeof useTranslation>['t'],
) {
  switch (code) {
    case 'authentication_required':
      return t('adminUsers.error.authenticationRequired');
    case 'origin_forbidden':
      return t('adminUsers.error.originForbidden');
    case 'forbidden':
      return t('adminUsers.error.forbidden');
    case 'cannot_target_self':
      return t('adminUsers.error.cannotTargetSelf');
    case 'admin_immutable':
      return t('adminUsers.error.adminImmutable');
    case 'last_admin':
      return t('adminUsers.error.lastAdmin');
    case 'user_not_found':
      return t('adminUsers.error.userNotFound');
    case 'invalid_request':
      return t('adminUsers.error.invalidRequest');
    case 'request_cancelled':
      return t('adminUsers.error.requestCancelled');
    case 'request_timeout':
      return t('adminUsers.error.requestTimeout');
    case 'dependency_unavailable':
    case 'invalid_response':
    case undefined:
      return t('adminUsers.error.unavailable');
  }
}
