import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { AdminStatsQueryError } from '../../app/composition/admin-stats-queries';
import type { createAdminStatsQueries } from '../../app/composition/admin-stats-queries';
import type { Metric, Snapshot } from '../../domain/admin/stats';
import {
  EmptyState,
  ErrorState,
  SegmentedControl,
  Skeleton,
  StatTile,
  Table,
  TrendChart,
} from '../ui';
import type { TableRow, TrendPoint } from '../ui';

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
  const lang = i18n.resolvedLanguage === 'en' ? 'en' : 'uk-UA';
  const num = new Intl.NumberFormat(lang);
  const money = new Intl.NumberFormat(lang, {
    style: 'currency',
    currency: 'UAH',
    maximumFractionDigits: 0,
  });
  const percent = new Intl.NumberFormat(lang, { style: 'percent', maximumFractionDigits: 0 });
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
      : percent.format(snapshot.engagement.avgProgress);

  const leaderboardRows: readonly TableRow[] = snapshot.instructors.map((instructor) => ({
    id: instructor.instructorId,
    cells: [
      instructor.displayName,
      num.format(instructor.publishedCourses),
      num.format(instructor.activeEnrollments),
      uah(instructor.revenueMinorUAH),
    ],
  }));

  return (
    <section className="admin-analytics">
      <div className="admin-list-heading">
        <div>
          <p className="eyebrow">{t('adminAnalytics.eyebrow')}</p>
          <h1>{t('adminAnalytics.title')}</h1>
        </div>
      </div>

      <div className="admin-analytics-grid">
        <StatTile
          label={t('adminAnalytics.totalAccounts')}
          value={num.format(snapshot.users.active + snapshot.users.blocked)}
        />
        <StatTile
          label={t('adminAnalytics.published')}
          value={num.format(snapshot.courses.published)}
        />
        <StatTile
          label={t('adminAnalytics.revenue')}
          value={uah(snapshot.orders.revenueMinorUAH)}
        />
        <StatTile
          label={t('adminAnalytics.openReports')}
          value={num.format(snapshot.reports.open)}
        />
      </div>

      <section className="admin-analytics-group">
        <h2>{t('adminAnalytics.sectionPeople')}</h2>
        <div className="admin-analytics-grid">
          <StatTile label={t('adminAnalytics.active')} value={num.format(snapshot.users.active)} />
          <StatTile
            label={t('adminAnalytics.blocked')}
            value={num.format(snapshot.users.blocked)}
          />
          <StatTile
            label={t('adminAnalytics.deleted')}
            value={num.format(snapshot.users.deleted)}
          />
          <StatTile
            label={t('adminAnalytics.verified')}
            value={num.format(snapshot.users.verified)}
          />
          <StatTile
            label={t('adminAnalytics.twoFactor')}
            value={num.format(snapshot.users.twoFactor)}
          />
          <StatTile
            label={t('adminAnalytics.students')}
            value={num.format(snapshot.users.byRole.student)}
          />
          {showInstructors ? (
            <StatTile
              label={t('adminAnalytics.instructors')}
              value={num.format(snapshot.users.byRole.instructor)}
            />
          ) : null}
          <StatTile
            label={t('adminAnalytics.admins')}
            value={num.format(snapshot.users.byRole.admin)}
          />
        </div>
      </section>

      <section className="admin-analytics-group">
        <h2>{t('adminAnalytics.sectionContent')}</h2>
        <div className="admin-analytics-grid">
          <StatTile label={t('adminAnalytics.draft')} value={num.format(snapshot.courses.draft)} />
          <StatTile
            label={t('adminAnalytics.published')}
            value={num.format(snapshot.courses.published)}
          />
          <StatTile
            label={t('adminAnalytics.archived')}
            value={num.format(snapshot.courses.archived)}
          />
        </div>
      </section>

      <section className="admin-analytics-group">
        <h2>{t('adminAnalytics.sectionAccess')}</h2>
        <div className="admin-analytics-grid">
          <StatTile
            label={t('adminAnalytics.grantsActive')}
            value={num.format(snapshot.grants.active)}
          />
          <StatTile
            label={t('adminAnalytics.grantsBlocked')}
            value={num.format(snapshot.grants.blocked)}
          />
          <StatTile
            label={t('adminAnalytics.grantsRevoked')}
            value={num.format(snapshot.grants.revoked)}
          />
          <StatTile
            label={t('adminAnalytics.ordersPending')}
            value={num.format(snapshot.orders.pending)}
          />
          <StatTile
            label={t('adminAnalytics.ordersPaid')}
            value={num.format(snapshot.orders.paid)}
          />
          <StatTile
            label={t('adminAnalytics.ordersFailed')}
            value={num.format(snapshot.orders.failed)}
          />
        </div>
      </section>

      <section className="admin-analytics-group">
        <h2>{t('adminAnalytics.sectionEngagement')}</h2>
        <div className="admin-analytics-grid">
          <StatTile
            label={t('adminAnalytics.activeLearners')}
            value={num.format(snapshot.engagement.activeLearners30d)}
          />
          <StatTile label={t('adminAnalytics.avgProgress')} value={avgProgress} />
          <StatTile
            label={t('adminAnalytics.completions')}
            value={num.format(snapshot.engagement.completions)}
          />
        </div>
      </section>

      <section className="admin-analytics-group">
        <h2>{t('adminAnalytics.sectionOperations')}</h2>
        <div className="admin-analytics-grid">
          <StatTile
            label={t('adminAnalytics.mediaInProgress')}
            value={num.format(snapshot.media.inProgress)}
          />
          <StatTile
            label={t('adminAnalytics.mediaReady')}
            value={num.format(snapshot.media.ready)}
          />
          <StatTile
            label={t('adminAnalytics.mediaFailed')}
            value={num.format(snapshot.media.failed)}
          />
          <StatTile
            label={t('adminAnalytics.totalDuration')}
            value={num.format(Math.round(snapshot.media.totalDurationSeconds))}
          />
          <StatTile
            label={t('adminAnalytics.avgDuration')}
            value={avgDuration}
            meta={t('adminAnalytics.durationCoverage', { count: snapshot.media.withDurationCount })}
          />
        </div>
      </section>

      <section className="admin-analytics-group">
        <h2>{t('adminAnalytics.sectionModeration')}</h2>
        <div className="admin-analytics-grid">
          <StatTile
            label={t('adminAnalytics.openReports')}
            value={num.format(snapshot.reports.open)}
          />
          <StatTile
            label={t('adminAnalytics.resolvedReports')}
            value={num.format(snapshot.reports.resolved)}
          />
        </div>
      </section>

      {showInstructors ? (
        <section className="admin-analytics-group">
          <h2>{t('adminAnalytics.sectionInstructors')}</h2>
          <Table
            label={t('adminAnalytics.leaderboardCaption')}
            columns={[
              { id: 'instructor', label: t('adminAnalytics.colInstructor') },
              { id: 'courses', label: t('adminAnalytics.colCourses') },
              { id: 'enrollments', label: t('adminAnalytics.colEnrollments') },
              { id: 'revenue', label: t('adminAnalytics.colRevenue') },
            ]}
            rows={leaderboardRows}
            emptyMessage={t('adminAnalytics.leaderboardEmpty')}
          />
        </section>
      ) : null}

      <TrendSection stats={stats} />

      <section className="admin-analytics-group">
        <h2>{t('adminAnalytics.sectionDeferred')}</h2>
        <EmptyState title={t('adminAnalytics.deferredTitle')}>
          <p>{t('adminAnalytics.deferredHint')}</p>
        </EmptyState>
      </section>
    </section>
  );
}

function TrendSection({ stats }: { readonly stats: StatsQueries }) {
  const { t } = useTranslation();
  const [metric, setMetric] = useState<Metric>('signups');
  const query = useQuery(stats.timeseries({ metric, bucket: 'day' }));
  const points: readonly TrendPoint[] = (query.data ?? []).map((point) => ({
    label: point.bucketStart.slice(0, 10),
    value: point.value,
  }));
  const options = metricOrder.map((value) => ({ id: value, label: t(metricKey[value]) }));

  return (
    <section className="admin-analytics-group">
      <h2>{t('adminAnalytics.sectionTrend')}</h2>
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
          mode={metric === 'revenue' ? 'bar' : 'line'}
          points={points}
          emptyLabel={t('adminAnalytics.trendEmpty')}
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
