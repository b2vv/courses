import type { ReactNode } from 'react';
import * as Aria from 'react-aria-components';
import { Button } from './controls';

function SortGlyph({ direction }: { direction: 'asc' | 'desc' }) {
  return (
    <svg viewBox="0 0 16 16" aria-hidden="true" focusable="false" className="table-sort-glyph">
      <path
        d={direction === 'asc' ? 'M4 10l4-4 4 4' : 'M4 6l4 4 4-4'}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

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

export type TableColumn = {
  id: string;
  label: string;
  pin?: 'end';
  rowHeader?: boolean;
  align?: 'start' | 'end';
};
export type TableRow = { id: string; cells: readonly ReactNode[] };
type TableProps = {
  label: string;
  columns: readonly TableColumn[];
  rows: readonly TableRow[];
  emptyMessage: string;
};

/* oxlint-disable jsx-a11y/no-redundant-roles, jsx-a11y/no-interactive-element-to-noninteractive-role -- Explicit table/row/cell/columnheader/rowheader roles are deliberate: the ≤52rem card layout sets `display:block` on these elements, which drops their implicit ARIA table roles in browsers — the explicit roles keep screen-reader table semantics and the e2e `getByRole('row')` contract intact. The lint heuristics do not model this responsive-table technique. */
export function Table({ label, columns, rows, emptyMessage }: TableProps) {
  return (
    // oxlint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- Keyboard users must be able to scroll wide tables.
    <section className="ui-table-scroll" aria-label={label} tabIndex={0}>
      <table className="ui-table" role="table">
        <caption>{label}</caption>
        <thead role="rowgroup">
          <tr role="row">
            {columns.map((column) => (
              <th
                key={column.id}
                scope="col"
                role="columnheader"
                data-pin={column.pin}
                {...(column.align === undefined ? {} : { 'data-align': column.align })}
              >
                {column.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody role="rowgroup">
          {rows.length === 0 ? (
            <tr role="row">
              <td role="cell" colSpan={columns.length}>
                {emptyMessage}
              </td>
            </tr>
          ) : (
            rows.map((row) => (
              <tr key={row.id} role="row">
                {columns.map((column, index) =>
                  column.rowHeader ? (
                    <th
                      key={column.id}
                      scope="row"
                      role="rowheader"
                      data-label={column.label}
                      data-pin={column.pin}
                      {...(column.align === undefined ? {} : { 'data-align': column.align })}
                    >
                      {row.cells[index]}
                    </th>
                  ) : (
                    <td
                      key={column.id}
                      role="cell"
                      data-label={column.label}
                      data-pin={column.pin}
                      {...(column.align === undefined ? {} : { 'data-align': column.align })}
                    >
                      {row.cells[index]}
                    </td>
                  ),
                )}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </section>
  );
}

/* oxlint-enable jsx-a11y/no-redundant-roles, jsx-a11y/no-interactive-element-to-noninteractive-role */

export type DataColumn<Row> = {
  key: string;
  label: string;
  /** Right-aligned, tabular figures (students, price). */
  numeric?: boolean;
  /** The sticky first column that names the row (icon tile + title). */
  identifier?: boolean;
  /**
   * Whether this column's header sorts the table. Defaults to true when the table is
   * sortable; set false for a column the owner cannot sort by (so its `key` is never sent
   * as a sort the server would reject).
   */
  sortable?: boolean;
  render: (row: Row) => ReactNode;
};
export type DataTableSort = { key: string; dir: 'asc' | 'desc' };

type DataTableProps<Row> = {
  /** Screen-reader caption and the figure the footer counts. */
  caption: string;
  columns: readonly DataColumn<Row>[];
  rows: readonly Row[];
  rowId: (row: Row) => string;
  /** Row actions; when omitted the actions column is not rendered. */
  actions?: (row: Row) => ReactNode;
  actionsLabel: string;
  /**
   * Sort is controlled and optional: pass both `sort` and `onSortChange` to make the
   * headers interactive, or neither to render plain labels — never a header that looks
   * sortable but does nothing. The table does not reorder rows; the owner (server or a
   * client sort) does, so a single page is never silently sorted out of its global order.
   */
  sort?: DataTableSort;
  onSortChange?: (sort: DataTableSort) => void;
  mobileSortLabel?: string;
  reverseSortLabel?: string;
  /** Shown inside the card instead of the table when there are no rows. */
  empty?: ReactNode;
  /** Pagination or other controls under the table, inside the card. */
  footer?: ReactNode;
};

function nextSort(sort: DataTableSort, key: string): DataTableSort {
  if (sort.key !== key) return { key, dir: 'asc' };
  return { key, dir: sort.dir === 'asc' ? 'desc' : 'asc' };
}

/* oxlint-disable jsx-a11y/no-redundant-roles, jsx-a11y/no-interactive-element-to-noninteractive-role -- The ≤52rem card layout sets `display:block` on table/row/cell elements, which drops their implicit ARIA table roles; the explicit roles keep screen-reader table semantics and the `getByRole` contract intact. The lint heuristics do not model this responsive-table technique. */
/**
 * The proto's editorial data table: an in-card table with sortable headers, a sticky
 * identifier and actions column, numeric alignment, a ≤52rem card/stack fallback and a
 * mobile sort control. Richer than `Table`, and scoped to `.admin-table-card` so the
 * simpler primitive stays untouched.
 */
export function DataTable<Row>({
  caption,
  columns,
  rows,
  rowId,
  actions,
  actionsLabel,
  sort,
  onSortChange,
  mobileSortLabel,
  reverseSortLabel,
  empty,
  footer,
}: DataTableProps<Row>) {
  const sortable = sort !== undefined && onSortChange !== undefined;
  const columnSortable = (column: DataColumn<Row>) => sortable && column.sortable !== false;
  const ariaSort = (key: string): 'ascending' | 'descending' | 'none' =>
    sortable && sort.key === key ? (sort.dir === 'asc' ? 'ascending' : 'descending') : 'none';
  const sortColumns = columns.filter(columnSortable);

  return (
    <section className="ui-card admin-table-card">
      {rows.length === 0 && empty !== undefined ? (
        empty
      ) : (
        <>
          {sortable ? (
            <div className="table-mobile-sort">
              <label>
                {mobileSortLabel}
                <select
                  className="ui-input"
                  aria-label={mobileSortLabel}
                  value={sort.key}
                  onChange={(event) => onSortChange({ key: event.target.value, dir: sort.dir })}
                >
                  {sortColumns.map((column) => (
                    <option key={column.key} value={column.key}>
                      {column.label}
                    </option>
                  ))}
                </select>
              </label>
              {/* The reverse control appears only when the consumer sorts both ways; a
                  key-only (fixed-direction) table omits it rather than show a dead toggle. */}
              {reverseSortLabel === undefined ? null : (
                <Button
                  variant="secondary"
                  className="admin-icon-button"
                  aria-label={reverseSortLabel}
                  onPress={() =>
                    onSortChange({ key: sort.key, dir: sort.dir === 'asc' ? 'desc' : 'asc' })
                  }
                >
                  <SortGlyph direction={sort.dir} />
                </Button>
              )}
            </div>
          ) : null}
          {/* oxlint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- Keyboard users must be able to scroll wide tables. */}
          <div className="ui-table-scroll" tabIndex={0}>
            <table className="ui-table" role="table">
              <caption className="sr-only">{caption}</caption>
              <thead role="rowgroup">
                <tr role="row">
                  {columns.map((column) => (
                    <th
                      key={column.key}
                      scope="col"
                      role="columnheader"
                      className={column.numeric ? 'is-numeric' : undefined}
                      aria-sort={columnSortable(column) ? ariaSort(column.key) : undefined}
                    >
                      {sortable && onSortChange !== undefined && column.sortable !== false ? (
                        <Button
                          variant="secondary"
                          className="table-sort"
                          onPress={() => onSortChange(nextSort(sort, column.key))}
                        >
                          {column.label}
                          <span className={sort.key === column.key ? 'sort-current' : 'sort-idle'}>
                            <SortGlyph
                              direction={
                                sort.key === column.key && sort.dir === 'desc' ? 'desc' : 'asc'
                              }
                            />
                          </span>
                        </Button>
                      ) : (
                        column.label
                      )}
                    </th>
                  ))}
                  {actions === undefined ? null : (
                    <th scope="col" role="columnheader" className="admin-table-actions">
                      {actionsLabel}
                    </th>
                  )}
                </tr>
              </thead>
              <tbody role="rowgroup">
                {rows.map((row) => (
                  <tr key={rowId(row)} role="row">
                    {columns.map((column) =>
                      column.identifier ? (
                        <th
                          key={column.key}
                          scope="row"
                          role="rowheader"
                          data-label={column.label}
                          className="admin-table-identifier"
                        >
                          {column.render(row)}
                        </th>
                      ) : (
                        <td
                          key={column.key}
                          role="cell"
                          data-label={column.label}
                          className={column.numeric ? 'is-numeric' : undefined}
                        >
                          {column.render(row)}
                        </td>
                      ),
                    )}
                    {actions === undefined ? null : (
                      <td role="cell" className="admin-table-actions" data-label={actionsLabel}>
                        <div className="table-action-group">{actions(row)}</div>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {footer === undefined ? null : footer}
        </>
      )}
    </section>
  );
}
/* oxlint-enable jsx-a11y/no-redundant-roles, jsx-a11y/no-interactive-element-to-noninteractive-role */

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
