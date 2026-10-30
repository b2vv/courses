import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../../../adapters/api/api-origin';
import { createApiTeachDocumentAuthoringRepository } from './admin-document-repository';
import type { LessonDocument } from '../domain/document';

const lessonId = '20000000-0000-4000-8000-000000000001';
const route = `${apiOrigin()}/api/v1/teach/lessons/${lessonId}`;
const document: LessonDocument = { schemaVersion: 1, locale: 'en', sections: [] };
const updatedAt = '2026-10-30T00:00:00Z';
const image = {
  id: '60000000-0000-4000-8000-000000000001',
  contentType: 'image/png',
  width: 10,
  height: 20,
};
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('teach generated operations send draft, save, publish and multipart image requests', async () => {
  const seen: string[] = [];
  server.use(
    http.get(`${route}/document`, ({ request }) => {
      seen.push('get');
      expect(new URL(request.url).searchParams.get('locale')).toBe('en');
      return HttpResponse.json({ document, updatedAt: null });
    }),
    http.put(`${route}/document`, async ({ request }) => {
      seen.push('save');
      expect(await request.json()).toEqual({ document });
      return HttpResponse.json({ updatedAt });
    }),
    http.post(`${route}/document/publish`, async ({ request }) => {
      seen.push('publish');
      expect(new URL(request.url).searchParams.get('locale')).toBe('en');
      expect(await request.json()).toEqual({ expectedUpdatedAt: updatedAt });
      return HttpResponse.json({ updatedAt });
    }),
    http.post(`${route}/images`, async ({ request }) => {
      seen.push('upload');
      const file = (await request.formData()).get('file');
      expect(file).toBeInstanceOf(File);
      if (file instanceof File) expect(await file.text()).toBe('image');
      return HttpResponse.json(image, { status: 201 });
    }),
  );
  const repo = createApiTeachDocumentAuthoringRepository();
  expect(await repo.getDraft(lessonId, 'en')).toEqual({
    ok: true,
    value: { document, updatedAt: null },
  });
  expect(await repo.saveDraft(lessonId, document, null, 'en')).toEqual({
    ok: true,
    value: { updatedAt },
  });
  expect(await repo.publish(lessonId, updatedAt, 'en')).toEqual({ ok: true, value: { updatedAt } });
  expect(
    await repo.uploadImage(lessonId, new File(['image'], 'x.png', { type: 'image/png' })),
  ).toEqual({
    ok: true,
    value: {
      imageId: image.id,
      contentType: image.contentType,
      width: image.width,
      height: image.height,
    },
  });
  expect(seen).toEqual(['get', 'save', 'publish', 'upload']);
});

test('non-owner forbidden remains named on every teach operation', async () => {
  const forbidden = () => HttpResponse.json({ error: { code: 'forbidden' } }, { status: 403 });
  server.use(
    http.get(`${route}/document`, forbidden),
    http.put(`${route}/document`, forbidden),
    http.post(`${route}/document/publish`, forbidden),
    http.post(`${route}/images`, forbidden),
  );
  const repo = createApiTeachDocumentAuthoringRepository();
  const results = await Promise.all([
    repo.getDraft(lessonId),
    repo.saveDraft(lessonId, document, null),
    repo.publish(lessonId, updatedAt),
    repo.uploadImage(lessonId, new File(['image'], 'x.png')),
  ]);
  for (const result of results) expect(result).toEqual({ ok: false, error: { code: 'forbidden' } });
});
