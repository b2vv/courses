import { describe, expect, it } from '@rstest/core';
import { toWebVTT, type TranscriptSegment } from './transcript';

function segment(
  startMs: number,
  endMs: number,
  text: string,
  speaker?: string,
): TranscriptSegment {
  return {
    id: `${startMs}`,
    startMs,
    endMs,
    text,
    documentIds: [],
    ...(speaker === undefined ? {} : { speaker }),
  };
}

describe('toWebVTT', () => {
  it('writes the header even for an empty transcript', () => {
    // A file without WEBVTT on the first line is not WebVTT at all, and the browser rejects the
    // whole track rather than showing nothing.
    expect(toWebVTT([])).toBe('WEBVTT\n\n');
  });

  it('formats times as hh:mm:ss.mmm', () => {
    const out = toWebVTT([segment(0, 1_500, 'перша'), segment(3_661_234, 3_662_000, 'пізніша')]);
    expect(out).toContain('00:00:00.000 --> 00:00:01.500');
    expect(out).toContain('01:01:01.234 --> 01:01:02.000');
  });

  it('carries the speaker as a voice span', () => {
    expect(toWebVTT([segment(0, 1000, 'слово', 'Лектор')])).toContain('<v Лектор>слово');
  });

  it('skips a zero-length segment instead of producing an unparseable file', () => {
    // The domain allows start === end (such a segment is simply never active), but WebVTT requires
    // the end to be strictly later. Emitting it would break the whole file, not one line.
    const out = toWebVTT([segment(1000, 1000, 'миттєва'), segment(2000, 3000, 'справжня')]);
    expect(out).not.toContain('миттєва');
    expect(out).toContain('справжня');
  });

  it('collapses a blank line inside cue text', () => {
    // A blank line ends a cue: without this every following cue would mean something else.
    const out = toWebVTT([
      segment(0, 1000, 'перший рядок\n\nдругий рядок'),
      segment(1000, 2000, 'наступна'),
    ]);
    expect(out.split('\n\n')).toHaveLength(3); // header, cue one, cue two
    expect(out).toContain('перший рядок\nдругий рядок');
  });

  it('writes text as text, never as markup', () => {
    const out = toWebVTT([segment(0, 1000, '<script>alert(1)</script>')]);
    expect(out).toContain('<script>alert(1)</script>');
    expect(out.startsWith('WEBVTT')).toBe(true);
  });
});
