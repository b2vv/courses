# visual-compare

Side-by-side visual comparison of our storefront against a design reference,
via Playwright. Captures the two true images (reference | ours) per surface and
width and stitches them into one sheet — the judge is the reader's eyes, not a
pixel diff (diffs mislead whenever layout, copy or framing legitimately differ).

## Run

    # local stack must be up (LOCAL_URL reachable), playwright installed in frontend/
    node tools/visual-compare/compare.mjs

    # overrides
    REF_URL=https://… LOCAL_URL=http://localhost:8080 PP_WIDTHS=1280,390 \
      PP_SCHEME=light node tools/visual-compare/compare.mjs

Sheets land in `tools/visual-compare/out/` (gitignored); raw single shots in
`out/raw/`. Edit the `SURFACES` array in `compare.mjs` to add pages/regions.
