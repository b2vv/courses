import { useCallback, useEffect, useRef, useState } from 'react';
import {
  forgetDismissal,
  rememberDismissal,
  wasDismissed,
} from '../../../adapters/storage/recovery-dismissal';
import {
  captureDirective,
  recoveryOffer,
  type RecoveryOffer,
  type TranscriptBaseline,
  type TranscriptDraftLine,
} from '../application/recovery';
import type { TranscriptHistoryStore } from '../ports/TranscriptHistoryStore';

/**
 * Idle delay before an unprompted capture, matching the interval the mechanism this is modelled
 * on has been running with. Capturing per keystroke was never the design: it would write
 * thousands of times per lesson to save the same text.
 */
export const idleCaptureMs = 90_000;

export type TranscriptRecovery = {
  readonly offer: RecoveryOffer;
  /** Restart the idle timer. Called on every edit; writes nothing by itself. */
  readonly noteChange: (lines: readonly TranscriptDraftLine[]) => void;
  /** Capture now, because the author left a field. */
  readonly captureNow: (lines: readonly TranscriptDraftLine[]) => void;
  /** Close the prompt for this mount, after the snapshot has been applied. */
  readonly accept: () => void;
  /** Close it and remember, so a remount does not ask again in this tab. */
  readonly dismiss: () => void;
  /** Only after a save the server accepted: drop the local copy and the dismissal with it. */
  readonly finalize: () => void;
};

export type TranscriptRecoveryOptions = {
  readonly historyKey: string;
  readonly store: TranscriptHistoryStore;
  readonly baseline: TranscriptBaseline;
};

/**
 * The editor holds its lines in component state, so today a reload loses them without a word.
 * This keeps a best-effort local copy and OFFERS it back - it never restores on its own, because
 * the store it reads may have been evicted, and because restoring silently over text the author
 * deliberately cleared would be worse than losing it.
 *
 * Every store call is caught. A refused or full store costs the author a recovery offer; it may
 * never cost them the save they were actually making.
 */
export function useTranscriptRecovery(options: TranscriptRecoveryOptions): TranscriptRecovery {
  const { historyKey, store, baseline } = options;

  const [offer, setOffer] = useState<RecoveryOffer>({ kind: 'none' });
  const idleTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Props are read through a ref so that the capture callbacks stay stable: a callback that
  // changed on every keystroke would restart the idle timer on every keystroke, and the idle
  // capture would then never fire while the author was typing steadily.
  //
  // The ref is filled from an effect rather than during render. Effects run before any event the
  // author can trigger, so every capture still sees the current props - and a ref written during
  // render is a value React is free to discard.
  const latest = useRef({ historyKey, store, baseline });
  useEffect(() => {
    latest.current = { historyKey, store, baseline };
  }, [historyKey, store, baseline]);

  useEffect(() => {
    // Read only after mount. IndexedDB does not exist during SSR, and the read is asynchronous,
    // so there is nothing for the first render to show anyway.
    let cancelled = false;
    void (async () => {
      try {
        const stored = await store.read(historyKey);
        if (cancelled) return;
        setOffer(recoveryOffer(stored, latest.current.baseline, wasDismissed(historyKey)));
      } catch {
        // No local copy is a normal outcome, not an error worth showing.
        if (!cancelled) setOffer({ kind: 'none' });
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [historyKey, store]);

  useEffect(
    () => () => {
      if (idleTimer.current !== null) clearTimeout(idleTimer.current);
    },
    [],
  );

  const capture = useCallback((lines: readonly TranscriptDraftLine[]) => {
    const current = latest.current;
    const directive = captureDirective(lines, current.baseline);
    const write =
      directive.kind === 'clear'
        ? current.store.clear(current.historyKey)
        : current.store.write({
            key: current.historyKey,
            capturedAt: Date.now(),
            ...(current.baseline.language === undefined
              ? {}
              : { language: current.baseline.language }),
            expectedUpdatedAt: current.baseline.expectedUpdatedAt,
            ...(current.baseline.draftUpdatedAt === undefined
              ? {}
              : { draftUpdatedAt: current.baseline.draftUpdatedAt }),
            lines,
          });
    write.catch(() => {
      // A quota breach is the realistic one here, because the domain permits 5000 segments of
      // 1000 runes. It is swallowed on purpose: the author is typing, not saving.
    });
  }, []);

  const captureNow = useCallback(
    (lines: readonly TranscriptDraftLine[]) => {
      if (idleTimer.current !== null) {
        clearTimeout(idleTimer.current);
        idleTimer.current = null;
      }
      capture(lines);
    },
    [capture],
  );

  const noteChange = useCallback(
    (lines: readonly TranscriptDraftLine[]) => {
      if (idleTimer.current !== null) clearTimeout(idleTimer.current);
      idleTimer.current = setTimeout(() => {
        idleTimer.current = null;
        capture(lines);
      }, idleCaptureMs);
    },
    [capture],
  );

  const accept = useCallback(() => {
    setOffer({ kind: 'none' });
  }, []);

  const dismiss = useCallback(() => {
    rememberDismissal(latest.current.historyKey);
    setOffer({ kind: 'none' });
  }, []);

  const finalize = useCallback(() => {
    const current = latest.current;
    if (idleTimer.current !== null) {
      clearTimeout(idleTimer.current);
      idleTimer.current = null;
    }
    forgetDismissal(current.historyKey);
    setOffer({ kind: 'none' });
    current.store.clear(current.historyKey).catch(() => {
      // Nothing to do: the save already succeeded, which is what the author asked for.
    });
  }, []);

  return { offer, noteChange, captureNow, accept, dismiss, finalize };
}
