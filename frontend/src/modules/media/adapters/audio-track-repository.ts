import { audioChunkBytes, audioUploadPath, maximumAudioBytes } from '../domain/audio-track';
import type {
  AudioTrack,
  AudioTrackError,
  AudioTrackResult,
  AudioUploadTicket,
} from '../domain/audio-track';
import type { AudioTrackRepository } from '../ports/AudioTrackRepository';
import {
  deleteAdminAudioTrack,
  listAdminAudioTracks,
  mintAdminAudioTicket,
  renameAdminAudioTrack,
  setDefaultAdminAudioTrack,
} from '../../../adapters/api/generated/courses';
import { errorSchema } from '../../../adapters/api/course-schema';
import { createApiTransport, unnamedStatusCode } from '../../../adapters/api/transport';
import type { TransportOptions, WireResponse } from '../../../adapters/api/transport';
import { audioTrackSchema, audioTracksSchema, audioUploadTicketSchema } from './audio-track-schema';
import { createTusUploads } from './tus-client';
import type { FingerprintStorage, TusResource, TusTransport } from './tus-client';

const invalid = {
  ok: false,
  error: { code: 'invalid_response' },
} satisfies AudioTrackResult<never>;

/** The tus upload target of an audio track, derived from its id alone (D-L2-2). */
export const audioTusResource: TusResource = {
  path: audioUploadPath,
  fingerprintKey: (trackId) => `courses.media.upload-fingerprint.v1.audio.${trackId}`,
  maxBytes: maximumAudioBytes,
  chunkBytes: audioChunkBytes,
  // 409 upload_offset_conflict: another tab or a stale retry is writing this track's upload.
  statusCode: (status) => (status === 409 ? 'upload_in_progress' : undefined),
};

function wireError(response: WireResponse): AudioTrackError {
  const parsed = errorSchema.safeParse(response.data);
  if (parsed.success) {
    switch (parsed.data.error.code) {
      case 'invalid_request':
      case 'unsupported_media':
      case 'upload_too_large':
        if (response.status === 400) return { code: parsed.data.error.code };
        break;
      case 'authentication_required':
        if (response.status === 401) return { code: 'authentication_required' };
        break;
      case 'origin_forbidden':
      case 'forbidden':
      case 'course_forbidden':
        if (response.status === 403) return { code: parsed.data.error.code };
        break;
      case 'video_not_found':
        if (response.status === 404) return { code: 'video_not_found' };
        break;
      case 'duplicate_language':
      case 'upload_in_progress':
        if (response.status === 409) return { code: parsed.data.error.code };
        break;
      case 'dependency_unavailable':
        if (response.status === 503) return { code: 'dependency_unavailable' };
        break;
      case 'internal_error':
        if (response.status === 500) return { code: 'dependency_unavailable' };
        break;
    }
  }
  return { code: unnamedStatusCode(response.status) };
}

function track(raw: unknown, expected: { videoId: string; trackId?: string }) {
  const parsed = audioTrackSchema.safeParse(raw);
  if (
    !parsed.success ||
    parsed.data.videoId !== expected.videoId ||
    (expected.trackId !== undefined && parsed.data.trackId !== expected.trackId)
  )
    return invalid;
  return { ok: true, value: parsed.data } satisfies AudioTrackResult<AudioTrack>;
}

const none = (): AudioTrackResult<null> => ({ ok: true, value: null });

/** Generated JSON calls and the tus upload of a track meet at this adapter boundary. */
export function createApiAudioTrackRepository(
  options: TransportOptions = {},
  storage?: FingerprintStorage,
  tusTransport?: TusTransport,
): AudioTrackRepository {
  const send = createApiTransport(options);
  const tus = createTusUploads(audioTusResource, options, storage, tusTransport);
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    expected: number,
    map: (data: unknown) => AudioTrackResult<T>,
    caller?: AbortSignal,
  ): Promise<AudioTrackResult<T>> {
    const outcome = await send(operation, execute, expected, caller);
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status === expected) return map(outcome.data);
    return { ok: false, error: wireError(outcome) };
  }

  return {
    list(videoId, signal) {
      return request<readonly AudioTrack[]>(
        'admin.audioTracks.list',
        (init) => listAdminAudioTracks(videoId, init),
        200,
        (raw) => {
          const parsed = audioTracksSchema.safeParse(raw);
          if (!parsed.success || parsed.data.items.some((item) => item.videoId !== videoId))
            return invalid;
          return { ok: true, value: parsed.data.items };
        },
        signal,
      );
    },
    mint(videoId, input, signal) {
      return request<AudioUploadTicket>(
        'admin.audioTracks.mint',
        (init) => mintAdminAudioTicket(videoId, input, init),
        201,
        (raw) => {
          const parsed = audioUploadTicketSchema.safeParse(raw);
          return parsed.success ? { ok: true, value: parsed.data } : invalid;
        },
        signal,
      );
    },
    rename(videoId, trackId, language, label, signal) {
      return request<AudioTrack>(
        'admin.audioTracks.rename',
        (init) => renameAdminAudioTrack(videoId, trackId, { language, label }, init),
        200,
        (raw) => track(raw, { videoId, trackId }),
        signal,
      );
    },
    setDefault(videoId, trackId, signal) {
      return request<null>(
        'admin.audioTracks.setDefault',
        (init) => setDefaultAdminAudioTrack(videoId, trackId, init),
        204,
        none,
        signal,
      );
    },
    remove(videoId, trackId, signal) {
      return request<null>(
        'admin.audioTracks.delete',
        (init) => deleteAdminAudioTrack(videoId, trackId, init),
        204,
        none,
        signal,
      );
    },
    async transfer(target, file, onProgress, signal) {
      const result = await tus.transfer(
        target.kind === 'new'
          ? { kind: 'new', id: target.ticket.trackId }
          : // A resume has no ticket: the size is the re-selected file's, and the tus policy only
            // proceeds when it matches the stored fingerprint and the server's Upload-Length.
            { kind: 'resume', id: target.trackId, declaredBytes: file.size },
        file,
        onProgress,
        signal,
      );
      // On a resume the size is the re-selected file's; a server Upload-Length that disagrees
      // means it is not the interrupted file, so the resume is unverified, not a bad request.
      if (target.kind === 'resume' && !result.ok && result.error.code === 'invalid_request')
        return { ok: false, error: { code: 'upload_resume_unverified' } };
      return result;
    },
    cancelUpload(trackId, signal) {
      return tus.cancel(trackId, signal);
    },
  };
}
