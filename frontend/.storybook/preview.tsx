import type { Decorator, Preview } from '@storybook/react';
import '../src/app/styles/main.css';

/**
 * Every story renders inside the admin surface (`data-surface="admin"`), which is
 * dark-only, so the NorthWind tokens resolve and the gloss/pills look as shipped.
 */
const withAdminSurface: Decorator = (Story) => (
  <div
    data-surface="admin"
    style={{
      minBlockSize: '100vh',
      padding: 'var(--space-6)',
      background: 'var(--color-background)',
      color: 'var(--color-text)',
    }}
  >
    <Story />
  </div>
);

const preview: Preview = {
  parameters: {
    layout: 'fullscreen',
    controls: { expanded: true },
  },
  decorators: [withAdminSurface],
};

export default preview;
