type AudioLanguage = { readonly language: string };
type StartingAudioTrack = AudioLanguage & { readonly isDefault: boolean };

/** Whole language tags, ignoring case. In particular en !== en-US and und !== und-x-original. */
export function exactAudioTrackIndex(
  tracks: readonly AudioLanguage[],
  language: string | null,
): number {
  if (language === null) return -1;
  return tracks.findIndex((track) => track.language.toLowerCase() === language.toLowerCase());
}

export function predictedAudioLanguage(
  tracks: readonly StartingAudioTrack[],
  remembered: string | null,
): string | null {
  return (
    tracks[exactAudioTrackIndex(tracks, remembered)]?.language ??
    tracks.find((track) => track.isDefault)?.language ??
    tracks[0]?.language ??
    null
  );
}
