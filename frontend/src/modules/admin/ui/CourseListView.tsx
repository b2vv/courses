import { useEffect, useRef, useState } from 'react';
import type { FormEvent, ReactNode } from 'react';
import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { formatLocale } from '../../../app/providers/format-locale';
import { AuthoringQueryError } from '../../../app/composition/admin-queries';
import type {
  AdminCourse,
  AdminCoursePage,
  AdminCourseSort,
  CourseStatus,
} from '../../../contracts/course';
import type { AdminListState } from '../../../app/router/admin-search';
import {
  Button,
  DataTable,
  Dialog,
  EmptyState,
  ErrorState,
  Link,
  PageHeader,
  Pagination,
  SegmentedControl,
  Skeleton,
  SummaryStrip,
  TextField,
} from '../../../presentation/ui';
import type { DataColumn, DataTableSort } from '../../../presentation/ui';

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
  onClearFilters: () => void;
  onDelete: (course: AdminCourse) => Promise<void>;
  onAuthenticationRequired: () => void;
  createAction?: ReactNode;
};

export function CourseListView({
  page,
  state,
  onStatusChange,
  onSearch,
  onSortChange,
  onPageChange,
  onClearFilters,
  onDelete,
  onAuthenticationRequired,
  createAction,
}: CourseListViewProps) {
  const { t, i18n } = useTranslation();
  const locale = formatLocale(i18n.resolvedLanguage);
  // One destructive action at a time; a ref catches a second press within a render
  // synchronously (the buttons stay enabled, never native-disabled which drops focus).
  const inFlight = useRef(false);
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<'status' | 'alert'>('status');
  const messageRef = useRef<HTMLDivElement>(null);
  // Deleting a row unmounts its action buttons, so the browser drops focus to <body>;
  // move it to the result message (a focusable live region) rather than the page top.
  useEffect(() => {
    if (!message) return;
    const active = document.activeElement;
    if (active === null || active === document.body) messageRef.current?.focus();
  }, [message]);

  async function runDelete(course: AdminCourse) {
    if (inFlight.current) return;
    inFlight.current = true;
    setMessage(undefined);
    try {
      await onDelete(course);
      setMessageKind('status');
      setMessage(t('administration.deletedComplete'));
    } catch (error) {
      const code = error instanceof AuthoringQueryError ? error.code : undefined;
      setMessageKind('alert');
      setMessage(deleteErrorMessage(code, t));
      if (code === 'authentication_required') onAuthenticationRequired();
    } finally {
      inFlight.current = false;
    }
  }

  const currentPage = Math.floor(page.pagination.offset / page.pagination.limit) + 1;
  const totalPages = Math.ceil(page.pagination.total / page.pagination.limit);
  const dateFormatter = new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
  const priceFormatter = new Intl.NumberFormat(locale, { maximumFractionDigits: 2 });
  const selectedStatus = state.status ?? 'all';
  // An empty list under an active search or status means "no matches, adjust the filters",
  // not "no courses exist yet" — the two states read differently and offer a different way out.
  const isFiltered = state.search !== '' || state.status !== undefined;

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
      {message && (
        <div className="admin-editor-message" role={messageKind} ref={messageRef} tabIndex={-1}>
          {message}
        </div>
      )}
      <DataTable<AdminCourse>
        caption={t('administration.courseTable')}
        columns={columns}
        rows={page.items}
        rowId={(course) => course.id}
        actionsLabel={t('administration.actions')}
        actions={(course) => <CourseRowActions course={course} onConfirmDelete={runDelete} />}
        sort={{ key: state.sort, dir: naturalDirection[state.sort] }}
        onSortChange={(next) => onSortChange(sortFromKey(next.key))}
        mobileSortLabel={t('administration.sortMobile')}
        empty={
          <EmptyState
            title={t(isFiltered ? 'administration.emptyFiltered' : 'administration.empty')}
          >
            <p>{t(isFiltered ? 'administration.emptyFilteredHint' : 'administration.emptyHint')}</p>
            {isFiltered && (
              <Button variant="secondary" onPress={onClearFilters}>
                {t('administration.clearFilters')}
              </Button>
            )}
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

function CourseRowActions({
  course,
  onConfirmDelete,
}: {
  course: AdminCourse;
  onConfirmDelete: (course: AdminCourse) => Promise<void>;
}) {
  const { t } = useTranslation();
  return (
    <>
      <Link
        href={`/admin/courses/${course.id}`}
        className="admin-row-action"
        aria-label={t('administration.editAction', { name: course.title })}
      >
        <ActionIcon name="edit" />
      </Link>
      {course.status === 'published' && (
        <Link
          href={`/courses/${course.slug}`}
          className="admin-row-action"
          aria-label={t('administration.previewAction', { name: course.title })}
        >
          <ActionIcon name="preview" />
        </Link>
      )}
      {course.status === 'draft' && (
        <Dialog
          triggerLabel={t('administration.deleteCourse')}
          trigger={
            <Button
              type="button"
              variant="danger"
              aria-label={t('administration.deleteAction', { name: course.title })}
            >
              <ActionIcon name="delete" />
            </Button>
          }
          title={t('administration.deleteCourse')}
          closeLabel={t('common.cancel')}
        >
          {(close) => (
            <>
              <p>{t('administration.confirmDeleteBody', { name: course.title })}</p>
              <Button
                variant="danger"
                onPress={() => {
                  void onConfirmDelete(course);
                  close();
                }}
              >
                {t('administration.confirmDeleteCourse')}
              </Button>
            </>
          )}
        </Dialog>
      )}
    </>
  );
}

function ActionIcon({ name }: { name: 'edit' | 'preview' | 'delete' }) {
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
      {name === 'edit' && <path d="M11.5 2.5l2 2L6 12l-2.5.5L4 10z" />}
      {name === 'preview' && (
        <>
          <path d="M1.5 8S4 3.5 8 3.5 14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8z" />
          <circle cx="8" cy="8" r="1.75" />
        </>
      )}
      {name === 'delete' && (
        <path d="M2.5 4.5h11M6 4.5V2.5h4v2M4 4.5l.5 9h7l.5-9M6.5 7v4M9.5 7v4" />
      )}
    </svg>
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

function deleteErrorMessage(
  code: string | undefined,
  t: ReturnType<typeof useTranslation>['t'],
): string {
  switch (code) {
    case 'course_not_deletable':
      return t('administration.deleteErrors.not_deletable');
    case 'course_modified':
      return t('administration.deleteErrors.modified');
    case 'course_not_found':
      return t('administration.deleteErrors.not_found');
    case 'authentication_required':
      return t('administration.deleteErrors.authentication_required');
    default:
      return t('administration.deleteErrors.default');
  }
}
