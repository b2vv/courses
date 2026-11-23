import type { CourseDetail } from '../../../contracts/course';
import type { Membership } from '../../enrollment';

/** What is known about the viewer when the course page decides its action box. */
export type CourseViewer =
  | { readonly status: 'pending' }
  | { readonly status: 'unavailable' }
  | { readonly status: 'anonymous' }
  | { readonly status: 'signed-in'; readonly role: 'student' | 'instructor' | 'admin' };

/** A fact read for a signed-in viewer: still loading, failed, or known. */
export type Fact<T> =
  | { readonly status: 'pending' }
  | { readonly status: 'unavailable' }
  | { readonly status: 'known'; readonly value: T };

/**
 * The one state the action box shows. Each variant is a row of the C3 table
 * (`docs/design-pass/plan.md` §16): the first row that matches wins.
 */
export type CourseAction =
  | { readonly kind: 'checking' }
  | { readonly kind: 'unavailable' }
  | { readonly kind: 'owner'; readonly comingSoon: boolean }
  | { readonly kind: 'admin'; readonly comingSoon: boolean }
  | { readonly kind: 'member'; readonly blocked: boolean; readonly comingSoon: boolean }
  | { readonly kind: 'coming-soon' }
  | { readonly kind: 'sign-in'; readonly paid: boolean }
  | { readonly kind: 'enroll' }
  | { readonly kind: 'buy' };

export function lessonTotal(course: CourseDetail): number {
  return (
    course.lessonCount ?? course.program.reduce((total, module) => total + module.lessons.length, 0)
  );
}

export function courseAction({
  course,
  viewer,
  membership,
  owner,
}: {
  readonly course: CourseDetail;
  readonly viewer: CourseViewer;
  /** Ignored unless the viewer is signed in. */
  readonly membership: Fact<Membership>;
  /** Ignored unless the viewer is a signed-in instructor. */
  readonly owner: Fact<boolean>;
}): CourseAction {
  const comingSoon = lessonTotal(course) === 0;
  if (viewer.status === 'pending') return { kind: 'checking' };
  if (viewer.status === 'unavailable') return { kind: 'unavailable' };
  if (viewer.status === 'anonymous')
    return comingSoon ? { kind: 'coming-soon' } : { kind: 'sign-in', paid: course.price !== null };
  // An administrator's box depends on neither fact (G8), so their loading cannot hold it.
  const facts =
    viewer.role === 'admin'
      ? []
      : viewer.role === 'instructor'
        ? [membership, owner]
        : [membership];
  if (facts.some((fact) => fact.status === 'unavailable')) return { kind: 'unavailable' };
  if (facts.some((fact) => fact.status === 'pending')) return { kind: 'checking' };
  if (viewer.role === 'instructor' && owner.status === 'known' && owner.value)
    return { kind: 'owner', comingSoon };
  if (viewer.role === 'admin') return { kind: 'admin', comingSoon };
  const state = membership.status === 'known' ? membership.value.state : null;
  if (state === 'active' || state === 'blocked')
    return { kind: 'member', blocked: state === 'blocked', comingSoon };
  if (comingSoon) return { kind: 'coming-soon' };
  return course.price === null ? { kind: 'enroll' } : { kind: 'buy' };
}
