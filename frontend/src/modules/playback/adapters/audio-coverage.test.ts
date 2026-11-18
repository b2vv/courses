import { expect, test } from '@rstest/core';
import { createAudioCoverage, type AudioFragment } from './audio-coverage';

function ranges(...spans: ReadonlyArray<readonly [number, number]>) {
  return {
    length: spans.length,
    start: (i: number) => spans[i]?.[0] ?? NaN,
    end: (i: number) => spans[i]?.[1] ?? NaN,
  };
}
const pending = { state: 'pending' };
const en = { state: 'confirmed', language: 'en' };
const frag = { start: 0, end: 10 };

function fixture() {
  const map = createAudioCoverage();
  const query = (selected = 'en', time = 5, buffered = ranges([0, 10])) =>
    map.selection(selected, time, () => buffered);
  function append(f: AudioFragment = frag, language = 'en') {
    map.appending(f);
    map.appended(f);
    map.buffered(f, language);
    map.endDispatch();
  }
  return { map, query, append };
}

test('APPENDING immediately invalidates only its interval; APPENDED alone never labels', () => {
  const { map, query, append } = fixture();
  append();
  const middle = { start: 3, end: 7 };
  map.appending(middle);
  expect(query()).toEqual(pending);
  expect(query('en', 2)).toEqual(en);
  expect(query('en', 8)).toEqual(en);
  map.appended(middle);
  expect(query()).toEqual(pending);
  map.buffered(middle, 'de');
  expect(query('de')).toEqual({ state: 'confirmed', language: 'de' });
  expect(query()).toEqual(pending);
});

test('FRAG_BUFFERED without APPENDED or without a language is unknown', () => {
  const { map, query } = fixture();
  map.appending(frag);
  map.buffered(frag, 'en');
  expect(query()).toEqual(pending);
  map.appending(frag);
  map.appended(frag);
  map.buffered(frag, undefined);
  expect(query()).toEqual(pending);
});

test('two physical executions (quota requeue without new APPENDING) remain one successful attempt', () => {
  const { map, query } = fixture();
  map.appending(frag);
  // A recoverable quota requeue emits no public ERROR and no new APPENDING.
  map.appended(frag);
  map.appended(frag);
  expect(query()).toEqual(pending);
  map.buffered(frag, 'en');
  expect(query()).toEqual(en);
});

test.each(['append', 'full', 'bufferAppendNoProgress', 'mediaSourceRequiresReset'])(
  '%s error taints the entire logical attempt despite subsequent APPENDED, then a fresh retry works',
  () => {
    const { map, query, append } = fixture();
    append();
    map.appending(frag);
    map.error(frag);
    map.endDispatch(); // asynchronous error and updateend are separate browser tasks
    map.appended(frag);
    map.buffered(frag, 'en');
    expect(query()).toEqual(pending);
    map.endDispatch();
    append();
    expect(query()).toEqual(en);
  },
);

test('nested ERROR in APPENDED handling, before or after our listener, cannot label', () => {
  for (const errorFirst of [true, false]) {
    const { map, query } = fixture();
    map.appending(frag);
    if (errorFirst) map.error(frag);
    map.appended(frag);
    map.buffered(frag, 'en');
    if (!errorFirst) map.error(frag);
    map.endDispatch();
    expect(query()).toEqual(pending);
  }
});

test('nested FRAG_BUFFERED before our APPENDED listener can complete; a later nested error revokes it', () => {
  const { map, query } = fixture();
  map.appending(frag);
  map.buffered(frag, 'en');
  map.appended(frag);
  expect(query()).toEqual(en);
  map.error(frag);
  expect(query()).toEqual(pending);
});

test('synchronous ERROR from an internal APPENDING listener before ours is retained', () => {
  const { map, query } = fixture();
  map.error(frag);
  map.appending(frag);
  map.appended(frag);
  map.buffered(frag, 'en');
  map.endDispatch();
  expect(query()).toEqual(pending);
});

test('empty attempts keep the existing label, including when the empty fragment is another language', () => {
  const { map, query, append } = fixture();
  append();
  map.buffered({ ...frag }, 'de');
  expect(query()).toEqual(en);
  expect(query('de')).toEqual(pending);
});

test('labels are intersected with actual audio ranges after flush, eviction and EOS (short audio tail)', () => {
  const { query, append } = fixture();
  append(); // append completes after flush was announced but before removal
  expect(query()).toEqual(en);
  expect(query('en', 5, ranges([0, 3], [7, 10]))).toEqual(pending);
  expect(query('en', 1, ranges())).toEqual(pending);
  expect(query('en', 9, ranges([0, 8]))).toEqual(pending);
  expect(query('en', 7.99, ranges([0, 8]))).toEqual(en);
  expect(query('en', 10)).toEqual(pending); // half-open interval
});

test('before BUFFER_CREATED or when SourceBuffer access throws InvalidStateError, all is unknown', () => {
  const { map, append } = fixture();
  append();
  expect(map.selection('en', 5, undefined)).toEqual(pending);
  expect(
    map.selection('en', 5, () => {
      throw new DOMException('removed', 'InvalidStateError');
    }),
  ).toEqual(pending);
  expect(
    map.selection('en', 5, () => ({
      length: 1,
      start() {
        throw new DOMException('removed', 'InvalidStateError');
      },
      end: () => 10,
    })),
  ).toEqual(pending);
});

test.each(['BUFFER_RESET', 'MEDIA_DETACHING', 'load'])(
  '%s discards labels and in-flight attempts',
  () => {
    const { map, query, append } = fixture();
    append();
    const next = { start: 10, end: 20 };
    map.appending(next);
    map.reset();
    map.appended(next);
    map.buffered(next, 'en');
    expect(query()).toEqual(pending);
    expect(query('en', 15, ranges([0, 20]))).toEqual(pending);
  },
);

test('cancellation en → de → en confirms en without SWITCHED; buffered de ahead never confirms en', () => {
  const { query, append } = fixture();
  append();
  append({ start: 10, end: 20 }, 'de');
  expect(query('de')).toEqual(pending);
  expect(query()).toEqual(en);
  expect(query('en', 15, ranges([0, 20]))).toEqual(pending);
  expect(query('de', 15, ranges([0, 20]))).toEqual({ state: 'confirmed', language: 'de' });
});

test('a late completion cannot restore labels over a newer overlapping append', () => {
  const { map, query } = fixture();
  const newer = { start: 3, end: 7 };
  map.appending(frag);
  map.buffered(frag, 'en'); // internal listener can emit this before our APPENDED listener
  map.appending(newer);
  map.appended(frag);
  expect(query('en', 2)).toEqual(en);
  expect(query()).toEqual(pending);
  map.appended(newer);
  map.buffered(newer, 'de');
  expect(query('de')).toEqual({ state: 'confirmed', language: 'de' });
  map.appended(frag);
  expect(query('en')).toEqual(pending);
});

test('each new APPENDING requires its own APPENDED, even for the same fragment object', () => {
  const { map, query } = fixture();
  map.appending(frag);
  map.appended(frag);
  map.appending(frag);
  map.buffered(frag, 'en');
  expect(query()).toEqual(pending);
});
