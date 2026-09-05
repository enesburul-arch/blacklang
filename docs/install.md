# BlackLang Install Paths

## Purpose

This document defines the official install paths for BlackLang.

It separates current local development usage from planned public distribution through GitHub Releases and npm.

## Current Local Development Install

Inside this repository, build the Windows CLI binary with:

```powershell
go build -o ..\..\dist\black.exe ./cmd/black
```

from:

```text
packages/cli/
```

Then verify:

```powershell
.\dist\black.exe version
.\dist\black.exe version --json
.\dist\black.exe --help
.\dist\black.exe format --check --json
.\dist\black.exe lint --json
.\dist\black.exe docs --all --json
.\dist\black.exe explain entity --json
.\dist\black.exe validate --json
```

## Current Local Release Artifact

The local release artifact path is:

```text
artifacts/releases/<version>/
```

Build the Windows release archive:

```powershell
.\scripts\build-release-windows.ps1
```

Generate checksums and release metadata:

```powershell
.\scripts\write-release-checksums.ps1
```

Current local dev artifact example:

```text
artifacts/releases/v0.1.0-dev/blacklang-v0.1.0-dev-windows-amd64.zip
artifacts/releases/v0.1.0-dev/checksums.sha256
artifacts/releases/v0.1.0-dev/release.blackdir
```

## Planned GitHub Releases Install

After public releases begin, the canonical release page should be:

```text
https://github.com/blacklang/blacklang/releases
```

Artifact URL shape:

```text
https://github.com/blacklang/blacklang/releases/download/<version>/<artifact>
```

Examples:

```text
https://github.com/blacklang/blacklang/releases/download/v0.2.0/blacklang-v0.2.0-windows-amd64.zip
https://github.com/blacklang/blacklang/releases/download/v0.2.0/checksums.sha256
https://github.com/blacklang/blacklang/releases/download/v0.2.0/release.blackdir
```

## Windows Manual Install

Planned manual install flow:

1. Download `blacklang-<version>-windows-amd64.zip`.
2. Download `checksums.sha256`.
3. Verify the archive SHA-256 hash.
4. Extract the archive.
5. Put the extracted folder on `PATH`, or copy `black.exe` into a folder already on `PATH`.
6. Run `black version`.
7. Run `black version --json`.
8. Run `black --help`.

PowerShell verification example:

```powershell
$version = "v0.2.0"
$file = "blacklang-$version-windows-amd64.zip"
$expected = (Select-String -Path checksums.sha256 -Pattern $file).Line.Split(" ")[0]
$actual = (Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actual -ne $expected) { throw "checksum mismatch" }
```

## Planned npm Install

The planned npm wrapper command is:

```bash
npx blacklang --help
```

Project-local install:

```bash
npm install -D blacklang
npx black validate --json
npx black build
```

The npm wrapper should:

- detect the current OS and architecture
- download the matching GitHub Release archive
- verify the archive checksum
- extract the native Go binary into the package vendor directory
- forward `blacklang` and `black` commands to the native binary

The npm package design is documented in:

```text
docs/npm-wrapper.md
```

## CI and Offline Install

CI and offline users may bypass npm download behavior with:

```bash
BLACKLANG_BINARY=/path/to/black npx blacklang --help
```

Rules:

- `BLACKLANG_BINARY` must point to a trusted binary.
- CI should verify the binary version before running project commands.
- CI should run `black lint --json` and `black validate --json` before `black build`.

## Project-Level Usage

Once BlackLang is installed, a project should include:

```text
blacklang.toml
```

Example:

```toml
version = "0.1"
target = "web"
source = "examples/warehouse/app.black"
out = "generated"
```

Then users and AI agents can run:

```bash
black inspect --json
black version --json
black format --check --json
black lint --json
black docs --all --json
black explain entity --json
black validate --json
black build
```

## Current Status

As of v0.2 planning, GitHub Releases and npm publishing are planned install paths.

The local scripts already produce the release archive, checksum file, and release manifest needed for those future channels.
