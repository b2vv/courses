// Public surface of the `identity` feature module. Everything another module, the composition
// roots, the router or the routes may use is re-exported here; nothing outside this folder imports
// the module's internals directly (enforced by scripts/check-architecture.mjs — no deep
// cross-module import).

// Domain: authentication/session types, roles, access predicates and pure helpers.
export * from './domain/identity';

// Ports: the I/O contract this feature depends on.
export type { SessionRepository } from './ports/SessionRepository';

// Application: the session query/mutation factory.
export { createSession } from './application/session';

// Adapters: the HTTP implementation wired by the composition roots.
export { createApiSessionRepository } from './adapters/session-repository';

// UI.
export { AccountSettingsView } from './ui/AccountSettingsView';
export { ConfirmEmailChangeView } from './ui/ConfirmEmailChangeView';
export { EmailVerificationBanner } from './ui/EmailVerificationBanner';
export { IdentityHeader } from './ui/IdentityHeader';
export { LoginView } from './ui/LoginView';
export { PasswordForgotView } from './ui/PasswordForgotView';
export { PasswordResetView } from './ui/PasswordResetView';
export { RegisterView } from './ui/RegisterView';
export { VerifyEmailView } from './ui/VerifyEmailView';
export {
  LearnAccessDenied,
  LearnRouteFailure,
  ProtectedRouteFailure,
  ProtectedRoutePending,
} from './ui/ProtectedRouteStates';
