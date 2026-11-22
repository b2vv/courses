import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { formatLocale } from '../../../app/providers/format-locale';
import { AdminStatsQueryError } from '../../../app/composition/admin-stats-queries';
import type { createAdminStatsQueries } from '../../../app/composition/admin-stats-queries';
import type { Metric, Snapshot } from '../domain/stats';
import {
  ErrorState,
  PageHeader,
  SegmentedBar,
  ShareMeter,
  SegmentedControl,
  Skeleton,
  StatTile,
  Table,
  TrendChart,
} from '../../../presentation/ui';
import type { SplitPart, SplitTone, TableRow, TrendPoint } from '../../../presentation/ui';

type StatsQueries = ReturnType<typeof createAdminStatsQueries>;

type AnalyticsViewProps = {
  readonly snapshot: Snapshot;
  readonly stats: StatsQueries;
};

const metricKey: Record<Metric, string> = {
  signups: 'adminAnalytics.metricSignups',
  enrollments: 'adminAnalytics.metricEnrollments',
  orders: 'adminAnalytics.metricOrders',
  revenue: 'adminAnalytics.metricRevenue',
};

const metricOrder: readonly Metric[] = ['signups', 'enrollments', 'orders', 'revenue'];

function isMetric(value: string): value is Metric {
  return (metricOrder as readonly string[]).includes(value);
}

export function AnalyticsView({ snapshot, stats }: AnalyticsViewProps) {
  const { t, i18n } = useTranslation();
  const lang = formatLocale(i18n.resolvedLanguage);
  const num = new Intl.NumberFormat(lang);
  const money = new Intl.NumberFormat(lang, {
    style: 'currency',
    currency: 'UAH',
    maximumFractionDigits: 0,
  });
  const none = t('adminAnalytics.none');

  const uah = (minor: number) => money.format(minor / 100);
  const showInstructors = !snapshot.platform.singleInstructor;
  const avgDuration =
    snapshot.media.avgDurationSeconds === null
      ? none
      : num.format(Math.round(snapshot.media.avgDurationSeconds));
  const avgProgress =
    snapshot.engagement.avgProgress === null
      ? none
      : new Intl.NumberFormat(lang, { style: 'percent', maximumFractionDigits: 0 }).format(
          snapshot.engagement.avgProgress,
        );

  const leaderboardRows: readonly TableRow[] = snapshot.instructors.map((instructor) => ({
    id: instructor.instructorId,
    cells: [
      instructor.displayName,
      num.format(instructor.publishedCourses),
      num.format(instructor.activeEnrollments),
      uah(instructor.revenueMinorUAH),
    ],
  }));

  const share = new Intl.NumberFormat(lang, { style: 'percent', maximumFractionDigits: 0 });
  const formatShare = (value: number) => share.format(value);
  const part = (id: string, labelKey: string, value: number, tone: SplitTone): SplitPart => ({
    id,
    label: t(labelKey),
    value,
    display: num.format(value),
    tone,
  });
  const total = (parts: readonly SplitPart[]) =>
    num.format(parts.reduce((sum, item) => sum + item.value, 0));
  const empty = t('adminAnalytics.splitEmpty');

  // The whole is the live accounts (active + blocked), the same figure as the headline tile;
  // deleted accounts are listed under the bar, outside the whole.
  const accounts = [
    part('active', 'adminAnalytics.active', snapshot.users.active, 'success'),
    part('blocked', 'adminAnalytics.blocked', snapshot.users.blocked, 'warning'),
  ];
  const roles = [
    part('student', 'adminAnalytics.students', snapshot.users.byRole.student, 'series'),
    ...(showInstructors
      ? [
          part(
            'instructor',
            'adminAnalytics.instructors',
            snapshot.users.byRole.instructor,
            'neutral',
          ),
        ]
      : []),
    part('admin', 'adminAnalytics.admins', snapshot.users.byRole.admin, 'muted'),
  ];
  const courses = [
    part('published', 'adminAnalytics.partPublished', snapshot.courses.published, 'success'),
    part('draft', 'adminAnalytics.draft', snapshot.courses.draft, 'info'),
    part('archived', 'adminAnalytics.archived', snapshot.courses.archived, 'muted'),
  ];
  const grants = [
    part('active', 'adminAnalytics.active', snapshot.grants.active, 'success'),
    part('blocked', 'adminAnalytics.blocked', snapshot.grants.blocked, 'warning'),
    part('revoked', 'adminAnalytics.partRevoked', snapshot.grants.revoked, 'danger'),
  ];
  const orders = [
    part('paid', 'adminAnalytics.partPaid', snapshot.orders.paid, 'success'),
    part('pending', 'adminAnalytics.partPending', snapshot.orders.pending, 'warning'),
    part('failed', 'adminAnalytics.partFailed', snapshot.orders.failed, 'danger'),
  ];
  const media = [
    part('ready', 'adminAnalytics.partReady', snapshot.media.ready, 'success'),
    part('processing', 'adminAnalytics.partProcessing', snapshot.media.inProgress, 'info'),
    part('failed', 'adminAnalytics.partErrored', snapshot.media.failed, 'danger'),
  ];
  // Verified/2FA count every non-deleted account (active + blocked), like the role counts.
  const liveAccounts = snapshot.users.active + snapshot.users.blocked;
  const ofAccounts = (value: number) =>
    t('adminAnalytics.ofAccounts', {
      value: num.format(value),
      total: num.format(liveAccounts),
      share: liveAccounts > 0 ? share.format(value / liveAccounts) : none,
    });

  return (
    <section className="admin-analytics">
      <PageHeader
        eyebrow={t('adminAnalytics.eyebrow')}
        title={t('adminAnalytics.title')}
        description={t('adminAnalytics.lede')}
      />

      <section className="admin-analytics-group" aria-labelledby="analytics-headline">
        <h2 id="analytics-headline" className="ui-visually-hidden">
          {t('adminAnalytics.sectionPlatform')}
        </h2>
        <div className="admin-analytics-grid admin-analytics-kpis">
          <StatTile
            label={t('adminAnalytics.totalAccounts')}
            value={num.format(snapshot.users.active + snapshot.users.blocked)}
          />
          <StatTile
            label={t('adminAnalytics.activeLearners')}
            value={num.format(snapshot.engagement.activeLearners30d)}
          />
          <StatTile
            label={t('adminAnalytics.revenue')}
            value={uah(snapshot.orders.revenueMinorUAH)}
          />
          <StatTile
            label={t('adminAnalytics.openReports')}
            value={num.format(snapshot.reports.open)}
            meta={t('adminAnalytics.resolvedMeta', {
              count: snapshot.reports.resolved,
              value: num.format(snapshot.reports.resolved),
            })}
          />
        </div>
      </section>

      <TrendSection stats={stats} />

      <section className="admin-analytics-group">
        <div className="group-heading">
          <h2>{t('adminAnalytics.sectionSplits')}</h2>
          <span>{t('adminAnalytics.subSplits')}</span>
        </div>
        <div className="admin-analytics-splits">
          <SegmentedBar
            title={t('adminAnalytics.splitAccounts')}
            totalLabel={total(accounts)}
            parts={accounts}
            formatShare={formatShare}
            emptyLabel={empty}
          >
            <dl className="ui-split-facts">
              <div>
                <dt>{t('adminAnalytics.deleted')}</dt>
                <dd>{num.format(snapshot.users.deleted)}</dd>
              </div>
            </dl>
            <ShareMeter
              label={t('adminAnalytics.verifiedShare')}
              valueLabel={ofAccounts(snapshot.users.verified)}
              ratio={liveAccounts > 0 ? snapshot.users.verified / liveAccounts : null}
              tone="success"
            />
            <ShareMeter
              label={t('adminAnalytics.twoFactorShare')}
              valueLabel={ofAccounts(snapshot.users.twoFactor)}
              ratio={liveAccounts > 0 ? snapshot.users.twoFactor / liveAccounts : null}
              tone="success"
            />
          </SegmentedBar>
          {/* A single-instructor platform withholds the instructor count (admin-stats 9a), so the
              listed roles are not the whole: no total, bar or shares then. */}
          <SegmentedBar
            title={t('adminAnalytics.splitRoles')}
            {...(showInstructors ? { totalLabel: total(roles) } : {})}
            parts={roles}
            formatShare={formatShare}
            emptyLabel={empty}
            isWhole={showInstructors}
          />
          <SegmentedBar
            title={t('adminAnalytics.splitCourses')}
            totalLabel={total(courses)}
            parts={courses}
            formatShare={formatShare}
            emptyLabel={empty}
          />
          <SegmentedBar
            title={t('adminAnalytics.splitGrants')}
            totalLabel={total(grants)}
            parts={grants}
            formatShare={formatShare}
            emptyLabel={empty}
          />
          <SegmentedBar
            title={t('adminAnalytics.splitOrders')}
            totalLabel={total(orders)}
            parts={orders}
            formatShare={formatShare}
            emptyLabel={empty}
          />
          <SegmentedBar
            title={t('adminAnalytics.splitMedia')}
            totalLabel={total(media)}
            parts={media}
            formatShare={formatShare}
            emptyLabel={empty}
          >
            <dl className="ui-split-facts">
              <div>
                <dt>{t('adminAnalytics.totalDuration')}</dt>
                <dd>{num.format(Math.round(snapshot.media.totalDurationSeconds))}</dd>
              </div>
              <div>
                <dt>{t('adminAnalytics.avgDuration')}</dt>
                <dd>
                  {avgDuration}{' '}
                  <span>
                    {t('adminAnalytics.durationCoverage', {
                      count: snapshot.media.withDurationCount,
                    })}
                  </span>
                </dd>
              </div>
            </dl>
          </SegmentedBar>
        </div>
      </section>

      <section className="admin-analytics-group">
        <div className="group-heading">
          <h2>{t('adminAnalytics.sectionEngagement')}</h2>
          <span>{t('adminAnalytics.subEngagement')}</span>
        </div>
        <div className="ui-split admin-analytics-engagement">
          <ShareMeter
            label={t('adminAnalytics.avgProgress')}
            valueLabel={avgProgress}
            ratio={snapshot.engagement.avgProgress}
            tone="info"
          />
          <p className="admin-analytics-note">
            {t('adminAnalytics.completionsNote', {
              count: snapshot.engagement.completions,
              value: num.format(snapshot.engagement.completions),
            })}
          </p>
        </div>
      </section>

      {showInstructors ? (
        <section className="admin-analytics-group">
          <div className="group-heading">
            <h2>{t('adminAnalytics.sectionInstructors')}</h2>
            <span>{t('adminAnalytics.subInstructors')}</span>
          </div>
          <Table
            label={t('adminAnalytics.leaderboardCaption')}
            columns={[
              { id: 'instructor', label: t('adminAnalytics.colInstructor') },
              { id: 'courses', label: t('adminAnalytics.colCourses'), align: 'end' },
              { id: 'enrollments', label: t('adminAnalytics.colEnrollments'), align: 'end' },
              { id: 'revenue', label: t('adminAnalytics.colRevenue'), align: 'end' },
            ]}
            rows={leaderboardRows}
            emptyMessage={t('adminAnalytics.leaderboardEmpty')}
          />
        </section>
      ) : null}

      <section className="admin-analytics-group admin-analytics-deferred">
        <h2>{t('adminAnalytics.sectionDeferred')}</h2>
        <p className="admin-analytics-note">
          <strong>{t('adminAnalytics.deferredTitle')}</strong>
          {' — '}
          {t('adminAnalytics.deferredHint')}
        </p>
      </section>
    </section>
  );
}

function TrendSection({ stats }: { readonly stats: StatsQueries }) {
  const { t, i18n } = useTranslation();
  const lang = formatLocale(i18n.resolvedLanguage);
  const num = new Intl.NumberFormat(lang);
  const money = new Intl.NumberFormat(lang, {
    style: 'currency',
    currency: 'UAH',
    maximumFractionDigits: 0,
  });
  const [metric, setMetric] = useState<Metric>('signups');
  const query = useQuery(stats.timeseries({ metric, bucket: 'day' }));
  const points: readonly TrendPoint[] = (query.data ?? []).map((point) => ({
    label: point.bucketStart.slice(0, 10),
    value: point.value,
  }));
  const options = metricOrder.map((value) => ({ id: value, label: t(metricKey[value]) }));
  // Revenue arrives in minor units (kopiyky); every other metric is a whole count.
  const isRevenue = metric === 'revenue';
  const formatValue = isRevenue
    ? (value: number) => money.format(value / 100)
    : (value: number) => num.format(value);

  return (
    <section className="admin-analytics-group">
      <div className="group-heading">
        <h2>{t('adminAnalytics.sectionTrend')}</h2>
        <span>{t('adminAnalytics.subTrend')}</span>
      </div>
      <SegmentedControl
        label={t('adminAnalytics.trendMetricLabel')}
        options={options}
        selectedKeys={new Set([metric])}
        onSelectionChange={(keys) => {
          const next = String([...keys][0] ?? 'signups');
          if (isMetric(next)) setMetric(next);
        }}
      />
      {query.isPending ? (
        <Skeleton label={t('adminAnalytics.loading')} lines={3} />
      ) : (
        <TrendChart
          title={t(metricKey[metric])}
          mode={isRevenue ? 'bar' : 'line'}
          points={points}
          emptyLabel={t('adminAnalytics.trendEmpty')}
          formatValue={formatValue}
          {...(isRevenue ? { unit: t('adminAnalytics.currencyUAH') } : {})}
        />
      )}
    </section>
  );
}

export function AnalyticsPending() {
  const { t } = useTranslation();
  return <Skeleton label={t('adminAnalytics.loading')} lines={6} />;
}

export function AnalyticsFailure({ error, reset }: ErrorComponentProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const degraded = error instanceof AdminStatsQueryError && error.code === 'dependency_unavailable';
  return (
    <ErrorState
      title={t(degraded ? 'adminAnalytics.degraded' : 'adminAnalytics.failure')}
      retryLabel={t('common.retry')}
      onRetry={() => void router.invalidate().then(reset)}
    >
      <p>{t(degraded ? 'adminAnalytics.degradedHint' : 'adminAnalytics.failureHint')}</p>
    </ErrorState>
  );
}
