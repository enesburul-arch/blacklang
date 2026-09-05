# BlackLang npm Wrapper Plan

## Purpose

The npm wrapper exists so users and AI agents can run BlackLang with:

```bash
npx blacklang --help
npx blacklang version --json
npx blacklang validate --json
npx blacklang build
```

The npm package should not reimplement BlackLang in JavaScript.

It is a small launcher that installs or finds the compiled Go CLI binary and forwards arguments to it.

## Package Name

Preferred package name:

```text
blacklang
```

Fallback package name if registry ownership requires it:

```text
@blacklang/cli
```

The final name must be checked at publish time. The wrapper design must work for either name.

## User-Facing Commands

The package should expose both commands:

```text
blacklang
black
```

Expected usage:

```bash
npx blacklang --help
npx blacklang init my-app
npx blacklang version --json
npx blacklang validate --json
npx blacklang build

npm install -D blacklang
npx black version --json
npx black validate --json
```

Rules:

- `npx blacklang` is the friendly first-run command.
- `black` is the CLI command users keep using inside projects.
- Both commands forward all arguments to the same native binary.

## Package Layout

Planned wrapper package:

```text
packages/npm/
├── package.json
├── README.md
├── bin/
│   └── blacklang.mjs
├── lib/
│   ├── platform.mjs
│   ├── release.mjs
│   └── checksum.mjs
├── scripts/
│   └── install.mjs
└── vendor/
    └── .gitkeep
```

## package.json Shape

```json
{
  "name": "blacklang",
  "version": "0.2.0",
  "description": "AI-native deterministic intent language CLI wrapper",
  "type": "module",
  "bin": {
    "blacklang": "./bin/blacklang.mjs",
    "black": "./bin/blacklang.mjs"
  },
  "scripts": {
    "postinstall": "node scripts/install.mjs"
  },
  "files": [
    "bin/",
    "lib/",
    "scripts/",
    "vendor/",
    "README.md"
  ]
}
```

Rules:

- npm package version should match the BlackLang CLI release version without the leading `v`.
- The wrapper must stay thin.
- The wrapper must not include `.black` source files.
- The wrapper must not require Python, Go, or npm packages after installation.
- The wrapper may use Node.js because npm/npx already require Node.

## Install Flow

`scripts/install.mjs` should:

1. Detect OS and CPU architecture.
2. Map Node platform names to release artifact names.
3. Download the matching archive from GitHub Releases.
4. Download `checksums.sha256` or `release.blackdir`.
5. Verify the archive SHA-256 hash.
6. Extract the archive into `vendor/<os>-<arch>/`.
7. Mark the binary executable on Linux/macOS.
8. Run `black version` to verify the binary.

Example mapping:

```text
win32 x64    -> windows amd64 -> black.exe
win32 arm64  -> windows arm64 -> black.exe
linux x64    -> linux amd64   -> black
linux arm64  -> linux arm64   -> black
darwin x64   -> darwin amd64  -> black
darwin arm64 -> darwin arm64  -> black
```

## Runtime Flow

`bin/blacklang.mjs` should:

1. Locate the installed native binary under `vendor/<os>-<arch>/`.
2. Respect `BLACKLANG_BINARY` when a user or CI provides a custom binary path.
3. Spawn the native binary with all user arguments.
4. Forward stdin, stdout, stderr, and exit code exactly.

Runtime must not parse BlackLang source itself.

## Release URL Source

The first implementation can use a fixed GitHub Releases URL pattern:

```text
https://github.com/blacklang/blacklang/releases/download/<version>/<artifact>
```

Future implementations may read a generated `release.blackdir` index from the docs site or package metadata.

## Checksum Rule

The wrapper must verify downloaded archives before extraction.

Accepted checksum sources:

- `checksums.sha256`
- `release.blackdir`

If both are present and disagree, installation must fail.

## Offline and CI Rules

The wrapper should support:

```bash
BLACKLANG_BINARY=/path/to/black npx blacklang --help
```

This allows CI systems, local development, and offline setups to use a trusted binary without downloading during install.

Future optional variables:

```text
BLACKLANG_RELEASE_BASE_URL
BLACKLANG_SKIP_DOWNLOAD
BLACKLANG_VENDOR_DIR
```

## Security Rules

- Never execute downloaded files before checksum verification.
- Never download `.black` source files as part of npm install.
- Never store secrets in npm package metadata.
- Fail closed when platform, URL, checksum, or extraction is unknown.
- Print clear errors with actionable suggestions.

## Future Alternative

If install scripts become undesirable, BlackLang can later move to platform packages:

```text
blacklang
@blacklang/windows-amd64
@blacklang/linux-amd64
@blacklang/darwin-arm64
```

In that model, the main `blacklang` package depends on optional platform packages.

The first v0.2 plan starts with the simpler release-download wrapper because it reuses the GitHub Release artifacts already defined in `docs/release-artifacts.md`.

User-facing install paths are documented in `docs/install.md`.
