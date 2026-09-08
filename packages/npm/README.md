# blacklang npm wrapper

This package is a thin launcher for the native BlackLang CLI.

It does not parse, validate, format, or build BlackLang source in JavaScript. It resolves a trusted native binary and forwards all arguments, stdin, stdout, stderr, and exit codes.

```bash
npx blacklang --help
npx blacklang ide --json
npx black --version
```

Use `BLACKLANG_BINARY` when a trusted local CLI binary already exists:

```bash
BLACKLANG_BINARY=/path/to/black npx blacklang validate app.black --json
```

Development validation:

```bash
cd packages/npm
npm test
```

Release installation downloads should be enabled only for published versions that have matching checksummed archives and detached Ed25519 signatures. Set `BLACKLANG_RELEASE_PUBLIC_KEY` or `BLACKLANG_RELEASE_PUBLIC_KEY_FILE` before enabling downloads. Local development can set `BLACKLANG_SKIP_DOWNLOAD=1`.
