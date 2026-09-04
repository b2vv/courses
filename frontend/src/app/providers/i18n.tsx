import { createInstance } from 'i18next';
import { I18nextProvider } from 'react-i18next';
import type { ReactNode } from 'react';

export type Locale = 'uk' | 'en';

export const resources = {
  uk: {
    translation: {
      common: {
        catalog: 'Каталог',
        login: 'Увійти',
        skip: 'До основного вмісту',
        mainNavigation: 'Основна навігація',
        tagline: 'Знання, які стають практикою.',
        homeLabel: 'CourseForge — головна',
        retry: 'Спробувати ще раз',
        loading: 'Завантаження',
        close: 'Закрити',
      },
      home: {
        label: 'Практичне навчання для розробників',
        title: 'Навички, які працюють у',
        description:
          'Від першого рішення до надійної системи. Поглиблюйте знання архітектури, бекенду та інтерфейсів на практиці.',
        action: 'Переглянути курси',
        art: 'Менше здогадок.\nБільше розуміння.',
      },
      catalog: {
        label: 'Навчання через практику',
        title: 'Каталог курсів',
        description: 'Оберіть тему для наступного кроку у розробці.',
        count_one: '{{count}} курс',
        count_few: '{{count}} курси',
        count_many: '{{count}} курсів',
        count_other: '{{count}} курсу',
        search: 'Знайти курс',
        placeholder: 'Назва або тема',
        sort: 'Сортування',
        recent: 'Спочатку нові',
        byTitle: 'За назвою',
        submit: 'Знайти',
        empty: 'Курсів не знайдено',
        emptyPage: 'На цій сторінці немає курсів',
        emptyHint: 'Спробуйте інший запит або поверніться до всіх курсів.',
        firstPage: 'На першу сторінку',
        pages: 'Сторінки каталогу',
        range: '{{from}}–{{to}} із {{total}}',
        zeroRange: '0 із {{total}}',
        previous: '← Попередня',
        next: 'Наступна →',
        kind: 'Практичний курс',
        details: 'Детальніше: {{title}}',
        loading: 'Завантажуємо курси',
        failure: 'Не вдалося завантажити курси',
        failureHint:
          'Перевірте з’єднання та спробуйте ще раз. Ваш запит збережено в адресі сторінки.',
        notFound: 'Курс не знайдено',
        notFoundHint: 'Можливо, його ще не опубліковано.',
        goToCatalog: 'Перейти до каталогу',
      },
      course: {
        all: '← Усі курси',
        about: 'Про курс',
        comingSoon: 'Навчання незабаром',
        unavailable:
          'Зараз можна ознайомитися з описом. Доступ до уроків і придбання курсу ще готуються.',
        continue: 'Продовжити перегляд',
      },
      locked: {
        title: 'Розділ поки недоступний',
        text: 'Кабінет та керування курсами ще готуються до відкриття.',
        action: 'Перейти до входу',
      },
      login: {
        label: 'Кабінет студента',
        title: 'Вхід у кабінет',
        description: 'Введіть пошту та пароль, щоб продовжити навчання.',
        email: 'Електронна пошта',
        emailPlaceholder: 'you@example.com',
        password: 'Пароль',
        submit: 'Увійти',
        submitting: 'Входимо…',
        signedInTitle: 'Ви вже увійшли',
        signedInText: '{{name}} — сеанс активний.',
        action: 'Переглянути курси',
        continue: 'Продовжити',
        checking: 'Перевіряємо сеанс',
        failure: 'Не вдалося перевірити сеанс',
        failureHint: 'Оновіть сторінку або спробуйте ще раз.',
        error: {
          invalid_credentials: 'Невірна пошта або пароль.',
          invalid_request: 'Не вдалося обробити запит. Перевірте введені дані.',
          origin_forbidden: 'Запит відхилено з міркувань безпеки. Оновіть сторінку.',
          authentication_required: 'Сеанс завершився. Увійдіть ще раз.',
          dependency_unavailable: 'Сервіс тимчасово недоступний. Спробуйте пізніше.',
          invalid_response: 'Отримано неочікувану відповідь. Спробуйте ще раз.',
          request_timeout: 'Перевищено час очікування. Спробуйте ще раз.',
          request_cancelled: 'Запит скасовано.',
        },
      },
      page: {
        notFound: 'Сторінку не знайдено',
        failure: 'Не вдалося відкрити сторінку',
        failureHint: 'Оновіть сторінку, щоб спробувати ще раз.',
      },
      ui: {
        selectPlaceholder: 'Оберіть варіант',
        showOptions: 'Показати варіанти',
        noOptions: 'Немає відповідних варіантів',
        noRecords: 'Ще немає записів',
        resultPages: 'Сторінки результатів',
        back: 'Назад',
        forward: 'Далі',
        noPages: 'Немає сторінок',
        pageOf: 'Сторінка {{page}} з {{total}}',
        closeMessage: 'Закрити повідомлення',
      },
    },
  },
  en: {
    translation: {
      common: {
        catalog: 'Courses',
        login: 'Log in',
        skip: 'Skip to main content',
        mainNavigation: 'Main navigation',
        tagline: 'Knowledge put into practice.',
        homeLabel: 'CourseForge — home',
        retry: 'Try again',
        loading: 'Loading',
        close: 'Close',
      },
      home: {
        label: 'Practical learning for developers',
        title: 'Skills that work in',
        description:
          'From the first decision to a reliable system. Deepen your understanding of architecture, backend engineering and interfaces through practice.',
        action: 'Browse courses',
        art: 'Less guessing.\nMore understanding.',
      },
      catalog: {
        label: 'Learn through practice',
        title: 'Course catalog',
        description: 'Choose a subject for your next step in software development.',
        count_one: '{{count}} course',
        count_other: '{{count}} courses',
        search: 'Find a course',
        placeholder: 'Title or subject',
        sort: 'Sort',
        recent: 'Newest first',
        byTitle: 'By title',
        submit: 'Search',
        empty: 'No courses found',
        emptyPage: 'There are no courses on this page',
        emptyHint: 'Try another query or return to all courses.',
        firstPage: 'First page',
        pages: 'Catalog pages',
        range: '{{from}}–{{to}} of {{total}}',
        zeroRange: '0 of {{total}}',
        previous: '← Previous',
        next: 'Next →',
        kind: 'Practical course',
        details: 'Course details: {{title}}',
        loading: 'Loading courses',
        failure: 'Could not load courses',
        failureHint: 'Check your connection and try again. Your query is preserved in the address.',
        notFound: 'Course not found',
        notFoundHint: 'It may not be published yet.',
        goToCatalog: 'Go to the catalog',
      },
      course: {
        all: '← All courses',
        about: 'About the course',
        comingSoon: 'Learning access is coming soon',
        unavailable:
          'You can review the course description now. Lessons and purchasing are still being prepared.',
        continue: 'Keep browsing',
      },
      locked: {
        title: 'This area is not available yet',
        text: 'Student and course management areas are still being prepared.',
        action: 'Go to login',
      },
      login: {
        label: 'Student area',
        title: 'Sign in',
        description: 'Enter your email and password to continue learning.',
        email: 'Email',
        emailPlaceholder: 'you@example.com',
        password: 'Password',
        submit: 'Log in',
        submitting: 'Signing in…',
        signedInTitle: 'You are already signed in',
        signedInText: '{{name}} — the session is active.',
        action: 'Browse courses',
        continue: 'Continue',
        checking: 'Checking your session',
        failure: 'Could not check your session',
        failureHint: 'Reload the page or try again.',
        error: {
          invalid_credentials: 'That email or password is not correct.',
          invalid_request: 'The request could not be processed. Check what you entered.',
          origin_forbidden: 'The request was refused for security reasons. Reload the page.',
          authentication_required: 'Your session ended. Sign in again.',
          dependency_unavailable: 'The service is temporarily unavailable. Try again later.',
          invalid_response: 'An unexpected response arrived. Try again.',
          request_timeout: 'The request timed out. Try again.',
          request_cancelled: 'The request was cancelled.',
        },
      },
      page: {
        notFound: 'Page not found',
        failure: 'Could not open the page',
        failureHint: 'Reload the page to try again.',
      },
      ui: {
        selectPlaceholder: 'Choose an option',
        showOptions: 'Show options',
        noOptions: 'No matching options',
        noRecords: 'No records yet',
        resultPages: 'Result pages',
        back: 'Back',
        forward: 'Next',
        noPages: 'No pages',
        pageOf: 'Page {{page}} of {{total}}',
        closeMessage: 'Dismiss message',
      },
    },
  },
} as const;

export function normalizeLocale(value: string | null | undefined): Locale | undefined {
  const language = value
    ?.trim()
    .toLowerCase()
    .split(/[-_,;]/)[0];
  return language === 'uk' || language === 'en' ? language : undefined;
}

export function cookieLocale(cookie: string | null | undefined): Locale | undefined {
  return normalizeLocale(
    cookie
      ?.split(';')
      .map((part) => part.trim())
      .find((part) => part.startsWith('courseforge_locale='))
      ?.split('=')[1],
  );
}

export function createI18n(locale: Locale = 'uk') {
  const instance = createInstance();
  void instance.init({
    lng: locale,
    fallbackLng: 'uk',
    supportedLngs: ['uk', 'en'],
    resources,
    initAsync: false,
    interpolation: { escapeValue: false },
  });
  return instance;
}
export function I18nProvider({
  instance,
  children,
}: {
  instance: ReturnType<typeof createI18n>;
  children: ReactNode;
}) {
  return <I18nextProvider i18n={instance}>{children}</I18nextProvider>;
}
