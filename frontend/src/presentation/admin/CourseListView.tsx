import { useState } from 'react';
import type { FormEvent, ReactNode } from 'react';
import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { formatLocale } from '../../app/providers/format-locale';
import type {
  AdminCourse,
  AdminCoursePage,
  AdminCourseSort,
  CourseStatus,
} from '../../domain/course/course';
import type { AdminListState } from '../../app/router/admin-search';
import {
  Button,
  DataTable,
  EmptyState,
  ErrorState,
  Link,
  PageHeader,
  Pagination,
  SegmentedControl,
  Skeleton,
  SummaryStrip,
  TextField,
} from '../ui';
import type { DataColumn, DataTableSort } from '../ui';

const statusVariant: Record<CourseStatus, string> = {
  draft: 'warning',
  published: 'success',
  archived: 'neutral',
};

// Each admin sort has one natural direction; the header reflects it and never fakes a toggle.
const naturalDirection: Record<AdminCourseSort, DataTableSort['dir']> = {
  recent: 'desc',
  title: 'asc',
  students: 'desc',
  price: 'desc',
};

type CourseListViewProps = {
  page: AdminCoursePage;
  state: AdminListState;
  onStatusChange: (status: CourseStatus | undefined) => void;
  onSearch: (query: string) => void;
  onSortChange: (sort: AdminCourseSort) => void;
  onPageChange: (page: number) => void;
  createAction?: ReactNode;
};

export function CourseListView({
  page,
  state,
  onStatusChange,
  onSearch,
  onSortChange,
  onPageChange,
  createAction,
}: CourseListViewProps) {
  const { t, i18n } = useTranslation();
  const locale = formatLocale(i18n.resolvedLanguage);
  const currentPage = Math.floor(page.pagination.offset / page.pagination.limit) + 1;
  const totalPages = Math.ceil(page.pagination.total / page.pagination.limit);
  const dateFormatter = new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
  const priceFormatter = new Intl.NumberFormat(locale, { maximumFractionDigits: 2 });
  const selectedStatus = state.status ?? 'all';

  const statusOptions = (['all', 'draft', 'published', 'archived'] as const).map((status) => ({
    id: status,
    label: t(status === 'all' ? 'administration.allStatuses' : `administration.${status}`),
  }));

  const columns: readonly DataColumn<AdminCourse>[] = [
    {
      key: 'title',
      label: t('administration.courseTitle'),
      identifier: true,
      render: (course) => (
        <Link href={`/admin/courses/${course.id}`} className="table-course-link">
          <span className="ui-icon-tile">
            <CourseGlyph />
          </span>
          <span>
            <strong>{course.title}</strong>
            <small>{t('administration.lessonsSub', { count: course.lessonCount })}</small>
          </span>
        </Link>
      ),
    },
    {
      key: 'status',
      label: t('administration.status'),
      sortable: false,
      render: (course) => (
        <span
          className={`ui-pill ui-pill--${statusVariant[course.status]}`}
          data-status={course.status}
        >
          {t(`administration.${course.status}`)}
        </span>
      ),
    },
    {
      key: 'students',
      label: t('administration.students'),
      numeric: true,
      render: (course) => course.studentCount,
    },
    {
      key: 'price',
      label: t('administration.priceColumn'),
      numeric: true,
      render: (course) =>
        course.price === null ? (
          <span className="course-price-free">{t('administration.free')}</span>
        ) : (
          <>
            {priceFormatter.format(course.price.minor / 100)}{' '}
            <small>{t('administration.priceUnit')}</small>
          </>
        ),
    },
    {
      key: 'recent',
      label: t('administration.updated'),
      render: (course) => (
        <time dateTime={course.updatedAt}>{dateFormatter.format(new Date(course.updatedAt))}</time>
      ),
    },
  ];

  return (
    <section className="admin-course-list" data-surface="admin">
      <PageHeader
        eyebrow={t('administration.eyebrow')}
        title={t('administration.courses')}
        description={t('administration.lede')}
        actions={createAction}
      />
      <SummaryStrip
        label={t('administration.summaryLabel')}
        items={[
          {
            id: 'total',
            value: String(page.statusCounts.total),
            label: t('administration.courses'),
          },
          {
            id: 'published',
            value: String(page.statusCounts.published),
            label: t('administration.published'),
            dot: true,
          },
          {
            id: 'drafts',
            value: String(page.statusCounts.draft),
            label: t('administration.summaryDrafts'),
          },
        ]}
      />
      <div className="admin-filter-bar">
        <CourseSearch key={state.search} initialQuery={state.search} onSearch={onSearch} />
        <SegmentedControl
          label={t('administration.statusTabs')}
          options={statusOptions}
          selectedKeys={new Set([selectedStatus])}
          onSelectionChange={(keys) => onStatusChange(statusFromKey(String([...keys][0] ?? 'all')))}
        />
        <span className="admin-filter-count">
          {t('administration.courseCount', { count: page.pagination.total })}
        </span>
      </div>
      <DataTable<AdminCourse>
        caption={t('administration.courseTable')}
        columns={columns}
        rows={page.items}
        rowId={(course) => course.id}
        actionsLabel={t('administration.actions')}
        sort={{ key: state.sort, dir: naturalDirection[state.sort] }}
        onSortChange={(next) => onSortChange(sortFromKey(next.key))}
        mobileSortLabel={t('administration.sortMobile')}
        empty={
          <EmptyState title={t('administration.empty')}>
            <p>{t('administration.emptyHint')}</p>
          </EmptyState>
        }
        footer={
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
        }
      />
    </section>
  );
}

function CourseSearch({
  initialQuery,
  onSearch,
}: {
  initialQuery: string;
  onSearch: (query: string) => void;
}) {
  const { t } = useTranslation();
  const [query, setQuery] = useState(initialQuery);
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSearch(query);
  }
  return (
    <search>
      <form className="admin-user-search" onSubmit={submit}>
        <TextField
          label={t('administration.searchLabel')}
          placeholder={t('administration.searchPlaceholder')}
          value={query}
          onChange={setQuery}
        />
        <Button type="submit">{t('administration.searchAction')}</Button>
      </form>
    </search>
  );
}

function CourseGlyph() {
  return (
    <svg
      viewBox="0 0 16 16"
      width="16"
      height="16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M3 3.5h6.5a2 2 0 0 1 2 2V13H5a2 2 0 0 0-2 2z" />
      <path d="M13 3.5h-1.5a2 2 0 0 0-2 2V13" />
    </svg>
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

function sortFromKey(key: string): AdminCourseSort {
  switch (key) {
    case 'title':
    case 'students':
    case 'price':
      return key;
    default:
      return 'recent';
  }
}
