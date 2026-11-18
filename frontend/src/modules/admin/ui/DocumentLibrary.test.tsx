import { expect, test } from '@rstest/core';
import { fireEvent, render, screen } from '@testing-library/react';
import type { TranscriptDocument, TranscriptErrorCode } from '../../transcript';
import { DocumentLibrary } from './DocumentLibrary';

const documents: TranscriptDocument[] = [
  { id: '60000000-0000-4000-8000-000000000001', label: 'Довідка', url: 'https://example.test/a' },
];

function library(
  actions: Partial<{
    add: (document: TranscriptDocument) => Promise<TranscriptErrorCode | null>;
    remove: (documentId: string) => Promise<TranscriptErrorCode | null>;
  }> = {},
) {
  return render(
    <DocumentLibrary
      documents={documents}
      actions={{
        add: actions.add ?? (async () => null),
        remove: actions.remove ?? (async () => null),
      }}
    />,
  );
}

test('an empty library says so rather than showing an empty list', () => {
  render(
    <DocumentLibrary
      documents={[]}
      actions={{ add: async () => null, remove: async () => null }}
    />,
  );
  expect(screen.getByText('Матеріалів ще немає.')).toBeTruthy();
});

test('a material is a real link that opens safely', () => {
  library();
  const link = screen.getByRole('link', { name: 'Довідка' });
  expect(link.getAttribute('href')).toBe('https://example.test/a');
  expect(link.getAttribute('rel')).toContain('noopener');
});

test('Ux14 reveals the form and explains missing fields through validation', async () => {
  let calls = 0;
  library({
    add: async () => {
      calls += 1;
      return null;
    },
  });
  expect(screen.queryByLabelText('Назва матеріалу')).toBeNull();
  const toggle = screen.getByRole('button', { name: 'Додати матеріал' });
  expect(toggle.getAttribute('aria-expanded')).toBe('false');
  fireEvent.click(toggle);
  expect(toggle.getAttribute('aria-expanded')).toBe('true');
  const add = screen.getByRole('button', { name: 'Зберегти матеріал' });
  expect(add.hasAttribute('disabled')).toBe(false);
  fireEvent.click(add);
  expect(await screen.findAllByText('Введіть значення.')).toHaveLength(2);
  expect(calls).toBe(0);
  fireEvent.click(toggle);
  expect(screen.queryByLabelText('Назва матеріалу')).toBeNull();
});

test('a referenced material names why it cannot be removed', async () => {
  // The refusal must say what to do next: an author who sees a generic failure has no idea that
  // unlinking the line is the way out.
  library({ remove: async () => 'transcript_document_referenced' });
  fireEvent.click(screen.getByRole('button', { name: 'Видалити матеріал «Довідка»' }));
  expect(
    await screen.findByText(
      'Матеріал використано в транскрипті. Спершу відв’яжіть його від реплік.',
    ),
  ).toBeTruthy();
});

test('a bad link is named by its rule, not by a generic failure', async () => {
  library({ add: async () => 'invalid_request' });
  fireEvent.click(screen.getByRole('button', { name: 'Додати матеріал' }));
  fireEvent.change(screen.getByLabelText('Назва матеріалу'), { target: { value: 'Погане' } });
  fireEvent.change(screen.getByLabelText('Посилання'), {
    target: { value: 'javascript:alert(1)' },
  });
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти матеріал' }));
  expect(
    await screen.findByText('Посилання має починатися з http:// або https:// і не вести в медіа.'),
  ).toBeTruthy();
});

test('the fields clear only after a successful add', async () => {
  let attempt = 0;
  library({
    add: async () => {
      attempt += 1;
      return attempt === 1 ? 'transcript_document_limit' : null;
    },
  });
  fireEvent.click(screen.getByRole('button', { name: 'Додати матеріал' }));
  const label = screen.getByLabelText('Назва матеріалу');
  fireEvent.change(label, { target: { value: 'Нове' } });
  fireEvent.change(screen.getByLabelText('Посилання'), {
    target: { value: 'https://example.test/b' },
  });

  fireEvent.click(screen.getByRole('button', { name: 'Зберегти матеріал' }));
  await screen.findByText('Забагато матеріалів: щонайбільше 200 на курс.');
  // A refused add must keep what the author typed: clearing it would make them retype a link the
  // server never objected to.
  expect((label as HTMLInputElement).value).toBe('Нове');

  fireEvent.click(screen.getByRole('button', { name: 'Зберегти матеріал' }));
  await new Promise((resolve) => setTimeout(resolve, 0));
  expect((label as HTMLInputElement).value).toBe('');
});
