# AI Agent Instructions

Follow these rules when working on BlackLang.

## Read First

Before changing language behavior, read:

1. `SPEC.md`
2. `BLACKLANG.md`
3. `ROADMAP.md`
4. `ROADMAP-v0.2.md`
5. `blacklang-web-yol-haritasi.md`

Before changing a BlackLang project, also check the local BlackLang version file when it exists:

```text
blacklang.toml
```

Before changing release packaging or install behavior, also read:

```text
docs/install.md
docs/release-artifacts.md
docs/npm-wrapper.md
```

## Editing Rules

- Prefer changing `.black` source examples and compiler code over generated output.
- Do not manually edit generated web application files unless debugging the generator.
- Keep the language deterministic.
- Do not add multiple syntaxes for the same behavior.
- Prefer clear syntax over overly compressed syntax.
- Preserve JSON output modes for AI agents.
- Prefer familiar programming words over invented symbols.
- Keep local learning docs short enough for AI agents to read at task start.
- Treat `.black` source files as high-value source assets.
- Do not put secrets, passwords, API keys, tokens, or private keys directly in `.black` files.
- Prefer environment references such as `env DATABASE_URL` for future secret-aware syntax.

## Validation Flow

After changing source or compiler behavior, run:

```bash
go test ./...
go run ./packages/cli/cmd/black --help
black format --check --json
black lint --json
black docs --all --json
black explain entity --json
```

When implemented, also run:

```bash
black validate --json
black build
```

## AI Learning Cost Rule

BlackLang is expected to be unknown to many AI models at first.

Do not assume the model knows the language from training. Use local docs, examples, JSON diagnostics, and CLI explain commands as the source of truth.

The goal is not zero learning cost. The goal is that repeated changes become cheaper than editing a normal web stack.
