import { test as setup } from '@playwright/test';
import { recreateApi } from './playback-profile';

// P08 доводить, що heartbeat поновлює володіння кілька разів за один перегляд. З
// продакшн-каденцією (heartbeat 30 с, TTL 180 с) це вимагало б понад двох хвилин
// відтворення в кожному прогоні, тож проєкт `playback` піднімає api з прискореним
// профілем, а його teardown повертає дефолти. `docker-compose.yml` НЕ змінюється:
// розробка й решта e2e бачать 30/180.
setup('apply the accelerated playback profile', () => {
  setup.setTimeout(180_000);
  recreateApi('2', '6');
});
