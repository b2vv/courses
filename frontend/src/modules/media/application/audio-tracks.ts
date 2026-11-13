import {
  audioChunkBytes,
  audioUploadPath,
  maximumAudioBytes,
  validAudioLabel,
  validAudioLanguage,
  validAudioTicketInput,
} from '../domain/audio-track';
import type {
  AudioTrackResult,
  AudioTransferProgress,
  AudioTransferTarget,
} from '../domain/audio-track';
import { validMediaId } from '../domain/video';
import type { AudioTrackRepository } from '../ports/AudioTrackRepository';

const invalid = { ok: false, error: { code: 'invalid_request' } } as const;

function validTransferTarget(target: AudioTransferTarget): boolean {
  if (target.kind === 'resume') return validMediaId(target.trackId);
  return (
    validMediaId(target.ticket.trackId) &&
    target.ticket.uploadUrl === audioUploadPath(target.ticket.trackId) &&
    target.ticket.chunkBytes === audioChunkBytes &&
    target.ticket.maxBytes === maximumAudioBytes
  );
}

/** Validates every input before it reaches the port; the server re-checks all of it. */
export function createAudioTracks(repository: AudioTrackRepository) {
  return {
    list(videoId: string, signal?: AbortSignal) {
      return validMediaId(videoId) ? repository.list(videoId, signal) : Promise.resolve(invalid);
    },
    mint(
      videoId: string,
      input: {
        readonly declaredBytes: number;
        readonly originalName: string;
        readonly language: string;
        readonly label: string;
      },
      signal?: AbortSignal,
    ) {
      return validMediaId(videoId) &&
        validAudioTicketInput(input.declaredBytes, input.originalName) &&
        validAudioLanguage(input.language) &&
        validAudioLabel(input.label)
        ? repository.mint(videoId, input, signal)
        : Promise.resolve(invalid);
    },
    rename(
      videoId: string,
      trackId: string,
      language: string,
      label: string,
      signal?: AbortSignal,
    ) {
      return validMediaId(videoId) &&
        validMediaId(trackId) &&
        validAudioLanguage(language) &&
        validAudioLabel(label)
        ? repository.rename(videoId, trackId, language, label, signal)
        : Promise.resolve(invalid);
    },
    setDefault(videoId: string, trackId: string, signal?: AbortSignal) {
      return validMediaId(videoId) && validMediaId(trackId)
        ? repository.setDefault(videoId, trackId, signal)
        : Promise.resolve(invalid);
    },
    remove(videoId: string, trackId: string, signal?: AbortSignal) {
      return validMediaId(videoId) && validMediaId(trackId)
        ? repository.remove(videoId, trackId, signal)
        : Promise.resolve(invalid);
    },
    transfer(
      target: AudioTransferTarget,
      file: File,
      onProgress: (progress: AudioTransferProgress) => void,
      signal?: AbortSignal,
    ): Promise<AudioTrackResult<null>> {
      return validTransferTarget(target) && file.size >= 1 && file.size <= maximumAudioBytes
        ? repository.transfer(target, file, onProgress, signal)
        : Promise.resolve(invalid);
    },
    cancelUpload(trackId: string, signal?: AbortSignal) {
      return validMediaId(trackId)
        ? repository.cancelUpload(trackId, signal)
        : Promise.resolve(invalid);
    },
  };
}
