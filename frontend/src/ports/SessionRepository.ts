import type {
  AuthenticatedUser,
  Credentials,
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
} from '../domain/identity/identity';

export interface SessionRepository {
  /** Browser-only: the server requires an Origin header a server-side call cannot send. */
  logIn(credentials: Credentials, signal?: AbortSignal): Promise<SessionResult<LoginResult>>;
  enrollTwoFactor(signal?: AbortSignal): Promise<SessionResult<TwoFactorEnrollment>>;
  confirmTwoFactor(code: string, signal?: AbortSignal): Promise<SessionResult<RecoveryCodes>>;
  completeTwoFactor(
    pendingId: string,
    code: string,
    signal?: AbortSignal,
  ): Promise<SessionResult<null>>;
  disableTwoFactor(
    password: string,
    code: string,
    signal?: AbortSignal,
  ): Promise<SessionResult<null>>;
  regenerateRecoveryCodes(
    code: string,
    signal?: AbortSignal,
  ): Promise<SessionResult<RecoveryCodes>>;
  register(input: RegisterInput, signal?: AbortSignal): Promise<SessionResult<AuthenticatedUser>>;
  verifyEmail(input: TokenInput, signal?: AbortSignal): Promise<SessionResult<null>>;
  resendVerification(
    input: Partial<EmailInput>,
    signal?: AbortSignal,
  ): Promise<SessionResult<null>>;
  forgotPassword(input: EmailInput, signal?: AbortSignal): Promise<SessionResult<null>>;
  resetPassword(input: PasswordResetInput, signal?: AbortSignal): Promise<SessionResult<null>>;
  changePassword(input: PasswordChangeInput, signal?: AbortSignal): Promise<SessionResult<null>>;
  requestEmailChange(input: EmailChangeInput, signal?: AbortSignal): Promise<SessionResult<null>>;
  confirmEmailChange(input: TokenInput, signal?: AbortSignal): Promise<SessionResult<null>>;
  /** Anonymous is a success carrying null, not an error; only transport faults fail. */
  currentUser(signal?: AbortSignal): Promise<SessionResult<AuthenticatedUser | null>>;
  /** Browser-only, and idempotent: revoking an absent session still succeeds. */
  logOut(signal?: AbortSignal): Promise<SessionResult<null>>;
}
