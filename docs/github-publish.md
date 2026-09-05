# GitHub Publish Policy

## Purpose

This document defines what should and should not be published when BlackLang is pushed to GitHub.

The public repository should contain the language source, compiler code, examples, documentation, benchmarks, scripts, and CI workflows. Local build outputs, generated applications, private notes, databases, and release artifacts should stay out of git.

## Publish

Commit these files and folders:

```text
.github/
benchmarks/
docs/
examples/
packages/
scripts/
.gitignore
AGENTS.md
BLACKLANG.md
CONTRIBUTING.md
LICENSE
README.md
ROADMAP.md
ROADMAP-v0.2.md
SPEC.md
blacklang-web-yol-haritasi.md
blacklang.toml
```

## Do Not Publish

Keep these files and folders out of git:

```text
.local-logs/
artifacts/
dist/
generated/
node_modules/
init-check-*/
blacklang-fikir.md
*.db
*.sqlite
*.sqlite3
.env
.env.*
```

## Reasoning

`generated/`, `dist/`, and `artifacts/` are reproducible outputs. They should be produced by the CLI, release scripts, or GitHub Actions instead of committed.

`blacklang-fikir.md` is a private ideation/history note. The public project should use `README.md`, `BLACKLANG.md`, `SPEC.md`, `ROADMAP.md`, `ROADMAP-v0.2.md`, and `blacklang-web-yol-haritasi.md` as cleaned public documentation.

Local databases and environment files may contain private runtime state or secrets and must not be committed.

## Before First Push

Run:

```bash
black format --check --json
black lint --json
black docs --all --json
black explain entity --json
go test ./...
```

From `packages/cli/`, run:

```bash
go test ./...
```

The root folder currently does not contain `.git`. Initialize git only after checking ignored files with:

```bash
git init
git status --ignored
```
