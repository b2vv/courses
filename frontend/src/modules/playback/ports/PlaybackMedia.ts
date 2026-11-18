export type PlaybackQuality = { readonly id: number; readonly label: string };
/**
 * One audio rendition of the lesson's master manifest (multi-audio S7). The language is its identity:
 * it is unique per asset (migration 00026), whereas the hls index can change between manifests.
 */
export type PlaybackAudioTrack = {
  readonly language: string;
  readonly label: string;
  readonly isDefault: boolean;
};
export type PlaybackMediaError = {
  readonly status: number | undefined;
  readonly retryable: boolean;
  // The address the failure came from, when the player knows it. A 403 only means the course is
  // closed to this student if it came from the prefix currently in use; one from a prefix that
  // has already been replaced says nothing about access - see `mediaFailureDirective`.
  readonly url: string | undefined;
};

export type CaptionSelection =
  | { readonly state: 'pending'; readonly requested?: string }
  | { readonly state: 'confirmed'; readonly language: string }
  | { readonly state: 'unavailable' };

export type PlaybackMediaCallbacks = {
  onReady(): void;
  onLevels(levels: readonly PlaybackQuality[]): void;
  onAudioTracks(tracks: readonly PlaybackAudioTrack[]): void;
  /** The track the player has selected (not necessarily audible yet). */
  onAudioTrack(language: string): void;
  /** Latest selected language, confirmed only when audio under the playhead has that language. */
  onCaptionSelection(selection: CaptionSelection): void;
  onError(error: PlaybackMediaError): void;
};

export interface PlaybackMedia {
  readonly kind: 'hls' | 'native';
  load(source: string): void;
  stopLoad(): void;
  selectLevel(level: number | 'auto'): void;
  /** Selects the track with exactly this language; false when the manifest has none. */
  selectAudioTrack(language: string): boolean;
  destroy(): void;
}

export type PlaybackMediaOptions = {
  /** The remembered audio language to start with when the manifest offers it exactly. */
  readonly audioLanguage: string | null;
};

export type CreatePlaybackMedia = (
  video: HTMLVideoElement,
  callbacks: PlaybackMediaCallbacks,
  options?: PlaybackMediaOptions,
) => PlaybackMedia;
