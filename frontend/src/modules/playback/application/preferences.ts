// Per-viewer player preferences, remembered across lessons (volume, mute, speed). localStorage is a
// convenience only: every access is guarded, and the player must render correctly when the store is
// empty, blocked (private window) or throwing. Nothing here is authoritative session state — that
// lives server-side through the heartbeat. These are the knobs YouTube/Hulu keep between videos.

export type PlayerPreferences = {
  readonly volume: number;
  readonly muted: boolean;
  readonly rate: number;
  readonly captionsOn: boolean;
  /** Advance to the next lesson when one ends — on by default, as YouTube/Hulu autoplay is. */
  readonly autoplayNext: boolean;
};

export const defaultPreferences: PlayerPreferences = {
  volume: 1,
  muted: false,
  rate: 1,
  captionsOn: false,
  autoplayNext: true,
};

const STORAGE_KEY = 'courses.player.prefs.v1';

// Rates the player offers; a stored rate outside this set is dropped rather than clamped, so an old
// or tampered value can never hand the student a speed no control can undo.
const ALLOWED_RATES = new Set([0.75, 1, 1.25, 1.5, 2]);

function boundedVolume(value: unknown): number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 1
    ? value
    : defaultPreferences.volume;
}

/**
 * Parse a stored preferences blob. Pure and total: any shape that is not exactly what we wrote falls
 * back field by field to the defaults, so a partial or corrupt value degrades to a working player
 * rather than throwing. Exported for the unit test — the read/write wrappers below own the I/O.
 */
export function parsePreferences(raw: string | null): PlayerPreferences {
  if (raw === null) return defaultPreferences;
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return defaultPreferences;
  }
  if (typeof parsed !== 'object' || parsed === null) return defaultPreferences;
  const record = parsed as Record<string, unknown>;
  return {
    volume: boundedVolume(record.volume),
    muted: typeof record.muted === 'boolean' ? record.muted : defaultPreferences.muted,
    rate:
      typeof record.rate === 'number' && ALLOWED_RATES.has(record.rate)
        ? record.rate
        : defaultPreferences.rate,
    captionsOn:
      typeof record.captionsOn === 'boolean' ? record.captionsOn : defaultPreferences.captionsOn,
    autoplayNext:
      typeof record.autoplayNext === 'boolean'
        ? record.autoplayNext
        : defaultPreferences.autoplayNext,
  };
}

export function serializePreferences(prefs: PlayerPreferences): string {
  return JSON.stringify({
    volume: prefs.volume,
    muted: prefs.muted,
    rate: prefs.rate,
    captionsOn: prefs.captionsOn,
    autoplayNext: prefs.autoplayNext,
  });
}

export function readPreferences(): PlayerPreferences {
  if (typeof window === 'undefined') return defaultPreferences;
  try {
    return parsePreferences(window.localStorage.getItem(STORAGE_KEY));
  } catch {
    return defaultPreferences;
  }
}

export function writePreferences(prefs: PlayerPreferences): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(STORAGE_KEY, serializePreferences(prefs));
  } catch {
    // A full, blocked or disabled store is not an error the student should see: the preference is
    // simply not remembered this time, and playback is unaffected.
  }
}
