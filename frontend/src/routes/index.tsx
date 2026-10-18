import { createFileRoute } from '@tanstack/react-router';
import { HomeView } from '../modules/catalog';
import { parseHomeSearch } from '../app/router/home-search';
export const Route = createFileRoute('/')({
  validateSearch: parseHomeSearch,
  component: HomeRoute,
});
function HomeRoute() {
  return <HomeView notice={Route.useSearch().notice} />;
}
