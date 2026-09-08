import { test as teardown } from '@playwright/test';
import { recreateApi } from './playback-profile';

// Повертає api до задокументованих дефолтів, щоб наступні проєкти й розробка не
// успадкували прискорений профіль P08.
teardown('restore the documented playback defaults', () => {
  teardown.setTimeout(180_000);
  recreateApi('30', '180');
});
