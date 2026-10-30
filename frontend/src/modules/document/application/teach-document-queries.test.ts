import { expect, rstest, test } from '@rstest/core';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import {
  DocumentAuthoringError,
  ImageUploadError,
  teachLessonDocumentQueryKey,
  createTeachDocumentQueries,
} from './admin-document-queries';
import type { LessonDocument } from '../domain/document';
import type { DocumentAuthoringRepository } from '../ports/DocumentAuthoringRepository';

const lessonId = '20000000-0000-4000-8000-000000000001';
const emptyDocument: LessonDocument = { schemaVersion: 1, locale: 'uk', sections: [] };

function repository(
  overrides: Partial<DocumentAuthoringRepository> = {},
): DocumentAuthoringRepository {
  return {
    getDraft: async () => ({ ok: true, value: { document: emptyDocument, updatedAt: null } }),
    saveDraft: async () => ({ ok: true, value: { updatedAt: '2026-10-13T00:00:00Z' } }),
    publish: async () => ({ ok: true, value: { updatedAt: '2026-10-13T00:00:00Z' } }),
    uploadImage: async () => ({
      ok: true,
      value: {
        imageId: '60000000-0000-4000-8000-000000000001',
        contentType: 'image/png',
        width: 10,
        height: 10,
      },
    }),
    ...overrides,
  };
}

function client() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

test('the draft query returns the loaded document', async () => {
  const queries = createTeachDocumentQueries(repository());
  const observer = new QueryObserver(client(), queries.document(lessonId));
  const unsubscribe = observer.subscribe(() => {});
  await observer.refetch();
  expect(observer.getCurrentResult().data).toEqual({ document: emptyDocument, updatedAt: null });
  unsubscribe();
});

test('a refusal becomes a typed DocumentAuthoringError', async () => {
  const queries = createTeachDocumentQueries(
    repository({ getDraft: async () => ({ ok: false, error: { code: 'document_conflict' } }) }),
  );
  const observer = new QueryObserver(client(), queries.document(lessonId));
  const unsubscribe = observer.subscribe(() => {});
  await observer.refetch();
  const error = observer.getCurrentResult().error;
  expect(error).toBeInstanceOf(DocumentAuthoringError);
  if (!(error instanceof DocumentAuthoringError)) throw new Error('Expected authoring error');
  expect(error.code).toBe('document_conflict');
  unsubscribe();
});

test('save invalidates the draft so the next save carries the fresh token', async () => {
  const c = client();
  const spy = rstest.spyOn(c, 'invalidateQueries');
  const queries = createTeachDocumentQueries(repository());
  await c
    .getMutationCache()
    .build(c, queries.save(c))
    .execute({ lessonId, document: emptyDocument, expectedUpdatedAt: null });
  expect(spy).toHaveBeenCalledWith({
    queryKey: teachLessonDocumentQueryKey(lessonId),
    exact: true,
  });
});

test('publish invalidates both the teach draft and every learner reader of the lesson', async () => {
  const c = client();
  const spy = rstest.spyOn(c, 'invalidateQueries');
  const queries = createTeachDocumentQueries(repository());
  await c
    .getMutationCache()
    .build(c, queries.publish(c))
    .execute({ lessonId, expectedUpdatedAt: '2026-10-13T00:00:00Z' });
  expect(spy).toHaveBeenCalledWith({
    queryKey: teachLessonDocumentQueryKey(lessonId),
    exact: true,
  });
  expect(spy).toHaveBeenCalledWith({ queryKey: ['learn', 'document', lessonId] });
});

test('an upload refusal becomes a typed ImageUploadError', async () => {
  const c = client();
  const queries = createTeachDocumentQueries(
    repository({ uploadImage: async () => ({ ok: false, error: { code: 'image_too_large' } }) }),
  );
  const file = new File([new Uint8Array([1])], 'x.png', { type: 'image/png' });
  await expect(
    c.getMutationCache().build(c, queries.uploadImage()).execute({ lessonId, file }),
  ).rejects.toBeInstanceOf(ImageUploadError);
});

test('a successful upload returns the stored image identity', async () => {
  const c = client();
  const queries = createTeachDocumentQueries(repository());
  const file = new File([new Uint8Array([1])], 'x.png', { type: 'image/png' });
  const uploaded = await c
    .getMutationCache()
    .build(c, queries.uploadImage())
    .execute({ lessonId, file });
  expect(uploaded.imageId).toBe('60000000-0000-4000-8000-000000000001');
  expect(uploaded.contentType).toBe('image/png');
});

test('teach keys are locale-scoped and all operations use the supplied repository', async () => {
  const repo = repository();
  const get = rstest.spyOn(repo, 'getDraft');
  const save = rstest.spyOn(repo, 'saveDraft');
  const publish = rstest.spyOn(repo, 'publish');
  const upload = rstest.spyOn(repo, 'uploadImage');
  const queries = createTeachDocumentQueries(repo);
  const c = client();
  expect(queries.documents).toBe(repo);
  expect(queries.document(lessonId).queryKey).toEqual(['teach', 'document', lessonId, null]);
  expect(queries.document(lessonId, 'en').queryKey).toEqual(['teach', 'document', lessonId, 'en']);
  expect(teachLessonDocumentQueryKey(lessonId, 'en')).toEqual([
    'teach',
    'document',
    lessonId,
    'en',
  ]);
  await c.fetchQuery(queries.document(lessonId, 'en'));
  expect(get).toHaveBeenCalledWith(lessonId, 'en', expect.any(AbortSignal));
  await c.getMutationCache().build(c, queries.save(c)).execute({
    lessonId,
    document: emptyDocument,
    expectedUpdatedAt: null,
    locale: 'uk',
  });
  expect(save).toHaveBeenCalledWith(lessonId, emptyDocument, null, 'uk');
  await c.getMutationCache().build(c, queries.publish(c)).execute({
    lessonId,
    expectedUpdatedAt: 'token',
    locale: 'en',
  });
  expect(publish).toHaveBeenCalledWith(lessonId, 'token', 'en');
  const file = new File(['image'], 'x.png', { type: 'image/png' });
  await c.getMutationCache().build(c, queries.uploadImage()).execute({ lessonId, file });
  expect(upload).toHaveBeenCalledWith(lessonId, file);
  c.clear();
});
