import { createFileRoute } from '@tanstack/react-router';
export const Route = createFileRoute('/login')({ component: Login });
function Login() {
  return (
    <section className="notice">
      <p className="eyebrow">Кабінет студента</p>
      <h1>Вхід незабаром</h1>
      <p>Зараз ви можете переглядати каталог. Вхід у кабінет ще недоступний.</p>
      <a href="/courses">Переглянути курси</a>
    </section>
  );
}
