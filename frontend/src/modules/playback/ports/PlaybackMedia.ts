export type PlaybackQuality = { readonly id: number; readonly label: string };
export type PlaybackMediaError = {
  readonly status: number | undefined;
  readonly retryable: boolean;
  // The address the failure came from, when the player knows it. A 403 only means the course is
  // closed to this student if it came from the prefix currently in use; one from a prefix that
  // has already been replaced says nothing about access - see `mediaFailureDirective`.
  readonly url: string | undefined;
};

export type PlaybackMediaCallbacks = {
  onReady(): void;
  onLevels(levels: readonly PlaybackQuality[]): void;
  onError(error: PlaybackMediaError): void;
};

export interface PlaybackMedia {
  readonly kind: 'hls' | 'native';
  load(source: string): void;
  stopLoad(): void;
  selectLevel(level: number | 'auto'): void;
  destroy(): void;
}

export type CreatePlaybackMedia = (
  video: HTMLVideoElement,
  callbacks: PlaybackMediaCallbacks,
) => PlaybackMedia;
