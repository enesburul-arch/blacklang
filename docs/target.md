# BlackLang Target Declaration

`target` declares which platform and generated stack a `.black` source file targets.

Draft v0.2 supports two generated application targets:

```black
target web {
  frontend react
  backend node
  database sqlite
}
```

```black
target api {
  backend node
  database sqlite
}
```

## Purpose

The target block makes the generator contract explicit. A source can ask for a generated React/Node web app, or for the same deterministic API/runtime surface without generated React and Vite files.

`target api` is useful for API services, webhooks, worker-backed projects, and AI-generated backends where pages still describe resource routes but no browser UI should be emitted.

## Rules

- Use one `target` block per project.
- `target web` supports `frontend react`, `backend node`, and `database sqlite`, `database postgres`, or `database mysql`.
- `target api` supports `backend node` and `database sqlite`, `database postgres`, or `database mysql`.
- Do not write `frontend` inside `target api`; API-only output has no React frontend.
- If `target` is omitted, the current generator keeps the legacy default: web, React, Node, SQLite.
- `database mysql` generates Prisma MySQL provider output, a PrismaMariaDb adapter client, MySQL seed upserts, and MySQL Docker Compose services. Generated rename migration runtime is still limited to SQLite and PostgreSQL.
- Mobile, desktop, and alternate backend targets remain unsupported until matching generator output exists.

## Generated Output

`target web` emits the full React + API app: Vite entry files, React pages, frontend smoke tests, optional browser-check/e2e/matrix files, Express routes, Prisma, OpenAPI, validation, seeds, jobs, and deploy files.

`target api` emits the server/runtime surface only:

```text
README.md
.env.example
package.json
openapi.json
prisma/schema.prisma
src/server.ts
src/types.ts
src/db.ts
src/setup-db.ts
src/api/*.ts
src/routes/*.ts
src/validation/*.ts
src/blacklang.contract.test.ts
src/blacklang.api.test.ts
```

When declarations exist, API-only output also emits seed files, job worker files, Docker/deploy files, ops routes, auth routes, and explicit API handlers. It intentionally omits `index.html`, `vite.config.ts`, React page/component files, `src/blacklang.frontend.test.tsx`, `src/blacklang.browser.test.tsx`, `src/blacklang.e2e.test.ts`, `src/blacklang.e2e.matrix.ts`, and `tests/browser-matrix.json`.

Top-level `test` declarations are browser-check intent and require `target web`. Under `target api`, validation reports `UNSUPPORTED_API_TARGET_TEST`.

## Agent Workflow

AI agents should run:

```bash
black docs target --json
black explain target --json
black inspect app.black --affected target --json
```

before editing target metadata.

When changing to `target api`, also inspect generated `package.json`, `src/server.ts`, `openapi.json`, and the absence of frontend/browser/matrix files after `black build`.
