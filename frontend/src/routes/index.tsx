import { createFileRoute, Link } from '@tanstack/react-router';
export const Route = createFileRoute('/')({ component: Home });
function Home() {
  return (
    <section className="intro">
      <p className="eyebrow">Бібліотека для розробників</p>
      <h1>
        Розібратись.
        <br />
        Спробувати. Побудувати.
      </h1>
      <p className="lede">Практичні курси про системи, які ви створюєте щодня.</p>
      <Link to="/courses" className="primary-link">
        Переглянути курси <span aria-hidden="true">↗</span>
      </Link>
    </section>
  );
}
