import { execFile } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';

const execute = promisify(execFile);
const root = fileURLToPath(new URL('../../', import.meta.url));

// FK deletion order (backend/migrations):
// 00011: grant_audit.grant_id -> course_grants.id and payment_event_id -> payment_events.id;
//        course_grants.course_id -> courses.id and order_id -> orders.id;
//        payment_events.order_id -> orders.id; orders.course_id -> courses.id (all NO ACTION).
// 00005: video_assets.lesson_id -> course_lessons.id is SET NULL, so delete assets first.
// Cascades relied on:
// 00003: course_modules.course_id -> courses.id; course_lessons.module_id -> course_modules.id.
// 00006: lesson_progress.lesson_id -> course_lessons.id.
// 00007: transcripts.lesson_id -> course_lessons.id;
//        transcript_segments.(lesson_id, language) -> transcripts.(lesson_id, language);
//        course_documents.course_id -> courses.id;
//        segment_documents.segment_id -> transcript_segments.id and document_id -> course_documents.id.
// 00008/00031: asr_drafts.lesson_id -> course_lessons.id, asset_id -> video_assets.id,
//              audio_track_id -> audio_tracks.id.
// 00009/00010: discussion_comments.lesson_id -> course_lessons.id;
//              discussion_comment_reports.comment_id -> discussion_comments.id.
// 00009: discussion_comments.parent_id -> discussion_comments.id is NO ACTION;
//        lesson cascade removes both parents and replies in the same statement.
// 00013: discussion_reviews.course_id -> courses.id;
//        discussion_review_reports.review_id -> discussion_reviews.id.
// 00015: participant_reports.course_id -> courses.id.
// 00018: lesson_documents.lesson_id and lesson_images.lesson_id -> course_lessons.id.
// 00025: audio_tracks.video_asset_id -> video_assets.id.
// 00010/00013: discussion_moderation_log has no FKs; delete its course-scoped rows explicitly.
// 00012/00017/00020: refunds are columns on orders, not a separate table.
// SQL deletion leaves image/video/audio bytes on the media volume (no filesystem cleanup).
const deleteCourseStatement = `exec psql -v ON_ERROR_STOP=1 -v "course_id=$1" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
BEGIN;
DELETE FROM grant_audit WHERE grant_id IN (SELECT id FROM course_grants WHERE course_id = :'course_id');
DELETE FROM course_grants WHERE course_id = :'course_id';
DELETE FROM payment_events WHERE order_id IN (SELECT id FROM orders WHERE course_id = :'course_id');
DELETE FROM orders WHERE course_id = :'course_id';
DELETE FROM discussion_moderation_log WHERE course_id = :'course_id';
DELETE FROM video_assets WHERE lesson_id IN (
  SELECT lessons.id FROM course_lessons lessons
  JOIN course_modules modules ON modules.id = lessons.module_id WHERE modules.course_id = :'course_id'
);
DELETE FROM courses WHERE id = :'course_id';
COMMIT;
SQL`;

export async function deleteE2ECourse(courseID: string) {
  if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(courseID)) {
    throw new Error('Course cleanup requires a UUID');
  }
  await execute(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'postgres',
      'sh',
      '-eu',
      '-c',
      deleteCourseStatement,
      'course-cleanup',
      courseID,
    ],
    { cwd: root, timeout: 60_000 },
  );
}
