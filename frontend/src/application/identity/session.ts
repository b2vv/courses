import { makeCredentials } from '../../domain/identity/identity';
import type {
  AuthenticatedUser,
  CredentialsInput,
  EmailChangeInput,
  EmailInput,
  PasswordChangeInput,
  PasswordResetInput,
  RegisterInput,
  SessionResult,
  LoginResult,
  TwoFactorEnrollment,
  RecoveryCodes,
  TokenInput,
} from '../../domain/identity/identity';
import type { SessionRepository } from '../../ports/SessionRepository';

export function createSession(repository: SessionRepository) {
  return {
    logIn(input: CredentialsInput, signal?: AbortSignal): Promise<SessionResult<LoginResult>> {
      const credentials = makeCredentials(input);
      return credentials.ok
        ? repository.logIn(credentials.value, signal)
        : Promise.resolve(credentials);
    },
    enrollTwoFactor(signal?: AbortSignal): Promise<SessionResult<TwoFactorEnrollment>> {
      return repository.enrollTwoFactor(signal);
    },
    confirmTwoFactor(code: string, signal?: AbortSignal): Promise<SessionResult<RecoveryCodes>> {
      return repository.confirmTwoFactor(code, signal);
    },
    completeTwoFactor(
      pendingId: string,
      code: string,
      signal?: AbortSignal,
    ): Promise<SessionResult<null>> {
      return repository.completeTwoFactor(pendingId, code, signal);
    },
    disableTwoFactor(
      password: string,
      code: string,
      signal?: AbortSignal,
    ): Promise<SessionResult<null>> {
      return repository.disableTwoFactor(password, code, signal);
    },
    regenerateRecoveryCodes(
      code: string,
      signal?: AbortSignal,
    ): Promise<SessionResult<RecoveryCodes>> {
      return repository.regenerateRecoveryCodes(code, signal);
    },
    register(
      input: RegisterInput,
      signal?: AbortSignal,
    ): Promise<SessionResult<AuthenticatedUser>> {
      return repository.register(input, signal);
    },
    verify(input: TokenInput, signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.verifyEmail(input, signal);
    },
    resend(input: Partial<EmailInput>, signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.resendVerification(input, signal);
    },
    forgot(input: EmailInput, signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.forgotPassword(input, signal);
    },
    reset(input: PasswordResetInput, signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.resetPassword(input, signal);
    },
    changePassword(input: PasswordChangeInput, signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.changePassword(input, signal);
    },
    requestEmailChange(
      input: EmailChangeInput,
      signal?: AbortSignal,
    ): Promise<SessionResult<null>> {
      return repository.requestEmailChange(input, signal);
    },
    confirmEmailChange(input: TokenInput, signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.confirmEmailChange(input, signal);
    },
    current(signal?: AbortSignal): Promise<SessionResult<AuthenticatedUser | null>> {
      return repository.currentUser(signal);
    },
    logOut(signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.logOut(signal);
    },
  };
}
