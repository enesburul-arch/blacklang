# BlackLang Target Declaration

`target` declares which platform and generated stack a `.black` source file targets.

In draft v0.1, the supported target is intentionally narrow:

```black
target web {
  frontend react
  backend node
  database sqlite
}
```

## Purpose

The target block makes the generator contract explicit. Instead of assuming that every `.black` file means the same web output, the source can say which stack it expects.

This is the foundation for future generator adapters such as API-only, mobile, desktop, PostgreSQL, or other runtimes.

## Rules

- Use one `target` block per project.
- Draft v0.1 supports only `target web`.
- Draft v0.1 supports only `frontend react`.
- Draft v0.1 supports only `backend node`.
- Draft v0.1 supports only `database sqlite`.
- If `target` is omitted, the current generator keeps the legacy default: web, React, Node, SQLite.
- Do not write unsupported targets until the compiler and generator can produce them.

## Agent Workflow

AI agents should run:

```bash
black docs target --json
black inspect app.black --affected target --json
```

before editing target metadata.

If the user asks for PostgreSQL, mobile, desktop, or another backend, the agent should treat that as a roadmap task unless matching generator support already exists.

