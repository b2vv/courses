import type { ReactNode } from 'react';

type PageHeaderProps = {
  /** Small uppercase label above the title (e.g. "CONTENT MANAGEMENT"). */
  eyebrow: string;
  title: string;
  /** Optional id on the <h1>, so a surrounding region can be `aria-labelledby` it. */
  titleId?: string;
  description?: string;
  /** Primary actions rendered at the end of the header (e.g. "Create course"). */
  actions?: ReactNode;
};

/**
 * The shared heading for every admin view: eyebrow, title, optional lede and a
 * trailing action cluster. One composition so all eight views read as one product.
 */
export function PageHeader({ eyebrow, title, titleId, description, actions }: PageHeaderProps) {
  return (
    <header className="admin-page-header">
      <div className="admin-page-heading">
        <p className="eyebrow">{eyebrow}</p>
        <h1 {...(titleId === undefined ? {} : { id: titleId })}>{title}</h1>
        {description === undefined ? null : <p className="admin-page-description">{description}</p>}
      </div>
      {actions === undefined ? null : <div className="admin-page-actions">{actions}</div>}
    </header>
  );
}

export type SummaryItem = {
  id: string;
  value: string;
  label: string;
  /** Marks the item with a leading accent dot (the proto's published marker). */
  dot?: boolean;
};

type SummaryStripProps = {
  /** Names the group for assistive technology; the figures read after it. */
  label: string;
  items: readonly SummaryItem[];
};

/**
 * A thin band of at-a-glance figures under the header. The counts stay a global
 * snapshot — a status or search filter narrows the table, not this strip.
 */
export function SummaryStrip({ label, items }: SummaryStripProps) {
  return (
    // oxlint-disable-next-line jsx-a11y/prefer-tag-over-role -- These are display figures, not form controls, so `role="group"` with a name is the right grouping; `fieldset`/`hgroup` would misdescribe them.
    <div className="admin-summary-strip" role="group" aria-label={label}>
      {items.map((item) => (
        <span key={item.id}>
          {item.dot ? <span className="summary-dot" aria-hidden="true" /> : null}
          <strong>{item.value}</strong>
          {item.label}
        </span>
      ))}
    </div>
  );
}
