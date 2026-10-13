// The learner-facing lesson document: an ordered list of typed sections rendered as semantic
// HTML. These types mirror the Go domain (internal/course/domain/document.go), which owns the
// schema; a browser cannot import a Go type, so the shape is repeated here as a discriminated
// union that the reader renders directly.

export type LessonMark = 'bold' | 'italic';

export type LessonSpan = {
  readonly text: string;
  readonly marks?: readonly LessonMark[];
  readonly href?: string;
};

export type LessonBlock =
  | { readonly type: 'paragraph'; readonly spans: readonly LessonSpan[] }
  | { readonly type: 'heading'; readonly level: 2 | 3; readonly spans: readonly LessonSpan[] }
  | {
      readonly type: 'list';
      readonly ordered: boolean;
      readonly items: readonly (readonly LessonSpan[])[];
    };

export type LessonRichText = { readonly blocks: readonly LessonBlock[] };

export type LessonImageContentType = 'image/jpeg' | 'image/png' | 'image/webp';

/** A row column is depth-1: only text or image, never another row or a video. */
export type LessonColumn =
  | { readonly type: 'text'; readonly id: string; readonly content: LessonRichText }
  | {
      readonly type: 'image';
      readonly id: string;
      readonly imageId: string;
      readonly contentType: LessonImageContentType;
      readonly alt: string;
    };

export type LessonSection =
  | { readonly type: 'video'; readonly id: string }
  | { readonly type: 'text'; readonly id: string; readonly content: LessonRichText }
  | { readonly type: 'task'; readonly id: string; readonly content: LessonRichText }
  | {
      readonly type: 'image';
      readonly id: string;
      readonly imageId: string;
      readonly contentType: LessonImageContentType;
      readonly alt: string;
    }
  | { readonly type: 'row'; readonly id: string; readonly columns: readonly LessonColumn[] };

export type LessonDocument = {
  readonly schemaVersion: 1;
  readonly locale: 'uk' | 'en';
  readonly sections: readonly LessonSection[];
};

export type DocumentErrorCode =
  | 'course_forbidden'
  | 'lesson_not_found'
  | 'authentication_required'
  | 'invalid_request'
  | 'dependency_unavailable'
  | 'unavailable';

export type DocumentResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: DocumentErrorCode } };

// The serve extension for an image content type. This mirrors the Go domain.ExtensionFor
// (internal/image/domain/image.go): nginx serves image bytes only at /media/i/<id>.<ext> and
// derives the Content-Type from the extension, so the reader must name it the same way the
// uploader stored it. The two are always consistent because both come from one upload response.
export function imageExtensionFor(contentType: LessonImageContentType): string {
  switch (contentType) {
    case 'image/jpeg':
      return 'jpg';
    case 'image/png':
      return 'png';
    case 'image/webp':
      return 'webp';
  }
}

export function imageServeUrl(imageId: string, contentType: LessonImageContentType): string {
  return `/media/i/${imageId}.${imageExtensionFor(contentType)}`;
}

// Whether the document has anything the reader renders. Video sections are played by the existing
// lesson player, not by the document reader, so a document that is only a video slot renders as
// empty — the same as an absent publication (scenario 4).
export function hasRenderableSections(document: LessonDocument): boolean {
  return document.sections.some((section) => section.type !== 'video');
}

// --- Authoring (admin editor side) ---------------------------------------------------------------
// Loading a draft, saving it under an optimistic lock, publishing it, and uploading an image to
// reference from an image section. Kept in this file because a domain file imports nothing.

export type DocumentAuthoringErrorCode =
  | 'invalid_lesson_document'
  | 'invalid_request'
  | 'forbidden'
  | 'course_forbidden'
  | 'lesson_not_found'
  | 'document_conflict'
  | 'authentication_required'
  | 'dependency_unavailable'
  | 'unavailable';

export type ImageUploadErrorCode =
  | 'invalid_image'
  | 'unsupported_image'
  | 'image_too_large'
  | 'image_too_many_pixels'
  | 'forbidden'
  | 'lesson_not_found'
  | 'authentication_required'
  | 'dependency_unavailable'
  | 'unavailable';

export type LessonDocumentDraft = {
  readonly document: LessonDocument;
  /** The optimistic-lock token to send back when saving or publishing; null when no row exists. */
  readonly updatedAt: string | null;
};

export type UploadedImage = {
  readonly imageId: string;
  readonly contentType: LessonImageContentType;
  readonly width: number;
  readonly height: number;
};

export type DocumentAuthoringResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: DocumentAuthoringErrorCode } };

export type ImageUploadResult =
  | { readonly ok: true; readonly value: UploadedImage }
  | { readonly ok: false; readonly error: { readonly code: ImageUploadErrorCode } };
