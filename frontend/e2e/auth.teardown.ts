import { expect, test } from '@playwright/test';
import { readLoginCount } from './auth';

test('B19 keeps host-originated logins within the nginx budget', async () => {
  const count = await readLoginCount();
  process.stdout.write(
    `B19 login count: total=${count.total}, maximum consecutive=${count.maximumConsecutive}\n`,
  );
  expect(count.total).toBeLessThanOrEqual(7);
  expect(count.maximumConsecutive).toBeLessThanOrEqual(3);
});
