import type { ReactNode } from 'react';
import { Link } from './controls';

type StateProps = { title: string; children?: ReactNode };

export function EmptyState({ title, children }: StateProps) {
  return (
    <section className="ui-state">
      <h2>{title}</h2>
      {children}
    </section>
  );
}

type ErrorStateProps = StateProps & { retryHref?: string };

export function ErrorState({ title, children, retryHref }: ErrorStateProps) {
  return (
    <section className="ui-state ui-state--error" role="alert">
      <h2>{title}</h2>
      {children}
      {retryHref && <Link href={retryHref}>Спробувати ще раз</Link>}
    </section>
  );
}

type SkeletonProps = { label?: string; lines?: number };

export function Skeleton({ label = 'Завантаження', lines = 3 }: SkeletonProps) {
  const lineCount = Number.isFinite(lines) ? Math.max(1, Math.min(10, Math.floor(lines))) : 3;
  return (
    <output className="ui-skeleton">
      <span className="ui-visually-hidden">{label}</span>
      {Array.from({ length: lineCount }, (_, index) => (
        <span key={`line-${index}`} className="ui-skeleton-line" aria-hidden="true" />
      ))}
    </output>
  );
}
