import { createFileRoute, notFound } from '@tanstack/react-router';

export const Route = createFileRoute('/admin/$')({
  ssr: false,
  beforeLoad: () => {
    throw notFound();
  },
});
