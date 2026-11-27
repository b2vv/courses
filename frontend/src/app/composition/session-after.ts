import { onlineManager, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState, useSyncExternalStore } from 'react';
import type { AuthenticatedUser } from '../../modules/identity';
import type { IdentityDependencies } from './identity';

export type SessionRead =
  | { readonly status: 'pending' }
  | { readonly status: 'done'; readonly user: AuthenticatedUser | null }
  | { readonly status: 'failed' };

/**
 * One session read of this screen's own, started when `trigger` names a new result
 * (design-pass ACCT V2). Background refetches of the shared session query do not change it, so
 * a form it revealed stays mounted with its focus. A response from a read already in flight is
 * taken as is (plan §18 Л1).
 */
export function useSessionReadAfter(
  identityDependencies: IdentityDependencies,
  trigger: string | undefined,
): SessionRead | undefined {
  const client = useQueryClient();
  const [read, setRead] = useState<{ readonly trigger: string; readonly value: SessionRead }>();
  useEffect(() => {
    if (trigger === undefined) return;
    let live = true;
    client.fetchQuery(identityDependencies.current()).then(
      (user) => {
        if (live) setRead({ trigger, value: { status: 'done', user } });
      },
      () => {
        if (live) setRead({ trigger, value: { status: 'failed' } });
      },
    );
    return () => {
      live = false;
    };
  }, [trigger, client, identityDependencies]);
  if (trigger === undefined) return undefined;
  return read?.trigger === trigger ? read.value : { status: 'pending' };
}

/** Whether TanStack Query thinks the network is up; paused requests wait for it (plan §18 Л6). */
export function useOnline(): boolean {
  return useSyncExternalStore(
    (listener) => onlineManager.subscribe(listener),
    () => onlineManager.isOnline(),
    () => true,
  );
}
