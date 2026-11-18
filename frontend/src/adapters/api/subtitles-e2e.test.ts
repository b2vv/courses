import { readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { expect, test } from '@rstest/core';

const e2e = path.resolve('e2e');
const subtitles = readFileSync(path.join(e2e, 'subtitles.spec.ts'), 'utf8');

// These are source constraints, checked without starting Playwright or touching the live stack.
test('subtitle e2e inherits the caller Docker environment without a machine-specific launcher', () => {
  for (const file of readdirSync(e2e).filter((file) => file.endsWith('.ts')))
    expect(readFileSync(path.join(e2e, file), 'utf8')).not.toMatch(/\/Users\/|\/home\/[^/]+\//);
  const reset = subtitles.slice(
    subtitles.indexOf('async function resetProgress()'),
    subtitles.indexOf('async function open('),
  );
  expect(reset).toMatch(/execute\(\s*'docker',\s*\[\s*'compose'/);
  expect(reset).not.toMatch(/\benv\s*:/);
});

test('E9 waits for advancing playback with a bounded poll instead of a fixed sleep', () => {
  const e9 = subtitles.slice(subtitles.indexOf("test('E9 "), subtitles.indexOf("test('E10 "));
  expect(e9).not.toContain('waitForTimeout');
  expect(e9).toMatch(/\.poll\([\s\S]*timeout:\s*[\d_]+[\s\S]*\.toBeGreaterThan\(before\)/);
});
