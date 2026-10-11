export type AdminNavIconKey =
  | 'overview'
  | 'people'
  | 'accounts'
  | 'moderation'
  | 'analytics'
  | 'refunds';

/** 16px stroke glyphs for the admin rail and sidebar links (NorthWind grammar). */
export function AdminNavIcon({ name }: { readonly name: AdminNavIconKey }) {
  return (
    <svg
      className="admin-nav-icon"
      viewBox="0 0 16 16"
      width="16"
      height="16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {name === 'overview' && (
        <>
          <rect x="2" y="2" width="5" height="5" rx="1" />
          <rect x="9" y="2" width="5" height="5" rx="1" />
          <rect x="2" y="9" width="5" height="5" rx="1" />
          <rect x="9" y="9" width="5" height="5" rx="1" />
        </>
      )}
      {name === 'people' && (
        <>
          <circle cx="6" cy="6" r="2.4" />
          <path d="M1.8 13.2c0-2.3 1.9-3.6 4.2-3.6s4.2 1.3 4.2 3.6" />
          <path d="M11 4.2c1.3.2 2.3 1.3 2.3 2.8 0 .9-.4 1.7-1 2.2M12.4 9.8c1.1.4 1.8 1.4 1.8 3.1" />
        </>
      )}
      {name === 'accounts' && (
        <>
          <rect x="2" y="3.5" width="12" height="9" rx="1.5" />
          <circle cx="6" cy="7.6" r="1.5" />
          <path d="M3.6 11c.3-1.2 1.3-1.8 2.4-1.8s2.1.6 2.4 1.8" />
          <path d="M10 6.6h2.4M10 9.4h2.4" />
        </>
      )}
      {name === 'moderation' && (
        <>
          <path d="M8 2 13 4v4c0 3.4-2.3 5.3-5 6-2.7-.7-5-2.6-5-6V4z" />
          <path d="m5.8 8 1.5 1.5L10.4 6.4" />
        </>
      )}
      {name === 'analytics' && (
        <>
          <path d="M2.5 13.5h11" />
          <rect x="3.5" y="8.5" width="2.2" height="4" rx="0.5" />
          <rect x="7" y="5.5" width="2.2" height="7" rx="0.5" />
          <rect x="10.5" y="9.5" width="2.2" height="3" rx="0.5" />
        </>
      )}
      {name === 'refunds' && (
        <>
          <circle cx="8" cy="8" r="5.5" />
          <path d="M6 5.5h3.2a1.6 1.6 0 0 1 0 3.2H6M6 8.7l2.2 2.2M6 5.5v5.4" />
        </>
      )}
    </svg>
  );
}
