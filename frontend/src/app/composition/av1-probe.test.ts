import { afterEach, expect, test } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import { primeAv1Probe, takeAv1ProbeResult } from './av1-probe';

const originalWindow = Object.getOwnPropertyDescriptor(globalThis, 'window');
const originalMediaSource = Object.getOwnPropertyDescriptor(globalThis, 'MediaSource');

function restoreGlobal(name: 'window' | 'MediaSource', descriptor: PropertyDescriptor | undefined) {
  if (descriptor === undefined) {
    Reflect.deleteProperty(globalThis, name);
    return;
  }
  Object.defineProperty(globalThis, name, descriptor);
}

afterEach(() => {
  restoreGlobal('window', originalWindow);
  restoreGlobal('MediaSource', originalMediaSource);
});

function installMediaSource(isTypeSupported: (contentType: string) => boolean): void {
  Object.defineProperty(globalThis, 'MediaSource', {
    configurable: true,
    value: { isTypeSupported },
  });
}

test('does not touch MediaSource when primed without a window', () => {
  let calls = 0;
  installMediaSource(() => {
    calls += 1;
    return true;
  });
  expect(Reflect.deleteProperty(globalThis, 'window')).toBe(true);
  const queryClient = new QueryClient();

  expect(() => primeAv1Probe(queryClient)).not.toThrow();

  expect(calls).toBe(0);
  expect(takeAv1ProbeResult(queryClient)).toBeUndefined();
});

test.each([
  ['supported', true],
  ['unsupported', false],
] as const)(
  'probes AV1 once per QueryClient page scope and reports %s once',
  (expected, support) => {
    const contentTypes: string[] = [];
    installMediaSource((contentType) => {
      contentTypes.push(contentType);
      return support;
    });
    const queryClient = new QueryClient();

    primeAv1Probe(queryClient);
    primeAv1Probe(queryClient);

    expect(takeAv1ProbeResult(queryClient)).toBe(expected);
    expect(takeAv1ProbeResult(queryClient)).toBeUndefined();
    expect(contentTypes).toEqual(['video/mp4; codecs="av01.0.05M.08"']);
  },
);

test.each(['missing API', 'throwing API'] as const)(
  'reports probe_failed without throwing for a %s',
  (failure) => {
    if (failure === 'missing API') {
      Reflect.deleteProperty(globalThis, 'MediaSource');
    } else {
      installMediaSource(() => {
        throw new Error('browser probe failed');
      });
    }
    const queryClient = new QueryClient();

    expect(() => primeAv1Probe(queryClient)).not.toThrow();
    expect(takeAv1ProbeResult(queryClient)).toBe('probe_failed');
    expect(takeAv1ProbeResult(queryClient)).toBeUndefined();
  },
);
