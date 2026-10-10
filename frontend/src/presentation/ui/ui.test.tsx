import { afterEach, describe, expect, it } from '@rstest/core';
import axe from 'axe-core';
import { cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {
  Button,
  Checkbox,
  ComboBox,
  Dialog,
  Delta,
  ErrorState,
  LoadMeter,
  Menu,
  Pagination,
  ProgressBar,
  Select,
  SegmentedControl,
  Skeleton,
  StatTile,
  Switch,
  Table,
  Tabs,
  TextField,
  Toast,
  TrendChart,
} from './index';

afterEach(cleanup);

describe('accessible UI primitives', () => {
  it('labels a text field and reaches the action with the keyboard', async () => {
    const user = userEvent.setup();
    let submitted = false;
    render(
      <>
        <TextField label="Назва курсу" description="Наприклад, Архітектура Go" />
        <Button
          onPress={() => {
            submitted = true;
          }}
        >
          Зберегти зміни
        </Button>
      </>,
    );
    await user.tab();
    expect(document.activeElement).toBe(screen.getByRole('textbox', { name: 'Назва курсу' }));
    await user.keyboard('Основи Go');
    expect(screen.getByRole('textbox').getAttribute('aria-describedby')).toBeTruthy();
    await user.tab();
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Зберегти зміни' }));
    await user.keyboard('{Enter}');
    expect(submitted).toBe(true);
  });

  it('names a modal, closes it with Escape, and restores focus to its trigger', async () => {
    const user = userEvent.setup();
    render(
      <Dialog triggerLabel="Редагувати назву" title="Назва курсу" closeLabel="Закрити">
        <TextField label="Нова назва" />
      </Dialog>,
    );
    const trigger = screen.getByRole('button', { name: 'Редагувати назву' });
    await user.click(trigger);
    const dialog = await screen.findByRole('dialog', { name: 'Назва курсу' });
    expect(dialog.contains(document.activeElement)).toBe(true);
    await user.tab();
    await user.tab();
    expect(dialog.contains(document.activeElement)).toBe(true);
    await user.keyboard('{Escape}');
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    await waitFor(() => expect(document.activeElement).toBe(trigger));
  });

  it('opens a labelled select and chooses a named option', async () => {
    const user = userEvent.setup();
    render(
      <Select
        label="Порядок курсів"
        options={[
          { id: 'recent', label: 'Нові спочатку' },
          { id: 'title', label: 'За назвою' },
        ]}
      />,
    );
    await user.tab();
    await user.keyboard('{Enter}');
    const listbox = await screen.findByRole('listbox', { name: 'Порядок курсів' });
    await user.click(within(listbox).getByRole('option', { name: 'За назвою' }));
    await waitFor(() => expect(screen.queryByRole('listbox')).toBeNull());
    expect(screen.getByRole('button').textContent).toContain('За назвою');
    expect(document.activeElement).toBe(screen.getByRole('button'));
  });

  it('toggles checkbox and switch with Space without depending on pointer input', async () => {
    const user = userEvent.setup();
    render(
      <>
        <Checkbox>Показувати завершені</Checkbox>
        <Switch>Темна тема</Switch>
      </>,
    );
    await user.tab();
    await user.keyboard(' ');
    expect(
      screen.getByRole('checkbox', { name: 'Показувати завершені', checked: true }),
    ).toBeTruthy();
    await user.tab();
    await user.keyboard(' ');
    expect(screen.getByRole('switch', { name: 'Темна тема', checked: true })).toBeTruthy();
  });

  it('disables pagination beyond the last page and still allows going back', async () => {
    const user = userEvent.setup();
    let requestedPage = 3;
    render(
      <Pagination
        label="Сторінки результатів"
        previousLabel="Назад"
        nextLabel="Далі"
        pageLabel="Сторінка 3 з 3"
        emptyLabel="Немає сторінок"
        page={3}
        totalPages={3}
        onPageChange={(page) => {
          requestedPage = page;
        }}
      />,
    );
    expect(screen.getByRole('button', { name: 'Далі' }).hasAttribute('disabled')).toBe(true);
    await user.click(screen.getByRole('button', { name: 'Назад' }));
    expect(requestedPage).toBe(2);
    expect(screen.getByRole('navigation', { name: 'Сторінки результатів' }).textContent).toContain(
      'Сторінка 3 з 3',
    );
  });

  it('renders only caller-supplied English labels for localized kit states', () => {
    const { container } = render(
      <>
        <Pagination
          label="Result pages"
          previousLabel="Previous"
          nextLabel="Next"
          pageLabel="Page 2 of 3"
          emptyLabel="No pages"
          page={2}
          totalPages={3}
          onPageChange={() => {}}
        />
        <Pagination
          label="Empty result pages"
          previousLabel="Previous"
          nextLabel="Next"
          pageLabel="Page 1 of 1"
          emptyLabel="No pages"
          page={1}
          totalPages={0}
          onPageChange={() => {}}
        />
        <Table
          label="Courses"
          columns={[{ id: 'title', label: 'Title' }]}
          rows={[]}
          emptyMessage="No records yet"
        />
        <Skeleton label="Loading courses" />
        <ErrorState title="Courses unavailable" retryHref="/courses" retryLabel="Try again">
          Reload the page.
        </ErrorState>
      </>,
    );

    const pages = screen.getByRole('navigation', { name: 'Result pages' });
    expect(within(pages).getByRole('button', { name: 'Previous' })).toBeTruthy();
    expect(within(pages).getByRole('button', { name: 'Next' })).toBeTruthy();
    expect(pages.textContent).toContain('Page 2 of 3');
    expect(screen.getByRole('navigation', { name: 'Empty result pages' }).textContent).toContain(
      'No pages',
    );
    expect(screen.getByRole('table', { name: 'Courses' }).textContent).toContain('No records yet');
    expect(screen.getByText('Loading courses')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Try again' }).getAttribute('href')).toBe('/courses');
    expect(container.textContent).not.toMatch(/[\u0400-\u04ff]/u);
    expect(container.innerHTML).not.toMatch(/[\u0400-\u04ff]/u);
  });

  it('filters combobox options while typing', async () => {
    const user = userEvent.setup();
    render(
      <ComboBox
        label="Мова курсу"
        options={[
          { id: 'uk', label: 'Українська' },
          { id: 'en', label: 'Англійська' },
        ]}
      />,
    );
    await user.type(screen.getByRole('combobox', { name: 'Мова курсу' }), 'Укра');
    const list = await screen.findByRole('listbox');
    expect(within(list).getByRole('option', { name: 'Українська' })).toBeTruthy();
    expect(within(list).queryByRole('option', { name: 'Англійська' })).toBeNull();
  });

  it('switches labelled tab panels with the keyboard', async () => {
    const user = userEvent.setup();
    render(
      <Tabs
        label="Розділи курсу"
        items={[
          { id: 'overview', label: 'Опис', content: 'Про курс' },
          { id: 'content', label: 'Програма', content: 'Теми занять' },
        ]}
      />,
    );
    await user.tab();
    await user.keyboard('{ArrowRight}');
    expect(screen.getByRole('tab', { name: 'Програма', selected: true })).toBe(
      document.activeElement,
    );
    expect(screen.getByRole('tabpanel', { name: 'Програма' }).textContent).toBe('Теми занять');
  });

  it('moves a single segmented selection with the keyboard', async () => {
    const user = userEvent.setup();
    render(
      <SegmentedControl
        label="Період статистики"
        options={[
          { id: '1h', label: '1 год' },
          { id: '24h', label: '24 год' },
          { id: '7d', label: '7 днів' },
        ]}
        defaultSelectedKeys={['24h']}
      />,
    );
    // Single selection starts at 24h and is exposed as a checked radio.
    expect(screen.getByRole('radio', { name: '24 год' }).getAttribute('aria-checked')).toBe('true');
    await user.tab();
    expect((document.activeElement as HTMLElement).getAttribute('role')).toBe('radio');
    // Arrow moves focus and the single selection follows it, staying exactly one.
    await user.keyboard('{ArrowRight}');
    const moved = document.activeElement as HTMLElement;
    expect(moved.getAttribute('role')).toBe('radio');
    expect(moved.getAttribute('aria-checked')).toBe('true');
    expect(screen.getAllByRole('radio', { checked: true })).toHaveLength(1);
  });

  it('exposes meter values and delta meaning without relying on color or arrows', () => {
    const { container } = render(
      <>
        <LoadMeter label="Заповнення курсу" value={82} tone="success" />
        <Delta label="Зросло на 18 відсотків" value="+18%" direction="up" tone="success" />
      </>,
    );
    const meter = screen.getByRole('meter', { name: 'Заповнення курсу' });
    expect(meter.getAttribute('aria-valuenow')).toBe('82');
    expect(meter.classList.contains('ui-meter--success')).toBe(true);
    expect(container.querySelector('.ui-meter-fill')?.getAttribute('style')).toContain('82%');
    const delta = screen.getByLabelText('Зросло на 18 відсотків');
    expect(delta.classList.contains('ui-delta--success')).toBe(true);
    expect(delta.textContent).toBe('↑+18%');
  });

  it('keeps unavailable menu actions disabled and dispatches the chosen action', async () => {
    const user = userEvent.setup();
    let action = '';
    render(
      <Menu
        label="Дії курсу"
        items={[
          { id: 'edit', label: 'Редагувати' },
          { id: 'publish', label: 'Опублікувати', isDisabled: true },
        ]}
        onAction={(key) => {
          action = String(key);
        }}
      />,
    );
    await user.click(screen.getByRole('button', { name: 'Дії курсу' }));
    expect(
      screen.getByRole('menuitem', { name: 'Опублікувати' }).getAttribute('aria-disabled'),
    ).toBe('true');
    await user.click(screen.getByRole('menuitem', { name: 'Редагувати' }));
    expect(action).toBe('edit');
  });

  it('associates validation errors with their field and preserves the retry URL', () => {
    render(
      <>
        <TextField label="Назва" isInvalid errorMessage="Вкажіть назву курсу" />
        <ErrorState
          title="Каталог тимчасово недоступний"
          retryHref="/courses?q=Go&sort=title&offset=20"
          retryLabel="Спробувати ще раз"
        >
          Перезавантажте сторінку.
        </ErrorState>
        <ProgressBar label="Обробка файлу" value={35} />
      </>,
    );
    const field = screen.getByRole('textbox', { name: 'Назва' });
    expect(field.getAttribute('aria-invalid')).toBe('true');
    const errorId = field.getAttribute('aria-describedby');
    expect(errorId).toBeTruthy();
    expect(errorId && document.getElementById(errorId)?.textContent).toBe('Вкажіть назву курсу');
    expect(screen.getByRole('link', { name: 'Спробувати ще раз' }).getAttribute('href')).toBe(
      '/courses?q=Go&sort=title&offset=20',
    );
    expect(
      screen.getByRole('progressbar', { name: 'Обробка файлу' }).getAttribute('aria-valuenow'),
    ).toBe('35');
  });

  it('announces a toast without stealing focus and lets users dismiss it', async () => {
    const user = userEvent.setup();
    let dismissed = false;
    render(
      <Toast
        message="Зміни збережено"
        onDismiss={() => {
          dismissed = true;
        }}
      />,
    );
    expect(screen.getByRole('status').textContent).toBe('Зміни збережено');
    await user.tab();
    await user.keyboard('{Enter}');
    expect(dismissed).toBe(true);
  });

  it('renders a stat tile with a value, meta and trailing content, axe-clean', async () => {
    const { container } = render(
      <StatTile label="Активні учні · 30 днів" value="4 820" meta="2 з 3 курсів">
        <span className="ui-stat-trailing">covered</span>
      </StatTile>,
    );
    expect(screen.getByText('4 820').textContent).toBe('4 820');
    expect(screen.getByText('2 з 3 курсів').textContent).toBe('2 з 3 курсів');
    expect(screen.getByText('covered').textContent).toBe('covered');
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it('renders a line trend whose values are readable as a list, axe-clean', async () => {
    const { container } = render(
      <TrendChart
        title="Реєстрації"
        mode="line"
        points={[
          { label: '2026-03-01', value: 4 },
          { label: '2026-03-02', value: 7 },
          { label: '2026-03-03', value: 2 },
        ]}
        emptyLabel="Немає даних у вікні"
      />,
    );
    expect(screen.getByTitle('Реєстрації').textContent).toBe('Реєстрації');
    expect(container.querySelector('svg')).not.toBeNull();
    expect(container.querySelector('.ui-chart-line')).not.toBeNull();
    expect(screen.getByText('2026-03-02: 7').textContent).toBe('2026-03-02: 7');
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it('renders a bar trend and keeps a flat all-zero series on the baseline', () => {
    render(
      <TrendChart
        title="Замовлення"
        mode="bar"
        points={[
          { label: '2026-03-01', value: 0 },
          { label: '2026-03-02', value: 0 },
        ]}
        emptyLabel="Немає даних у вікні"
      />,
    );
    const bars = document.querySelectorAll('.ui-chart-bar');
    expect(bars).toHaveLength(2);
    bars.forEach((bar) => expect(bar.getAttribute('height')).toBe('0'));
  });

  it('shows an empty state instead of an axis when the window is quiet, axe-clean', async () => {
    const { container } = render(
      <TrendChart title="Дохід" mode="line" points={[]} emptyLabel="Немає даних у вікні" />,
    );
    expect(screen.getByRole('status').textContent).toBe('Немає даних у вікні');
    expect(container.querySelector('svg')).toBeNull();
    expect((await axe.run(container)).violations).toEqual([]);
  });
});
