import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { formatLocale } from '../../app/providers/format-locale';
import type { ReactNode } from 'react';
import type { AdminCourse, AdminCoursePage, CourseStatus } from '../../domain/course/course';
import type { AdminListState } from '../../app/router/admin-search';
import { EmptyState, ErrorState, Link, Pagination, Skeleton, Table, Tabs } from '../ui';

type CourseListViewProps = {
  page: AdminCoursePage;
  state: AdminListState;
  onStatusChange: (status: CourseStatus | undefined) => void;
  onPageChange: (page: number) => void;
  createAction?: ReactNode;
};

export function CourseListView({
  page,
  state,
  onStatusChange,
  onPageChange,
  createAction,
}: CourseListViewProps) {
  const { t, i18n } = useTranslation();
  const currentPage = Math.floor(page.pagination.offset / page.pagination.limit) + 1;
  const totalPages = Math.ceil(page.pagination.total / page.pagination.limit);
  const dateFormatter = new Intl.DateTimeFormat(formatLocale(i18n.resolvedLanguage), {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
  const rows = page.items.map((course) => courseRow(course, dateFormatter, t));
  const selectedStatus = state.status ?? 'all';
  const results = (
    <div className="admin-list-results">
      {page.items.length === 0 ? (
        <EmptyState title={t('administration.empty')}>
          <p>{t('administration.emptyHint')}</p>
        </EmptyState>
      ) : (
        <Table
          label={t('administration.courseTable')}
          columns={[
            { id: 'title', label: t('administration.courseTitle') },
            { id: 'slug', label: t('administration.slug') },
            { id: 'status', label: t('administration.status') },
            { id: 'owner', label: t('administration.owner') },
            { id: 'updated', label: t('administration.updated') },
          ]}
          rows={rows}
          emptyMessage={t('ui.noRecords')}
        />
      )}
      <Pagination
        label={t('administration.pages')}
        previousLabel={t('ui.back')}
        nextLabel={t('ui.forward')}
        pageLabel={t('ui.pageOf', { page: currentPage, total: totalPages })}
        emptyLabel={t('ui.noPages')}
        page={currentPage}
        totalPages={totalPages}
        onPageChange={onPageChange}
      />
    </div>
  );
  const tabs = [
    {
      id: 'all',
      label: t('administration.allStatuses'),
      content: selectedStatus === 'all' ? results : null,
    },
    {
      id: 'draft',
      label: t('administration.draft'),
      content: selectedStatus === 'draft' ? results : null,
    },
    {
      id: 'published',
      label: t('administration.published'),
      content: selectedStatus === 'published' ? results : null,
    },
    {
      id: 'archived',
      label: t('administration.archived'),
      content: selectedStatus === 'archived' ? results : null,
    },
  ];

  return (
    <section className="admin-course-list">
      <div className="admin-list-heading">
        <div>
          <p className="eyebrow">{t('administration.courses')}</p>
          <h1>{t('administration.title')}</h1>
        </div>
        <div className="admin-list-actions">
          <span className="admin-list-count">
            {t('administration.courseCount', { count: page.pagination.total })}
          </span>
          {createAction}
        </div>
      </div>
      <Tabs
        label={t('administration.statusTabs')}
        selectedKey={selectedStatus}
        onSelectionChange={(key) => onStatusChange(statusFromKey(String(key)))}
        items={tabs}
      />
    </section>
  );
}

export function CourseListPending() {
  const { t } = useTranslation();
  return <Skeleton label={t('administration.loading')} lines={5} />;
}

export function CourseListFailure({ reset }: ErrorComponentProps) {
  const { t } = useTranslation();
  const router = useRouter();
  return (
    <ErrorState
      title={t('administration.failure')}
      retryLabel={t('common.retry')}
      onRetry={() => void router.invalidate().then(reset)}
    >
      <p>{t('administration.failureHint')}</p>
    </ErrorState>
  );
}

function courseRow(
  course: AdminCourse,
  dateFormatter: Intl.DateTimeFormat,
  t: ReturnType<typeof useTranslation>['t'],
) {
  return {
    id: course.id,
    cells: [
      <Link key="title" href={`/admin/courses/${course.id}`}>
        {course.title}
      </Link>,
      <code key="slug">{course.slug}</code>,
      <span key="status" className="admin-course-status" data-status={course.status}>
        {statusLabel(course.status, t)}
      </span>,
      <code key="owner">{course.instructorId}</code>,
      <time key="updated" dateTime={course.updatedAt}>
        {dateFormatter.format(new Date(course.updatedAt))}
      </time>,
    ],
  };
}

function statusFromKey(key: string): CourseStatus | undefined {
  switch (key) {
    case 'draft':
    case 'published':
    case 'archived':
      return key;
    default:
      return undefined;
  }
}

function statusLabel(status: CourseStatus, t: ReturnType<typeof useTranslation>['t']) {
  switch (status) {
    case 'draft':
      return t('administration.draft');
    case 'published':
      return t('administration.published');
    case 'archived':
      return t('administration.archived');
  }
}
