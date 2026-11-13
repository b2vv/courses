// An audio track belongs to one video asset. The wire types track and asset states with one
// media enum, so a track's state set equals the asset's, although a track itself never sits in
// awaiting_audio. Domain files import nothing outside `contracts/`, hence the explicit lists.
export const audioTrackStates = [
  'created',
  'uploading',
  'queued',
  'processing',
  'awaiting_audio',
  'ready',
  'failed',
] as const;
export type AudioTrackState = (typeof audioTrackStates)[number];

export const audioTrackKinds = ['embedded', 'uploaded'] as const;
export type AudioTrackKind = (typeof audioTrackKinds)[number];

export type AudioTrack = {
  readonly trackId: string;
  readonly videoId: string;
  readonly kind: AudioTrackKind;
  readonly language: string;
  readonly label: string;
  readonly isDefault: boolean;
  readonly state: AudioTrackState;
  readonly updatedAt: string;
};

export type AudioUploadTicket = {
  readonly trackId: string;
  readonly uploadUrl: string;
  readonly chunkBytes: number;
  readonly maxBytes: number;
  readonly expiresAt: string;
};

/**
 * A new transfer starts from a freshly minted ticket; a resume starts from the track id alone,
 * because after a reload no ticket exists — the upload target is derived from the id and the
 * size is checked against the stored fingerprint and the server's Upload-Length (D-L2-2).
 */
export type AudioTransferTarget =
  | { readonly kind: 'new'; readonly ticket: AudioUploadTicket }
  | { readonly kind: 'resume'; readonly trackId: string };

export type AudioTransferProgress = {
  readonly receivedBytes: number;
  readonly declaredBytes: number;
};

export type AudioTrackErrorCode =
  | 'invalid_request'
  | 'authentication_required'
  | 'origin_forbidden'
  | 'forbidden'
  | 'course_forbidden'
  | 'duplicate_language'
  | 'upload_in_progress'
  | 'unsupported_media'
  | 'upload_too_large'
  | 'video_not_found'
  | 'upload_resume_unverified'
  | 'dependency_unavailable'
  | 'invalid_response'
  | 'request_timeout'
  | 'request_cancelled';
export type AudioTrackError = { readonly code: AudioTrackErrorCode };
export type AudioTrackResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: AudioTrackError };

export const maximumAudioBytes = 500 * 1024 * 1024;
export const audioChunkBytes = 8 * 1024 * 1024;
export const acceptedAudioFileTypes = ['.m4a', '.mp3', '.aac'] as const;

/**
 * The languages an author picks from (owner decision 2026-11-13). `und` must stay: the embedded
 * track of a video is created with language "und" and the author renames it from there.
 */
export const audioTrackLanguages = ['uk', 'en', 'pl', 'de', 'fr', 'es', 'und'] as const;

export const maximumAudioLanguageLength = 35;
export const maximumAudioLabelLength = 64;

export function audioUploadPath(trackId: string): string {
  return `/uploads/audio/${trackId}`;
}

/** Mirrors the server rule: 1..35 ASCII letters, digits or hyphens. */
export function validAudioLanguage(value: string): boolean {
  return value.length <= maximumAudioLanguageLength && /^[A-Za-z0-9-]+$/.test(value);
}

/**
 * Mirrors the server rule: 1..64 characters with no double quote or control character — the
 * label is written verbatim into an HLS quoted-string, which has no escape.
 */
export function validAudioLabel(value: string): boolean {
  const characters = [...value];
  if (characters.length < 1 || characters.length > maximumAudioLabelLength) return false;
  return characters.every((character) => {
    const code = character.codePointAt(0) ?? 0;
    return character !== '"' && code >= 0x20 && code !== 0x7f;
  });
}

export function validAudioTicketInput(declaredBytes: number, originalName: string): boolean {
  return (
    Number.isSafeInteger(declaredBytes) &&
    declaredBytes >= 1 &&
    declaredBytes <= maximumAudioBytes &&
    [...originalName].length >= 1 &&
    [...originalName].length <= 255 &&
    !originalName.includes('\u0000')
  );
}

/**
 * The offered languages a track may still take: one track per (asset, language), failed tracks
 * excluded (migration 00026), so a failed upload frees its language. A track's own language stays
 * offered while it is edited, even when it lies outside the offered list (data from before it).
 */
export function availableAudioLanguages(
  tracks: readonly AudioTrack[],
  editedTrackId?: string,
): readonly string[] {
  const taken = new Set(
    tracks
      .filter((track) => track.state !== 'failed' && track.trackId !== editedTrackId)
      .map((track) => track.language),
  );
  const own = tracks.find((track) => track.trackId === editedTrackId)?.language;
  const offered: string[] = audioTrackLanguages.filter((language) => !taken.has(language));
  if (own !== undefined && !offered.includes(own)) offered.unshift(own);
  return offered;
}

/** A created or uploading track with no transfer running in this tab was interrupted. */
export function isInterruptedAudioTrack(track: AudioTrack): boolean {
  return track.kind === 'uploaded' && (track.state === 'created' || track.state === 'uploading');
}

/** Only a Ready track that is not already the default can become the default. */
export function canSetDefaultAudioTrack(track: AudioTrack): boolean {
  return track.state === 'ready' && !track.isDefault;
}

/**
 * Mirrors the server deletion rule: on a Ready asset the last Ready track cannot be removed,
 * because a Ready lesson must always keep audible playback.
 */
export function canDeleteAudioTrack(
  tracks: readonly AudioTrack[],
  trackId: string,
  assetState: AudioTrackState,
): boolean {
  if (!tracks.some((track) => track.trackId === trackId)) return false;
  if (assetState !== 'ready') return true;
  return tracks.some((track) => track.trackId !== trackId && track.state === 'ready');
}

/**
 * The sorted `trackId:state` of tracks that reached a terminal state. When it changes — including
 * the first observation — the asset state may have changed too, because the worker promotes or
 * fails the asset together with the track (D-L2-3).
 */
export function terminalAudioSignature(tracks: readonly AudioTrack[]): string {
  return tracks
    .filter((track) => track.state === 'ready' || track.state === 'failed')
    .map((track) => `${track.trackId}:${track.state}`)
    .sort()
    .join(',');
}
