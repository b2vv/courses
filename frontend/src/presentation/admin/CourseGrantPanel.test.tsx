import { describe, expect, it } from '@rstest/core';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import type { GrantErrorCode } from '../../domain/purchase/grant';
import { CourseGrantPanel } from './CourseGrantPanel';

const student = { id: '40000000-0000-4000-8000-000000000001', label: 'Student — s@x.test' };

function renderPanel(answer: GrantErrorCode | null, calls: Array<[string, string]>) {
  render(
    <I18nProvider instance={createI18n('en')}>
      <CourseGrantPanel
        users={[student]}
        onSearch={() => {}}
        onGrant={async (userId, reason) => {
          calls.push([userId, reason]);
          return answer;
        }}
      />
    </I18nProvider>,
  );
}

async function choose(browser: ReturnType<typeof userEvent.setup>) {
  await browser.type(screen.getByRole('combobox', { name: 'User' }), 'Stud');
  await browser.click(await screen.findByRole('option', { name: student.label }));
}

describe('course grant panel', () => {
  it('Pay10 asks for a user and a reason before sending anything', async () => {
    const browser = userEvent.setup();
    const calls: Array<[string, string]> = [];
    renderPanel(null, calls);

    await browser.click(screen.getByRole('button', { name: 'Grant access' }));
    expect(screen.getByText('Choose a user from the list.')).toBeTruthy();
    expect(screen.getByText('State the reason for the grant.')).toBeTruthy();
    await choose(browser);
    await browser.type(screen.getByRole('textbox', { name: 'Reason' }), '   ');
    await browser.click(screen.getByRole('button', { name: 'Grant access' }));
    expect(calls).toEqual([]);

    await browser.type(screen.getByRole('textbox', { name: 'Reason' }), ' Scholarship ');
    await browser.click(screen.getByRole('button', { name: 'Grant access' }));
    await waitFor(() => expect(calls).toEqual([[student.id, 'Scholarship']]));
    expect(await screen.findByText(`Access granted: ${student.label}.`)).toBeTruthy();
  });

  it('says plainly when the user already has access', async () => {
    const browser = userEvent.setup();
    renderPanel('course_already_owned', []);
    await choose(browser);
    await browser.type(screen.getByRole('textbox', { name: 'Reason' }), 'Again');
    await browser.click(screen.getByRole('button', { name: 'Grant access' }));
    expect((await screen.findByRole('alert')).textContent).toBe(
      'This user already has access to the course.',
    );
  });

  it('keeps the chosen user when the input then shows its label', async () => {
    // The server matches a name or an email, never the "name — email" label the input shows
    // after a choice; a search for that label comes back empty.
    const calls: string[] = [];
    function Harness() {
      const [search, setSearch] = useState('');
      const users = ['Student', 's@x.test'].some((part) => part.includes(search)) ? [student] : [];
      return (
        <CourseGrantPanel
          users={search === '' || users.length > 0 ? users : []}
          onSearch={setSearch}
          onGrant={async (userId) => {
            calls.push(userId);
            return calls.length === 1 ? null : 'course_already_owned';
          }}
        />
      );
    }
    const browser = userEvent.setup();
    render(
      <I18nProvider instance={createI18n('en')}>
        <Harness />
      </I18nProvider>,
    );
    await choose(browser);
    await browser.type(screen.getByRole('textbox', { name: 'Reason' }), 'First');
    await browser.click(screen.getByRole('button', { name: 'Grant access' }));
    expect(await screen.findByText(`Access granted: ${student.label}.`)).toBeTruthy();
    await browser.type(screen.getByRole('textbox', { name: 'Reason' }), 'Second');
    await browser.click(screen.getByRole('button', { name: 'Grant access' }));
    expect((await screen.findByRole('alert')).textContent).toBe(
      'This user already has access to the course.',
    );
    expect(calls).toEqual([student.id, student.id]);
  });
});
