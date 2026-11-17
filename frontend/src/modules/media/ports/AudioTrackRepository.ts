import type {
  AudioTrack,
  AudioTracks,
  AudioTrackResult,
  AudioTransferProgress,
  AudioTransferTarget,
  AudioUploadTicket,
} from '../domain/audio-track';

export interface AudioTrackRepository {
  list(videoId: string, signal?: AbortSignal): Promise<AudioTrackResult<AudioTracks>>;
  mint(
    videoId: string,
    input: {
      readonly declaredBytes: number;
      readonly originalName: string;
      readonly language: string;
      readonly label: string;
      readonly multilingual?: boolean;
    },
    signal?: AbortSignal,
  ): Promise<AudioTrackResult<AudioUploadTicket>>;
  rename(
    videoId: string,
    trackId: string,
    language: string,
    label: string,
    multilingual: boolean,
    signal?: AbortSignal,
  ): Promise<AudioTrackResult<AudioTrack>>;
  setDefault(
    videoId: string,
    trackId: string,
    signal?: AbortSignal,
  ): Promise<AudioTrackResult<null>>;
  remove(videoId: string, trackId: string, signal?: AbortSignal): Promise<AudioTrackResult<null>>;
  transfer(
    target: AudioTransferTarget,
    file: File,
    onProgress: (progress: AudioTransferProgress) => void,
    signal?: AbortSignal,
  ): Promise<AudioTrackResult<null>>;
  cancelUpload(trackId: string, signal?: AbortSignal): Promise<AudioTrackResult<null>>;
}
