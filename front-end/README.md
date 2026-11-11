# Anywhere Console

Vite 5 + React 17 + TypeScript 5 + Ant Design 4.

## Scripts

```shell
npm ci          # install dependencies
npm run dev     # Vite dev server on http://localhost:3000
npm run build   # type-check (tsc) + production build into build/
npm run preview # preview the production build
npm run lint    # ESLint for src/**/*.{ts,tsx}
npm test        # Vitest + jsdom
```

The API is expected on the same origin by default. For local development against a standalone
server, create `.env.local` with:

```
VITE_API_PREFIX=https://localhost:1114
```

The server's `ui_config.debug` mode allows CORS from `http://localhost:3000` and accepts the
session cookie for that origin.

The production bundle is copied to `../static` by `make ui` at the repository root.
