import type { CaptionSelection } from '../ports/PlaybackMedia';

export type AudioFragment = { readonly start: number; readonly end: number };
type Attempt = {
  appended: boolean;
  failed: boolean;
  completed: boolean;
  language: string | undefined;
};
export type AudioRanges = {
  readonly length: number;
  start(index: number): number;
  end(index: number): number;
};

/** No hls.js state: fragment objects identify logical attempts, never physical append executions. */
export function createAudioCoverage() {
  type Span = AudioFragment & { readonly attempt: Attempt };
  let spans: Span[] = [];
  const attempts = new Map<AudioFragment, Attempt>();
  // An internal hls listener may emit ERROR before our outer APPENDING listener runs.
  const dispatchErrors = new Set<AudioFragment>();

  function forget(frag: AudioFragment) {
    spans = spans.flatMap((label) => {
      if (label.end <= frag.start || label.start >= frag.end) return [label];
      const kept: Span[] = [];
      if (label.start < frag.start) kept.push({ ...label, end: frag.start });
      if (label.end > frag.end) kept.push({ ...label, start: frag.end });
      return kept;
    });
  }

  return {
    appending(frag: AudioFragment) {
      forget(frag);
      const attempt: Attempt = {
        appended: false,
        failed: dispatchErrors.has(frag),
        completed: false,
        language: undefined,
      };
      attempts.set(frag, attempt);
      // Span ownership prevents a late completion from relabelling a newer overlapping append.
      spans.push({ start: frag.start, end: frag.end, attempt });
    },
    appended(frag: AudioFragment) {
      const attempt = attempts.get(frag);
      if (!attempt) return;
      attempt.appended = true;
    },
    buffered(frag: AudioFragment, language: string | undefined) {
      const attempt = attempts.get(frag);
      // Empty/gap/no-media fragments never change an existing label.
      if (!attempt) return;
      attempt.completed = true;
      attempt.language = language;
    },
    error(frag: AudioFragment) {
      dispatchErrors.add(frag);
      const attempt = attempts.get(frag);
      if (!attempt) return;
      attempt.failed = true;
    },
    endDispatch() {
      dispatchErrors.clear();
      for (const [frag, attempt] of attempts) {
        if (attempt.completed) attempts.delete(frag);
      }
    },
    reset() {
      spans = [];
      attempts.clear();
      dispatchErrors.clear();
    },
    selection(
      selected: string | null,
      currentTime: number,
      readBuffered: (() => AudioRanges) | undefined,
    ): CaptionSelection {
      if (selected === null || readBuffered === undefined) return { state: 'pending' };
      try {
        const buffered = readBuffered();
        const attempt = spans.find(
          (range) => range.start <= currentTime && currentTime < range.end,
        )?.attempt;
        if (
          !attempt?.completed ||
          !attempt.appended ||
          attempt.failed ||
          attempt.language?.toLowerCase() !== selected.toLowerCase()
        )
          return { state: 'pending' };
        for (let i = 0; i < buffered.length; i += 1) {
          if (buffered.start(i) <= currentTime && currentTime < buffered.end(i))
            return { state: 'confirmed', language: selected };
        }
      } catch {
        // Removed SourceBuffers throw InvalidStateError. A failed read proves no coverage.
      }
      return { state: 'pending' };
    },
  };
}
