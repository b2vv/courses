/** Display only: round valid raw milliseconds to the nearest 100 ms, with ties up. */
export function formatTranscriptTime(ms: number): string | undefined {
  if (!Number.isInteger(ms) || ms < 0) return undefined;
  const deciseconds = Math.round(ms / 100);
  const seconds = Math.floor(deciseconds / 10);
  const minutes = Math.floor(seconds / 60);
  const tail = `${String(seconds % 60).padStart(2, '0')}.${deciseconds % 10}`;
  // Unpadded minutes keep short lessons compact; hours use two-digit minute/second fields.
  return minutes < 60
    ? `${minutes}:${tail}`
    : `${Math.floor(minutes / 60)}:${String(minutes % 60).padStart(2, '0')}:${tail}`;
}
