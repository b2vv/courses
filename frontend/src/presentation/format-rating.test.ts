import { expect, test } from '@rstest/core';
import { createI18n } from '../app/providers/i18n';
import { formatRating } from './format-rating';

test('formats rating averages to one decimal in the current locale', () => {
  expect(formatRating(5, 'uk')).toBe('5,0');
  expect(formatRating(5, 'en')).toBe('5.0');
  expect(formatRating(4.7, 'uk')).toBe('4,7');
});

test('lesson and rating counts use the locale plural forms', async () => {
  const uk = createI18n('uk');
  expect([1, 2, 5, 11, 21, 22, 25].map((count) => uk.t('ratings.lessons', { count }))).toEqual([
    '1 урок',
    '2 уроки',
    '5 уроків',
    '11 уроків',
    '21 урок',
    '22 уроки',
    '25 уроків',
  ]);
  expect([3, 5, 21].map((count) => uk.t('ratings.summary', { average: '4,7', count }))).toEqual([
    'Рейтинг 4,7 з 5, 3 оцінки',
    'Рейтинг 4,7 з 5, 5 оцінок',
    'Рейтинг 4,7 з 5, 21 оцінка',
  ]);
  const en = createI18n('en');
  expect([1, 2].map((count) => en.t('ratings.lessons', { count }))).toEqual([
    '1 lesson',
    '2 lessons',
  ]);
  expect(en.t('ratings.summary', { average: '4.7', count: 3 })).toBe(
    'Rated 4.7 out of 5, 3 ratings',
  );
});
