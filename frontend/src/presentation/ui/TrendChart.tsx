export type TrendPoint = { readonly label: string; readonly value: number };

type TrendChartProps = {
  /** Visible caption and the basis of the chart's accessible name. */
  title: string;
  mode: 'line' | 'bar';
  points: readonly TrendPoint[];
  /** Shown instead of the plot when there is no data (a quiet window). */
  emptyLabel: string;
  /** Formats a value for the y-axis labels and the screen-reader list; defaults to the number. */
  formatValue?: (value: number) => string;
  /** Optional small label above the title (e.g. a series grouping). */
  eyebrow?: string;
  /** Optional legend/axis unit (e.g. "UAH"); renders a legend dot when set. */
  unit?: string;
};

const viewWidth = 800;
const viewHeight = 170;
const padding = 4;
/** Horizontal grid lines and y-axis labels, counting the zero line. */
const tickCount = 4;
const maxXLabels = 6;

/** Round a positive maximum up to a round number so the axis top reads cleanly. */
function niceCeil(value: number): number {
  if (value <= 0) return 1;
  const pow = 10 ** Math.floor(Math.log10(value));
  for (const step of [1, 1.5, 2, 2.5, 3, 4, 5, 7.5, 10]) {
    if (step * pow >= value) return step * pow;
  }
  return 10 * pow;
}

function geometry(points: readonly TrendPoint[], top: number) {
  const innerWidth = viewWidth - padding * 2;
  const innerHeight = viewHeight - padding * 2;
  const step = points.length <= 1 ? 0 : innerWidth / (points.length - 1);
  const x = (index: number) => (points.length <= 1 ? viewWidth / 2 : padding + step * index);
  const y = (value: number) => padding + innerHeight * (1 - value / top);
  return { x, y };
}

/** Evenly sample up to `count` points so the x-axis stays legible with long series. */
function sampleLabels(points: readonly TrendPoint[], count: number): readonly TrendPoint[] {
  if (points.length <= count) return points;
  const stride = (points.length - 1) / (count - 1);
  const sampled: TrendPoint[] = [];
  for (let i = 0; i < count; i += 1) {
    const point = points[Math.round(i * stride)];
    if (point !== undefined) sampled.push(point);
  }
  return sampled;
}

/**
 * One inline-SVG trend chart in two modes (line or bar) — no charting library,
 * per the project's dependency rules. The plot carries an accessible name and a
 * visually-hidden data list so a screen reader reads the series as values, not as
 * an image; the y/x labels and grid lines are decorative (aria-hidden) scaffolding
 * over that same data. An empty series shows the empty label rather than an axis
 * with nothing on it.
 */
export function TrendChart({
  title,
  mode,
  points,
  emptyLabel,
  formatValue,
  eyebrow,
  unit,
}: TrendChartProps) {
  const format = formatValue ?? ((value: number) => String(value));

  const heading = (
    <div className="ui-chart-heading">
      <div>
        {eyebrow === undefined ? null : <span className="eyebrow">{eyebrow}</span>}
        <h3 className="ui-chart-title">{title}</h3>
      </div>
      {unit === undefined ? null : (
        <span className="ui-chart-legend">
          <span className="ui-chart-dot" aria-hidden="true" />
          {unit}
        </span>
      )}
    </div>
  );

  if (points.length === 0) {
    return (
      <figure className="ui-chart ui-chart--empty">
        {heading}
        <output className="ui-chart-empty">{emptyLabel}</output>
      </figure>
    );
  }

  const max = points.reduce((highest, point) => Math.max(highest, point.value), 0);
  const top = niceCeil(max);
  const { x, y } = geometry(points, top);
  const baseline = viewHeight - padding;
  const barWidth = Math.max(2, (viewWidth - padding * 2) / points.length / 1.6);
  // Tick values from the axis top down to zero, evenly spaced.
  const ticks = Array.from(
    { length: tickCount },
    (_unused, i) => (top * (tickCount - 1 - i)) / (tickCount - 1),
  );
  const linePoints = points.map((point, index) => `${x(index)},${y(point.value)}`).join(' ');
  const areaPath = `M ${points
    .map((point, index) => `${x(index)} ${y(point.value)}`)
    .join(' L ')} L ${x(points.length - 1)} ${baseline} L ${x(0)} ${baseline} Z`;

  return (
    <figure className="ui-chart">
      {heading}
      <div className="ui-chart-layout">
        <div className="ui-chart-y" aria-hidden="true">
          {ticks.map((tick) => (
            <span key={tick}>{format(Math.round(tick))}</span>
          ))}
        </div>
        <div className="ui-chart-plot-wrap">
          <svg
            className="ui-chart-plot"
            viewBox={`0 0 ${viewWidth} ${viewHeight}`}
            preserveAspectRatio="none"
          >
            <title>{title}</title>
            {ticks.map((tick) => (
              <line
                key={tick}
                className="ui-chart-grid"
                x1={padding}
                y1={y(tick)}
                x2={viewWidth - padding}
                y2={y(tick)}
              />
            ))}
            {mode === 'line' ? (
              <>
                <path className="ui-chart-area" d={areaPath} />
                <polyline className="ui-chart-line" points={linePoints} />
              </>
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
          <div className="ui-chart-x" aria-hidden="true">
            {sampleLabels(points, maxXLabels).map((point) => (
              <span key={point.label}>{point.label}</span>
            ))}
          </div>
        </div>
      </div>
      <ul className="ui-visually-hidden">
        {points.map((point) => (
          <li key={point.label}>{`${point.label}: ${format(point.value)}`}</li>
        ))}
      </ul>
    </figure>
  );
}
