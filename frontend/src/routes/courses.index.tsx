import { createFileRoute } from '@tanstack/react-router';
export const Route = createFileRoute('/courses/')({ component: CatalogShell });
function CatalogShell() {
  return (
    <section>
      <p className="eyebrow">Бібліотека</p>
      <h1>Каталог курсів</h1>
      <p className="lede">Курси незабаром з’являться тут.</p>
    </section>
  );
}
