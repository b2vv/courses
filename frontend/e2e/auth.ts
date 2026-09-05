import { appendFile, mkdir, readFile, writeFile } from 'node:fs/promises';
import path from 'node:path';
import type { Page } from '@playwright/test';

export const authFixtures = {
  student: {
    email: 'student@courseforge.local',
    password: 'courseforge-student-demo',
    name: 'Студентка Курсу',
  },
  instructor: {
    email: 'instructor@courseforge.local',
    password: 'courseforge-instructor-demo',
    name: 'Викладач Курсу',
  },
  admin: {
    email: 'admin@courseforge.local',
    password: 'courseforge-admin-demo',
    name: 'Адміністратор Курсу',
  },
} as const;

export type AuthRole = keyof typeof authFixtures;

const authDirectory = path.join(import.meta.dirname, '..', 'playwright', '.auth');
const loginCountPath = path.join(authDirectory, 'login-count.txt');
type RequestKind = 'login' | 'other';

export function authStatePath(role: AuthRole) {
  return path.join(authDirectory, `${role}.json`);
}

export async function resetLoginCount() {
  await mkdir(authDirectory, { recursive: true });
  await writeFile(loginCountPath, '');
}

export async function recordRequest(kind: RequestKind) {
  await appendFile(loginCountPath, `${kind}\n`);
}

export function trackPageRequests(page: Page) {
  let records = Promise.resolve();
  page.on('request', (request) => {
    const isLogin =
      new URL(request.url()).pathname === '/api/v1/auth/login' && request.method() === 'POST';
    records = records.then(() => recordRequest(isLogin ? 'login' : 'other'));
  });
  return () => records;
}

export async function readLoginCount() {
  const requests = (await readFile(loginCountPath, 'utf8')).split('\n').filter(Boolean);
  let total = 0;
  let current = 0;
  let maximumConsecutive = 0;

  for (const request of requests) {
    if (request !== 'login') {
      current = 0;
      continue;
    }
    total += 1;
    current += 1;
    maximumConsecutive = Math.max(maximumConsecutive, current);
  }

  return { total, maximumConsecutive };
}
