import { useEffect, useState } from 'react';
import {
  detachedTranscriptSnapshots,
  isLessonHistoryKey,
  transcriptHistoryKey,
  type TranscriptRecoveryTab,
  type TranscriptSnapshot,
} from '../application/recovery';
import type { TranscriptHistoryStore } from '../ports/TranscriptHistoryStore';

/** One inventory read per lesson/tab set, shared by every mounted editor panel. */
export function useDetachedTranscriptRecovery(
  lessonId: string,
  history: { readonly authorId: string; readonly store: TranscriptHistoryStore } | undefined,
  tabs: readonly TranscriptRecoveryTab[],
  enabled = true,
) {
  const store = history?.store;
  const authorId = history?.authorId;
  const [read, setRead] = useState<{
    lessonKey: string;
    tabSet: string;
    store: TranscriptHistoryStore;
    snapshots: readonly TranscriptSnapshot[];
  }>();
  const lessonKey = transcriptHistoryKey({ userId: authorId ?? '', lessonId });
  // Semantic dependency: parents may compute a new array while the current tab set is unchanged.
  const tabSet = JSON.stringify(tabs);
  useEffect(() => {
    let cancelled = false;
    if (store === undefined || !enabled) return;
    void (async () => {
      try {
        const keys = (await store.keys()).filter((key) => isLessonHistoryKey(key, lessonKey));
        const stored = await Promise.all(
          keys.map(async (key) => {
            try {
              const snapshot = await store.read(key);
              return snapshot?.key === key ? snapshot : null;
            } catch {
              return null;
            }
          }),
        );
        if (!cancelled)
          setRead({
            lessonKey,
            tabSet,
            store,
            snapshots: stored.filter((snapshot) => snapshot !== null),
          });
      } catch {
        // Local recovery is best-effort; a refused inventory must not interrupt editing.
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [store, lessonKey, tabSet, enabled]);
  const snapshots =
    enabled && read?.lessonKey === lessonKey && read.tabSet === tabSet && read.store === store
      ? read.snapshots
      : [];
  return detachedTranscriptSnapshots(
    lessonKey,
    snapshots.map((snapshot) => snapshot.key),
    snapshots,
    tabs,
  );
}
