const prefix = 'transcript-recovery-dismissed:';

/**
 * Dismissal lives in `sessionStorage`, so it dies with the tab. That is the whole point: an
 * author who waves the prompt away today should still be offered work recovered from a crash
 * tomorrow, and a dismissal persisted on disk would silently disarm the mechanism for good.
 *
 * Guarded exactly like `modules/media/adapters/tus-client.ts`: absent during SSR, and able to throw in a
 * private window, so both are treated as "no memory of a dismissal".
 */
function storage(): Storage | undefined {
  try {
    return typeof window === 'undefined' ? undefined : window.sessionStorage;
  } catch {
    return undefined;
  }
}

export function wasDismissed(key: string): boolean {
  try {
    return storage()?.getItem(`${prefix}${key}`) === '1';
  } catch {
    return false;
  }
}

export function rememberDismissal(key: string): void {
  try {
    storage()?.setItem(`${prefix}${key}`, '1');
  } catch {
    // A refused write costs the author one extra prompt, never their text.
  }
}

export function forgetDismissal(key: string): void {
  try {
    storage()?.removeItem(`${prefix}${key}`);
  } catch {
    // As above.
  }
}
