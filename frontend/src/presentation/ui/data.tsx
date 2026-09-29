import type { ReactNode } from 'react';
import * as Aria from 'react-aria-components';
import { Button } from './controls';

type ProgressBarProps = Omit<Aria.ProgressBarProps, 'children' | 'className'> & { label: string };

export function ProgressBar({ label, ...props }: ProgressBarProps) {
  return (
    <Aria.ProgressBar {...props} className="ui-progress">
      {({ percentage, valueText }) => (
        <>
          <div className="ui-progress-label">
            <Aria.Label>{label}</Aria.Label>
            <span>{valueText}</span>
          </div>
          <div className="ui-progress-track">
            <div className="ui-progress-fill" style={{ width: `${percentage ?? 35}%` }} />
          </div>
        </>
      )}
    </Aria.ProgressBar>
  );
}

export type MeterTone = 'neutral' | 'success' | 'warning' | 'danger' | 'info';

type LoadMeterProps = Omit<Aria.MeterProps, 'children' | 'className'> & {
  label: string;
  tone?: MeterTone;
};

export function LoadMeter({ label, tone = 'neutral', ...props }: LoadMeterProps) {
  return (
    <Aria.Meter {...props} className={`ui-meter ui-meter--${tone}`}>
      {({ percentage, valueText }) => (
        <>
          <div className="ui-meter-label">
            <Aria.Label>{label}</Aria.Label>
            <span className="ui-meter-value">{valueText}</span>
          </div>
          <div className="ui-meter-track">
            <div className="ui-meter-fill" style={{ width: `${percentage}%` }} />
          </div>
        </>
      )}
    </Aria.Meter>
  );
}

type DeltaProps = {
  label: string;
  value: string;
  direction: 'up' | 'down';
  tone: 'success' | 'danger';
};

export function Delta({ label, value, direction, tone }: DeltaProps) {
  return (
    <span className={`ui-delta ui-delta--${tone}`} aria-label={label}>
      <span aria-hidden="true">{direction === 'up' ? '↑' : '↓'}</span>
      <span aria-hidden="true">{value}</span>
    </span>
  );
}

export type TableColumn = { id: string; label: string };
export type TableRow = { id: string; cells: readonly ReactNode[] };
type TableProps = {
  label: string;
  columns: readonly TableColumn[];
  rows: readonly TableRow[];
  emptyMessage: string;
};

export function Table({ label, columns, rows, emptyMessage }: TableProps) {
  return (
    // oxlint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- Keyboard users must be able to scroll wide tables.
    <section className="ui-table-scroll" aria-label={label} tabIndex={0}>
      <table className="ui-table">
        <caption>{label}</caption>
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column.id} scope="col">
                {column.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td colSpan={columns.length}>{emptyMessage}</td>
            </tr>
          ) : (
            rows.map((row) => (
              <tr key={row.id}>
                {columns.map((column, index) => (
                  <td key={column.id}>{row.cells[index]}</td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </section>
  );
}

type PaginationProps = {
  label: string;
  previousLabel: string;
  nextLabel: string;
  pageLabel: string;
  emptyLabel: string;
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
};

export function Pagination({
  label,
  previousLabel,
  nextLabel,
  pageLabel,
  emptyLabel,
  page,
  totalPages,
  onPageChange,
}: PaginationProps) {
  return (
    <nav aria-label={label} className="ui-pagination">
      <Button
        variant="secondary"
        isDisabled={page <= 1 || totalPages === 0}
        onPress={() => onPageChange(page - 1)}
      >
        {previousLabel}
      </Button>
      <span className="ui-pagination-label" aria-live="polite">
        {totalPages === 0 ? emptyLabel : pageLabel}
      </span>
      <Button
        variant="secondary"
        isDisabled={page >= totalPages}
        onPress={() => onPageChange(page + 1)}
      >
        {nextLabel}
      </Button>
    </nav>
  );
}
