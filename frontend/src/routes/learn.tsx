import { createFileRoute } from '@tanstack/react-router';
import { LockedArea } from '../presentation/LockedArea';
export const Route = createFileRoute('/learn')({ ssr: false, component: LockedArea });
