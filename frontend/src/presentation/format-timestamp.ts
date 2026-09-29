export function formatTimestamp(positionMs: number): string {
  const total = Math.max(0, Math.floor(positionMs / 1000));
  const minutes = Math.floor(total / 60);
  const seconds = total % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}
