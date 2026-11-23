import { describe, expect, it } from '@rstest/core';
import type { CourseDetail } from '../../../contracts/course';
import { courseAction, type CourseViewer, type Fact } from './course-action';
import type { Membership } from '../../enrollment';

const course: CourseDetail = {
  id: 'c1',
  slug: 'go',
  title: 'Go',
  summary: 's',
  description: 'd',
  publishedAt: '2026-09-01T09:00:00Z',
  price: null,
  paid: false,
  lessonCount: 2,
  program: [{ title: 'M', lessons: [{ title: 'a' }, { title: 'b' }] }],
};
const paid = { ...course, price: { minor: 189000, currency: 'UAH' as const }, paid: true };
const empty = { ...course, lessonCount: 0, program: [] };
const known = <T>(value: T): Fact<T> => ({ status: 'known', value });
const pending = { status: 'pending' } as const;
const unavailable = { status: 'unavailable' } as const;
const none = known<Membership>({ state: null });
const signedIn = (role: 'student' | 'instructor' | 'admin'): CourseViewer => ({
  status: 'signed-in',
  role,
});

describe('courseAction — the C3 action table (plan §16 A1–A9)', () => {
  it('A1/A2: holds actions while the viewer is unknown and reports a failure', () => {
    expect(
      courseAction({ course, viewer: { status: 'pending' }, membership: none, owner: known(false) })
        .kind,
    ).toBe('checking');
    expect(
      courseAction({
        course,
        viewer: { status: 'unavailable' },
        membership: none,
        owner: known(false),
      }).kind,
    ).toBe('unavailable');
    expect(
      courseAction({ course, viewer: signedIn('student'), membership: pending, owner: pending })
        .kind,
    ).toBe('checking');
    expect(
      courseAction({ course, viewer: signedIn('student'), membership: unavailable, owner: pending })
        .kind,
    ).toBe('unavailable');
    expect(
      courseAction({ course, viewer: signedIn('instructor'), membership: none, owner: pending })
        .kind,
    ).toBe('checking');
    expect(
      courseAction({ course, viewer: signedIn('instructor'), membership: none, owner: unavailable })
        .kind,
    ).toBe('unavailable');
  });

  it('A3: the author sees their course, also without lessons', () => {
    expect(
      courseAction({
        course,
        viewer: signedIn('instructor'),
        membership: none,
        owner: known(true),
      }),
    ).toEqual({ kind: 'owner', comingSoon: false });
    expect(
      courseAction({
        course: empty,
        viewer: signedIn('instructor'),
        membership: none,
        owner: known(true),
      }),
    ).toEqual({ kind: 'owner', comingSoon: true });
  });

  it('A4: an administrator keeps the exemption and is not held by membership', () => {
    expect(
      courseAction({ course, viewer: signedIn('admin'), membership: pending, owner: pending }),
    ).toEqual({ kind: 'admin', comingSoon: false });
    expect(
      courseAction({
        course: empty,
        viewer: signedIn('admin'),
        membership: unavailable,
        owner: pending,
      }),
    ).toEqual({ kind: 'admin', comingSoon: true });
  });

  it('A5: active and blocked are members, revoked is not', () => {
    expect(
      courseAction({
        course,
        viewer: signedIn('student'),
        membership: known({ state: 'active' }),
        owner: pending,
      }),
    ).toEqual({ kind: 'member', blocked: false, comingSoon: false });
    expect(
      courseAction({
        course: empty,
        viewer: signedIn('student'),
        membership: known({ state: 'blocked' }),
        owner: pending,
      }),
    ).toEqual({ kind: 'member', blocked: true, comingSoon: true });
    expect(
      courseAction({
        course,
        viewer: signedIn('student'),
        membership: known({ state: 'revoked' }),
        owner: pending,
      }).kind,
    ).toBe('enroll');
  });

  it('A6: a course without lessons offers no action to anyone else', () => {
    for (const viewer of [{ status: 'anonymous' } as const, signedIn('student')])
      expect(courseAction({ course: empty, viewer, membership: none, owner: pending }).kind).toBe(
        'coming-soon',
      );
    expect(
      courseAction({
        course: { ...paid, lessonCount: 0, program: [] },
        viewer: signedIn('instructor'),
        membership: none,
        owner: known(false),
      }).kind,
    ).toBe('coming-soon');
  });

  it('A7–A9: guest signs in; a signed-in non-member enrols or buys', () => {
    expect(
      courseAction({
        course,
        viewer: { status: 'anonymous' },
        membership: pending,
        owner: pending,
      }),
    ).toEqual({ kind: 'sign-in', paid: false });
    expect(
      courseAction({
        course: paid,
        viewer: { status: 'anonymous' },
        membership: pending,
        owner: pending,
      }),
    ).toEqual({ kind: 'sign-in', paid: true });
    expect(
      courseAction({ course, viewer: signedIn('student'), membership: none, owner: pending }).kind,
    ).toBe('enroll');
    expect(
      courseAction({ course: paid, viewer: signedIn('student'), membership: none, owner: pending })
        .kind,
    ).toBe('buy');
    // An instructor who is not the author is any other user.
    expect(
      courseAction({
        course: paid,
        viewer: signedIn('instructor'),
        membership: known({ state: 'revoked' }),
        owner: known(false),
      }).kind,
    ).toBe('buy');
  });

  it('counts lessons from the program when the count is absent', () => {
    const { lessonCount: _drop, ...withoutCount } = course;
    expect(
      courseAction({
        course: withoutCount,
        viewer: signedIn('student'),
        membership: none,
        owner: pending,
      }).kind,
    ).toBe('enroll');
    const { lessonCount: _drop2, ...emptyWithoutCount } = empty;
    expect(
      courseAction({
        course: emptyWithoutCount,
        viewer: signedIn('student'),
        membership: none,
        owner: pending,
      }).kind,
    ).toBe('coming-soon');
  });
});
