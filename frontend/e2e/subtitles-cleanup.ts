import { execFile } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';

const execute = promisify(execFile);
const root = fileURLToPath(new URL('../../', import.meta.url));
// Grants/audit do not cascade (00011). Assets would survive as detached Ready rows (00005).
// Lessons cascade progress (00006), transcripts/segments (00007); assets cascade audio (00025).
const deleteCourseStatement = `exec psql -v ON_ERROR_STOP=1 -v "course_id=$1" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
BEGIN;
DELETE FROM grant_audit WHERE grant_id IN (SELECT id FROM course_grants WHERE course_id = :'course_id');
DELETE FROM course_grants WHERE course_id = :'course_id';
DELETE FROM video_assets WHERE lesson_id IN (
  SELECT lessons.id FROM course_lessons lessons
  JOIN course_modules modules ON modules.id = lessons.module_id WHERE modules.course_id = :'course_id'
);
DELETE FROM courses WHERE id = :'course_id';
COMMIT;
SQL`;

export async function cleanupSubtitlesCourse(courseID: string) {
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
      'subtitles-cleanup',
      courseID,
    ],
    { cwd: root, timeout: 60_000 },
  );
}
