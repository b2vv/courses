import { afterEach, expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nProvider, createI18n } from '../../../app/providers/i18n';
import type { PurchaseOrder } from '../domain/purchase';
import { OrderStatusView, submitCheckoutForm } from './OrderStatusView';

const pending: PurchaseOrder = {
  id: '60000000-0000-4000-8000-000000000001',
  courseId: '10000000-0000-4000-8000-000000000001',
  amount: { minor: 189000, currency: 'UAH' },
  status: 'pending',
  createdAt: '2026-10-01T10:00:00Z',
  paidAt: null,
};

let originalRequestSubmit: HTMLFormElement['requestSubmit'];

afterEach(() => {
  HTMLFormElement.prototype.requestSubmit = originalRequestSubmit;
});

test('pending unpaid orders offer the hosted payment redirect', () => {
  render(
    <I18nProvider instance={createI18n('uk')}>
      <OrderStatusView
        order={pending}
        courseHref="/learn/go-basics"
        isRetrying={false}
        onRetry={() => undefined}
        isCheckingOut={false}
        onCheckout={() => undefined}
      />
    </I18nProvider>,
  );

  expect(screen.getByRole('button', { name: 'Перейти до оплати' })).toBeTruthy();
});

test('the payment redirect builds and submits a native form with provider fields', async () => {
  let submitted: HTMLFormElement | undefined;
  originalRequestSubmit = HTMLFormElement.prototype.requestSubmit;
  HTMLFormElement.prototype.requestSubmit = () => {
    submitted = document.querySelector('form') ?? undefined;
  };
  const user = userEvent.setup();
  render(
    <I18nProvider instance={createI18n('uk')}>
      <OrderStatusView
        order={pending}
        courseHref="/learn/go-basics"
        isRetrying={false}
        onRetry={() => undefined}
        isCheckingOut={false}
        onCheckout={() =>
          submitCheckoutForm({
            action: 'https://secure.wayforpay.com/pay',
            fields: [
              { name: 'merchantAccount', value: 'merchant' },
              { name: 'merchantSignature', value: 'signature' },
            ],
          })
        }
      />
    </I18nProvider>,
  );

  await user.click(screen.getByRole('button', { name: 'Перейти до оплати' }));
  if (submitted === undefined) throw new Error('checkout form was not submitted');
  expect(submitted.method).toBe('post');
  expect(submitted.action).toBe('https://secure.wayforpay.com/pay');
  expect(
    [...submitted.querySelectorAll('input')].map((input) => [input.name, input.value]),
  ).toEqual([
    ['merchantAccount', 'merchant'],
    ['merchantSignature', 'signature'],
  ]);
});
