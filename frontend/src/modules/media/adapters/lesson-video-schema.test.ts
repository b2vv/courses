import { describe, expect, test } from '@rstest/core';
import { courseVideosSchema, videoUploadTicketSchema } from './lesson-video-schema';

const video = {
  lessonId: '10000000-0000-4000-8000-000000000001',
  videoId: '50000000-0000-4000-8000-000000000001',
  state: 'processing',
  generation: 1,
  originalName: 'lesson.mp4',
  receivedBytes: 12,
  declaredBytes: 12,
  updatedAt: '2026-09-06T12:00:00.123456Z',
};

describe('lesson video wire schemas', () => {
  test('accepts the exact ticket and video response shapes', () => {
    expect(
      videoUploadTicketSchema.safeParse({
        videoId: video.videoId,
        uploadUrl: `/uploads/${video.videoId}`,
        chunkBytes: 8 * 1024 * 1024,
        maxBytes: 2 * 1024 * 1024 * 1024,
        expiresAt: '2026-09-07T12:00:00.123456Z',
      }).success,
    ).toBe(true);
    expect(courseVideosSchema.safeParse({ items: [video] }).success).toBe(true);
    expect(
      courseVideosSchema.safeParse({
        items: [{ ...video, originalName: `${'🎬'.repeat(251)}.mp4` }],
      }).success,
    ).toBe(true);
  });

  test.each([
    { items: [{ ...video, receivedBytes: 13 }] },
    { items: [{ ...video, storageKey: 'source/private' }] },
  ])('refuses malformed or inconsistent video data', (raw) => {
    expect(courseVideosSchema.safeParse(raw).success).toBe(false);
  });

  test('refuses a ticket whose upload path does not belong to its video', () => {
    expect(
      videoUploadTicketSchema.safeParse({
        videoId: video.videoId,
        uploadUrl: '/uploads/50000000-0000-4000-8000-000000000002',
        chunkBytes: 8 * 1024 * 1024,
        maxBytes: 2 * 1024 * 1024 * 1024,
        expiresAt: '2026-09-07T12:00:00.123456Z',
      }).success,
    ).toBe(false);
  });
});
