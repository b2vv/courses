import { maximumVideoBytes, videoChunkBytes } from '../domain/video';
import type { VideoTransferProgress, VideoTransferTarget } from '../domain/video';
import { apiOrigin } from '../../../adapters/api/api-origin';
import { createRequestControl, unnamedStatusCode } from '../../../adapters/api/transport';
import type { TransportFailureCode, TransportOptions } from '../../../adapters/api/transport';

const tusVersion = '1.0.0';
const chunkTimeoutMs = 120_000;
const fingerprintKeyPrefix = 'courses.media.upload-fingerprint.v1.';

type TusMethod = 'HEAD' | 'PATCH' | 'DELETE';

/** Every way a tus exchange can fail; a subset of each media error union that uses it. */
export type TusErrorCode =
  | TransportFailureCode
  | 'invalid_request'
  | 'unsupported_media'
  | 'upload_too_large'
  | 'video_not_found'
  | 'upload_in_progress'
  | 'upload_resume_unverified';
type TusResult =
  | { readonly ok: true; readonly value: null }
  | { readonly ok: false; readonly error: { readonly code: TusErrorCode } };

/**
 * What one kind of tus upload resource looks like on this origin: where its bytes go, how the
 * resume fingerprint is keyed, and which size/chunk limits the server enforces. The id is the
 * resource's own id (a video id or an audio track id).
 */
export type TusResource = {
  readonly path: (id: string) => string;
  readonly fingerprintKey: (id: string) => string;
  readonly maxBytes: number;
  readonly chunkBytes: number;
  /** A resource-specific reading of a tus status, tried before the shared mapping. */
  readonly statusCode?: (status: number) => TusErrorCode | undefined;
};

/** A transfer of one resource: start a freshly minted upload, or resume an interrupted one. */
export type TusTransferTarget =
  | { readonly kind: 'new'; readonly id: string }
  | { readonly kind: 'resume'; readonly id: string; readonly declaredBytes: number };

export const videoTusResource: TusResource = {
  path: (videoId) => `/uploads/${videoId}`,
  fingerprintKey: (videoId) => `${fingerprintKeyPrefix}${videoId}`,
  maxBytes: maximumVideoBytes,
  chunkBytes: videoChunkBytes,
};
type TusResponse =
  | { readonly ok: true; readonly response: Response }
  | { readonly ok: false; readonly code: TusErrorCode };

export type TusTransport = {
  send(
    method: TusMethod,
    id: string,
    body?: Blob,
    offset?: number,
    caller?: AbortSignal,
  ): Promise<TusResponse>;
};

export type UploadFingerprint = {
  readonly originalName: string;
  readonly size: number;
  readonly lastModified: number;
};

export type FingerprintStorage = {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
};

type FingerprintSource = UploadFingerprint | Pick<File, 'name' | 'size' | 'lastModified'>;

function fingerprint(source: FingerprintSource): UploadFingerprint {
  return 'originalName' in source
    ? source
    : {
        originalName: source.name,
        size: source.size,
        lastModified: source.lastModified,
      };
}

function isUploadFingerprint(value: unknown): value is UploadFingerprint {
  if (typeof value !== 'object' || value === null) return false;
  return (
    'originalName' in value &&
    typeof value.originalName === 'string' &&
    'size' in value &&
    typeof value.size === 'number' &&
    Number.isSafeInteger(value.size) &&
    'lastModified' in value &&
    typeof value.lastModified === 'number' &&
    Number.isSafeInteger(value.lastModified)
  );
}

function browserStorage(): FingerprintStorage | undefined {
  try {
    return typeof window === 'undefined' ? undefined : window.localStorage;
  } catch {
    return undefined;
  }
}

function readFingerprint(storage: FingerprintStorage, key: string): UploadFingerprint | null {
  try {
    const encoded = storage.getItem(key);
    if (encoded === null) return null;
    const decoded: unknown = JSON.parse(encoded);
    return isUploadFingerprint(decoded) ? decoded : null;
  } catch {
    return null;
  }
}

function storeFingerprint(storage: FingerprintStorage, key: string, file: File): boolean {
  try {
    storage.setItem(key, JSON.stringify(fingerprint(file)));
    return true;
  } catch {
    return false;
  }
}

function removeFingerprint(storage: FingerprintStorage | undefined, key: string): void {
  try {
    storage?.removeItem(key);
  } catch {
    // A successful server-side completion or cancellation must not be masked by storage failure.
  }
}

function headerInteger(headers: Headers, name: string): number | null {
  const value = headers.get(name);
  if (value === null || !/^\d+$/.test(value)) return null;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) ? parsed : null;
}

export function sameUploadFingerprint(left: FingerprintSource, right: FingerprintSource): boolean {
  const first = fingerprint(left);
  const second = fingerprint(right);
  return (
    first.originalName === second.originalName &&
    first.size === second.size &&
    first.lastModified === second.lastModified
  );
}

export function buildTusHeaders(method: TusMethod, offset = 0): Headers {
  const headers = new Headers({ 'Tus-Resumable': tusVersion });
  if (method === 'PATCH') {
    headers.set('Content-Type', 'application/offset+octet-stream');
    headers.set('Upload-Offset', String(offset));
  }
  return headers;
}

export function parseTusOffset(headers: Headers): number | null {
  if (headers.get('Tus-Resumable') !== tusVersion) return null;
  return headerInteger(headers, 'Upload-Offset');
}

export function parseTusHead(
  headers: Headers,
): { readonly offset: number; readonly declaredBytes: number } | null {
  const offset = parseTusOffset(headers);
  const declaredBytes = headerInteger(headers, 'Upload-Length');
  return offset !== null && declaredBytes !== null && offset <= declaredBytes
    ? { offset, declaredBytes }
    : null;
}

export function tusStatusCode(status: number): TusErrorCode {
  switch (status) {
    case 400:
      return 'unsupported_media';
    case 404:
      return 'video_not_found';
    case 413:
      return 'upload_too_large';
    case 503:
    case 507:
      return 'dependency_unavailable';
    default:
      return unnamedStatusCode(status);
  }
}

export function nextTusChunk(
  offset: number,
  declaredBytes: number,
  chunkBytes: number = videoChunkBytes,
): { readonly start: number; readonly end: number } | null {
  if (
    !Number.isSafeInteger(offset) ||
    !Number.isSafeInteger(declaredBytes) ||
    offset < 0 ||
    declaredBytes < 1 ||
    offset >= declaredBytes
  )
    return null;
  return { start: offset, end: Math.min(offset + chunkBytes, declaredBytes) };
}

/** Binary tus transport: one independent deadline per HEAD, PATCH, or DELETE exchange. */
export function createTusTransport(
  options: TransportOptions = {},
  resource: TusResource = videoTusResource,
) {
  return {
    async send(
      method: TusMethod,
      id: string,
      body?: Blob,
      offset = 0,
      caller?: AbortSignal,
    ): Promise<TusResponse> {
      const control = createRequestControl(options, caller, chunkTimeoutMs);
      try {
        const response = await fetch(`${apiOrigin()}${resource.path(id)}`, {
          method,
          headers: buildTusHeaders(method, offset),
          signal: control.signal,
          credentials: 'omit',
          redirect: 'error',
          cache: 'no-store',
          ...(body === undefined ? {} : { body }),
        });
        const expected = method === 'HEAD' ? 200 : 204;
        return response.status === expected
          ? { ok: true, response }
          : {
              ok: false,
              code: resource.statusCode?.(response.status) ?? tusStatusCode(response.status),
            };
      } catch (error) {
        return { ok: false, code: control.failureCode(error) };
      }
    },
  };
}

/** Resumable upload policy layered over the binary tus transport, for one resource kind. */
export function createTusUploads(
  resource: TusResource,
  options: TransportOptions = {},
  storage: FingerprintStorage | undefined = browserStorage(),
  transport: TusTransport = createTusTransport(options, resource),
) {
  return {
    async transfer(
      target: TusTransferTarget,
      file: File,
      onProgress: (progress: VideoTransferProgress) => void,
      signal?: AbortSignal,
    ): Promise<TusResult> {
      const id = target.id;
      const key = resource.fingerprintKey(id);
      const expectedBytes = target.kind === 'new' ? file.size : target.declaredBytes;
      let existingFingerprint: UploadFingerprint | null = null;
      if (target.kind === 'new') {
        existingFingerprint = storage === undefined ? null : readFingerprint(storage, key);
        if (existingFingerprint !== null && !sameUploadFingerprint(existingFingerprint, file))
          return { ok: false, error: { code: 'upload_resume_unverified' } };
        if (storage === undefined || !storeFingerprint(storage, key, file))
          return { ok: false, error: { code: 'dependency_unavailable' } };
      } else {
        const stored = storage === undefined ? null : readFingerprint(storage, key);
        if (stored === null || !sameUploadFingerprint(stored, file))
          return { ok: false, error: { code: 'upload_resume_unverified' } };
      }

      const head = await transport.send('HEAD', id, undefined, 0, signal);
      if (!head.ok) return { ok: false, error: { code: head.code } };
      const metadata = parseTusHead(head.response.headers);
      if (metadata === null) return { ok: false, error: { code: 'invalid_response' } };
      if (target.kind === 'new' && existingFingerprint === null && metadata.offset > 0) {
        removeFingerprint(storage, key);
        return { ok: false, error: { code: 'upload_resume_unverified' } };
      }
      if (
        metadata.declaredBytes !== expectedBytes ||
        metadata.declaredBytes !== file.size ||
        metadata.declaredBytes < 1 ||
        metadata.declaredBytes > resource.maxBytes
      )
        return { ok: false, error: { code: 'invalid_request' } };

      let receivedBytes = metadata.offset;
      onProgress({ receivedBytes, declaredBytes: metadata.declaredBytes });
      for (
        let chunk = nextTusChunk(receivedBytes, metadata.declaredBytes, resource.chunkBytes);
        chunk !== null;
      ) {
        const patched = await transport.send(
          'PATCH',
          id,
          file.slice(chunk.start, chunk.end),
          chunk.start,
          signal,
        );
        if (!patched.ok) return { ok: false, error: { code: patched.code } };
        const next = parseTusOffset(patched.response.headers);
        if (next !== chunk.end) return { ok: false, error: { code: 'invalid_response' } };
        receivedBytes = next;
        onProgress({ receivedBytes, declaredBytes: metadata.declaredBytes });
        chunk = nextTusChunk(receivedBytes, metadata.declaredBytes, resource.chunkBytes);
      }
      removeFingerprint(storage, key);
      return { ok: true, value: null };
    },
    async cancel(id: string, signal?: AbortSignal): Promise<TusResult> {
      const response = await transport.send('DELETE', id, undefined, 0, signal);
      if (!response.ok) return { ok: false, error: { code: response.code } };
      removeFingerprint(storage, resource.fingerprintKey(id));
      return { ok: true, value: null };
    },
  };
}

/** The lesson-video upload: the tus policy bound to the video resource. */
export function createTusClient(
  options: TransportOptions = {},
  storage: FingerprintStorage | undefined = browserStorage(),
  transport: TusTransport = createTusTransport(options),
) {
  const uploads = createTusUploads(videoTusResource, options, storage, transport);
  return {
    transfer(
      target: VideoTransferTarget,
      file: File,
      onProgress: (progress: VideoTransferProgress) => void,
      signal?: AbortSignal,
    ): Promise<TusResult> {
      return uploads.transfer(
        target.kind === 'new'
          ? { kind: 'new', id: target.ticket.videoId }
          : { kind: 'resume', id: target.videoId, declaredBytes: target.declaredBytes },
        file,
        onProgress,
        signal,
      );
    },
    cancel(videoId: string, signal?: AbortSignal): Promise<TusResult> {
      return uploads.cancel(videoId, signal);
    },
  };
}
