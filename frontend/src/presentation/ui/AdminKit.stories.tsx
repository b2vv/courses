import type { Meta, StoryObj } from '@storybook/react';
import { Button, Delta, LoadMeter, SegmentedControl, StatTile, Table, TrendChart } from './index';
import { AdminNavIcon } from '../admin/AdminNavIcon';

/**
 * The admin UI kit rendered on the NorthWind dark surface (the preview wraps every
 * story in `data-surface="admin"`). This is the live, in-repo counterpart to the
 * browsable design-system artifact. Admin paths only.
 */
const meta: Meta = {
  title: 'Admin/Kit',
  parameters: { layout: 'padded' },
};
export default meta;

type Story = StoryObj;

export const Buttons: Story = {
  render: () => (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-3)', alignItems: 'center' }}>
      <Button variant="primary">Створити курс</Button>
      <Button variant="secondary">Призупинити</Button>
      <Button variant="danger">Видалити назавжди</Button>
    </div>
  ),
};

export const StatusPills: Story = {
  render: () => (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-2)', alignItems: 'center' }}>
      <span className="ui-pill ui-pill--success">Активний</span>
      <span className="ui-pill ui-pill--info">У процесі</span>
      <span className="ui-pill ui-pill--danger">Заблокований</span>
      <span className="ui-pill ui-pill--neutral">Неактивний</span>
    </div>
  ),
};

export const Meters: Story = {
  render: () => (
    <div style={{ display: 'grid', gap: 'var(--space-4)', maxWidth: '22rem' }}>
      <LoadMeter label="Прогрес" tone="success" value={72} />
      <LoadMeter label="Заповнення" tone="warning" value={79} />
      <LoadMeter label="Навантаження" tone="danger" value={91} />
      <LoadMeter label="Базовий" tone="neutral" value={38} />
    </div>
  ),
};

export const Metrics: Story = {
  render: () => (
    <div className="ui-stat" style={{ maxWidth: '26rem' }}>
      <span
        className="ui-stat-label ui-mono"
        style={{ textTransform: 'uppercase', letterSpacing: '0.08em' }}
      >
        Активні студенти · Усі курси
      </span>
      <div
        className="ui-stat-row"
        style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'center' }}
      >
        <span className="ui-stat-value">4 820</span>
        <Delta label="проти минулого тижня" value="6.2%" direction="up" tone="success" />
      </div>
    </div>
  ),
};

export const StatTiles: Story = {
  render: () => (
    <div
      style={{
        display: 'grid',
        gap: 'var(--space-4)',
        gridTemplateColumns: 'repeat(auto-fit, minmax(14rem, 1fr))',
      }}
    >
      <StatTile label="Усього акаунтів" value="4 820" />
      <StatTile label="Активні учні · 30 днів" value="1 204" meta="по курсах з прогресом" />
      <StatTile label="Середня тривалість відео" value="—" meta="нема покриття" />
      <StatTile label="Дохід, грн" value="128 400">
        <Delta label="проти минулого місяця" value="6.2%" direction="up" tone="success" />
      </StatTile>
    </div>
  ),
};

export const Trends: Story = {
  render: () => (
    <div
      style={{
        display: 'grid',
        gap: 'var(--space-6)',
        gridTemplateColumns: 'repeat(auto-fit, minmax(18rem, 1fr))',
      }}
    >
      <TrendChart
        title="Реєстрації · день"
        mode="line"
        emptyLabel="Немає даних у вікні"
        points={[
          { label: '01.03', value: 4 },
          { label: '02.03', value: 9 },
          { label: '03.03', value: 6 },
          { label: '04.03', value: 12 },
          { label: '05.03', value: 8 },
        ]}
      />
      <TrendChart
        title="Замовлення · тиждень"
        mode="bar"
        emptyLabel="Немає даних у вікні"
        points={[
          { label: 'Т1', value: 3 },
          { label: 'Т2', value: 7 },
          { label: 'Т3', value: 5 },
          { label: 'Т4', value: 9 },
        ]}
      />
      <TrendChart title="Дохід · день" mode="line" emptyLabel="Немає даних у вікні" points={[]} />
    </div>
  ),
};

export const Roster: Story = {
  render: () => (
    <Table
      label="Ростер учнів"
      columns={[
        { id: 'learner', label: 'Учень' },
        { id: 'progress', label: 'Прогрес' },
        { id: 'lessons', label: 'Уроки' },
        { id: 'status', label: 'Статус' },
      ]}
      rows={[
        {
          id: '1',
          cells: [
            <span className="ui-mono" key="a">
              olena.k
            </span>,
            <LoadMeter key="b" label="Прогрес olena.k" tone="success" value={72} />,
            <span className="ui-mono" key="c">
              13/18
            </span>,
            <span className="ui-pill ui-pill--success" key="d">
              Активний
            </span>,
          ],
        },
        {
          id: '2',
          cells: [
            <span className="ui-mono" key="a">
              marko.p
            </span>,
            <LoadMeter key="b" label="Прогрес marko.p" tone="warning" value={41} />,
            <span className="ui-mono" key="c">
              7/18
            </span>,
            <span className="ui-pill ui-pill--neutral" key="d">
              Застопорений
            </span>,
          ],
        },
        {
          id: '3',
          cells: [
            <span className="ui-mono" key="a">
              ihor.v
            </span>,
            <LoadMeter key="b" label="Прогрес ihor.v" tone="danger" value={12} />,
            <span className="ui-mono" key="c">
              2/18
            </span>,
            <span className="ui-pill ui-pill--danger" key="d">
              Заблокований
            </span>,
          ],
        },
      ]}
      emptyMessage="Ще немає учнів"
    />
  ),
};

export const Segmented: Story = {
  render: () => (
    <SegmentedControl
      label="Діапазон"
      defaultSelectedKeys={['24h']}
      options={[
        { id: '1h', label: '1г' },
        { id: '24h', label: '24г' },
        { id: '7d', label: '7д' },
        { id: '30d', label: '30д' },
      ]}
    />
  ),
};

/**
 * Admin shell chrome — the real `.admin-rail` + `.admin-sidebar` structure from
 * AdminShell (S0). The active destination carries the frosted blue gradient +
 * gloss + glow on BOTH the rail glyph and the sidebar link. Links are inert `<a>`
 * (no router in Storybook); classes, icons and markup mirror AdminShell exactly.
 */
export const Shell: Story = {
  render: () => (
    <div className="admin-shell" style={{ minBlockSize: 0, maxWidth: '19rem' }}>
      <div
        className="admin-rail"
        aria-hidden="true"
        style={{ position: 'static', blockSize: 'auto' }}
      >
        <a className="admin-rail-brand" href="/admin" tabIndex={-1}>
          <span aria-hidden="true">{'</>'}</span>
        </a>
        <div className="admin-rail-items">
          <a
            className="admin-rail-item"
            href="/admin"
            aria-current="page"
            tabIndex={-1}
            title="Курси"
          >
            <AdminNavIcon name="overview" />
          </a>
          <a className="admin-rail-item" href="/admin/learners" tabIndex={-1} title="Учні">
            <AdminNavIcon name="people" />
          </a>
          <a className="admin-rail-item" href="/admin/users" tabIndex={-1} title="Користувачі">
            <AdminNavIcon name="accounts" />
          </a>
          <a
            className="admin-rail-item"
            href="/admin/participant-reports"
            tabIndex={-1}
            title="Скарги"
          >
            <AdminNavIcon name="moderation" />
          </a>
        </div>
      </div>
      <aside
        className="admin-sidebar"
        style={{ position: 'static', blockSize: 'auto', gridColumn: 'span 2' }}
      >
        <div className="admin-workspace">
          <span className="admin-workspace-name">CourseForge</span>
          <span className="admin-workspace-area">Адмінка</span>
        </div>
        <nav className="admin-nav" aria-label="Адмінка">
          <div className="admin-nav-section">
            <a className="admin-nav-link" href="/admin" aria-current="page">
              <AdminNavIcon name="overview" />
              <span>Курси</span>
            </a>
          </div>
          <div className="admin-nav-section">
            <span className="admin-nav-section-label">Люди</span>
            <a className="admin-nav-link" href="/admin/learners">
              <AdminNavIcon name="people" />
              <span>Учні</span>
            </a>
            <a className="admin-nav-link" href="/admin/users">
              <AdminNavIcon name="accounts" />
              <span>Користувачі</span>
            </a>
          </div>
          <div className="admin-nav-section">
            <span className="admin-nav-section-label">Модерація</span>
            <a className="admin-nav-link" href="/admin/participant-reports">
              <AdminNavIcon name="moderation" />
              <span>Скарги учасників</span>
            </a>
          </div>
        </nav>
        <a className="admin-to-client" href="/">
          <span aria-hidden="true">←</span> На клієнт
        </a>
      </aside>
    </div>
  ),
};
