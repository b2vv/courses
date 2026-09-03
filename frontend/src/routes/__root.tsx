import {
  createRootRouteWithContext,
  HeadContent,
  Outlet,
  Scripts,
  Link,
} from '@tanstack/react-router';
import type { QueryClient } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import '../app/styles/main.css';

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  head: () => ({
    meta: [
      { charSet: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { title: 'Courses — навчання через практику' },
      {
        name: 'description',
        content: 'Курси для розробників: архітектура, інструменти й практика.',
      },
    ],
  }),
  shellComponent: Document,
  component: Layout,
  notFoundComponent: () => (
    <section className="notice">
      <h1>Сторінку не знайдено</h1>
      <a href="/courses">Перейти до каталогу</a>
    </section>
  ),
  errorComponent: () => (
    <section className="notice" role="alert">
      <h1>Не вдалося відкрити сторінку</h1>
      <p>Оновіть сторінку, щоб спробувати ще раз.</p>
      <form method="get">
        <button type="submit">Спробувати ще раз</button>
      </form>
    </section>
  ),
});

function Document({ children }: { children: ReactNode }) {
  return (
    <html lang="uk">
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  );
}

function Layout() {
  return (
    <>
      <a className="skip-link" href="#main">
        До основного вмісту
      </a>
      <header className="site-header">
        <div className="shell header-inner">
          <Link to="/" className="wordmark" aria-label="Courses — головна">
            courses<span aria-hidden="true"> / </span>
          </Link>
          <nav aria-label="Основна навігація">
            <Link to="/courses">Каталог</Link>
            <Link to="/login">Увійти</Link>
          </nav>
        </div>
      </header>
      <main id="main" className="shell main-content" tabIndex={-1}>
        <Outlet />
      </main>
      <footer className="shell site-footer">
        <span>Courses</span>
        <span>Знання, які стають практикою.</span>
      </footer>
    </>
  );
}
