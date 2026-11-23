import type { ReactNode } from 'react';

/** One layout for a page that could not be shown: a short label, the heading, what happened and
 * the ways back as buttons (G0 #6, #7). */
export function StatusPage({
  label,
  title,
  hint,
  isAlert = false,
  children,
}: {
  label?: string;
  title: string;
  hint: string;
  isAlert?: boolean;
  children: ReactNode;
}) {
  return (
    <section className="status-page" {...(isAlert ? { role: 'alert' } : {})}>
      {label === undefined ? null : <p className="status-page-label">{label}</p>}
      <h1>{title}</h1>
      <p className="status-page-hint">{hint}</p>
      <div className="status-page-actions">{children}</div>
    </section>
  );
}
