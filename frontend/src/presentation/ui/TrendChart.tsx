export type TrendPoint = { readonly label: string; readonly value: number };

type TrendChartProps = {
  /** Visible caption and the basis of the chart's accessible name. */
  title: string;
  mode: 'line' | 'bar';
  points: readonly TrendPoint[];
  /** Shown instead of the plot when there is no data (a quiet window). */
  emptyLabel: string;
  /** Formats a value for the screen-reader data list; defaults to the number. */
  formatValue?: (value: number) => string;
};

const viewWidth = 320;
const viewHeight = 120;
const padding = 4;

function geometry(points: readonly TrendPoint[]) {
  const max = points.reduce((highest, point) => Math.max(highest, point.value), 0);
  const span = max === 0 ? 1 : max;
  const innerWidth = viewWidth - padding * 2;
  const innerHeight = viewHeight - padding * 2;
  const step = points.length <= 1 ? 0 : innerWidth / (points.length - 1);
  const x = (index: number) => (points.length <= 1 ? viewWidth / 2 : padding + step * index);
  const y = (value: number) => padding + innerHeight * (1 - value / span);
  return { innerWidth, innerHeight, x, y };
}

/**
 * One inline-SVG trend chart in two modes (line or bar) — no charting library,
 * per the project's dependency rules. The plot is decorative: the chart carries
 * an accessible name and a visually-hidden data list so a screen reader reads
 * the series as values, not as an image. An empty series shows the empty label
 * rather than an axis with nothing on it.
 */
export function TrendChart({ title, mode, points, emptyLabel, formatValue }: TrendChartProps) {
  const format = formatValue ?? ((value: number) => String(value));

  if (points.length === 0) {
    return (
      <figure className="ui-chart ui-chart--empty">
        <figcaption className="ui-chart-title">{title}</figcaption>
        <output className="ui-chart-empty">{emptyLabel}</output>
      </figure>
    );
  }

  const { x, y } = geometry(points);
  const baseline = viewHeight - padding;
  const barWidth = Math.max(2, (viewWidth - padding * 2) / points.length / 1.6);

  return (
    <figure className="ui-chart">
      <figcaption className="ui-chart-title">{title}</figcaption>
      <svg
        className="ui-chart-plot"
        viewBox={`0 0 ${viewWidth} ${viewHeight}`}
        preserveAspectRatio="none"
      >
        <title>{title}</title>
        <line
          className="ui-chart-axis"
          x1={padding}
          y1={baseline}
          x2={viewWidth - padding}
          y2={baseline}
        />
        {mode === 'line' ? (
          <polyline
            className="ui-chart-line"
            points={points.map((point, index) => `${x(index)},${y(point.value)}`).join(' ')}
          />
        ) : (
          points.map((point, index) => (
            <rect
              key={point.label}
              className="ui-chart-bar"
              x={x(index) - barWidth / 2}
              y={y(point.value)}
              width={barWidth}
              height={Math.max(0, baseline - y(point.value))}
            />
          ))
        )}
      </svg>
      <ul className="ui-visually-hidden">
        {points.map((point) => (
          <li key={point.label}>{`${point.label}: ${format(point.value)}`}</li>
        ))}
      </ul>
    </figure>
  );
}
