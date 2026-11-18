import type { CaptionSelection } from '../ports/PlaybackMedia';

export type AudioFragment = { readonly start: number; readonly end: number };
type Attempt = {
  appended: boolean;
  failed: boolean;
  completed: boolean;
  language: string | undefined;
  /** Languages overwritten directly or inherited through an uncertain overlapping attempt. */
  replaced: readonly string[];
  /** Media time before which this attempt's audio may not be audible yet. */
  settleUntil: number | undefined;
};

/**
 * Replacing audio of ANOTHER language just ahead of the playhead does not change what is heard at
 * once: Chromium keeps already decoded samples of the old track in its decode/render pipeline. MS4 E4b
 * (2026-11-18, Chromium on Kiri, audio tap of the real player) heard the old track for 0.62-0.93 s of
 * media time after the replacing append completed. The selection stays pending for this window after
 * such a replacement, so a caption never shows over the old language; 1.5 s keeps a margin over the
 * measured lag. Same-language appends (ordinary playback) are not affected.
 */
export const REPLACEMENT_SETTLE_SECONDS = 1.5;
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
      const replaced = new Set<string>();
      let settleUntil: number | undefined;
      for (const label of spans) {
        if (label.end <= frag.start || label.start >= frag.end) continue;
        const previous = label.attempt;
        const languages =
          previous.completed && previous.appended && !previous.failed
            ? previous.language === undefined
              ? []
              : [previous.language.toLowerCase()]
            : previous.replaced;
        for (const language of languages) replaced.add(language);
        if (previous.settleUntil !== undefined)
          settleUntil = Math.max(settleUntil ?? -Infinity, previous.settleUntil);
      }
      forget(frag);
      const attempt: Attempt = {
        appended: false,
        failed: dispatchErrors.has(frag),
        completed: false,
        language: undefined,
        replaced: [...replaced],
        settleUntil,
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
    buffered(frag: AudioFragment, language: string | undefined, currentTime: number) {
      const attempt = attempts.get(frag);
      // Empty/gap/no-media fragments never change an existing label.
      if (!attempt) return;
      attempt.completed = true;
      attempt.language = language;
      const replacesAnother = attempt.replaced.some(
        (previous) => previous !== language?.toLowerCase(),
      );
      if (
        replacesAnother &&
        frag.end > currentTime &&
        frag.start < currentTime + REPLACEMENT_SETTLE_SECONDS
      )
        attempt.settleUntil = Math.max(
          attempt.settleUntil ?? -Infinity,
          currentTime + REPLACEMENT_SETTLE_SECONDS,
        );
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
          (attempt.settleUntil !== undefined && currentTime < attempt.settleUntil) ||
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
