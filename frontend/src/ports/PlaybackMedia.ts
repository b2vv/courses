export type PlaybackQuality = { readonly id: number; readonly label: string };
export type PlaybackMediaError = {
  readonly status: number | undefined;
  readonly retryable: boolean;
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
