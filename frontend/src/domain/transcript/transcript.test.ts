import { describe, expect, it } from '@rstest/core';
import {
  activeSegmentIndex,
  documentsOf,
  searchSegments,
  type LessonTranscript,
  type TranscriptSegment,
} from './transcript';

function segment(startMs: number, endMs: number, text = 'текст', documentIds: string[] = []) {
  return {
    id: `${startMs}-${endMs}`,
    startMs,
    endMs,
    text,
    documentIds,
  } satisfies TranscriptSegment;
}

describe('activeSegmentIndex', () => {
  // This mirrors the Go domain rule exactly. If the two ever disagree, the highlight in the
  // player stops matching anything the server derives from the same times.
  const segments = [segment(0, 1000), segment(500, 1500), segment(2000, 2000)];

  it.each([
    { position: 0, expected: 0 },
    { position: 999, expected: 0 },
    { position: 500, expected: 0 }, // overlap: the earliest wins
    { position: 1000, expected: 1 }, // half-open: the boundary belongs to the next one only
    { position: 1499, expected: 1 },
    { position: 1500, expected: -1 }, // the gap highlights nothing
    { position: 2000, expected: -1 }, // a zero-length segment is never active
  ])('position $position highlights $expected', ({ position, expected }) => {
    expect(activeSegmentIndex(segments, position)).toBe(expected);
  });

  it('refuses a position that is not a number rather than highlighting the first segment', () => {
    expect(activeSegmentIndex(segments, Number.NaN)).toBe(-1);
    expect(activeSegmentIndex(segments, Number.POSITIVE_INFINITY)).toBe(-1);
  });

  it('highlights nothing in an empty transcript', () => {
    expect(activeSegmentIndex([], 0)).toBe(-1);
  });
});

describe('searchSegments', () => {
  const segments = [
    segment(0, 1, 'Перший рядок'),
    segment(1, 2, 'другий РЯДОК'),
    segment(2, 3, 'інше'),
  ];

  it('matches case-insensitively across the whole lesson', () => {
    expect(searchSegments(segments, 'рядок')).toHaveLength(2);
    expect(searchSegments(segments, 'РЯДОК')).toHaveLength(2);
  });

  it('treats an empty or blank query as no search rather than as everything', () => {
    // "No results" and "not searching" are different states; returning everything for a blank
    // query would collapse them and make the empty-result state unreachable.
    expect(searchSegments(segments, '')).toHaveLength(0);
    expect(searchSegments(segments, '   ')).toHaveLength(0);
  });

  it('returns nothing when the word is absent', () => {
    expect(searchSegments(segments, 'відсутнє')).toHaveLength(0);
  });
});

describe('documentsOf', () => {
  const transcript: LessonTranscript = {
    lessonId: 'l',
    language: 'uk',
    updatedAt: '2026-09-18T00:00:00Z',
    stale: false,
    segments: [],
    documents: [{ id: 'a', label: 'A', url: 'https://example.test/a' }],
  };

  it('resolves references against the library carried once', () => {
    expect(documentsOf(transcript, segment(0, 1, 't', ['a']))).toHaveLength(1);
  });

  it('drops a reference with no library entry instead of rendering a hole', () => {
    expect(documentsOf(transcript, segment(0, 1, 't', ['missing']))).toHaveLength(0);
  });
});
