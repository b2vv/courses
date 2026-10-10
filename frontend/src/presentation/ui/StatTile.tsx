import type { ReactNode } from 'react';

type StatTileProps = {
  /** Short uppercase caption, e.g. "Active learners · 30 days". */
  label: string;
  /** Already-formatted value; the caller owns number and locale formatting. */
  value: string;
  /** Optional secondary line, e.g. an honest coverage note. */
  meta?: string;
  /** Optional trailing content on the value row, e.g. a <Delta />. */
  children?: ReactNode;
};

/**
 * A KPI tile over the existing `.ui-stat` tokens. It renders a caption, a large
 * value and an optional meta line; the value is a string so the caller decides
 * formatting (and can show an em dash when a metric has no coverage).
 */
export function StatTile({ label, value, meta, children }: StatTileProps) {
  return (
    <div className="ui-stat">
      <span className="ui-stat-label ui-mono">{label}</span>
      <div className="ui-stat-row">
        <span className="ui-stat-value">{value}</span>
        {children}
      </div>
      {meta === undefined ? null : <span className="ui-stat-meta">{meta}</span>}
    </div>
  );
}
