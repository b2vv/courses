import { createFileRoute } from '@tanstack/react-router';
import { LockedArea } from '../presentation/LockedArea';
export const Route = createFileRoute('/admin')({ ssr: false, component: LockedArea });
