import { expect, test } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import userEvent from '@testing-library/user-event';
import type { OwnReview, PublicReviewPage } from '../../contracts/course';
import { ReviewBlock } from './ReviewBlock';

const review = {
  id: '75000000-0000-4000-8000-000000000001',
  authorDisplayName: 'Олена',
  stars: 4,
  body: 'Дуже корисний курс для щоденної практики.',
  completedLessons: 1,
  totalLessons: 2,
  percent: 50,
  createdAt: '2026-10-01T09:00:00Z',
  editedAt: null,
};
const page: PublicReviewPage = { rating: null, items: [], nextCursor: null };
function renderBlock(own: OwnReview, onSave = async () => {}, onDelete = async () => {}) {
  return render(
    <ReviewBlock
      own={own}
      reviews={page}
      reviewsError={false}
      hasMore={false}
      loadingMore={false}
      nextPageError={false}
      onMore={() => {}}
      onRetry={() => {}}
      onRetryMore={() => {}}
      onSave={onSave}
      onDelete={onDelete}
    />,
  );
}

test('disables review submission at star and text boundaries and announces the counter', async () => {
  const user = userEvent.setup();
  renderBlock({ review: null, eligibility: 'ok' });
  const submit = screen.getByRole('button', { name: 'Зберегти' });
  const body = screen.getByRole('textbox', { name: 'Текст відгуку' });
  expect((submit as HTMLButtonElement).disabled).toBe(true);
  expect(screen.getByText('Мінімум 20 символів')).toBeTruthy();
  await user.click(screen.getByRole('radio', { name: '5 з 5' }));
  await user.type(body, 'a'.repeat(19));
  expect((submit as HTMLButtonElement).disabled).toBe(true);
  await user.type(body, 'a');
  expect((submit as HTMLButtonElement).disabled).toBe(false);
  // Paste the bulk: typing 1980 keystrokes outruns the test timeout and leaks into the next test.
  await user.click(body);
  await user.paste('a'.repeat(1980));
  expect((submit as HTMLButtonElement).disabled).toBe(false);
  await user.type(body, 'a');
  expect((submit as HTMLButtonElement).disabled).toBe(true);
  // The visible counter is silent; only crossing the limit is announced politely.
  expect(screen.getByText('2001 / 2000').getAttribute('aria-live')).toBeNull();
  expect(screen.getByText('Текст довший за 2000 символів').getAttribute('aria-live')).toBe(
    'polite',
  );
});
test('submits, renders the saved review, and opens edit prefilled', async () => {
  const user = userEvent.setup();
  let saved = false;
  const { rerender } = renderBlock({ review: null, eligibility: 'ok' }, async () => {
    saved = true;
  });
  await user.click(screen.getByRole('radio', { name: '4 з 5' }));
  await user.type(screen.getByRole('textbox'), review.body);
  await user.click(screen.getByRole('button', { name: 'Зберегти' }));
  expect(saved).toBe(true);
  rerender(
    <ReviewBlock
      own={{ review, eligibility: 'ok' }}
      reviews={page}
      reviewsError={false}
      hasMore={false}
      loadingMore={false}
      nextPageError={false}
      onMore={() => {}}
      onRetry={() => {}}
      onRetryMore={() => {}}
      onSave={async () => {}}
      onDelete={async () => {}}
    />,
  );
  await user.click(screen.getByRole('button', { name: 'Редагувати' }));
  expect((screen.getByRole('textbox') as HTMLTextAreaElement).value).toBe(review.body);
});
test('deletes only through the dialog', async () => {
  const user = userEvent.setup();
  let deleted = 0;
  renderBlock(
    { review, eligibility: 'ok' },
    async () => {},
    async () => {
      deleted += 1;
    },
  );
  await user.click(screen.getByRole('button', { name: 'Видалити' }));
  expect(screen.getByRole('dialog', { name: 'Видалити відгук?' })).toBeTruthy();
  await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Видалити' }));
  expect(deleted).toBe(1);
});
test('shows every eligibility refusal without a form', () => {
  for (const code of [
    'review_owner_forbidden',
    'review_purchase_required',
    'review_not_started',
    'review_not_available',
    'review_removed',
  ]) {
    const { unmount } = renderBlock({ review: null, eligibility: code });
    expect(screen.queryByRole('textbox')).toBeNull();
    unmount();
  }
});
test('keeps the typed text when a save fails (Rt29)', async () => {
  const user = userEvent.setup();
  renderBlock({ review: null, eligibility: 'ok' }, async () => {
    throw new Error('dependency_unavailable');
  });
  await user.click(screen.getByRole('radio', { name: '3 з 5' }));
  const body = screen.getByRole('textbox');
  await user.type(body, review.body);
  await user.click(screen.getByRole('button', { name: 'Зберегти' }));
  expect((body as HTMLTextAreaElement).value).toBe(review.body);
  expect(screen.getByRole('alert').textContent).toContain('Не вдалося зберегти');
});
test('names a save refused because the review was removed meanwhile', async () => {
  const user = userEvent.setup();
  renderBlock({ review: null, eligibility: 'ok' }, async () => {
    throw new Error('review_removed');
  });
  await user.click(screen.getByRole('radio', { name: '3 з 5' }));
  await user.click(screen.getByRole('textbox'));
  await user.paste(review.body);
  await user.click(screen.getByRole('button', { name: 'Зберегти' }));
  expect(screen.getByRole('alert').textContent).toContain('Відгук знято модератором');
});
test('names the rating radio group and all star choices', () => {
  renderBlock({ review: null, eligibility: 'ok' });
  expect(screen.getByRole('radiogroup', { name: 'Оцінка' })).toBeTruthy();
  for (let stars = 1; stars <= 5; stars += 1)
    expect(screen.getByRole('radio', { name: `${stars} з 5` })).toBeTruthy();
});

test('counts runes after trimming, as the server does', async () => {
  const user = userEvent.setup();
  renderBlock({ review: null, eligibility: 'ok' });
  await user.click(screen.getByRole('radio', { name: '5 з 5' }));
  await user.click(screen.getByRole('textbox'));
  // Ten emoji are twenty UTF-16 units but ten characters: still below the minimum.
  await user.paste('😀'.repeat(10));
  expect(screen.getByText('10 / 2000')).toBeTruthy();
  expect(screen.getByRole('button', { name: 'Зберегти' }).hasAttribute('disabled')).toBe(true);
});

test('names the star choices in the current locale', () => {
  render(
    <I18nProvider instance={createI18n('en')}>
      <ReviewBlock
        own={{ review: null, eligibility: 'ok' }}
        reviews={page}
        reviewsError={false}
        hasMore={false}
        loadingMore={false}
        nextPageError={false}
        onMore={() => {}}
        onRetry={() => {}}
        onRetryMore={() => {}}
        onSave={async () => {}}
        onDelete={async () => {}}
      />
    </I18nProvider>,
  );
  expect(screen.getByRole('radio', { name: '4 out of 5' })).toBeTruthy();
});

test('shows the reason instead of Edit when the author lost the right (G4)', () => {
  renderBlock({ review, eligibility: 'review_purchase_required' });
  expect(screen.queryByRole('button', { name: 'Редагувати' })).toBeNull();
  expect(screen.getByText('Відгук можна залишити після покупки курсу')).toBeTruthy();
});

test('names a review removed by moderation and offers no form', () => {
  renderBlock({ review: null, eligibility: 'review_removed' });
  expect(screen.getByText('Відгук знято модератором')).toBeTruthy();
  expect(screen.queryByRole('textbox')).toBeNull();
});

test('shows a delete failure instead of swallowing it', async () => {
  const user = userEvent.setup();
  renderBlock(
    { review, eligibility: 'ok' },
    async () => {},
    async () => {
      throw new Error('dependency_unavailable');
    },
  );
  await user.click(screen.getByRole('button', { name: 'Видалити' }));
  const dialog = screen.getByRole('dialog');
  await user.click(within(dialog).getByRole('button', { name: 'Видалити' }));
  expect(await screen.findByText('Не вдалося видалити відгук, спробуйте ще')).toBeTruthy();
});

test('offers a retry when the own review cannot be loaded', async () => {
  const user = userEvent.setup();
  let retried = 0;
  render(
    <ReviewBlock
      own={undefined}
      ownError
      onRetryOwn={() => {
        retried += 1;
      }}
      reviews={page}
      reviewsError={false}
      hasMore={false}
      loadingMore={false}
      nextPageError={false}
      onMore={() => {}}
      onRetry={() => {}}
      onRetryMore={() => {}}
      onSave={async () => {}}
      onDelete={async () => {}}
    />,
  );
  expect(screen.getByRole('alert').textContent).toContain('Не вдалося завантажити ваш відгук');
  await user.click(screen.getByRole('button', { name: 'Спробувати ще' }));
  expect(retried).toBe(1);
});

for (const [code, message] of [
  ['enrollment_required', 'Щоб залишити відгук або поскаржитися на відгук, запишіться на курс.'],
  [
    'enrollment_blocked',
    'Вас заблоковано в цьому курсі — відгуки та скарги на відгуки недоступні.',
  ],
] as const) {
  test(`shows the ${code} message after submitting a review`, async () => {
    const user = userEvent.setup();
    renderBlock({ review: null, eligibility: 'ok' }, async () => {
      throw new Error(code);
    });
    await user.click(screen.getByRole('radio', { name: '3 з 5' }));
    await user.click(screen.getByRole('textbox'));
    await user.paste(review.body);
    await user.click(screen.getByRole('button', { name: 'Зберегти' }));
    expect((await screen.findByRole('alert')).textContent).toBe(message);
  });
}
