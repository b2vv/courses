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
  TokenInput,
} from '../../domain/identity/identity';
import type { SessionRepository } from '../../ports/SessionRepository';

export function createSession(repository: SessionRepository) {
  return {
    logIn(
      input: CredentialsInput,
      signal?: AbortSignal,
    ): Promise<SessionResult<AuthenticatedUser>> {
      const credentials = makeCredentials(input);
      return credentials.ok
        ? repository.logIn(credentials.value, signal)
        : Promise.resolve(credentials);
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
