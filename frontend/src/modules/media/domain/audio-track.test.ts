import { describe, expect, test } from '@rstest/core';
import {
  audioTrackLanguages,
  availableAudioLanguages,
  canDeleteAudioTrack,
  canSetDefaultAudioTrack,
  isInterruptedAudioTrack,
  terminalAudioSignature,
  validAudioLabel,
  validAudioLanguage,
  validAudioTicketInput,
} from './audio-track';
import type { AudioTrack } from './audio-track';

const videoId = '50000000-0000-4000-8000-000000000001';
function track(trackId: string, state: AudioTrack['state'], isDefault = false): AudioTrack {
  return {
    trackId,
    videoId,
    kind: 'uploaded',
    language: 'en',
    label: 'English',
    isDefault,
    state,
    updatedAt: '2026-11-13T12:00:00.000000Z',
  };
}
const a = '60000000-0000-4000-8000-00000000000a';
const b = '60000000-0000-4000-8000-00000000000b';

describe('audio track domain', () => {
  test('every offered language passes the server language rule, and und is offered', () => {
    expect(audioTrackLanguages).toContain('und');
    for (const language of audioTrackLanguages) expect(validAudioLanguage(language)).toBe(true);
  });

  test.each(['uk', 'pt-BR', 'x'.repeat(35)])('accepts language %s', (value) => {
    expect(validAudioLanguage(value)).toBe(true);
  });

  test.each(['', 'x'.repeat(36), 'uk_UA', 'en us', 'укр'])('refuses language %j', (value) => {
    expect(validAudioLanguage(value)).toBe(false);
  });

  test.each(['English', 'Українська', '🎧'.repeat(64)])('accepts label %j', (value) => {
    expect(validAudioLabel(value)).toBe(true);
  });

  test.each(['', '🎧'.repeat(65), 'Say "hi"', 'line\nbreak', 'del\u007f'])(
    'refuses label %j, which would corrupt the HLS quoted-string',
    (value) => {
      expect(validAudioLabel(value)).toBe(false);
    },
  );

  test('the upload ticket input stays within the 500 MiB audio ceiling', () => {
    expect(validAudioTicketInput(500 * 1024 * 1024, 'dub.m4a')).toBe(true);
    expect(validAudioTicketInput(500 * 1024 * 1024 + 1, 'dub.m4a')).toBe(false);
    expect(validAudioTicketInput(0, 'dub.m4a')).toBe(false);
    expect(validAudioTicketInput(3, '')).toBe(false);
  });

  test('only a Ready non-default track can become the default', () => {
    expect(canSetDefaultAudioTrack(track(a, 'ready'))).toBe(true);
    expect(canSetDefaultAudioTrack(track(a, 'ready', true))).toBe(false);
    expect(canSetDefaultAudioTrack(track(a, 'processing'))).toBe(false);
  });

  test('a Ready asset keeps its last Ready track; any other deletion is allowed', () => {
    const lastReady = [track(a, 'ready', true), track(b, 'processing')];
    expect(canDeleteAudioTrack(lastReady, a, 'ready')).toBe(false);
    expect(canDeleteAudioTrack(lastReady, b, 'ready')).toBe(true);
    expect(canDeleteAudioTrack(lastReady, a, 'awaiting_audio')).toBe(true);
    expect(canDeleteAudioTrack([track(a, 'ready', true), track(b, 'ready')], a, 'ready')).toBe(
      true,
    );
    expect(canDeleteAudioTrack(lastReady, '60000000-0000-4000-8000-00000000000c', 'ready')).toBe(
      false,
    );
  });

  test('the terminal signature ignores order and non-terminal tracks', () => {
    expect(terminalAudioSignature([])).toBe('');
    expect(terminalAudioSignature([track(a, 'processing')])).toBe('');
    expect(terminalAudioSignature([track(b, 'failed'), track(a, 'ready')])).toBe(
      terminalAudioSignature([track(a, 'ready'), track(b, 'failed')]),
    );
    expect(terminalAudioSignature([track(a, 'ready')])).not.toBe(
      terminalAudioSignature([track(a, 'failed')]),
    );
  });

  test('a language is offered once per asset; a failed track frees it; an edit keeps its own', () => {
    const dub = { ...track(a, 'ready'), language: 'en' };
    const failed = { ...track(b, 'failed'), language: 'pl' };
    expect(availableAudioLanguages([dub, failed])).toEqual(['uk', 'pl', 'de', 'fr', 'es', 'und']);
    expect(availableAudioLanguages([dub], a)).toContain('en');
    const legacy = { ...track(a, 'ready'), language: 'pt-BR' };
    expect(availableAudioLanguages([legacy], a)[0]).toBe('pt-BR');
    expect(availableAudioLanguages([legacy])).not.toContain('pt-BR');
  });

  test('only an uploaded created/uploading track counts as interrupted', () => {
    expect(isInterruptedAudioTrack(track(a, 'uploading'))).toBe(true);
    expect(isInterruptedAudioTrack(track(a, 'created'))).toBe(true);
    expect(isInterruptedAudioTrack(track(a, 'queued'))).toBe(false);
    expect(isInterruptedAudioTrack({ ...track(a, 'created'), kind: 'embedded' })).toBe(false);
  });
});
