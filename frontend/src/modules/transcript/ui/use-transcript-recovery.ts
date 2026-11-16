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
  readonly isLoaded: boolean;
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
  /** Flush pending author edits, then suspend capture during a save/dialog. */
  readonly setCapturing: (enabled: boolean) => void;
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
  const [loaded, setLoaded] = useState<{
    historyKey: string;
    store: TranscriptHistoryStore;
  } | null>(null);
  const isLoaded = loaded?.historyKey === historyKey && loaded.store === store;
  const hasRead = useRef(false);
  // Ownership is scoped to this key/store visit, never inferred from a later storage read.
  const ownership = useRef<{
    historyKey: string;
    store: TranscriptHistoryStore;
    seen?: { capturedAt: number; ready: Promise<void> };
    written?: { capturedAt: number; ready: Promise<void> };
  }>({ historyKey, store });
  const pendingLines = useRef<readonly TranscriptDraftLine[]>([]);
  const capturing = useRef(true);
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
    hasRead.current = false;
    const owner = { historyKey, store };
    ownership.current = owner;
    void (async () => {
      try {
        const stored = await store.read(historyKey);
        if (cancelled) return;
        if (stored !== null)
          ownership.current.seen = { capturedAt: stored.capturedAt, ready: Promise.resolve() };
        const next = recoveryOffer(stored, latest.current.baseline, wasDismissed(historyKey));
        setOffer(next);
      } catch {
        // No local copy is a normal outcome, not an error worth showing.
        if (!cancelled) {
          setOffer({ kind: 'none' });
        }
      } finally {
        if (!cancelled) {
          hasRead.current = true;
          setLoaded({ historyKey, store });
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [historyKey, store]);

  const retire = useCallback(
    async (
      owner: typeof ownership.current,
      version: typeof ownership.current.seen,
      refresh: boolean,
    ) => {
      if (version === undefined) return;
      try {
        // A write may still be completing when blur/Save asks to retire it.
        await version.ready;
        if (await owner.store.clearIfVersion(owner.historyKey, version.capturedAt)) return;
        if (!refresh) return; // Save leaves a newer foreign capture for the next visit.
        const stored = await owner.store.read(owner.historyKey);
        if (ownership.current !== owner) return;
        if (stored !== null) {
          owner.seen = { capturedAt: stored.capturedAt, ready: Promise.resolve() };
          forgetDismissal(owner.historyKey);
        }
        setOffer(recoveryOffer(stored, latest.current.baseline, false));
      } catch {
        // Refused storage never prevents editing or a successful server Save.
      }
    },
    [],
  );

  const capture = useCallback(
    (lines: readonly TranscriptDraftLine[]) => {
      if (!capturing.current) return;
      const current = latest.current;
      const owner = ownership.current;
      if (owner.historyKey !== current.historyKey || owner.store !== current.store) return;
      const directive = captureDirective(lines, current.baseline);
      if (directive.kind === 'clear') {
        if (hasRead.current) void retire(owner, owner.seen, true);
        return;
      }
      try {
        const capturedAt = Date.now();
        const ready = current.store.write({
          key: current.historyKey,
          capturedAt,
          ...(current.baseline.language === undefined
            ? {}
            : { language: current.baseline.language }),
          ...(current.baseline.baseLanguage === undefined
            ? {}
            : { baseLanguage: current.baseline.baseLanguage }),
          expectedUpdatedAt: current.baseline.expectedUpdatedAt,
          ...(current.baseline.draftUpdatedAt === undefined
            ? {}
            : { draftUpdatedAt: current.baseline.draftUpdatedAt }),
          ...(current.baseline.acceptedSource === undefined
            ? {}
            : { acceptedSource: current.baseline.acceptedSource }),
          lines,
        });
        owner.seen = owner.written = { capturedAt, ready };
        void ready
          .then(() => forgetDismissal(current.historyKey))
          .catch(() => {
            // Best-effort capture: a refused/full store must never interrupt typing.
          });
      } catch {
        // A store may also refuse synchronously; capture remains best-effort.
      }
    },
    [retire],
  );

  useEffect(
    () => () => {
      if (idleTimer.current === null) return;
      clearTimeout(idleTimer.current);
      idleTimer.current = null;
      capture(pendingLines.current);
    },
    [capture],
  );

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
      idleTimer.current = null;
      if (!capturing.current) return;
      pendingLines.current = lines;
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
    const owner = ownership.current;
    rememberDismissal(owner.historyKey);
    setOffer({ kind: 'none' });
    void retire(owner, owner.seen, true);
  }, [retire]);

  const finalize = useCallback(() => {
    if (!hasRead.current) return;
    const current = latest.current;
    if (idleTimer.current !== null) {
      clearTimeout(idleTimer.current);
      idleTimer.current = null;
    }
    forgetDismissal(current.historyKey);
    setOffer({ kind: 'none' });
    const owner = ownership.current;
    void retire(owner, owner.written, false);
  }, [retire]);

  const setCapturing = useCallback(
    (enabled: boolean) => {
      if (!enabled && idleTimer.current !== null) {
        clearTimeout(idleTimer.current);
        idleTimer.current = null;
        capture(pendingLines.current);
      }
      capturing.current = enabled;
    },
    [capture],
  );

  return {
    offer: isLoaded ? offer : { kind: 'none' },
    isLoaded,
    noteChange,
    captureNow,
    accept,
    dismiss,
    finalize,
    setCapturing,
  };
}
