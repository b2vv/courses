import { test, expect } from '@rstest/core';
import { mkdtempSync, readFileSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve, join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { z } from 'zod';

const resultSchema = z.object({
  volumes: z.array(z.tuple([z.string(), z.string(), z.string()])),
  calls: z.array(z.array(z.string())),
  lists: z.number(),
});

for (const name of [
  'partial resolver failure',
  'absent lookup failure',
  'sorted matches',
  'resolved summary',
  'project key collision',
  'missing volumes diagnostic',
  'renamed label match',
]) {
  test(`e2e-reset ${name}`, () => {
    const project = name === 'project key collision' ? 'postgres-data' : 'resolved-project';
    const configured: Record<string, { name: string } | null> = {
      'postgres-data': { name: 'custom-postgres-data' },
      'redis-data': { name: 'custom-redis-data' },
      'media-data': { name: 'custom-media-data' },
      'prometheus-data': { name: 'custom-prometheus-data' },
    };
    const volumes = [
      ['custom-postgres-data', '', ''],
      ['custom-redis-data', '', ''],
      ['custom-media-data', '', ''],
      ['custom-prometheus-data', project, 'prometheus-data'],
      ['foreign', 'another-project', 'postgres-data'],
    ];
    if (name === 'partial resolver failure') configured.zbad = null;
    if (name === 'absent lookup failure') volumes.splice(0);
    if (name === 'sorted matches')
      volumes.unshift(
        ['alias-2', project, 'postgres-data'],
        ['alias-10', project, 'postgres-data'],
      );
    if (name === 'renamed label match') volumes[1] = ['renamed-redis', project, 'redis-data'];
    const dir = mkdtempSync(join(tmpdir(), 'courses-reset-'));
    try {
      const statePath = join(dir, 'state.json');
      writeFileSync(
        statePath,
        JSON.stringify({
          config:
            name === 'missing volumes diagnostic'
              ? { name: project }
              : { name: project, volumes: configured },
          volumes,
          calls: [],
          lists: 0,
          failFirstList: name === 'absent lookup failure',
        }),
      );
      writeFileSync(join(dir, 'docker'), '#!/bin/sh\nexec "$RESET_NODE" "$RESET_HELPER" "$@"\n', {
        mode: 0o700,
      });
      const run = spawnSync(
        'make',
        ['--no-print-directory', '-f', resolve('..', 'Makefile'), 'e2e-reset'],
        {
          cwd: dir,
          encoding: 'utf8',
          timeout: 20000,
          env: {
            ...process.env,
            PATH: `${dir}:${process.env.PATH}`,
            RESET_NODE: process.execPath,
            RESET_HELPER: resolve('src/adapters/api/test-fixtures/reset-docker.mjs'),
            RESET_TEST_STATE: statePath,
            COMPOSE_PROJECT_NAME: 'wrong-environment-project',
            MAKEFLAGS: '',
            MFLAGS: '',
          },
        },
      );
      expect(run.error).toBeUndefined();
      const output = run.stdout + run.stderr;
      const final = resultSchema.parse(JSON.parse(readFileSync(statePath, 'utf8')));
      if (name === 'partial resolver failure') {
        expect(run.status, output).not.toBe(0);
        expect(final.lists, 'resolver failure must abort before listing').toBe(0);
      } else if (name === 'absent lookup failure') {
        expect(run.status, 'failed empty lookup must fail').not.toBe(0);
        expect(output).toContain('cannot list volumes');
      } else if (name === 'missing volumes diagnostic') {
        expect(run.status).not.toBe(0);
        expect(final.lists).toBe(0);
        expect(output).toContain('postgres-data is not a volume of project resolved-project');
      } else {
        expect(run.status, output).toBe(0);
        expect(final.volumes, 'only prometheus and foreign volumes remain').toEqual([
          ['custom-prometheus-data', project, 'prometheus-data'],
          ['foreign', 'another-project', 'postgres-data'],
        ]);
        expect(output).toContain(`e2e-reset: project ${project} done (prometheus-data kept)`);
        if (name === 'sorted matches') {
          const removal = final.calls.find((call) => call[0] === 'volume' && call[1] === 'rm');
          expect(removal).toEqual(['volume', 'rm', 'alias-10', 'alias-2', 'custom-postgres-data']);
        }
      }
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  });
}
