import type { QueryClient } from '@tanstack/react-query';
import type { Av1ProbeResult } from '../../domain/playback/playback';

const av1ProbeKey = ['playback', 'av1-probe'] as const;
const av1ContentType = 'video/mp4; codecs="av01.0.05M.08"';

interface Av1ProbeState {
  readonly result: Av1ProbeResult;
  readonly reported: boolean;
}

function probeAv1Support(): Av1ProbeResult {
  try {
    if (typeof MediaSource === 'undefined' || typeof MediaSource.isTypeSupported !== 'function') {
      return 'probe_failed';
    }
    return MediaSource.isTypeSupported(av1ContentType) ? 'supported' : 'unsupported';
  } catch {
    return 'probe_failed';
  }
}

export function primeAv1Probe(queryClient: QueryClient): void {
  if (typeof window === 'undefined' || queryClient.getQueryData(av1ProbeKey) !== undefined) return;
  queryClient.setQueryDefaults(av1ProbeKey, { gcTime: Number.POSITIVE_INFINITY });
  queryClient.setQueryData<Av1ProbeState>(av1ProbeKey, {
    result: probeAv1Support(),
    reported: false,
  });
}

export function takeAv1ProbeResult(queryClient: QueryClient): Av1ProbeResult | undefined {
  const state = queryClient.getQueryData<Av1ProbeState>(av1ProbeKey);
  if (state === undefined || state.reported) return undefined;
  queryClient.setQueryData<Av1ProbeState>(av1ProbeKey, { ...state, reported: true });
  return state.result;
}
