import type { ReactNode } from 'react';

/** Colour role of one part. Status tones keep their reserved meaning; `neutral`/`muted` are plain greys. */
export type SplitTone = 'success' | 'warning' | 'danger' | 'info' | 'neutral' | 'muted';

export type SplitPart = {
  readonly id: string;
  readonly label: string;
  readonly value: number;
  /** Already-formatted value; the caller owns number and locale formatting. */
  readonly display: string;
  readonly tone: SplitTone;
};

type SegmentedBarProps = {
  /** Card heading: the whole being split (e.g. "Accounts"). */
  title: string;
  /** Formatted total shown beside the heading; omit when the listed parts are not the whole. */
  totalLabel?: string;
  parts: readonly SplitPart[];
  /** Formats a part's share of the whole (0..1), e.g. a percent. */
  formatShare: (share: number) => string;
  /** Shown under an empty track when every part is zero. */
  emptyLabel: string;
  /**
   * False when the listed parts do not add up to the whole (a part is withheld): then
   * no bar and no shares are drawn, because either would misstate the split.
   */
  isWhole?: boolean;
  /** Optional footer content under the list (e.g. secondary figures). */
  children?: ReactNode;
};

/**
 * One whole split into a few parts: a thin segmented bar over a list that names
 * every part with a swatch, its figure and its share. The list carries every number,
 * so the bar is decorative (`aria-hidden`) and identity never rests on colour alone.
 * A zero part is listed but not drawn; an all-zero whole shows an empty track and a
 * note instead of an invented split.
 */
export function SegmentedBar({
  title,
  totalLabel,
  parts,
  formatShare,
  emptyLabel,
  isWhole = true,
  children,
}: SegmentedBarProps) {
  const total = parts.reduce((sum, part) => sum + part.value, 0);
  return (
    <article className="ui-split">
      <header className="ui-split-head">
        <h3>{title}</h3>
        {totalLabel === undefined ? null : <span className="ui-split-total">{totalLabel}</span>}
      </header>
      {isWhole ? (
        <div className="ui-split-bar" aria-hidden="true">
          {total > 0
            ? parts
                .filter((part) => part.value > 0)
                .map((part) => (
                  <span
                    key={part.id}
                    className={`ui-split-seg ui-split-tone--${part.tone}`}
                    style={{ flexGrow: part.value }}
                  />
                ))
            : null}
        </div>
      ) : null}
      {total === 0 ? <p className="ui-split-empty">{emptyLabel}</p> : null}
      <ul className="ui-split-list">
        {parts.map((part) => (
          <li key={part.id}>
            <span className={`ui-split-swatch ui-split-tone--${part.tone}`} aria-hidden="true" />
            <span className="ui-split-name">{part.label}</span>
            <span className="ui-split-value">{part.display}</span>
            {isWhole ? (
              <span className="ui-split-share">
                {total > 0 ? formatShare(part.value / total) : ''}
              </span>
            ) : null}
          </li>
        ))}
      </ul>
      {children}
    </article>
  );
}

type ShareMeterProps = {
  label: string;
  /** Already-formatted figure, e.g. "4 of 5 active · 80%" or an em dash. */
  valueLabel: string;
  /** Filled share 0..1; null draws no bar, because a bar would imply a value. */
  ratio: number | null;
  tone: 'success' | 'warning' | 'danger' | 'info';
};

/**
 * One named share against its whole: the label and the formatted figure carry the
 * value as text, and the track under them is decorative. Used where several shares
 * stand in a list, each against its own whole.
 */
export function ShareMeter({ label, valueLabel, ratio, tone }: ShareMeterProps) {
  return (
    <div className={`ui-meter ui-meter--${tone}`}>
      <div className="ui-meter-label">
        <span>{label}</span>
        <span className="ui-meter-value">{valueLabel}</span>
      </div>
      {ratio === null ? null : (
        <div className="ui-meter-track" aria-hidden="true">
          <div
            className="ui-meter-fill"
            style={{ width: `${Math.min(Math.max(ratio, 0), 1) * 100}%` }}
          />
        </div>
      )}
    </div>
  );
}
