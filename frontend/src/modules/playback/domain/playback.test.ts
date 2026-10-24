import { describe, expect, test } from '@rstest/core';
import { nextLesson, type LearningCourse, type LearningLesson } from './playback';

function lesson(
  id: string,
  position: number,
  overrides: Partial<LearningLesson> = {},
): LearningLesson {
  return {
    id,
    title: `Lesson ${id}`,
    videoState: 'ready',
    position,
    completed: false,
    clientSeq: 0,
    isPreview: false,
    locked: false,
    ...overrides,
  };
}

function course(modules: LearningCourse['modules']): LearningCourse {
  return {
    id: 'c1',
    slug: 'c1',
    title: 'Course',
    summary: '',
    description: '',
    status: 'published',
    price: null,
    modules,
    continueLessonId: null,
  };
}

describe('nextLesson', () => {
  test('returns the immediately following lesson in outline order, crossing module boundaries', () => {
    const c = course([
      { id: 'm1', title: 'One', lessons: [lesson('a', 1), lesson('b', 2)] },
      { id: 'm2', title: 'Two', lessons: [lesson('c', 1)] },
    ]);
    expect(nextLesson(c, 'a')?.id).toBe('b');
    // The last lesson of a module advances into the first of the next, matching the rendered order.
    expect(nextLesson(c, 'b')?.id).toBe('c');
  });

  test('returns null for the last lesson and for an id not in the course', () => {
    const c = course([{ id: 'm1', title: 'One', lessons: [lesson('a', 1), lesson('b', 2)] }]);
    expect(nextLesson(c, 'b')).toBeNull();
    expect(nextLesson(c, 'missing')).toBeNull();
  });

  test('does not re-derive playability — it returns the neighbour even when locked or not ready', () => {
    // The caller gates on locked/videoState; the helper only orders the tree, so a locked neighbour
    // is still "next". Asserting this keeps the responsibility boundary from drifting into the helper.
    const c = course([
      {
        id: 'm1',
        title: 'One',
        lessons: [lesson('a', 1), lesson('b', 2, { locked: true, videoState: 'processing' })],
      },
    ]);
    const next = nextLesson(c, 'a');
    expect(next?.id).toBe('b');
    expect(next?.locked).toBe(true);
  });
});
