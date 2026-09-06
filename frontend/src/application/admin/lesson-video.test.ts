import { describe, expect, test } from '@rstest/core';
import type { LessonVideoRepository } from '../../ports/LessonVideoRepository';
import { maximumVideoBytes, videoChunkBytes } from '../../domain/media/video';
import { createLessonVideos } from './lesson-video';

const id = '10000000-0000-4000-8000-000000000001';
const target = {
  kind: 'new',
  ticket: {
    videoId: id,
    uploadUrl: `/uploads/${id}`,
    chunkBytes: videoChunkBytes,
    maxBytes: maximumVideoBytes,
    expiresAt: '2026-09-07T12:00:00.123456Z',
  },
} as const;

function repository(overrides: Partial<LessonVideoRepository> = {}): LessonVideoRepository {
  const base: LessonVideoRepository = {
    mint: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    listByCourse: async () => ({ ok: true, value: [] }),
    retry: async () => ({ ok: false, error: { code: 'video_not_found' } }),
    transfer: async () => ({ ok: true, value: null }),
    cancelUpload: async () => ({ ok: true, value: null }),
  };
  return { ...base, ...overrides };
}

describe('lesson video use cases', () => {
  test('refuses malformed inputs before crossing the repository seam', async () => {
    let calls = 0;
    const videos = createLessonVideos(
      repository({
        mint: async () => {
          calls += 1;
          return { ok: false, error: { code: 'dependency_unavailable' } };
        },
        transfer: async () => {
          calls += 1;
          return { ok: true, value: null };
        },
      }),
    );
    expect(await videos.mint('not-an-id', 0, '')).toEqual({
      ok: false,
      error: { code: 'invalid_request' },
    });
    expect(await videos.transfer(target, new File([], 'empty.mp4'), () => {})).toEqual({
      ok: false,
      error: { code: 'invalid_request' },
    });
    expect(calls).toBe(0);
  });

  test('returns repository business outcomes unchanged', async () => {
    const videos = createLessonVideos(
      repository({
        mint: async () => ({ ok: false, error: { code: 'upload_in_progress' } }),
        transfer: async () => ({ ok: false, error: { code: 'unsupported_media' } }),
      }),
    );
    expect(await videos.mint(id, 12, 'lesson.mp4')).toEqual({
      ok: false,
      error: { code: 'upload_in_progress' },
    });
    expect(
      await videos.transfer(
        target,
        new File(['bad'], 'lesson.mp4', { lastModified: 1234 }),
        () => {},
      ),
    ).toEqual({
      ok: false,
      error: { code: 'unsupported_media' },
    });
  });
});
