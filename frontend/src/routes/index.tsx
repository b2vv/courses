import { createFileRoute } from '@tanstack/react-router';
import { HomeView } from '../presentation/storefront/HomeView';
export const Route = createFileRoute('/')({ component: HomeView });
