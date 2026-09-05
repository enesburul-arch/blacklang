# BlackLang Deployment

BlackLang deployment intent describes how the generated web app should run in production.

```black
deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
}
```

Rules:

- Use one top-level `deploy` block.
- Draft v0.1 supports `target docker`.
- `port env NAME default PORT` makes the generated server read its listen port from the environment.
- `env NAME required|optional` documents runtime configuration without storing secrets in `.black` source.
- Environment variable names use uppercase letters, numbers, and underscores.

Generated web behavior when `target docker` is declared:

- Writes `Dockerfile`.
- Writes `.dockerignore`.
- Writes `docker-compose.yml`.
- Adds `PORT` to `.env.example`.
- Adds a `start` script to `package.json`.
- Makes `src/server.ts` read the configured port environment variable.
- Serves the built Vite frontend from `dist` after API routes.

Draft v0.1 keeps the generated database runtime on SQLite. PostgreSQL deployment syntax should be added only after the generator can produce a matching PostgreSQL runtime.
