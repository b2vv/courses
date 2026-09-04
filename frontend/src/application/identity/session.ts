import { makeCredentials } from '../../domain/identity/identity';
import type {
  AuthenticatedUser,
  CredentialsInput,
  SessionResult,
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
    current(signal?: AbortSignal): Promise<SessionResult<AuthenticatedUser | null>> {
      return repository.currentUser(signal);
    },
    logOut(signal?: AbortSignal): Promise<SessionResult<null>> {
      return repository.logOut(signal);
    },
  };
}
