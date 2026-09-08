package main

import (
	"sort"
	"strings"
)

var docs = map[string]DocEntry{
	"version": {
		Keyword: "version",
		Purpose: "Prints the installed BlackLang CLI version for humans, AI agents, and CI tools.",
		Syntax:  "black version | black version --json",
		Example: `black version
black version --json`,
		AgentNotes: []string{
			"Use black version --json in installers, CI, and AI agent startup checks.",
			"The JSON version field is the CLI version string.",
			"Plain black version remains stable for human terminal use and release scripts.",
		},
		Errors: []string{},
	},
	"format": {
		Keyword: "format",
		Purpose: "Formats .black source files into the deterministic project style.",
		Syntax:  "black format [file] [--check] [--stdout] [--json]",
		Example: `black format
black format examples/warehouse/app.black --check --json
black format app.black --stdout`,
		AgentNotes: []string{
			"Use black format --check --json before committing AI-written source changes.",
			"Use --stdout when you need the formatted source without modifying files.",
			"When no file is provided, BlackLang reads blacklang.toml and formats the configured source file.",
			"Format does not rewrite .black.enc files; decrypt to a trusted plaintext workspace, format, and re-encrypt.",
			"The JSON changed field tells agents whether the file already matched the deterministic style.",
		},
		Errors: []string{"FILE_READ_ERROR", "FILE_WRITE_ERROR", "FORMAT_REQUIRED", "UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE", "UNCLOSED_STRING", "UNEXPECTED_CHARACTER"},
	},
	"lint": {
		Keyword: "lint",
		Purpose: "Checks .black source formatting, syntax, semantic validity, and source-security findings without writing files.",
		Syntax:  "black lint [file] [--json]",
		Example: `black lint
black lint examples/warehouse/app.black --json`,
		AgentNotes: []string{
			"Use black lint --json after editing .black source and before running black build.",
			"The checks array reports format, parse, validate, and security status separately.",
			"The findings array contains diagnostics the agent should fix before build.",
			"Lint does not write files; use black format to apply formatting changes.",
			"Lint can read .black.enc source in memory when the key environment variable named in the encrypted header is set.",
		},
		Errors: []string{"FILE_READ_ERROR", "FORMAT_REQUIRED", "MISSING_ENCRYPTION_KEY", "INVALID_ENCRYPTED_SOURCE", "DECRYPTION_FAILED", "UNCLOSED_STRING", "UNEXPECTED_CHARACTER", "MISSING_APP", "UNKNOWN_TABLE_COLUMN", "HARDCODED_DATABASE_URL", "HARDCODED_TOKEN", "HARDCODED_PRIVATE_KEY"},
	},
	"inspect": {
		Keyword: "inspect",
		Purpose: "Prints project structure or a focused affected graph for AI agents before editing.",
		Syntax:  "black inspect [file] [--json|--ir] | black inspect [file] --affected <symbol> --json",
		Example: `black inspect examples/warehouse/app.black --json
black inspect examples/warehouse/app.black --ir
black inspect examples/warehouse/app.black --affected Product.stock --json
black inspect examples/warehouse/app.black --affected Product.index --json
black inspect examples/warehouse/app.black --affected migration --json
black inspect examples/warehouse/app.black --affected DemoProducts --json
black inspect examples/warehouse/app.black --affected InventoryIntegration --json
black inspect examples/warehouse/app.black --affected seed --json
black inspect examples/warehouse/app.black --affected i18n --json
black inspect examples/warehouse/app.black --affected ops --json`,
		AgentNotes: []string{
			"Use black inspect --json or --ir at project start to learn the current app structure.",
			"Use --affected before changing an entity, field, index, query, job, action, transaction, service, seed, migration, page, role, workflow, state, component, api, target, deploy, ops, i18n, security, database, auth, or app symbol.",
			"The affected JSON lists source symbols and generated files that should be validated after the edit.",
			"Unknown symbols return UNKNOWN_AFFECTED_SYMBOL instead of asking the model to guess.",
		},
		Errors: []string{"FILE_READ_ERROR", "MISSING_AFFECTED_SYMBOL", "UNKNOWN_AFFECTED_SYMBOL"},
	},
	"migrate": {
		Keyword: "migrate",
		Purpose: "Builds a read-only schema migration plan between two .black sources before deploying database shape changes.",
		Syntax:  "black migrate plan <old.black> <new.black> [--json|--ir]",
		Example: `black migrate plan app-v1.black app-v2.black --json
black migrate plan app-v1.black app-v2.black --ir`,
		AgentNotes: []string{
			"Run migrate plan before deploying entity field, relation, uniqueness, default, required, or index changes to an existing database.",
			"The command is read-only; it does not connect to or mutate a database.",
			"Use changes[].risk to separate safe, manual, and destructive database work.",
			"Field and entity renames are not guessed; declare first-class migration rename intent in the new .black source.",
			"Generated db:migrate:plan reports database reachability, ledger state, pending renames, and unsafe checks as JSON without printing database secrets.",
			"Generated db:migrate applies only declared rename migrations; generated db:setup/db:push still applies them before the current schema is created or pushed.",
			"Safe false does not mean parse failure; inspect errors[] to distinguish invalid source from manual/destructive migration work.",
		},
		Errors: []string{"MISSING_SCHEMA_MIGRATION_FILES", "FILE_READ_ERROR", "UNCLOSED_STRING", "UNEXPECTED_CHARACTER", "MISSING_APP", "DUPLICATE_ENTITY", "UNKNOWN_INDEX_FIELD", "INVALID_MIGRATION_RENAME", "UNKNOWN_MIGRATION_RENAME_TARGET"},
	},
	"migration": {
		Keyword: "migration",
		Purpose: "Declares explicit, deterministic schema rename intent that generated setup can apply before creating or pushing the current database schema.",
		Syntax:  "migration Name { rename entity OldEntity to NewEntity | rename field Entity.oldField to newField }",
		Example: `migration RenameProductName {
  rename entity ProductItem to Product
  rename field Product.title to name
}`,
		AgentNotes: []string{
			"Put migration blocks in the new/current .black source when an entity or stored field was renamed.",
			"The `to` entity or field must exist in the current source; the `from` entity or field must not remain in the current source.",
			"Generated setup applies entity renames before field renames inside a migration so table renames can safely precede column renames.",
			"Use black migrate plan old.black new.black --json to confirm renames appear as rename-table or rename-column instead of add/drop.",
			"Migration blocks generate migrations/manifest.json, migrations/*.sql, src/migrate.ts, prisma/schema.prisma ledger metadata, src/setup-db.ts runtime application, and db:migrate:plan/db:migrate package scripts.",
		},
		Errors: []string{"INVALID_MIGRATION_DECLARATION", "INVALID_MIGRATION_RENAME", "INVALID_MIGRATION_FIELD_RENAME", "INVALID_MIGRATION_RENAME_KIND", "UNCLOSED_MIGRATION", "DUPLICATE_MIGRATION", "MISSING_MIGRATION_RENAME", "UNKNOWN_MIGRATION_RENAME_ENTITY", "UNKNOWN_MIGRATION_RENAME_TARGET", "UNSUPPORTED_MIGRATION_RENAME_TARGET", "UNSUPPORTED_TARGET_DATABASE_MIGRATION", "MIGRATION_RENAME_SOURCE_STILL_EXISTS", "DUPLICATE_MIGRATION_RENAME"},
	},
	"seed": {
		Keyword: "seed",
		Purpose: "Declares deterministic local/demo fixture rows that generated database setup can apply safely.",
		Syntax:  "seed <Name> { source <Entity>; row <RowKey> { field literal; relationField ref <OtherRowKey> } }",
		Example: `seed DemoProducts {
  source Product

  row DemoProductLow {
    sku "LOW-001"
    name "Low Stock Widget"
    stock 3
  }
}

seed DemoOrders {
  source Order

  row DemoOrderDraft {
    customer ref DemoCustomerAcme
    total 125.50
  }
}`,
		AgentNotes: []string{
			"Seed is a top-level declaration.",
			"Each seed block targets one source entity and may contain one or more named rows.",
			"Row keys become stable generated id values, so rerunning db:setup/db:seed is idempotent for declared rows.",
			"Scalar values use typed literals: quoted strings for text/email/date/datetime, finite numbers for numeric fields, and true/false for booleans.",
			"Relation fields must use `ref RowKey`, and the referenced row must be declared for the related entity.",
			"Seed rows validate required fields, stored field existence, scalar types, relation refs, basic field constraints, and duplicate unique values before build.",
			"Generated web output emits src/seed.ts and wires db:setup to run db:seed after schema setup when seeds are declared.",
			"Do not put secrets, passwords, tokens, or production credentials in seed literals.",
		},
		Errors: []string{"INVALID_SEED_DECLARATION", "INVALID_SEED_SOURCE", "DUPLICATE_SEED_SOURCE", "INVALID_SEED_ROW", "INVALID_SEED_VALUE", "UNEXPECTED_SEED_TOKEN", "UNEXPECTED_SEED_ROW_TOKEN", "UNCLOSED_SEED", "UNCLOSED_SEED_ROW", "DUPLICATE_SEED", "INVALID_SEED_NAME", "SEED_NAME_COLLISION", "MISSING_SEED_SOURCE", "UNKNOWN_SEED_SOURCE", "DUPLICATE_SEED_ROW", "INVALID_SEED_FIELD", "DUPLICATE_SEED_FIELD", "UNSUPPORTED_SEED_FIELD", "UNKNOWN_SEED_FIELD", "UNSUPPORTED_SEED_REF", "SEED_VALUE_TYPE_MISMATCH", "SEED_VALUE_CONSTRAINT_MISMATCH", "DUPLICATE_SEED_UNIQUE_VALUE", "MISSING_SEED_FIELD", "SEED_RELATION_REQUIRES_REF", "UNKNOWN_SEED_REF"},
	},
	"diagnostics": {
		Keyword: "diagnostics",
		Purpose: "Documents stable BlackLang diagnostic codes and repair strategy for humans, AI agents, and CI tools.",
		Syntax:  "docs/diagnostics.md | black docs diagnostics --json",
		Example: `black validate app.black --json
black lint app.black --json
black docs diagnostics --json`,
		AgentNotes: []string{
			"Branch on diagnostic code, not message text.",
			"Fix parser diagnostics before semantic diagnostics.",
			"Fix source-security diagnostics before packaging or deployment.",
			"Use docs/diagnostics.md as the full stable local reference.",
		},
		Errors: []string{"FILE_READ_ERROR", "UNKNOWN_TABLE_COLUMN", "UNKNOWN_AFFECTED_SYMBOL", "HARDCODED_TOKEN", "MISSING_ENCRYPTION_KEY", "MISSING_DECRYPT_STDOUT"},
	},
	"agent": {
		Keyword: "agent",
		Purpose: "Prints the project startup checklist an AI agent should follow before editing BlackLang source.",
		Syntax:  "black agent startup [file] [--json|--ir]",
		Example: `black agent startup --json
black agent startup examples/warehouse/app.black --json
black agent startup --ir`,
		AgentNotes: []string{
			"Use black agent startup --json as the first command after entering an unfamiliar BlackLang project.",
			"The readFirst array tells the agent which local files to read before editing.",
			"The checklist array gives a deterministic project entry workflow.",
			"The commands array lists validation, inspection, build, and source-security commands with resolved source/out paths.",
			"If the configured source is .black.enc, set the header-declared key environment variable before parse, lint, validate, inspect, benchmark, or build.",
			"If success is false, use the errors array before making source changes.",
		},
		Errors: []string{"UNKNOWN_AGENT_COMMAND", "FILE_READ_ERROR", "FORMAT_REQUIRED", "MISSING_ENCRYPTION_KEY", "HARDCODED_TOKEN"},
	},
	"agent-contract": {
		Keyword: "agent-contract",
		Purpose: "Documents the current capability boundary and official workflow for AI agents using BlackLang.",
		Syntax:  "docs/ai-agent-contract.md | black docs agent-contract --json",
		Example: `black agent startup --json
black docs agent-contract --json
black docs --all --json`,
		AgentNotes: []string{
			"Official BlackLang work uses .black source files, .blackthm theme/profile files, and the black CLI.",
			"Do not invent unsupported syntax or present a normal HTML/JavaScript prototype as official BlackLang output.",
			"`script type=\"text/black\"` is not supported unless an official browser runtime exists.",
			"Current BlackLang is strongest for CRUD/admin-style generated web applications, not arbitrary frontend calculator/game logic.",
			"Entity computed display fields are supported, and bounded local value/if-else logic is supported inside custom actions and explicit API update blocks.",
			"Calculator-style client local expression state, arbitrary click handlers, and general-purpose variables outside bounded action/API handlers are not yet supported.",
			"Declared custom row-level actions are supported when bound through page actions; arbitrary frontend event handlers are not yet supported.",
			"If a task is outside the current boundary, state the limitation and either label a normal web prototype clearly or add the missing compiler feature first.",
		},
		Errors: []string{"UNKNOWN_DOC_KEYWORD"},
	},
	"theme": {
		Keyword: "theme",
		Purpose: "Inspects and migration-checks .blackthm UI theme/profile files used by AI agents and the CSS generator.",
		Syntax:  "black theme inspect [file] [--json|--ir] | black theme migrate <old.blackthm> <new.blackthm> [--json|--ir]",
		Example: `black theme inspect --json
black theme inspect examples/warehouse/theme.blackthm --json
black theme inspect examples/warehouse/theme.blackthm --ir
black theme migrate old.blackthm new.blackthm --json`,
		AgentNotes: []string{
			"Use .blackthm for deterministic UI tokens and mode slot profile metadata.",
			"Set theme = \"path/to/theme.blackthm\" in blacklang.toml so agents can discover it.",
			"Theme inspect returns profile.rules so agents know how compact inline UI values will be read.",
			"Theme inspect returns profile.modeGroups for standard box, text, table, and button semantics.",
			"Theme migrate compares an old and new theme and reports whether existing inline UI slot positions remain safe.",
			"When locked is true, baseline lines freeze existing mode slot prefixes.",
			"Hex colors should be quoted because # starts a comment outside strings.",
			"`ui <mode> = <slot...>;` declares the generator reading order for one UI mode.",
			"black build uses the configured .blackthm profile when mapping compact inline UI values to CSS properties.",
			"Treat .blackthm files as source assets, not generated output.",
		},
		Errors: []string{"FILE_READ_ERROR", "INVALID_THEME_DECLARATION", "MISSING_THEME_VERSION", "MISSING_UI_PROFILE", "INVALID_UI_BASELINE", "DUPLICATE_UI_BASELINE", "INVALID_UI_MODE", "DUPLICATE_UI_MODE", "DUPLICATE_UI_SLOT", "MISSING_STANDARD_UI_MODE", "MISSING_UI_LOCK_BASELINE", "LOCKED_UI_MODE_REMOVED", "NON_APPEND_ONLY_UI_SLOT", "MISSING_THEME_MIGRATION_FILES", "THEME_MIGRATION_NAME_CHANGED", "THEME_MIGRATION_TARGET_CHANGED", "THEME_VERSION_REGRESSION", "UI_PROFILE_MIGRATION_NAME_CHANGED", "UI_PROFILE_VERSION_REGRESSION", "UI_PROFILE_UNLOCKED", "UI_MODE_REMOVED", "UI_SLOT_MIGRATION_BREAK"},
	},
	"theme-migration": {
		Keyword: "theme-migration",
		Purpose: "Checks whether a new .blackthm profile can safely replace an old one without remapping existing inline UI values.",
		Syntax:  "black theme migrate <old.blackthm> <new.blackthm> [--json|--ir]",
		Example: `black theme migrate theme-v1.blackthm theme-v2.blackthm --json
black theme migrate theme-v1.blackthm theme-v2.blackthm --ir`,
		AgentNotes: []string{
			"Run theme migrate before replacing a theme used by existing .black source files.",
			"Existing theme name, target, profile name, and version direction must remain compatible.",
			"Every old UI mode must still exist in the new profile.",
			"Every old mode slot sequence must remain the exact prefix of the new mode slot sequence.",
			"Appending new slots at the end is safe and appears as slot-appended in changes.",
			"Inserting, reordering, or removing old slots is unsafe because compact UI values are positional.",
			"Use --json for CI and agent repair loops; use --ir for compact context.",
		},
		Errors: []string{"FILE_READ_ERROR", "MISSING_THEME_MIGRATION_FILES", "THEME_MIGRATION_NAME_CHANGED", "THEME_MIGRATION_TARGET_CHANGED", "THEME_VERSION_REGRESSION", "UI_PROFILE_MIGRATION_NAME_CHANGED", "UI_PROFILE_VERSION_REGRESSION", "UI_PROFILE_UNLOCKED", "UI_MODE_REMOVED", "UI_SLOT_MIGRATION_BREAK"},
	},
	"ui-profile": {
		Keyword: "ui-profile",
		Purpose: "Documents compact positional UI mode slot rules used by .blackthm profiles.",
		Syntax:  "baseline <mode> <slot...>; ui <mode> = <slot...>; -> ui <mode> <values...> [| <mode> <values...>...]",
		Example: `profile UICompact {
  version 2
  baseline box color width style pt pr pb pl radius place
  ui box = color width style pt pr pb pl radius place shadow;
  ui text = color size weight align;
}

form {
  fields email, password
  ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
}`,
		AgentNotes: []string{
			"Read profile.modes[].slots from black theme inspect --json before writing inline UI intent.",
			"Use `ui <mode> = <slot...>;` to declare the generator reading order for that mode.",
			"Slots are positional and are read left to right.",
			"Trailing missing values use CSS generation defaults.",
			"Extra values are errors because they cannot map to a known slot.",
			"Each slot name may appear only once inside a mode.",
			"After a profile is locked, existing slots are immutable and new slots are append-only.",
			"Locked profiles require baseline lines so the compiler can verify mode slots still start with the frozen order.",
			"Web profiles must include the standard box, text, table, and button mode groups.",
			"Use black theme migrate <old.blackthm> <new.blackthm> --json before replacing a profile used by existing source.",
		},
		Errors: []string{"INVALID_UI_BASELINE", "DUPLICATE_UI_BASELINE", "INVALID_UI_MODE", "DUPLICATE_UI_MODE", "DUPLICATE_UI_SLOT", "MISSING_STANDARD_UI_MODE", "MISSING_UI_LOCK_BASELINE", "LOCKED_UI_MODE_REMOVED", "NON_APPEND_ONLY_UI_SLOT", "UI_MODE_REMOVED", "UI_SLOT_MIGRATION_BREAK"},
	},
	"ui-modes": {
		Keyword: "ui-modes",
		Purpose: "Documents the standard BlackLang UI mode groups used by web theme profiles.",
		Syntax:  "ui box = <slots...>; ui text = <slots...>; ui table = <slots...>; ui button = <slots...>;",
		Example: `profile UICompact {
  version 1
  ui box = color width style pt pr pb pl radius place;
  ui text = color size weight align;
  ui table = color width style density zebra;
  ui button = bg color radius size variant;
}`,
		AgentNotes: []string{
			"box is for container border, spacing, radius, and placement.",
			"text is for typography on labels, headings, helper text, and body copy.",
			"table is for table-specific border, density, and row pattern styling.",
			"button is for action control styling such as submit, create, edit, and delete buttons.",
			"black theme inspect --json returns profile.modeGroups with purpose, appliesTo, and defaultSlots.",
			"Missing standard modes return MISSING_STANDARD_UI_MODE.",
		},
		Errors: []string{"MISSING_STANDARD_UI_MODE"},
	},
	"ui": {
		Keyword: "ui",
		Purpose: "Declares compact inline UI intent near fields, forms, tables, and action buttons.",
		Syntax:  "ui <mode> <values...> [| <mode> <values...>...]",
		Example: `entity Product {
  name text required ui text "#172026" 14 semibold left
}

page Products {
  source Product

  table {
    id ProductsTable
    class inventoryTable
    columns name
    ui table border 1 solid compact true
  }

  form {
    id ProductForm
    class inventoryForm
    fields name
    ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
  }

  actions create
  action create id CreateProductButton
  action create class primaryAction
  action create ui button primary white 6 md solid
}`,
		AgentNotes: []string{
			"Field-level ui is trailing metadata; keep normal field modifiers before ui.",
			"Field UI currently accepts box and text modes.",
			"Form UI currently accepts box, text, and button modes.",
			"Table UI currently accepts box, text, and table modes.",
			"Action button UI currently accepts button mode with `action <name> ui button ...`.",
			"Use `id Identifier` and `class ClassName...` inside table/form blocks for explicit generated UI identity.",
			"Use `action <name> id Identifier` and `action <name> class ClassName...` for action button identity.",
			"Generated IDs and classes are normalized to kebab-case; repeated row action IDs get generated suffixes.",
			"Values are compact positional data; read .blackthm profile.modes[].slots before generating them.",
			"Web builds generate stable .bl-ui-* CSS classes from supported inline UI intent.",
			"Full .blackthm token resolution is a later extension; current builds use standard v0.2 slots and safe defaults.",
		},
		Errors: []string{"INVALID_UI_INTENT", "INVALID_ACTION_UI", "INVALID_ACTION_INTENT", "INVALID_UI_ID", "INVALID_UI_CLASS", "DUPLICATE_UI_ID", "DUPLICATE_UI_CLASS", "UNSUPPORTED_UI_MODE", "UNSUPPORTED_UI_TARGET_MODE", "DUPLICATE_UI_INTENT", "UNKNOWN_ACTION_UI"},
	},
	"docs": {
		Keyword: "docs",
		Purpose: "Prints compact BlackLang reference entries for one keyword or every known keyword.",
		Syntax:  "black docs [keyword] [--json|--ir] | black docs --all --json",
		Example: `black docs entity --json
black docs --all --json
black docs --all --ir`,
		AgentNotes: []string{
			"Use black docs --all --json when an AI agent needs the complete compact local reference.",
			"Use black docs <keyword> --json when only one concept is relevant to the current edit.",
			"The --all JSON output is sorted by keyword for deterministic agent context.",
			"Prefer local docs output over guessing BlackLang syntax from model memory.",
		},
		Errors: []string{"UNKNOWN_DOC_KEYWORD"},
	},
	"ecosystem": {
		Keyword: "ecosystem",
		Purpose: "Prints deterministic release, package, registry, adapter, marketplace, extension, public index, and trust-policy discovery metadata for AI agents and tooling.",
		Syntax:  "black ecosystem [--json|--ir]",
		Example: `black ecosystem --json
black ecosystem --ir`,
		AgentNotes: []string{
			"Use black ecosystem --json when an agent needs to know which release scripts, package wrappers, registries, editor extensions, deploy adapters, provider marketplaces, and built-in target adapters currently exist.",
			"Use release.trust to read the signature algorithm, detached signature file pattern, public key env names, transparency log metadata, key rotation policy, and strict verification command before public install paths are trusted.",
			"Use registries[] to find package-index.blackdir and package trust policy files before npm/editor package publishing.",
			"Use marketplaces[] to find provider adapter indexes before cloud, observability, editor, or target adapter work.",
			"Use publicIndex to find the prepared website/ecosystem-index.json source for future hosted package/provider adapter discovery.",
			"Use trustWorkflow.requiredChecks before enabling an external package, adapter, or provider publish path.",
			"Provider-specific behavior belongs in extensions or adapters until parser, validator, docs, JSON/BlackIR, diagnostics, and affected graph support make the syntax official.",
			"The npm wrapper in packages/npm is a thin launcher; it must forward to the native CLI rather than reimplement BlackLang.",
			"The VS Code bridge in editors/vscode-blacklang consumes black ide output instead of maintaining stale local language rules.",
			"Editor package channel metadata covers VS Code, Open VSX, and Cursor-compatible VSIX publication readiness without publishing externally.",
		},
		Errors: []string{},
	},
	"package-registry": {
		Keyword: "package-registry",
		Purpose: "Documents the prepared local package registry manifest and trust checks for npm/editor package publishing, multi-editor channels, release transparency, and key rotation.",
		Syntax:  "packages/registry/package-index.blackdir | packages/registry/public-index.blackdir | packages/registry/trust-policy.blackdir | packages/registry/release-transparency.blackdir | packages/registry/key-rotation-policy.blackdir | black ecosystem --json",
		Example: `black ecosystem --json
black docs package-registry --json
node packages/registry/scripts/validate-registry.mjs`,
		AgentNotes: []string{
			"Use package-registry docs before npm, npx, VS Code, Open VSX, Cursor-compatible, or editor extension publishing work.",
			"Stable package IDs live in packages/registry/package-index.blackdir.",
			"The prepared public ecosystem index source is declared in packages/registry/public-index.blackdir and emitted at website/ecosystem-index.json.",
			"Editor package channel IDs include vscode:blacklang-vscode, openvsx:blacklang-vscode, and cursor:blacklang-vscode.",
			"Public publishing requires release.blackdir, checksums.sha256, detached Ed25519 signatures, release transparency log metadata, key rotation policy metadata, and release trust verification from the release artifact workflow.",
			"Release transparency and key rotation policy files are prepared-local metadata; they do not store secret values or private signing keys.",
			"Package wrappers and editor bridges must call the native black CLI instead of reimplementing BlackLang language behavior.",
			"Run node packages/registry/scripts/validate-registry.mjs before any external registry publish step.",
		},
		Errors: []string{},
	},
	"release-trust": {
		Keyword: "release-trust",
		Purpose: "Documents signed release verification, transparency log policy, and key rotation policy for CLI archives, npm wrapper downloads, editor packages, and provider adapter packages.",
		Syntax:  "node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict | packages/registry/release-transparency.blackdir | packages/registry/key-rotation-policy.blackdir | black ecosystem --json",
		Example: `node scripts/verify-release-trust.mjs artifacts/releases/v0.2.0 --json
node scripts/verify-release-trust.mjs artifacts/releases/v0.2.0 --json --strict
black ecosystem --json`,
		AgentNotes: []string{
			"Use release-trust docs before enabling GitHub Releases, npm downloads, editor package publishing, or provider adapter marketplace publishing.",
			"Release archives must match release.blackdir, checksums.sha256, detached Ed25519 .sig files, and a trusted public key.",
			"Public release entries should also have append-only transparency metadata with release, channel, artifact, sha256, signature, keyId, previousEntryHash, entryHash, and publishedAt fields.",
			"Key rotation policy uses sha256-public-key-spki-prefix key ids, a 30 day overlap rule, and an append-only revocation file.",
			"Public keys are read from BLACKLANG_RELEASE_PUBLIC_KEY or BLACKLANG_RELEASE_PUBLIC_KEY_FILE; private signing keys must never be stored in .black source, manifests, or package metadata.",
			"Run node scripts/verify-release-trust.mjs <release-dir> --json --strict and node packages/registry/scripts/validate-registry.mjs before trusting public install paths.",
			"Use black ecosystem --json release.trust for the current signature algorithm, signature file pattern, public key env names, transparency log metadata, key rotation policy, and verification command.",
		},
		Errors: []string{},
	},
	"adapter-marketplace": {
		Keyword: "adapter-marketplace",
		Purpose: "Documents the prepared local provider adapter marketplace manifest and trust rules for adapter installation or execution.",
		Syntax:  "adapters/marketplace/adapter-index.blackdir | adapters/marketplace/trust-policy.blackdir | black ecosystem --json",
		Example: `black ecosystem --json
black docs adapter-marketplace --json
node packages/registry/scripts/validate-registry.mjs`,
		AgentNotes: []string{
			"Use adapter-marketplace docs before adding provider-specific deploy, observability, editor, storage, email, or payment behavior.",
			"Stable adapter IDs live in adapters/marketplace/adapter-index.blackdir.",
			"Prepared hosted index metadata lives in packages/registry/public-index.blackdir and website/ecosystem-index.json.",
			"Editor adapter IDs include editor:vscode, editor:open-vsx, and editor:cursor-compatible.",
			"Provider credentials, endpoints, DSNs, app names, regions, and tokens must stay outside .black source.",
			"Provider adapter execution must start with read-only manifest, plan, or preflight checks before any external mutation, and apply mode must be explicit.",
			"Provider adapter packages must pass the signed release trust workflow before public marketplace publishing.",
			"Marketplace entries must include docs, capabilities, status, source ownership, trust flags, and validation commands.",
		},
		Errors: []string{},
	},
	"editor-marketplace": {
		Keyword: "editor-marketplace",
		Purpose: "Documents prepared multi-editor package channel metadata for the compiler-owned BlackLang editor bridge.",
		Syntax:  "packages/registry/package-index.blackdir | adapters/marketplace/adapter-index.blackdir | black ecosystem --json | black docs ide --json",
		Example: `black docs editor-marketplace --json
black ecosystem --json
node packages/registry/scripts/validate-registry.mjs
cd editors/vscode-blacklang && npm test
cd editors/vscode-blacklang && npm run package:vsix`,
		AgentNotes: []string{
			"Use editor-marketplace docs before VS Code, Open VSX, Cursor-compatible, or other editor package channel work.",
			"Current editor package channel IDs are vscode:blacklang-vscode, openvsx:blacklang-vscode, and cursor:blacklang-vscode.",
			"Current editor adapter IDs are editor:vscode, editor:open-vsx, and editor:cursor-compatible.",
			"All current editor channels reuse editors/vscode-blacklang and must consume black ide --json plus black ide diagnostics <file> --json.",
			"Public editor marketplace publication requires the same release.blackdir, checksums.sha256, detached Ed25519 signature, release transparency, and key rotation trust workflow as CLI and npm releases.",
			"Publishing tokens and marketplace credentials must stay outside .black source, generated output, manifests, and package metadata.",
		},
		Errors: []string{},
	},
	"explain": {
		Keyword: "explain",
		Purpose: "Prints an action-oriented explanation for one BlackLang keyword so AI agents can edit with less guessing.",
		Syntax:  "black explain <keyword> --json",
		Example: `black explain entity --json
black explain table --json
black explain syntax --json`,
		AgentNotes: []string{
			"Use black explain <keyword> --json when one concept needs deeper task guidance than black docs <keyword> --json.",
			"The output includes agentSteps, agentNotes, related keywords, and errorCodes.",
			"Explain is read-only and never writes project files.",
			"Prefer black docs --all --json when the agent needs the whole compact language reference.",
		},
		Errors: []string{"UNKNOWN_EXPLAIN_KEYWORD"},
	},
	"ide": {
		Keyword: "ide",
		Purpose: "Exports compiler-owned IDE metadata, completions, snippets, diagnostic codes, and live source diagnostics for editor integrations and AI agents.",
		Syntax:  "black ide [--json|--ir] | black ide diagnostics [file] [--json|--ir]",
		Example: `black ide --json
black ide --ir
black ide diagnostics examples/warehouse/app.black --json
black ide diagnostics examples/warehouse/app.black --ir`,
		AgentNotes: []string{
			"Use black ide --json when an editor, LSP bridge, or AI agent needs stable BlackLang keyword completions, snippets, file extensions, and diagnostic code metadata.",
			"Use black ide diagnostics <file> --json for editor-friendly zero-based ranges without rewriting the source file.",
			"IDE diagnostics reports valid=false when parse, validate, format, or source-security findings exist, but still returns success=true when the file was readable and diagnostics were produced.",
			"Completion and snippet data is deterministic and compiler-owned; do not copy stale snippet lists into editor integrations.",
		},
		Errors: []string{"UNKNOWN_IDE_COMMAND", "FILE_READ_ERROR", "FORMAT_REQUIRED", "UNCLOSED_STRING", "UNEXPECTED_CHARACTER", "MISSING_APP", "UNKNOWN_SOURCE_ENTITY", "HARDCODED_DATABASE_URL", "HARDCODED_TOKEN", "HARDCODED_PRIVATE_KEY"},
	},
	"syntax": {
		Keyword: "syntax",
		Purpose: "Explains the minimal current BlackLang source structure.",
		Syntax:  "app Name; target web { frontend react; backend node; database sqlite|postgres|mysql } | target api { backend node; database sqlite|postgres|mysql }; entity Name { field type modifiers... }; query Name { source Entity ... }; job Name { schedule every N minutes|hours|days; run query QueryName }; api Name { method GET; path \"/api/name\" }; service Name { api APIName }; page Name { source Entity ... }",
		Example: `app Warehouse
target web {
  frontend react
  backend node
  database sqlite
}
entity Product {
  sku text required unique
}
page Products {
  source Product

  view {
    order form, table, detail
  }
}`,
		AgentNotes: []string{
			"Read blacklang.toml first to find source and output paths.",
			"Use target web with frontend react, backend node, and database sqlite, postgres, or mysql for generated React apps.",
			"Use target api with backend node and database sqlite, postgres, or mysql when the project should emit API/runtime output without React or Vite files.",
			"Prefer changing .black source files instead of generated files.",
			"Run black validate --ir after edits.",
		},
		Errors: []string{"UNEXPECTED_TOP_LEVEL", "INVALID_TARGET_DECLARATION", "INVALID_ENTITY_DECLARATION", "INVALID_PAGE_DECLARATION"},
	},
	"app": {
		Keyword: "app",
		Purpose: "Declares the application name.",
		Syntax:  "app <PascalCaseName>",
		Example: "app Warehouse",
		AgentNotes: []string{
			"Use one app declaration per project.",
			"The app name appears in generated metadata and UI.",
		},
		Errors: []string{"MISSING_APP", "DUPLICATE_APP", "INVALID_APP_DECLARATION"},
	},
	"database": {
		Keyword: "database",
		Purpose: "Declares secret-safe database connection intent.",
		Syntax:  "database { url env <ENV_NAME> }",
		Example: `database {
  url env DATABASE_URL
}`,
		AgentNotes: []string{
			"Database is a top-level declaration.",
			"Current draft accepts one database block per project.",
			"Database url must reference an environment variable with `url env NAME`.",
			"The generated runtime reads that environment variable; when omitted, it uses DATABASE_URL.",
			"Do not write real connection strings, passwords, or tokens directly in .black source.",
			"Run black migrate plan old.black new.black --json before deploying schema-affecting entity changes to an existing database.",
			"Generated production deployments should avoid shipping protected .black source files.",
		},
		Errors: []string{"INVALID_DATABASE_DECLARATION", "DUPLICATE_DATABASE", "INVALID_DATABASE_URL", "MISSING_DATABASE_URL", "INVALID_ENV_NAME"},
	},
	"target": {
		Keyword: "target",
		Purpose: "Declares which application platform and generated stack the current .black source targets.",
		Syntax:  "target web { frontend react; backend node; database sqlite|postgres|mysql } | target api { backend node; database sqlite|postgres|mysql }",
		Example: `target api {
  backend node
  database sqlite
}`,
		AgentNotes: []string{
			"Use one target block per project.",
			"Draft v0.2 supports target web with frontend react, backend node, and database sqlite, postgres, or mysql.",
			"Draft v0.2 also supports target api with backend node and database sqlite, postgres, or mysql; omit the frontend line in API-only sources.",
			"API-only generation emits Express, Prisma, OpenAPI, validation, API smoke tests, optional seed/job/deploy output, and no React, Vite, frontend smoke, browser-check, browser e2e, or browser matrix files.",
			"Use `database postgres` for generated Prisma PostgreSQL provider, PrismaPg adapter, and PostgreSQL Docker Compose service output.",
			"Use `database mysql` for generated Prisma MySQL provider, PrismaMariaDb adapter, and MySQL Docker Compose service output; generated rename migration runtime remains limited to sqlite and postgres.",
			"When omitted, the current generator keeps the legacy default web/react/node/sqlite stack.",
			"Do not write mobile, desktop, or alternate backend targets until the generator supports them.",
			"The target block is the future handoff point for generator plugins and adapters.",
			"Use black inspect --affected target --json before changing target metadata in an existing project.",
		},
		Errors: []string{"INVALID_TARGET_DECLARATION", "DUPLICATE_TARGET", "UNCLOSED_TARGET", "UNEXPECTED_TARGET_TOKEN", "INVALID_TARGET_FRONTEND", "DUPLICATE_TARGET_FRONTEND", "MISSING_TARGET_FRONTEND", "UNSUPPORTED_TARGET_FRONTEND", "UNSUPPORTED_API_TARGET_TEST", "INVALID_TARGET_BACKEND", "DUPLICATE_TARGET_BACKEND", "MISSING_TARGET_BACKEND", "UNSUPPORTED_TARGET_BACKEND", "INVALID_TARGET_DATABASE", "DUPLICATE_TARGET_DATABASE", "MISSING_TARGET_DATABASE", "UNSUPPORTED_TARGET_DATABASE", "UNSUPPORTED_TARGET_DATABASE_MIGRATION", "MISSING_TARGET_NAME", "UNSUPPORTED_TARGET"},
	},
	"cors": {
		Keyword: "cors",
		Purpose: "Declares browser cross-origin API access intent from an environment-managed origin list.",
		Syntax:  "security { cors { origins env <ENV_NAME> credentials true|false } }",
		Example: `security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}`,
		AgentNotes: []string{
			"CORS is declared inside the top-level security block.",
			"Origins must reference an environment variable; do not hardcode production domains into .black source.",
			"The generated server reads comma-separated origins from the configured environment variable.",
			"Use credentials true when cookie auth must work across allowed browser origins.",
			"Generated CORS middleware rejects browser origins that are not listed in the environment value.",
		},
		Errors: []string{"INVALID_SECURITY_DECLARATION", "DUPLICATE_SECURITY", "INVALID_CORS_DECLARATION", "DUPLICATE_CORS", "INVALID_CORS_ORIGINS", "DUPLICATE_CORS_ORIGINS", "MISSING_CORS_ORIGINS", "INVALID_CORS_CREDENTIALS", "DUPLICATE_CORS_CREDENTIALS", "INVALID_ENV_NAME"},
	},
	"deploy": {
		Keyword: "deploy",
		Purpose: "Declares production deployment intent, generated environment wiring, and provider CLI execution boundaries for web targets.",
		Syntax:  "deploy { target docker; port env <ENV_NAME> default <PORT>; env <ENV_NAME> required|optional; preview local; rollback keep <COUNT>; cloud <fly|render|railway> app env <ENV_NAME> [region env <ENV_NAME>] }",
		Example: `deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
  preview local
  rollback keep 3
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}`,
		AgentNotes: []string{
			"Use one deploy block per project.",
			"Current draft supports target docker.",
			"`port env NAME default PORT` makes the generated server read its listen port from an environment variable.",
			"`env NAME required|optional` documents deployment-time environment variables without placing secrets in .black source.",
			"When target docker is declared, the web generator writes Dockerfile, .dockerignore, and docker-compose.yml.",
			"Generated output also writes security/secrets.json, scripts/secrets-plan.mjs, scripts/secrets-provider.mjs, and security:secrets:* package scripts for read-only secret/config readiness and provider preflight checks.",
			"`preview local` writes docker-compose.preview.yml, BLACKLANG_PREVIEW_PORT in .env.example, and deploy:preview package scripts for an isolated local preview stack.",
			"`rollback keep COUNT` writes deploy/manifest.json, deploy/rollback.json, scripts/rollback-plan.mjs, and deploy:rollback:plan package script.",
			"`cloud PROVIDER app env NAME region env NAME` writes deploy/cloud.json, deploy/manifest.json cloud metadata, scripts/cloud-plan.mjs, scripts/cloud-exec.mjs, and deploy:cloud:* package scripts.",
			"Cloud provider app and region values must be environment references; do not place provider app names, tokens, or regions directly in .black source.",
			"Rollback plan output is read-only metadata; it reports retained releases and a candidate without mutating infrastructure.",
			"Cloud plan output is read-only metadata; cloud preflight checks required env and provider CLI availability without mutating infrastructure.",
			"Cloud execution requires npm run deploy:cloud:exec, which calls scripts/cloud-exec.mjs --apply and redacts configured environment values from captured provider output.",
			"When ops health is declared, generated Dockerfile and docker-compose.yml include a healthcheck that probes that path.",
			"Generated Docker Compose keeps SQLite data under /app/data for sqlite targets and emits a PostgreSQL service for postgres targets.",
		},
		Errors: []string{"INVALID_DEPLOY_DECLARATION", "DUPLICATE_DEPLOY", "UNCLOSED_DEPLOY", "UNEXPECTED_DEPLOY_TOKEN", "INVALID_DEPLOY_TARGET", "DUPLICATE_DEPLOY_TARGET", "MISSING_DEPLOY_TARGET", "UNSUPPORTED_DEPLOY_TARGET", "INVALID_DEPLOY_PORT", "DUPLICATE_DEPLOY_PORT", "MISSING_DEPLOY_PORT_ENV", "INVALID_DEPLOY_PORT_DEFAULT", "INVALID_DEPLOY_ENV", "DUPLICATE_DEPLOY_ENV", "UNSUPPORTED_DEPLOY_ENV_MODE", "INVALID_DEPLOY_PREVIEW", "DUPLICATE_DEPLOY_PREVIEW", "UNSUPPORTED_DEPLOY_PREVIEW", "INVALID_DEPLOY_ROLLBACK", "DUPLICATE_DEPLOY_ROLLBACK", "UNSUPPORTED_DEPLOY_ROLLBACK", "INVALID_DEPLOY_ROLLBACK_KEEP", "INVALID_DEPLOY_CLOUD", "DUPLICATE_DEPLOY_CLOUD", "DEPLOY_CLOUD_REQUIRES_DOCKER", "UNSUPPORTED_DEPLOY_CLOUD_PROVIDER", "MISSING_DEPLOY_CLOUD_APP_ENV", "INVALID_ENV_NAME"},
	},
	"ops": {
		Keyword: "ops",
		Purpose: "Declares deterministic runtime operation signals for generated web deployments.",
		Syntax: `ops {
  health path <absolutePath>
  readiness path <absolutePath>
  metrics path <absolutePath>
  logging requests
  observe webhook|otlp endpoint env <ENV_NAME>
}`,
		Example: `ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}`,
		AgentNotes: []string{
			"Use one ops block per project.",
			"`health path` generates a public GET endpoint that reports process uptime and app/version metadata.",
			"`readiness path` generates a public GET endpoint that checks database connectivity and returns 503 when the database is unavailable.",
			"`metrics path` generates a public GET endpoint with in-process request, error, status-code, uptime, and start-time counters.",
			"`logging requests` emits one structured JSON console log line when each observed request finishes.",
			"`observe webhook endpoint env NAME` posts non-blocking structured request events to an env-referenced endpoint when that environment variable is set.",
			"`observe otlp endpoint env NAME` posts non-blocking OTLP HTTP JSON trace payloads to an env-referenced endpoint when that environment variable is set.",
			"Generated observe middleware creates or propagates W3C traceparent context and adds the response traceparent header.",
			"Observability endpoint values must come from environment variables; do not hardcode vendor URLs or tokens in .black source.",
			"Ops paths must be root-level public paths such as /healthz, /readyz, or /metrics; they cannot be /api, /openapi.json, or contain query strings, fragments, braces, whitespace, or unsafe characters.",
			"When deploy target docker is declared with a health endpoint, generated Dockerfile and docker-compose.yml include a Node-based healthcheck for that path.",
			"Ops endpoints are generated before API auth and CSRF middleware so deployment probes are not blocked by application login.",
			"Generated OpenAPI includes x-blacklang-observability metadata when observe is declared.",
			"Generated output includes ops/observability.json with provider, endpoint env, signal, trace context, delivery, and exporter metadata.",
		},
		Errors: []string{"INVALID_OPS_DECLARATION", "DUPLICATE_OPS", "UNCLOSED_OPS", "UNEXPECTED_OPS_TOKEN", "INVALID_OPS_HEALTH", "DUPLICATE_OPS_HEALTH", "INVALID_OPS_READINESS", "DUPLICATE_OPS_READINESS", "INVALID_OPS_METRICS", "DUPLICATE_OPS_METRICS", "INVALID_OPS_LOGGING", "DUPLICATE_OPS_LOGGING", "INVALID_OPS_OBSERVE", "DUPLICATE_OPS_OBSERVE", "MISSING_OPS_SIGNAL", "UNSUPPORTED_OPS_LOGGING", "UNSUPPORTED_OPS_OBSERVE_PROVIDER", "MISSING_OPS_OBSERVE_ENDPOINT_ENV", "INVALID_OPS_PATH", "DUPLICATE_OPS_PATH", "INVALID_ENV_NAME"},
	},
	"i18n": {
		Keyword: "i18n",
		Purpose: "Declares supported locales, the initial locale, runtime field text switching, UI copy labels, and locale-aware display formatting.",
		Syntax:  "i18n { default <locale>; locales <locale...> }; label <Entity>.<field>|app.<key>|page.<Page>|action.<key>|table.<key>|status.<key> { <locale> \"Text\" }; placeholder|help|message <Entity>.<field> { <locale> \"Text\" }",
		Example: `i18n {
  default tr
  locales tr, en
}

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}

label app.title {
  tr "Depo"
  en "Warehouse"
}

label page.Products {
  tr "Ürünler"
  en "Products"
}

label action.create {
  tr "Oluştur"
  en "Create"
}

placeholder Product.name {
  tr "Ürün adını gir"
  en "Enter product name"
}

help Product.name {
  tr "Listelerde görünen ürün adı"
  en "Visible product name in lists"
}

message Product.stock {
  tr "Geçerli bir stok adedi gir"
  en "Enter a valid stock count"
}`,
		AgentNotes: []string{
			"Use one i18n block per project.",
			"The default locale must be included in locales.",
			"Locale names can use letters, numbers, underscores, and hyphens, such as tr, en, or en-US.",
			"Top-level label blocks target stored fields and computed display fields with Entity.field.",
			"Top-level label blocks can also target generated UI copy with app.<key>, page.<Page>, action.<key>, action.<CustomAction>, action.<transition>, action.new.<Entity>, action.create.<Entity>, action.edit.<Entity>, action.view.<Entity>, table.<key>, and status.<key>.",
			"Top-level placeholder, help, and message blocks target stored form fields with Entity.field.",
			"When more than one locale is declared, generated web UI includes a language selector and passes the active locale to pages.",
			"Generated app chrome, navigation page names, table tools, status labels, CRUD buttons, custom action buttons, workflow transition buttons, and table/detail/form/filter/column labels re-render when the user changes locale.",
			"Generated form placeholders, help text, and field-level validation messages re-render when translated blocks exist.",
			"Generated table and detail number, integer, decimal, money, date, and datetime values use Intl formatting for the active locale.",
			"Generated app shell derives lang and dir from the active locale for basic RTL support.",
			"Runtime text falls back to the default locale translation, then inline field metadata, then deterministic generated fallback text.",
		},
		Errors: []string{"INVALID_I18N_DECLARATION", "DUPLICATE_I18N", "INVALID_I18N_DEFAULT", "DUPLICATE_I18N_DEFAULT", "MISSING_I18N_DEFAULT", "INVALID_I18N_LOCALES", "DUPLICATE_I18N_LOCALES", "MISSING_I18N_LOCALES", "INVALID_LOCALE", "DUPLICATE_LOCALE", "UNKNOWN_DEFAULT_LOCALE", "INVALID_LABEL_DECLARATION", "INVALID_LABEL_TRANSLATION", "UNCLOSED_LABEL", "MISSING_I18N", "DUPLICATE_LABEL_TARGET", "INVALID_LABEL_TARGET", "UNKNOWN_LABEL_TARGET", "MISSING_LABEL_TRANSLATION", "UNKNOWN_LABEL_LOCALE", "DUPLICATE_LABEL_LOCALE", "MISSING_DEFAULT_LABEL_TRANSLATION", "INVALID_PLACEHOLDER_DECLARATION", "INVALID_PLACEHOLDER_TRANSLATION", "UNCLOSED_PLACEHOLDER", "DUPLICATE_PLACEHOLDER_TARGET", "INVALID_PLACEHOLDER_TARGET", "UNKNOWN_PLACEHOLDER_TARGET", "UNSUPPORTED_PLACEHOLDER_TARGET", "MISSING_PLACEHOLDER_TRANSLATION", "UNKNOWN_PLACEHOLDER_LOCALE", "DUPLICATE_PLACEHOLDER_LOCALE", "MISSING_DEFAULT_PLACEHOLDER_TRANSLATION", "INVALID_HELP_DECLARATION", "INVALID_HELP_TRANSLATION", "UNCLOSED_HELP", "DUPLICATE_HELP_TARGET", "INVALID_HELP_TARGET", "UNKNOWN_HELP_TARGET", "UNSUPPORTED_HELP_TARGET", "MISSING_HELP_TRANSLATION", "UNKNOWN_HELP_LOCALE", "DUPLICATE_HELP_LOCALE", "MISSING_DEFAULT_HELP_TRANSLATION", "INVALID_MESSAGE_DECLARATION", "INVALID_MESSAGE_TRANSLATION", "UNCLOSED_MESSAGE", "DUPLICATE_MESSAGE_TARGET", "INVALID_MESSAGE_TARGET", "UNKNOWN_MESSAGE_TARGET", "UNSUPPORTED_MESSAGE_TARGET", "MISSING_MESSAGE_TRANSLATION", "UNKNOWN_MESSAGE_LOCALE", "DUPLICATE_MESSAGE_LOCALE", "MISSING_DEFAULT_MESSAGE_TRANSLATION"},
	},
	"label": {
		Keyword: "label",
		Purpose: "Sets fallback field labels, per-locale field labels, computed display labels, or generated UI copy labels.",
		Syntax:  "fieldName type label \"Text\" | label <Entity>.<field>|app.<key>|page.<Page>|action.<key>|table.<key>|status.<key> { <locale> \"Text\" }",
		Example: `name text required label "Product Name"

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}

label app.title {
  tr "Depo"
  en "Warehouse"
}`,
		AgentNotes: []string{
			"Use the field modifier form for a single fallback label.",
			"Use the top-level block form with i18n when multiple locales are needed.",
			"Top-level label translations override the field modifier for generated runtime UI field labels.",
			"Use Entity.computedField to translate computed display field labels.",
			"Use app.*, page.*, action.*, table.*, and status.* label targets for generated app chrome, navigation, table tools, status copy, CRUD buttons, custom actions, and workflow transition buttons.",
			"Keep storage field names stable; change displayed text with label metadata.",
		},
		Errors: []string{"MISSING_LABEL_VALUE", "INVALID_LABEL_DECLARATION", "INVALID_LABEL_TRANSLATION", "DUPLICATE_LABEL_TARGET", "INVALID_LABEL_TARGET", "UNKNOWN_LABEL_TARGET", "MISSING_LABEL_TRANSLATION", "UNKNOWN_LABEL_LOCALE", "DUPLICATE_LABEL_LOCALE", "MISSING_DEFAULT_LABEL_TRANSLATION"},
	},
	"placeholder": {
		Keyword: "placeholder",
		Purpose: "Sets fallback form input hints or per-locale generated form placeholder text.",
		Syntax:  "fieldName type placeholder \"Text\" | placeholder <Entity>.<storedField> { <locale> \"Text\" }",
		Example: `name text required placeholder "Enter product name"

placeholder Product.name {
  tr "Ürün adını gir"
  en "Enter product name"
}`,
		AgentNotes: []string{
			"Use the field modifier form for a single fallback input hint.",
			"Use the top-level block form with i18n when generated form placeholders need multiple locales.",
			"Top-level placeholder translations target stored fields only because computed display fields are not form inputs.",
			"Generated relation select fields use placeholder text as the empty option label.",
		},
		Errors: []string{"MISSING_PLACEHOLDER_VALUE", "INVALID_PLACEHOLDER_DECLARATION", "INVALID_PLACEHOLDER_TRANSLATION", "UNCLOSED_PLACEHOLDER", "MISSING_I18N", "DUPLICATE_PLACEHOLDER_TARGET", "INVALID_PLACEHOLDER_TARGET", "UNKNOWN_PLACEHOLDER_TARGET", "UNSUPPORTED_PLACEHOLDER_TARGET", "MISSING_PLACEHOLDER_TRANSLATION", "UNKNOWN_PLACEHOLDER_LOCALE", "DUPLICATE_PLACEHOLDER_LOCALE", "MISSING_DEFAULT_PLACEHOLDER_TRANSLATION"},
	},
	"help": {
		Keyword: "help",
		Purpose: "Sets fallback field help text or per-locale generated form help text.",
		Syntax:  "fieldName type help \"Text\" | help <Entity>.<storedField> { <locale> \"Text\" }",
		Example: `name text required help "Visible product name in lists"

help Product.name {
  tr "Listelerde görünen ürün adı"
  en "Visible product name in lists"
}`,
		AgentNotes: []string{
			"Use the field modifier form for a single fallback helper note.",
			"Use the top-level block form with i18n when generated form help text needs multiple locales.",
			"Top-level help translations target stored fields only because computed display fields are not form inputs.",
		},
		Errors: []string{"MISSING_HELP_VALUE", "INVALID_HELP_DECLARATION", "INVALID_HELP_TRANSLATION", "UNCLOSED_HELP", "MISSING_I18N", "DUPLICATE_HELP_TARGET", "INVALID_HELP_TARGET", "UNKNOWN_HELP_TARGET", "UNSUPPORTED_HELP_TARGET", "MISSING_HELP_TRANSLATION", "UNKNOWN_HELP_LOCALE", "DUPLICATE_HELP_LOCALE", "MISSING_DEFAULT_HELP_TRANSLATION"},
	},
	"message": {
		Keyword: "message",
		Purpose: "Sets fallback field validation messages or per-locale generated form field validation messages.",
		Syntax:  "fieldName type message \"Text\" | message <Entity>.<storedField> { <locale> \"Text\" }",
		Example: `stock number min 0 message "Enter a valid stock count"

message Product.stock {
  tr "Geçerli bir stok adedi gir"
  en "Enter a valid stock count"
}`,
		AgentNotes: []string{
			"Use the field modifier form for a single fallback field-level validation message.",
			"Use the top-level block form with i18n when generated frontend form field validation messages need multiple locales.",
			"Top-level message translations target stored fields only and apply to generated frontend field-level validation for that field.",
			"Entity-level validate ... message text remains explicit validation metadata in the current draft.",
		},
		Errors: []string{"MISSING_MESSAGE_VALUE", "INVALID_MESSAGE_DECLARATION", "INVALID_MESSAGE_TRANSLATION", "UNCLOSED_MESSAGE", "MISSING_I18N", "DUPLICATE_MESSAGE_TARGET", "INVALID_MESSAGE_TARGET", "UNKNOWN_MESSAGE_TARGET", "UNSUPPORTED_MESSAGE_TARGET", "MISSING_MESSAGE_TRANSLATION", "UNKNOWN_MESSAGE_LOCALE", "DUPLICATE_MESSAGE_LOCALE", "MISSING_DEFAULT_MESSAGE_TRANSLATION"},
	},
	"entity": {
		Keyword: "entity",
		Purpose: "Declares stored application data, fields, local media fields, relation load policy, read-only computed display fields, database index intent, and entity row policy intent.",
		Syntax:  "entity <PascalCaseName> { <fieldName> <primitive|EntityName> <modifiers...>; relationField EntityName load list|detail|query|mutation|none; computed <name> <type> = <numericExpression>; index <field...>; policy owner|tenant <requiredTextField> }",
		Example: `entity Product {
  tenantId text required default "default"
  sku text required unique
  name text required label "Product Name" placeholder "Enter product name" help "Shown under the input"
  photo image optional accept "image/*" label "Photo"
  stock number default 0
  price money default 0
  computed inventoryValue money = stock * price label "Inventory Value"
  policy tenant tenantId
  index sku, stock
}

entity Order {
  product Product required load list detail query
}`,
		AgentNotes: []string{
			"Use entity for data that should be stored or shown in pages.",
			"Fields are referenced by table, form, and search blocks.",
			"number and integer fields use the current generated integer runtime; use decimal or money for fractional values.",
			"Use image for visual media and file for generic local uploads; generated forms store data URL strings and do not configure provider storage.",
			"Use relation fields by naming another entity as the field type.",
			"Use load list, detail, query, mutation, or none on relation fields to control generated response attachment contexts; omitted load keeps the backward-compatible default of all contexts.",
			"Use computed for read-only display values derived from stored number-like fields.",
			"Use index field or index fieldA, fieldB for deterministic stored-field database indexes.",
			"Use policy owner ownerId or policy tenant tenantId to scope generated list, query, detail, mutation, archive, delete, and custom action routes by authenticated user context.",
			"Use label \"Text\" to control generated UI field labels.",
			"Use placeholder \"Text\" to control generated input hints.",
			"Use help \"Text\" to show persistent guidance under generated inputs.",
			"Use regex \"pattern\", url, and message \"Text\" for advanced validation intent.",
			"Use validate left <= right message \"Text\" inside an entity for cross-field validation.",
			"Use validate field required when otherField == value message \"Text\" for conditional required validation.",
		},
		Errors: []string{"DUPLICATE_ENTITY", "DUPLICATE_FIELD", "DUPLICATE_COMPUTED_FIELD", "UNSUPPORTED_FIELD_TYPE", "UNSUPPORTED_FIELD_MODIFIER", "UNSUPPORTED_RELATION_LOAD_FIELD", "MISSING_RELATION_LOAD_SCOPE", "UNSUPPORTED_RELATION_LOAD_SCOPE", "DUPLICATE_RELATION_LOAD_SCOPE", "DUPLICATE_RELATION_LOAD", "CONFLICTING_RELATION_LOAD_SCOPE", "INVALID_COMPUTED_FIELD", "INVALID_COMPUTED_EXPRESSION", "UNSUPPORTED_COMPUTED_FIELD_TYPE", "UNSUPPORTED_COMPUTED_OPERATOR", "UNKNOWN_COMPUTED_FIELD", "INCOMPATIBLE_COMPUTED_FIELD", "UNSUPPORTED_COMPUTED_FIELD_MODIFIER", "INVALID_ENTITY_INDEX", "UNKNOWN_INDEX_FIELD", "UNSUPPORTED_COMPUTED_INDEX_FIELD", "DUPLICATE_INDEX_FIELD", "DUPLICATE_ENTITY_INDEX", "UNSUPPORTED_ENTITY_INDEX_FIELD_COUNT", "INVALID_ENTITY_POLICY", "UNSUPPORTED_ENTITY_POLICY", "AUTH_REQUIRED_FOR_ENTITY_POLICY", "DUPLICATE_ENTITY_POLICY", "DUPLICATE_ENTITY_POLICY_FIELD", "UNKNOWN_ENTITY_POLICY_FIELD", "UNSUPPORTED_ENTITY_POLICY_FIELD", "MISSING_ENTITY_POLICY_REQUIRED_FIELD", "UNSUPPORTED_ENTITY_POLICY_UNIQUE_FIELD", "MISSING_LABEL_VALUE", "MISSING_PLACEHOLDER_VALUE", "MISSING_HELP_VALUE", "MISSING_ACCEPT_VALUE", "UNSUPPORTED_ACCEPT_CONSTRAINT", "UNSUPPORTED_IMAGE_ACCEPT", "INVALID_REGEX_CONSTRAINT", "UNSUPPORTED_URL_CONSTRAINT", "MISSING_MESSAGE_VALUE", "INVALID_ENTITY_VALIDATION", "UNKNOWN_VALIDATION_FIELD", "UNSUPPORTED_VALIDATION_OPERATOR", "INCOMPATIBLE_VALIDATION_FIELDS", "MISSING_VALIDATION_CONDITION", "INVALID_VALIDATION_LITERAL"},
	},
	"relation-load": {
		Keyword: "relation-load",
		Purpose: "Controls which generated API response contexts attach relation objects for entity-typed fields.",
		Syntax:  "<fieldName> <EntityName> [required|optional] load list|detail|query|mutation|none",
		Example: `entity Customer {
  name text required
}

entity Order {
  customer Customer required load detail query label "Customer"
}`,
		AgentNotes: []string{
			"Use load only on relation fields whose type is another entity.",
			"Supported scopes are list, detail, query, mutation, and none.",
			"list controls base page collection routes, detail controls item read routes, query controls bound custom query list routes, and mutation controls create/update/archive/restore/custom-action/workflow responses.",
			"Omitting load keeps the backward-compatible default: the relation is included in list, detail, query, and mutation responses.",
			"Use load none by itself to return only the generated relation ID field and never attach the relation object in generated API responses.",
			"Generated list and query routes batch relation IDs into one target lookup per loaded relation field; detail and mutation responses use the same helper for one item.",
			"Permission-aware generated routes attach a relation object only when the current role can read the source relation field and sanitize attached target records with the target entity read policy.",
			"Generated UI falls back to the relation ID when a relation object is not loaded, so table/detail rendering remains deterministic.",
			"Generated OpenAPI operations expose x-blacklang-relation-load-context and x-blacklang-relation-load metadata.",
			"Relation load policy does not add joins to custom query filters; query where/sort/aggregate still use stored primitive fields only.",
		},
		Errors: []string{"UNSUPPORTED_RELATION_LOAD_FIELD", "MISSING_RELATION_LOAD_SCOPE", "UNSUPPORTED_RELATION_LOAD_SCOPE", "DUPLICATE_RELATION_LOAD_SCOPE", "DUPLICATE_RELATION_LOAD", "CONFLICTING_RELATION_LOAD_SCOPE", "UNSUPPORTED_AUTH_USER_FIELD_MODIFIER"},
	},
	"media": {
		Keyword: "media",
		Purpose: "Declares deterministic local file and image fields for generated web forms, table/detail display, API validation, and OpenAPI metadata.",
		Syntax:  "<fieldName> file|image [required|optional] [accept \"MIME hint\"] [label \"Text\"] [help \"Text\"] [message \"Text\"]",
		Example: `entity Product {
  photo image optional accept "image/*" label "Photo" help "Local product image"
  specSheet file optional accept "application/pdf" label "Spec Sheet"
}

page Products {
  source Product
  table {
    columns photo, specSheet
  }
  form {
    fields photo, specSheet
  }
}`,
		AgentNotes: []string{
			"Use image for browser-rendered visual previews and file for generic downloadable attachments.",
			"Generated forms emit file inputs, read the selected file as a data URL, and submit that string through the normal JSON API validation path.",
			"Generated database schemas store file and image fields as nullable or required strings; no upload provider, bucket, token, or secret is declared in .black.",
			"accept is optional and only valid on file or image fields. image defaults to accept \"image/*\" when no accept value is declared.",
			"Image accept values must include image/. Use file with accept \"application/pdf\" or another MIME hint for non-image documents.",
			"Custom action media inputs and auth user media fields are not generated in this MVP; model media as entity fields.",
			"Generated OpenAPI schemas include x-blacklang-media and x-blacklang-encoding data-url metadata for AI agents and client generators.",
			"Use black inspect --affected Entity.mediaField --json, black docs media --json, and black explain media --json before changing media fields.",
		},
		Errors: []string{"UNSUPPORTED_FIELD_TYPE", "UNSUPPORTED_FIELD_MODIFIER", "MISSING_ACCEPT_VALUE", "UNSUPPORTED_ACCEPT_CONSTRAINT", "UNSUPPORTED_IMAGE_ACCEPT", "MISSING_LABEL_VALUE", "MISSING_HELP_VALUE", "MISSING_MESSAGE_VALUE", "UNSUPPORTED_ACTION_INPUT_TYPE", "UNSUPPORTED_AUTH_USER_FIELD_TYPE", "UNSUPPORTED_COMPUTED_FORM_FIELD", "UNKNOWN_TABLE_COLUMN", "UNKNOWN_FORM_FIELD"},
	},
	"policy": {
		Keyword: "policy",
		Purpose: "Declares ownership and tenant row scope for generated web routes without exposing policy fields as user-editable inputs.",
		Syntax:  "entity <Name> { <ownerId|tenantId> text required; policy owner <field>; policy tenant <field> }",
		Example: `auth {
  strategy emailPassword
  session cookie
  user {
    name text required
    email email required
  }
}

entity Order {
  tenantId text required default "default"
  ownerId text required default "system"
  total money default 0
  policy tenant tenantId
  policy owner ownerId
}

page Orders {
  source Order
  actions create, edit, delete
  table {
    columns total
  }
  form {
    fields total
  }
}`,
		AgentNotes: []string{
			"Declare policy lines inside an entity. The policy field itself must be declared as a stored text required field on that entity.",
			"Use policy owner ownerId to bind rows to the authenticated user's id. Use policy tenant tenantId to bind rows to the authenticated user's tenantId; generated auth defaults new users to \"default\" and the generated Users page can update tenantId when tenant policies exist.",
			"Policy fields are generated database columns because they are normal stored fields, but generated forms, API client input types, and custom actions cannot set them.",
			"Generated create and update validation stamps policy fields from current user context. Generated list, query, detail, archive, restore, delete, workflow transition, and custom action routes apply the same row scope.",
			"Policies require an auth block. Policy fields must be stored text fields, required, and non-unique; computed fields, relation fields, and unknown fields are rejected.",
			"Queries remain list-selection rules; entity policies are the row-authorization layer that query and CRUD routes share.",
			"Use black inspect --affected policy --json or --affected Entity.policyField --json before changing policy fields.",
		},
		Errors: []string{"INVALID_ENTITY_POLICY", "UNSUPPORTED_ENTITY_POLICY", "AUTH_REQUIRED_FOR_ENTITY_POLICY", "DUPLICATE_ENTITY_POLICY", "DUPLICATE_ENTITY_POLICY_FIELD", "UNKNOWN_ENTITY_POLICY_FIELD", "UNSUPPORTED_ENTITY_POLICY_FIELD", "MISSING_ENTITY_POLICY_REQUIRED_FIELD", "UNSUPPORTED_ENTITY_POLICY_UNIQUE_FIELD", "UNSUPPORTED_POLICY_FORM_FIELD", "UNSUPPORTED_ACTION_POLICY_FIELD"},
	},
	"index": {
		Keyword: "index",
		Purpose: "Declares deterministic database index intent for stored entity fields.",
		Syntax:  "entity <Name> { index <storedField> | index <storedFieldA>, <storedFieldB> }",
		Example: `entity Order {
  customer Customer required
  status text default draft
  createdAt datetime
  index customer, status
}`,
		AgentNotes: []string{
			"Declare index lines inside an entity block.",
			"Indexes may reference stored primitive fields or relation fields declared on the same entity.",
			"Computed display fields cannot be indexed because they are not database columns.",
			"Generated Prisma schema uses @@index with a deterministic map name.",
			"Generated SQLite setup creates CREATE INDEX IF NOT EXISTS statements.",
			"Use black inspect --affected Entity.index --json before changing index declarations.",
			"Index declarations do not change generated React pages or API response schemas.",
		},
		Errors: []string{"INVALID_ENTITY_INDEX", "UNKNOWN_INDEX_FIELD", "UNSUPPORTED_COMPUTED_INDEX_FIELD", "DUPLICATE_INDEX_FIELD", "DUPLICATE_ENTITY_INDEX", "UNSUPPORTED_ENTITY_INDEX_FIELD_COUNT"},
	},
	"query": {
		Keyword: "query",
		Purpose: "Declares a reusable, deterministic stored-record query and binds it to a page list.",
		Syntax: `query <Name> {
  source <Entity>
  where <field> <operator> <literal>
  aggregate <name> count
  aggregate <name> <sum|avg|min|max> <numericField>
  sort <field> <asc|desc>
  limit <1..1000>
}
page <Name> {
  source <Entity>
  query <QueryName>
  table { columns <field...> }
}`,
		Example: `app Inventory
entity Product {
  name text required
  stock number default 0
}
query LowStockProducts {
  source Product
  where stock < 10
  aggregate lowStockCount count
  aggregate totalStock sum stock
  sort stock asc
  limit 50
}
page LowStock {
  source Product
  query LowStockProducts
  table {
    columns name, stock
  }
}`,
		AgentNotes: []string{
			"Declare queries only at top level. A page keeps source Entity and selects one query Name with the same source.",
			"Use stored primitive fields only; computed, relation and generated system fields, parameters, joins, raw SQL, and raw expressions are unsupported.",
			"where is optional; repeat where lines for AND. Use == or != for all supported types; < <= > >= require numeric, date or datetime fields.",
			"aggregate is optional and repeatable. count has no field; sum, avg, min, and max require one stored number, integer, decimal, or money field.",
			"Quote text/email/file/image/date/datetime literals. Dates use YYYY-MM-DD; datetimes use RFC3339 with an explicit timezone. Numbers and true/false are unquoted; null and field-to-field comparisons are unsupported.",
			"number and integer use signed int32 whole values; decimal and money use finite decimal literals. Exponent, hexadecimal, and non-finite number spellings are unsupported.",
			"sort and limit are optional and may occur once. Default limit is 100; explicit limit is 1..1000. Default order is id asc, appended as a tie breaker to explicit sort.",
			"GET /api/<lowercase-page>/query applies row policy, archive policy, and where filters, then sort, then limit in the database. GET /api/<lowercase-page>/query/summary applies the same row/archive/where rules and returns declared aggregate values without sort or limit.",
			"Table search/filter/sort/pagination work within the returned subset. Base CRUD list and relation selectors keep their existing behavior.",
			"The query endpoints preserve page/auth/entity permissions and require read permission on every filter/sort/aggregate field; hidden query dependencies return 403.",
			"Queries select lists; they do not declare ownership or tenant authorization. Use entity policy lines for row scope shared by list, query, detail, and mutation routes. An unbound query exposes no endpoint.",
			"Use black inspect --affected QueryName --json and --affected Entity.field --json before edits. Do not confuse this with api-block query parameters (contract metadata).",
		},
		Errors: []string{"INVALID_QUERY_DECLARATION", "INVALID_QUERY_NAME", "DUPLICATE_QUERY", "QUERY_NAME_COLLISION", "MISSING_QUERY_SOURCE", "INVALID_QUERY_SOURCE", "DUPLICATE_QUERY_SOURCE", "UNKNOWN_QUERY_SOURCE", "INVALID_QUERY_WHERE", "DUPLICATE_QUERY_WHERE", "INVALID_QUERY_LITERAL", "QUERY_LITERAL_TYPE_MISMATCH", "INVALID_QUERY_FIELD", "UNKNOWN_QUERY_FIELD", "UNSUPPORTED_QUERY_FIELD", "UNSUPPORTED_QUERY_OPERATOR", "INVALID_QUERY_AGGREGATE", "DUPLICATE_QUERY_AGGREGATE", "UNSUPPORTED_QUERY_AGGREGATE", "MISSING_QUERY_AGGREGATE_FIELD", "UNSUPPORTED_QUERY_AGGREGATE_FIELD", "INVALID_QUERY_SORT", "DUPLICATE_QUERY_SORT", "UNSUPPORTED_QUERY_SORT_DIRECTION", "INVALID_QUERY_LIMIT", "DUPLICATE_QUERY_LIMIT", "UNEXPECTED_QUERY_TOKEN", "UNCLOSED_QUERY", "INVALID_PAGE_QUERY", "DUPLICATE_PAGE_QUERY", "UNKNOWN_PAGE_QUERY", "PAGE_QUERY_SOURCE_MISMATCH"},
	},
	"job": {
		Keyword: "job",
		Purpose: "Declares a deterministic generated background worker job that runs a stored-field query on a schedule.",
		Syntax: `job <Name> {
  schedule every <integer> minutes|hours|days
  run query <QueryName>
}`,
		Example: `query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

job LowStockMonitor {
  schedule every 15 minutes
  run query LowStockProducts
}`,
		AgentNotes: []string{
			"Declare jobs only at top level with one PascalCase name. The MVP accepts one schedule line and one run line.",
			"`schedule every N minutes|hours|days` is the single supported schedule shape. Supported ranges are 1..1440 minutes, 1..168 hours, or 1..365 days.",
			"`run query QueryName` is the only run mode in this MVP. The referenced query must be a declared top-level query.",
			"The generated worker applies the query's where, sort, and limit rules, selects only IDs, and logs compact JSON metadata with count/limit/source/query.",
			"Generated output includes jobs/manifest.json, src/worker.ts, package scripts jobs:run and jobs:loop, and OpenAPI root x-blacklang-jobs metadata.",
			"Jobs run in internal worker scope and expose no public endpoint. Queue providers, retries, delayed queues, mutating jobs, external calls, and provider schedulers are future adapter work.",
			"Use black inspect --affected JobName --json or --affected QueryName --json before changing a job or its query. Run generated npm run jobs:run after build for runtime proof.",
		},
		Errors: []string{"INVALID_JOB_DECLARATION", "INVALID_JOB_NAME", "DUPLICATE_JOB", "JOB_NAME_COLLISION", "INVALID_JOB_SCHEDULE", "DUPLICATE_JOB_SCHEDULE", "MISSING_JOB_SCHEDULE", "UNSUPPORTED_JOB_SCHEDULE", "UNSUPPORTED_JOB_SCHEDULE_UNIT", "INVALID_JOB_SCHEDULE_INTERVAL", "INVALID_JOB_RUN", "DUPLICATE_JOB_RUN", "MISSING_JOB_RUN", "UNSUPPORTED_JOB_RUN", "INVALID_JOB_QUERY", "UNKNOWN_JOB_QUERY", "UNKNOWN_JOB_QUERY_SOURCE", "UNEXPECTED_JOB_TOKEN", "UNCLOSED_JOB"},
	},
	"action": {
		Keyword: "action",
		Purpose: "Declares a deterministic row-level custom mutation and exposes it through page actions.",
		Syntax: `action <Name> {
  source <Entity>
  input <name> <fieldType> <modifiers...>
  value <name> = <value|numericExpression>
  if <condition>
    set <storedField> = <value|numericExpression>
  else
    set <storedField> = <value|numericExpression>
  set <storedField> = <value|numericExpression>
  allow <RoleName...|authenticated>
  success "Message"
}
page <Name> {
  source <Entity>
  actions <crudAction...>, <ActionName>
}`,
		Example: `action RestockProduct {
  source Product
  input quantity number required min 1 label "Quantity"
  value restockValue = quantity
  if restockValue > 0 and stock >= 0
    set stock = stock + restockValue
  else
    set stock = stock
  allow Admin, Worker
  success "Stock updated"
}

page Products {
  source Product
  actions edit, RestockProduct
}`,
		AgentNotes: []string{
			"Declare custom actions only at top level with one PascalCase name. Bind them by name inside a page actions list whose source matches the action source.",
			"Custom actions are row-level mutations. A bound page generates POST /api/<lowercase-page>/:id/actions/<lowercase-action> plus a typed API client method and a generated form panel.",
			"Inputs use normal primitive field types and validation modifiers. number/integer inputs validate as whole numbers; decimal/money inputs allow finite fractional values.",
			"Input names must not collide with source fields, so identifiers in set expressions are unambiguous.",
			"value declares an ordered local value that later value, set, and if statements can reference. Local value names must not collide with inputs, source fields, computed fields, or reserved words.",
			"if uses `if condition` with indentation-based then/else branches. else aligns with the matching if. Branch statements may be value, set, or nested if.",
			"Condition comparisons use ==, !=, <, <=, >, and >=. Ordered comparisons currently require numeric values. Join comparisons with and, or, not, and parentheses; precedence is not before and before or.",
			"set can assign one stored primitive source field or use bounded numeric expressions with source fields, inputs, local values, typed literals, +, -, *, /, parentheses, and deterministic precedence. Computed fields, relation fields, policy fields, joins, function calls, loops, and side effects are unsupported in the MVP.",
			"Use a top-level transaction block when the generated action route must run source row lookup, field update, and audit log write inside one Prisma transaction.",
			"allow is optional. When present it requires auth and narrows who can run the action in addition to page access, update permission, field-level update checks, and any entity row policy.",
			"Unbound custom actions do not create endpoints. Query-bound pages refresh from the server after a custom action instead of guessing list membership locally.",
			"Use black inspect --affected ActionName --json before editing an action; use black docs action --json or black explain action --json when an agent needs the compact contract.",
		},
		Errors: []string{"INVALID_ACTION_DECLARATION", "INVALID_ACTION_NAME", "DUPLICATE_ACTION", "ACTION_NAME_COLLISION", "MISSING_ACTION_SOURCE", "INVALID_ACTION_SOURCE", "DUPLICATE_ACTION_SOURCE", "UNKNOWN_ACTION_SOURCE", "UNSUPPORTED_ACTION_TRANSACTION", "INVALID_ACTION_INPUT", "DUPLICATE_ACTION_INPUT", "ACTION_INPUT_FIELD_COLLISION", "UNSUPPORTED_ACTION_INPUT_TYPE", "UNSUPPORTED_ACTION_INPUT_MODIFIER", "MISSING_DEFAULT_VALUE", "MISSING_LABEL_VALUE", "MISSING_PLACEHOLDER_VALUE", "MISSING_HELP_VALUE", "MISSING_MESSAGE_VALUE", "INVALID_ACTION_SET", "MISSING_ACTION_SET", "DUPLICATE_ACTION_SET", "UNKNOWN_ACTION_FIELD", "UNSUPPORTED_ACTION_FIELD", "UNSUPPORTED_ACTION_POLICY_FIELD", "INVALID_ACTION_VALUE", "INVALID_ACTION_VALUE_NAME", "DUPLICATE_ACTION_VALUE", "ACTION_VALUE_NAME_COLLISION", "UNKNOWN_ACTION_VALUE", "UNSUPPORTED_ACTION_REFERENCE", "ACTION_VALUE_TYPE_MISMATCH", "INCOMPATIBLE_ACTION_EXPRESSION", "UNSUPPORTED_ACTION_OPERATOR", "INVALID_ACTION_EXPRESSION", "INVALID_ACTION_IF", "MISSING_ACTION_IF_BODY", "INVALID_ACTION_ELSE", "MISSING_ACTION_ELSE_BODY", "UNSUPPORTED_ACTION_CONDITION_OPERATOR", "ACTION_CONDITION_TYPE_MISMATCH", "INCOMPATIBLE_ACTION_CONDITION", "INVALID_ACTION_ALLOW", "DUPLICATE_ACTION_ALLOW", "AUTH_REQUIRED_FOR_ACTION_ALLOW", "UNKNOWN_ACTION_ALLOW_ROLE", "INVALID_ACTION_SUCCESS", "DUPLICATE_ACTION_SUCCESS", "UNEXPECTED_ACTION_TOKEN", "UNEXPECTED_ACTION_ELSE", "UNEXPECTED_ACTION_BRANCH_TOKEN", "UNCLOSED_ACTION", "PAGE_ACTION_SOURCE_MISMATCH", "UNSUPPORTED_ACTION"},
	},
	"transaction": {
		Keyword: "transaction",
		Purpose: "Declares an atomic generated runtime boundary for selected custom actions and explicit API update handlers.",
		Syntax: `transaction <Name> {
  action <ActionName>
  api <APIName>
}`,
		Example: `action RestockProduct {
  source Product
  input quantity number required min 1
  set stock = stock + quantity
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
}

transaction RestockAtomic {
  action RestockProduct
  api StockWebhook
}`,
		AgentNotes: []string{
			"Declare transaction blocks only at top level with one PascalCase name.",
			"Transaction blocks do not create endpoints and do not declare mutations. They bind existing action/API update declarations to a generated Prisma transaction boundary.",
			"Each action or api target may appear in at most one transaction block.",
			"API targets must be explicit APIs with an update handler; read-only declared APIs do not need transaction boundaries.",
			"Generated OpenAPI operations include x-blacklang-transaction and x-blacklang-transaction-name metadata for targeted routes.",
			"Use black inspect --affected TransactionName --json before editing a transaction block; use black docs action --json or black docs api --json before changing the targeted mutation.",
		},
		Errors: []string{"INVALID_TRANSACTION_DECLARATION", "INVALID_TRANSACTION_NAME", "TRANSACTION_NAME_COLLISION", "DUPLICATE_TRANSACTION", "EMPTY_TRANSACTION", "INVALID_TRANSACTION_ACTION", "INVALID_TRANSACTION_API", "DUPLICATE_TRANSACTION_TARGET", "UNKNOWN_TRANSACTION_ACTION", "UNKNOWN_TRANSACTION_API", "UNSUPPORTED_TRANSACTION_API", "DUPLICATE_TRANSACTION_BINDING", "UNEXPECTED_TRANSACTION_TOKEN", "UNCLOSED_TRANSACTION"},
	},
	"service": {
		Keyword: "service",
		Purpose: "Groups existing explicit API declarations into a deterministic generated service module and OpenAPI service boundary.",
		Syntax: `service <Name> {
  api <APIName>
}`,
		Example: `api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
}

service InventoryIntegration {
  api StockWebhook
}`,
		AgentNotes: []string{
			"Declare service blocks only at top level with one PascalCase name.",
			"Service blocks do not create endpoints and do not declare request handlers. They bind existing explicit api declarations to a generated service manifest, service module, and OpenAPI tag/extension metadata.",
			"Each api target may appear in at most one service block so inspect --affected output stays unambiguous.",
			"Service targets must be explicit top-level api declarations; generated page CRUD, query, action, auth, ops, and workflow routes are not service targets in the MVP.",
			"Generated output includes services/manifest.json, src/services/<service>.ts, OpenAPI x-blacklang-services, route-level x-blacklang-service, and generated contract assertions.",
			"Use black inspect --affected ServiceName --json before changing service grouping; use black inspect --affected APIName --json before changing a targeted API.",
		},
		Errors: []string{"INVALID_SERVICE_DECLARATION", "INVALID_SERVICE_NAME", "SERVICE_NAME_COLLISION", "DUPLICATE_SERVICE", "EMPTY_SERVICE", "INVALID_SERVICE_API", "UNKNOWN_SERVICE_API", "DUPLICATE_SERVICE_TARGET", "DUPLICATE_SERVICE_BINDING", "UNEXPECTED_SERVICE_TOKEN", "UNCLOSED_SERVICE"},
	},
	"computed": {
		Keyword: "computed",
		Purpose: "Declares a read-only display field calculated from stored number-like entity fields.",
		Syntax:  "computed <name> <number|integer|decimal|money> = <numericExpression> [label \"Text\"] [help \"Text\"]",
		Example: `entity Product {
  stock number default 0
  price money default 0
  tax money default 0
  computed inventoryValue money = stock * (price + tax) label "Inventory Value"
}

page Products {
  source Product

  table {
    columns stock, price, inventoryValue
  }
}`,
		AgentNotes: []string{
			"Computed fields are not database columns and are not submitted by generated forms.",
			"Computed fields can be shown in table columns and generated detail views.",
			"Computed expressions support +, -, *, /, parentheses, and deterministic operator precedence.",
			"Operands must be stored number-like fields on the same entity or numeric literals.",
			"Computed fields are hidden when any referenced source field is hidden by field-level permissions.",
			"Search, filter, and sort over computed fields are planned for a later data logic phase.",
		},
		Errors: []string{"INVALID_COMPUTED_FIELD", "INVALID_COMPUTED_EXPRESSION", "DUPLICATE_COMPUTED_FIELD", "DUPLICATE_FIELD", "UNSUPPORTED_COMPUTED_FIELD_TYPE", "UNSUPPORTED_COMPUTED_OPERATOR", "UNKNOWN_COMPUTED_FIELD", "INCOMPATIBLE_COMPUTED_FIELD", "UNSUPPORTED_COMPUTED_FIELD_MODIFIER", "UNSUPPORTED_COMPUTED_FORM_FIELD", "UNSUPPORTED_COMPUTED_SEARCH_FIELD", "UNSUPPORTED_COMPUTED_FILTER_FIELD", "UNSUPPORTED_COMPUTED_SORT_FIELD"},
	},
	"page": {
		Keyword: "page",
		Purpose: "Declares a generated web screen bound to one source entity.",
		Syntax:  "page <PascalCaseName> { layout <LayoutName>; source <EntityName>; query <QueryName>; view {...}; table {...}; form {...}; actions ... }",
		Example: `page Products {
  layout AdminLayout
  source Product
  view {
    order table, detail, form
    compose tabs gap md
    tab List sections table
    tab Record sections detail, form
  }
  actions create, edit, delete, archive, restore
}`,
		AgentNotes: []string{
			"Every page should have a source entity in v0.1.",
			"Use layout when the page should belong to an explicit generated app shell.",
			"Use query QueryName when the page should list a deterministic query subset for the same source entity.",
			"Use actions to expose CRUD actions and bound top-level custom actions for the page source.",
			"Use view when table, detail, form, and component sections should render in a specific DOM/visual order or page-level composition.",
			"Use view component sections when a declared component should render as a reusable page panel bound to selected, first, or each source record data.",
			"Use compose tabs and tab declarations when generated table/detail/form sections should render behind tab controls.",
			"Use section display modal or drawer when generated detail/form panels should open as overlays.",
			"Use view groups when contiguous inline sections should share a nested generated layout wrapper.",
			"Use view triggers when generated section interactions should switch tabs, open panels, or return to the table deterministically.",
			"Generated React pages are based on page blocks.",
		},
		Errors: []string{"DUPLICATE_PAGE", "MISSING_PAGE_SOURCE", "UNKNOWN_SOURCE_ENTITY", "UNKNOWN_PAGE_LAYOUT", "INVALID_PAGE_QUERY", "DUPLICATE_PAGE_QUERY", "UNKNOWN_PAGE_QUERY", "PAGE_QUERY_SOURCE_MISMATCH", "PAGE_ACTION_SOURCE_MISMATCH", "DUPLICATE_VIEW", "MISSING_VIEW_ORDER", "UNSUPPORTED_VIEW_SECTION", "DUPLICATE_VIEW_SECTION", "UNKNOWN_VIEW_COMPONENT", "MISSING_VIEW_COMPONENT_BIND", "UNSUPPORTED_VIEW_COMPONENT_BIND", "UNKNOWN_VIEW_COMPONENT_INPUT_FIELD", "VIEW_COMPONENT_INPUT_TYPE_MISMATCH", "UNSUPPORTED_VIEW_COMPONENT_INPUT", "CONFLICTING_VIEW_COMPONENT_SECTION", "UNSUPPORTED_VIEW_SECTION_BIND", "UNSUPPORTED_VIEW_COMPOSE_MODE", "UNSUPPORTED_VIEW_SECTION_SPAN", "UNSUPPORTED_VIEW_SECTION_DISPLAY", "UNSUPPORTED_VIEW_SECTION_SIDE", "INVALID_VIEW_GROUP", "DUPLICATE_VIEW_GROUP", "UNSUPPORTED_VIEW_GROUP_SECTION", "INVALID_VIEW_TAB", "MISSING_VIEW_TABS", "DUPLICATE_VIEW_TAB", "MISSING_VIEW_TAB_SECTION", "INVALID_VIEW_TRIGGER", "DUPLICATE_VIEW_TRIGGER", "UNSUPPORTED_VIEW_TRIGGER_SECTION", "UNSUPPORTED_VIEW_TRIGGER_EVENT", "UNSUPPORTED_VIEW_TRIGGER", "UNSUPPORTED_VIEW_TRIGGER_ACTION"},
	},
	"view": {
		Keyword: "view",
		Purpose: "Declares generated page section DOM order, reusable component sections, layout composition, tabs, modal/drawer display, and deterministic section interaction intent without editing generated React or CSS.",
		Syntax:  "view { order table, StockSummary, detail, form; compose stack|grid|tabs [columns 1..4] [gap sm|md|lg] [stackAt sm|md|lg|none]; section table|detail|form [span 1..4] [display inline|modal|drawer] [side left|right] [title \"Text\"]; section <Name> component <Component> bind selected|first|each [span 1..4] [title \"Text\"]; group <Name> sections table, StockSummary [compose stack|grid] [columns 1..4] [gap sm|md|lg] [span 1..4] [title \"Text\"]; tab <Name> sections table, StockSummary; trigger <section> on rowSelect|createStart|editStart|saveSuccess|close }",
		Example: `page Products {
  source Product

  view {
    order table, StockSummary, StockCards, detail, form
    compose grid columns 2 gap md stackAt md
    section table span 2
    section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
    section StockCards component StockBadge bind each span 1 title "Stock Cards"
    section detail display drawer side right title "Product Details"
    section form display modal title "Product Form"
    trigger detail on rowSelect
    trigger form on createStart
    trigger form on editStart
    trigger detail on saveSuccess
    trigger table on close
  }

  table {
    columns sku, name, stock
  }

  form {
    fields sku, name, stock
  }
}`,
		AgentNotes: []string{
			"View is declared inside a page block.",
			"Current web output supports table, detail, form, and explicitly declared component sections.",
			"Listed sections render first; omitted supported sections are appended in the default order table, detail, form.",
			"Generated JSX/DOM section order follows the effective view order; CSS order rules remain stable metadata and a layout backstop.",
			"Use section <Name> component <Component> bind selected to render a reusable component from the selected detail record.",
			"Use section <Name> component <Component> bind first to render a reusable component from the first loaded list record.",
			"Use section <Name> component <Component> bind each to render the component once per loaded list/query record.",
			"Component section inputs bind by matching input name and type to stored or computed fields on the page source entity.",
			"Component sections render inline in this MVP and are hidden when their required field permissions are hidden.",
			"compose stack keeps vertical section flow; compose grid emits deterministic CSS grid rules and responsive breakpoint media queries.",
			"compose tabs emits generated React tab state and tab controls for table/detail/form/component sections.",
			"grid columns defaults to 2 when omitted; stackAt defaults to md for grid and may be none.",
			"Grid stackAt sm may step 4 columns down through lg/md/sm; stackAt md and lg collapse earlier.",
			"Section spans that exceed a responsive column count are clamped at that breakpoint.",
			"section display inline keeps normal page flow; display modal and display drawer create generated overlay wrappers for detail/form sections.",
			"Drawer side defaults to right and accepts side left or side right only with display drawer.",
			"section title \"Text\" sets the generated modal/drawer heading.",
			"group <Name> sections ... wraps contiguous inline sections, including built-in or component sections, in a generated nested layout container.",
			"group compose stack keeps grouped sections vertical; group compose grid supports columns 1..4 and gap sm|md|lg.",
			"Group span applies to the group wrapper inside an outer grid.",
			"tabs accepts gap but does not accept columns, stackAt, modal/drawer section display, or groups.",
			"tab <Name> sections ... is valid only with compose tabs.",
			"Every ordered section must appear in exactly one tab.",
			"trigger <section> on <event> binds generated page events to supported section state changes without arbitrary JavaScript.",
			"Supported trigger events are rowSelect, createStart, editStart, saveSuccess, and close.",
			"Use trigger detail on rowSelect to switch/show the detail section after a record is loaded.",
			"Use trigger form on createStart or editStart to switch/show the form section during create/edit flows.",
			"Use trigger detail on saveSuccess or trigger table on saveSuccess to choose the post-save section.",
			"Use trigger table on close to return tabbed or overlay workflows to the list section.",
			"section span applies a grid-column span to one supported built-in or component section.",
			"Generated web output adds stable bl-view-section-*, bl-view-component-section, bl-view-component-*, bl-view-group-*, bl-view-compose-*, bl-view-span-*, bl-view-display-*, bl-view-side-*, and bl-view-has-triggers classes.",
			"Arbitrary coordinates and arbitrary frontend event handlers remain later features.",
		},
		Errors: []string{"INVALID_VIEW_DECLARATION", "INVALID_VIEW_ORDER", "DUPLICATE_VIEW", "DUPLICATE_VIEW_ORDER", "DUPLICATE_VIEW_COMPOSE", "DUPLICATE_VIEW_COMPOSE_OPTION", "MISSING_VIEW_ORDER", "INVALID_VIEW_COMPOSE", "INVALID_VIEW_COMPOSE_COLUMNS", "INVALID_VIEW_COMPOSE_OPTION", "INVALID_VIEW_SECTION", "INVALID_VIEW_SECTION_OPTION", "INVALID_VIEW_SECTION_SPAN", "INVALID_VIEW_GROUP", "INVALID_VIEW_GROUP_COLUMNS", "INVALID_VIEW_GROUP_OPTION", "INVALID_VIEW_GROUP_SPAN", "INVALID_VIEW_TAB", "INVALID_VIEW_TRIGGER", "UNSUPPORTED_VIEW_SECTION", "DUPLICATE_VIEW_SECTION", "DUPLICATE_VIEW_SECTION_OPTION", "UNKNOWN_VIEW_COMPONENT", "MISSING_VIEW_COMPONENT_BIND", "UNSUPPORTED_VIEW_COMPONENT_BIND", "UNKNOWN_VIEW_COMPONENT_INPUT_FIELD", "VIEW_COMPONENT_INPUT_TYPE_MISMATCH", "UNSUPPORTED_VIEW_COMPONENT_INPUT", "CONFLICTING_VIEW_COMPONENT_SECTION", "INVALID_VIEW_COMPONENT_SECTION", "UNSUPPORTED_VIEW_COMPONENT_DISPLAY", "UNSUPPORTED_VIEW_COMPONENT_SIDE", "UNSUPPORTED_VIEW_SECTION_BIND", "DUPLICATE_VIEW_GROUP", "DUPLICATE_VIEW_GROUP_OPTION", "DUPLICATE_VIEW_GROUP_SECTION", "MISSING_VIEW_GROUP_SECTION", "UNSUPPORTED_VIEW_COMPOSE_MODE", "UNSUPPORTED_VIEW_COMPOSE_COLUMNS", "UNSUPPORTED_VIEW_GAP", "UNSUPPORTED_VIEW_STACK_AT", "UNSUPPORTED_VIEW_SECTION_SPAN", "UNSUPPORTED_VIEW_SECTION_DISPLAY", "UNSUPPORTED_VIEW_SECTION_SIDE", "UNSUPPORTED_VIEW_GROUP", "UNSUPPORTED_VIEW_GROUP_COMPOSE_MODE", "UNSUPPORTED_VIEW_GROUP_COMPOSE_COLUMNS", "UNSUPPORTED_VIEW_GROUP_GAP", "UNSUPPORTED_VIEW_GROUP_SECTION", "UNSUPPORTED_VIEW_GROUP_SECTION_DISPLAY", "UNSUPPORTED_VIEW_GROUP_SPAN", "UNSUPPORTED_VIEW_TAB", "MISSING_VIEW_TABS", "DUPLICATE_VIEW_TAB", "UNSUPPORTED_VIEW_TAB_SECTION", "DUPLICATE_VIEW_TAB_SECTION", "MISSING_VIEW_TAB_SECTION", "DUPLICATE_VIEW_TRIGGER", "UNSUPPORTED_VIEW_TRIGGER_SECTION", "UNSUPPORTED_VIEW_TRIGGER_EVENT", "UNSUPPORTED_VIEW_TRIGGER", "UNSUPPORTED_VIEW_TRIGGER_ACTION", "UNCLOSED_VIEW", "UNEXPECTED_VIEW_TOKEN"},
	},
	"layout": {
		Keyword: "layout",
		Purpose: "Declares a generated application shell and sidebar navigation order.",
		Syntax:  "layout <PascalCaseName> { sidebar { item <PageName> } }",
		Example: `layout AdminLayout {
  sidebar {
    item Products
    item Customers
    item Orders
  }
}`,
		AgentNotes: []string{
			"Layout is a top-level declaration.",
			"Sidebar item values must match page names.",
			"Pages can reference a layout with `layout AdminLayout` inside the page block.",
		},
		Errors: []string{"DUPLICATE_LAYOUT", "UNKNOWN_PAGE_LAYOUT", "UNKNOWN_SIDEBAR_ITEM", "DUPLICATE_SIDEBAR_ITEM"},
	},
	"auth": {
		Keyword: "auth",
		Purpose: "Declares authentication intent for generated applications.",
		Syntax:  "auth { strategy emailPassword; session cookie; user { field type modifiers... } }",
		Example: `auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}`,
		AgentNotes: []string{
			"Auth is a top-level declaration.",
			"Draft v0.1 parses auth intent and generates a basic login/register UI shell.",
			"Draft v0.1 generates register, login, logout, and current-user API endpoints.",
			"Draft v0.1 stores cookie sessions and hashes passwords in generated auth routes.",
			"Generated CRUD API routes require a valid cookie session when auth exists.",
			"Use emailPassword and cookie in v0.1.",
		},
		Errors: []string{"INVALID_AUTH_DECLARATION", "DUPLICATE_AUTH", "MISSING_AUTH_STRATEGY", "UNSUPPORTED_AUTH_STRATEGY", "MISSING_AUTH_SESSION", "UNSUPPORTED_AUTH_SESSION", "MISSING_AUTH_USER"},
	},
	"role": {
		Keyword: "role",
		Purpose: "Declares named permission groups for generated applications.",
		Syntax:  "role <Name> { allow|deny <action> <Resource> }",
		Example: `role Admin {
  allow all
}

role Worker {
  allow read Product
  deny read Product price
}`,
		AgentNotes: []string{
			"Role is a top-level declaration.",
			"Draft v0.1 parses, validates, stores, and checks role-based page and action access. Generated auth supports one or more assigned roles per user while keeping a primary role for compatibility.",
			"Newly registered users receive the first declared role by default.",
			"When roles exist, draft v0.1 generates a basic Users role management page for the first declared role. If tenant policies exist, the same page can update user tenantId values.",
			"Permission checks evaluate all assigned user roles. Matching deny rules override allow rules.",
			"Field names after a permission resource scope that permission to specific fields.",
			"Generated update payloads keep only fields allowed by field-level update permissions.",
			"Use page access to attach roles to pages.",
		},
		Errors: []string{"INVALID_ROLE_DECLARATION", "DUPLICATE_ROLE", "UNSUPPORTED_PERMISSION_ACTION", "MISSING_PERMISSION_RESOURCE", "UNKNOWN_PERMISSION_RESOURCE", "UNKNOWN_PERMISSION_FIELD"},
	},
	"access": {
		Keyword: "access",
		Purpose: "Declares which roles or auth state may access a page.",
		Syntax:  "access <RoleName...|authenticated>",
		Example: `page Products {
  source Product
  access Admin, Worker
}`,
		AgentNotes: []string{
			"Access is currently defined inside page blocks.",
			"Access values must be existing role names or authenticated.",
			"Draft v0.1 generates page-level route guards for access.",
			"Draft v0.1 also uses role permission actions to guard generated CRUD endpoints and UI controls.",
			"Field-level read and update permission rules are enforced by generated routes and UI.",
		},
		Errors: []string{"UNKNOWN_ACCESS_ROLE", "AUTH_REQUIRED_FOR_ACCESS"},
	},
	"api": {
		Keyword: "api",
		Purpose: "Declares explicit REST API contracts, generated declared-runtime routes, typed body fields, and bounded update handlers.",
		Syntax:  "api <Name> { method <GET|POST|PUT|PATCH|DELETE>; path \"/api/path/{id}\"; param id text; query limit integer; body sku text required; update Entity where field == body.sku set stock = body.stock; update Entity where field == body.sku { value incoming = body.quantity; if incoming > 0 and tenantId == body.tenantId ... }; respond accepted|updated; public|private; webhook }",
		Example: `api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body tenantId text required
  body sku text required
  body quantity number required min 0
  body packSize number required min 1
  update Product where sku == body.sku {
    value incoming = body.quantity / body.packSize
    if incoming > 0 and tenantId == body.tenantId
      set stock = stock + incoming
    else
      set stock = stock
  }
  respond accepted
  webhook
  public
}`,
		AgentNotes: []string{
			"API is a top-level declaration.",
			"Explicit api blocks are included in JSON, BlackIR, inspect output, generated/openapi.json, generated server routes, and generated API smoke tests.",
			"Explicit API paths must use a safe `/api/...` path and cannot reuse generated page, auth, query, action, or workflow routes.",
			"Path parameters use `{name}` in path and must have matching `param name type` lines.",
			"Generated runtime routes validate declared path, query, and body values and return a deterministic declared response.",
			"Body fields are supported for POST, PUT, and PATCH and may include file and image values as data URL or absolute http(s) strings.",
			"Body fields may use required, optional, min, max, length, regex, url, and message modifiers; accept belongs to entity media form fields.",
			"Use update Entity where field == body.name set field = body.name for the shortest bounded handler. The where field must be id or a stored unique field; set targets must be stored primitive non-policy fields.",
			"Use update Entity where field == body.name { ... } when a handler needs value, if, else, or nested if statements.",
			"Handler operands use body.name, param.name, source fields in set expressions, local values, typed string/number/boolean literals, and bounded numeric expressions with +, -, *, /, parentheses, and deterministic precedence.",
			"value declares an ordered local value that later value, set, and if statements can reference. Local value names must not collide with body fields, params, source fields, computed fields, or reserved words.",
			"if uses indentation-based then/else branches. Condition comparisons use ==, !=, <, <=, >, and >=; ordered comparisons currently require numeric values. Join comparisons with and, or, not, and parentheses; precedence is not before and before or.",
			"Use a top-level transaction block to wrap an explicit API update handler in a generated Prisma transaction.",
			"Use a top-level service block to group existing explicit APIs into generated service manifest/module metadata without redefining endpoints.",
			"Public handlers on tenant-scoped entities must include the tenant policy field in body or path params. Public handlers cannot update owner-scoped entities.",
			"Use public for unauthenticated routes and private for authenticated routes; private routes are enforced when auth exists.",
			"Use webhook to mark inbound webhook endpoint contracts; generated webhook routes return 202 accepted.",
		},
		Errors: []string{"INVALID_API_DECLARATION", "MISSING_API_METHOD", "UNSUPPORTED_API_METHOD", "MISSING_API_PATH", "INVALID_API_PATH", "MISSING_API_PATH_PARAM", "UNUSED_API_PARAM", "DUPLICATE_API", "DUPLICATE_API_ROUTE", "UNSUPPORTED_API_QUERY_TYPE", "UNSUPPORTED_API_PARAM_TYPE", "INVALID_API_BODY", "UNSUPPORTED_API_BODY_METHOD", "UNSUPPORTED_API_BODY_TYPE", "UNSUPPORTED_API_BODY_MODIFIER", "DUPLICATE_API_BODY", "INVALID_API_UPDATE", "UNCLOSED_API_UPDATE", "UNSUPPORTED_API_HANDLER_METHOD", "UNBOUNDED_API_UPDATE", "MISSING_API_HANDLER_POLICY_SCOPE", "UNSUPPORTED_API_HANDLER_POLICY_SCOPE", "UNSUPPORTED_API_HANDLER_POLICY_FIELD", "UNKNOWN_API_HANDLER_VALUE", "INVALID_API_HANDLER_VALUE_NAME", "DUPLICATE_API_HANDLER_VALUE", "API_HANDLER_VALUE_NAME_COLLISION", "API_HANDLER_VALUE_TYPE_MISMATCH", "INVALID_API_HANDLER_IF", "MISSING_API_HANDLER_IF_BODY", "INVALID_API_HANDLER_ELSE", "MISSING_API_HANDLER_ELSE_BODY", "UNSUPPORTED_API_HANDLER_CONDITION_OPERATOR", "API_HANDLER_CONDITION_TYPE_MISMATCH", "INCOMPATIBLE_API_HANDLER_CONDITION", "INVALID_API_RESPOND", "UNSUPPORTED_API_RESPOND"},
	},
	"table": {
		Keyword: "table",
		Purpose: "Defines list columns, searchable fields, field filters, default sort order, pagination, and generated UI identity for a page.",
		Syntax:  "table { id <Identifier>; class <ClassName...>; columns <field...>; search <field...>; filter <field...>; sort <field> <asc|desc>; paginate <number>; ui <mode> <values...> }",
		Example: `table {
  id ProductsTable
  class inventoryTable
  columns sku, name, stock, inventoryValue
  search sku, name
  filter stock
  sort stock desc
  paginate 25
  ui table border 1 solid compact true
}`,
		AgentNotes: []string{
			"Columns must exist on the page source entity or be computed fields on that entity.",
			"Search fields must be searchable types in v0.1.",
			"Filter fields must exist on the page source entity.",
			"Sort field must exist on the page source entity.",
			"Computed fields can be listed in columns, but search/filter/sort support comes later.",
			"Sort direction must be asc or desc.",
			"Paginate size must be a positive whole number.",
			"Inline UI intent can be declared with table, box, or text modes.",
			"Explicit id and class values are normalized to kebab-case in generated HTML.",
		},
		Errors: []string{"UNKNOWN_TABLE_COLUMN", "UNKNOWN_SEARCH_FIELD", "UNSEARCHABLE_FIELD_TYPE", "UNKNOWN_FILTER_FIELD", "UNKNOWN_SORT_FIELD", "UNSUPPORTED_COMPUTED_SEARCH_FIELD", "UNSUPPORTED_COMPUTED_FILTER_FIELD", "UNSUPPORTED_COMPUTED_SORT_FIELD", "UNSUPPORTED_SORT_DIRECTION", "INVALID_TABLE_PAGINATION", "UNSUPPORTED_PAGE_SIZE", "INVALID_UI_ID", "INVALID_UI_CLASS", "DUPLICATE_UI_ID", "DUPLICATE_UI_CLASS", "INVALID_UI_INTENT", "UNSUPPORTED_UI_MODE", "UNSUPPORTED_UI_TARGET_MODE", "DUPLICATE_UI_INTENT"},
	},
	"filter": {
		Keyword: "filter",
		Purpose: "Declares source fields that should get generated table filter inputs.",
		Syntax:  "filter <field...>",
		Example: "filter customer, status",
		AgentNotes: []string{
			"Filter is currently defined inside table blocks.",
			"Generated React lists apply field filters after search and before sort.",
			"Relation filters use the readable relation label when available.",
			"Computed field filtering is planned for a later data logic phase.",
		},
		Errors: []string{"UNKNOWN_FILTER_FIELD", "UNSUPPORTED_COMPUTED_FILTER_FIELD"},
	},
	"paginate": {
		Keyword: "paginate",
		Purpose: "Declares how many records a generated table should show per page.",
		Syntax:  "paginate <positiveNumber>",
		Example: "paginate 25",
		AgentNotes: []string{
			"Pagination is currently defined inside table blocks.",
			"Generated React lists apply pagination after search and sort.",
			"Use a small positive whole number for compact admin lists.",
		},
		Errors: []string{"INVALID_TABLE_PAGINATION", "UNSUPPORTED_PAGE_SIZE"},
	},
	"form": {
		Keyword: "form",
		Purpose: "Defines generated input fields and generated UI identity for create and edit UI.",
		Syntax:  "form { id <Identifier>; class <ClassName...>; fields <field...>; ui <mode> <values...> }",
		Example: `form {
  id ProductForm
  class inventoryForm
  fields sku, name, stock
  ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
}`,
		AgentNotes: []string{
			"Form fields must exist on the page source entity.",
			"Required/default modifiers affect generated form behavior.",
			"Field label modifiers affect generated form labels.",
			"Field placeholder modifiers affect generated input placeholders.",
			"Field help modifiers generate persistent field notes.",
			"Field min/max modifiers generate numeric frontend and API validation.",
			"Field length min..max modifiers generate text/email frontend and API validation.",
			"Field regex \"pattern\" modifiers generate text/email pattern validation.",
			"Field url modifiers generate URL validation for text fields.",
			"Field message \"Text\" modifiers override generated validation text for that field.",
			"Computed fields are read-only display values and cannot be listed in form fields.",
			"Inline UI intent can be declared with box, text, or button modes.",
			"Explicit id and class values are normalized to kebab-case in generated HTML.",
		},
		Errors: []string{"UNKNOWN_FORM_FIELD", "UNSUPPORTED_COMPUTED_FORM_FIELD", "UNEXPECTED_FORM_TOKEN", "MISSING_CONSTRAINT_VALUE", "INVALID_NUMERIC_CONSTRAINT", "INVALID_LENGTH_CONSTRAINT", "INVALID_REGEX_CONSTRAINT", "UNSUPPORTED_URL_CONSTRAINT", "MISSING_MESSAGE_VALUE", "INVALID_UI_ID", "INVALID_UI_CLASS", "DUPLICATE_UI_ID", "DUPLICATE_UI_CLASS", "INVALID_UI_INTENT", "UNSUPPORTED_UI_MODE", "UNSUPPORTED_UI_TARGET_MODE", "DUPLICATE_UI_INTENT"},
	},
	"actions": {
		Keyword: "actions",
		Purpose: "Declares which CRUD and custom actions a generated page exposes, plus optional generated action button identity.",
		Syntax:  "actions create, edit, delete, archive, restore, <CustomActionName>; action <name> id <Identifier>; action <name> class <ClassName...>; action <name> ui button <values...>",
		Example: `actions create, edit, delete, archive, restore, RestockProduct
action create id CreateProductButton
action create class primaryAction
action create ui button primary white 6 md solid
action RestockProduct id RestockProductButton`,
		AgentNotes: []string{
			"v0.1 supports create, edit, delete, archive, and restore.",
			"PascalCase names in actions must refer to declared top-level custom actions with the same source entity as the page.",
			"Do not invent action names without declaring a custom action or adding validator and generator support.",
			"Use `action <name> ui button ...` to attach inline UI intent to one generated action button.",
			"Use `action <name> id ...` and `action <name> class ...` to attach stable generated button identity.",
			"Generated repeated row button IDs get suffixes so DOM IDs stay unique.",
		},
		Errors: []string{"UNSUPPORTED_ACTION", "DUPLICATE_PAGE_ACTION", "PAGE_ACTION_SOURCE_MISMATCH", "INVALID_ACTION_UI", "INVALID_ACTION_INTENT", "UNKNOWN_ACTION_UI", "INVALID_UI_ID", "INVALID_UI_CLASS", "DUPLICATE_UI_ID", "DUPLICATE_UI_CLASS", "INVALID_UI_INTENT", "UNSUPPORTED_UI_MODE", "UNSUPPORTED_UI_TARGET_MODE", "DUPLICATE_UI_INTENT"},
	},
	"search": {
		Keyword: "search",
		Purpose: "Declares fields used for text search in generated list UI.",
		Syntax:  "search <field...>",
		Example: "search sku, name",
		AgentNotes: []string{
			"Search is currently defined inside table blocks.",
			"v0.1 supports text, email, and entity reference search fields.",
			"Computed field search is planned for a later data logic phase.",
		},
		Errors: []string{"UNKNOWN_SEARCH_FIELD", "UNSEARCHABLE_FIELD_TYPE", "UNSUPPORTED_COMPUTED_SEARCH_FIELD"},
	},
	"blackir": {
		Keyword: "blackir",
		Purpose: "Compact AI-readable intermediate representation for BlackLang projects.",
		Syntax:  "black <command> --ir",
		Example: `blackir 0.1
app Warehouse
entity Product
  sku text required unique`,
		AgentNotes: []string{
			"Use --ir for compact agent-facing output.",
			"Use --json when integrating with standard external tools.",
		},
		Errors: []string{},
	},
	"openapi": {
		Keyword: "openapi",
		Purpose: "Describes the generated REST API contract for web targets.",
		Syntax:  "black build",
		Example: `generated/openapi.json
GET /healthz
GET /readyz
GET /metrics
GET /api/products
POST /api/products
GET /api/products/{id}
POST /api/products/{id}/actions/restockproduct`,
		AgentNotes: []string{
			"OpenAPI output is generated from ops endpoints, page source entities, page CRUD actions, bound queries, explicit api blocks, and bound custom actions.",
			"Ops operations include x-blacklang-ops and x-blacklang-public metadata for health, readiness, and metrics probes.",
			"Explicit API operations include typed body schemas when body fields are declared, plus x-blacklang-handler, x-blacklang-update, and optional transaction metadata when a bounded update handler is declared.",
			"Service-bound explicit API operations include route-level x-blacklang-service metadata, OpenAPI tags, and root x-blacklang-services metadata.",
			"Custom action operations include x-blacklang-action, x-blacklang-source, x-blacklang-mutates-fields, optional x-blacklang-transaction and x-blacklang-transaction-name, request input schema, and response schema.",
			"Read generated/openapi.json when integrating external clients or AI tools.",
			"Run npm test inside the generated app after black build to smoke test OpenAPI paths, schemas, validation, bound queries, bound custom actions, ops endpoints, and basic API server behavior; target web also checks React render behavior.",
			"Do not edit generated/openapi.json manually; change .black source or the generator.",
		},
		Errors: []string{},
	},
	"generated-test": {
		Keyword: "generated-test",
		Purpose: "Describes the deterministic contract/API/frontend/browser-check, browser e2e, and browser e2e matrix tests emitted in generated web or API-only projects.",
		Syntax:  "black build && cd generated && npm test [&& npm run test:e2e:plan && npm run test:e2e:matrix for target web browser tests]",
		Example: `black build examples/warehouse/app.black --out generated
cd generated
npm test
npm run test:e2e:plan
npm run test:e2e:matrix`,
		AgentNotes: []string{
			"Generated target web package.json includes a test script that runs npm run db:generate, src/blacklang.contract.test.ts, src/blacklang.api.test.ts, and src/blacklang.frontend.test.tsx with tsx; when source test declarations exist, it also runs src/blacklang.browser.test.tsx.",
			"Generated target api package.json runs npm run db:generate, src/blacklang.contract.test.ts, and src/blacklang.api.test.ts only; API-only sources do not emit React, frontend smoke, browser-check, browser e2e, or browser matrix tests.",
			"When source test declarations exist in target web, generated package.json also includes test:e2e, test:e2e:plan, test:e2e:matrix, and test:all; test:e2e runs one selected browser and test:e2e:matrix runs every available matrix target after db:generate.",
			"The generated test reads openapi.json and checks ops paths, entity schemas, page CRUD paths, bound query and query summary paths, bound custom action paths, explicit API typed body schemas, update handler metadata, and action metadata.",
			"The generated test imports validation functions and checks representative valid and invalid entity/action payloads.",
			"The generated API smoke test starts createApp() on a random localhost port and checks /openapi.json, declared ops endpoints, explicit API request validation, anonymous API status, and generated CORS behavior when configured.",
			"When seed declarations exist, generated package.json includes db:seed and db:setup runs schema setup followed by deterministic fixture upserts.",
			"The generated frontend smoke test renders App with react-dom/server to catch React import and render failures.",
			"First-class test declarations generate browser-check expectations against page metadata, generated action lists, and generated text/render catalogs.",
			"The generated browser e2e test launches an installed Chrome, Chromium, or Edge executable, creates an ephemeral SQLite test database by default, runs setup and seed modules, registers a generated auth user when auth is enabled, and checks declared text/page/action expectations in the real DOM.",
			"Generated src/blacklang.e2e.matrix.ts and tests/browser-matrix.json provide a read-only matrix plan plus a deterministic matrix run across custom, Chrome, Edge, and Chromium executable candidates.",
			"Set BLACKLANG_E2E_BROWSER_PATH for a custom Chromium-based executable; set BLACKLANG_E2E_CHROME_PATH, BLACKLANG_E2E_EDGE_PATH, or BLACKLANG_E2E_CHROMIUM_PATH for explicit matrix targets; set BLACKLANG_E2E_DATABASE_URL when a non-default e2e database is required.",
			"Do not edit the generated test by hand; change .black source or the generator.",
		},
		Errors: []string{},
	},
	"test": {
		Keyword: "test",
		Purpose: "Declares deterministic generated browser-check expectations for a page.",
		Syntax: `test <Name> {
  page <PageName>
  expect text "Generated text"
  expect page <PageName>
  expect action <actionName>
}`,
		Example: `test WarehouseBrowserSmoke {
  page Products
  expect text "Depo"
  expect page LowStock
  expect action RestockProduct
}`,
		AgentNotes: []string{
			"Test is a top-level declaration for generated browser-check, browser e2e, and browser matrix intent.",
			"Top-level test declarations require target web; target api reports UNSUPPORTED_API_TARGET_TEST because no browser runtime is generated.",
			"Each test targets exactly one generated page with `page PageName`.",
			"expect text checks generated render/catalog data in npm test and visible DOM text in npm run test:e2e or npm run test:e2e:matrix.",
			"expect page checks generated page metadata in npm test and real navigation/page labels in npm run test:e2e or npm run test:e2e:matrix.",
			"expect action checks generated page action metadata in npm test and visible action buttons in npm run test:e2e or npm run test:e2e:matrix.",
			"Generated web output emits src/blacklang.browser.test.tsx and appends it to npm test when tests are declared; it also emits src/blacklang.e2e.test.ts, src/blacklang.e2e.matrix.ts, tests/browser-matrix.json, and test:e2e/test:e2e:plan/test:e2e:matrix/test:all scripts.",
			"Browser e2e and matrix runs are deterministic and bounded: they do not execute arbitrary frontend code from .black; they only run generated auth, setup/seed, navigation, text, page, and action checks.",
		},
		Errors: []string{"INVALID_TEST_DECLARATION", "INVALID_TEST_PAGE", "DUPLICATE_TEST_PAGE", "INVALID_TEST_EXPECT", "UNEXPECTED_TEST_TOKEN", "UNCLOSED_TEST", "DUPLICATE_TEST", "INVALID_TEST_NAME", "TEST_NAME_COLLISION", "MISSING_TEST_PAGE", "UNKNOWN_TEST_PAGE", "MISSING_TEST_EXPECT", "DUPLICATE_TEST_EXPECT", "INVALID_TEST_EXPECT_TEXT", "UNKNOWN_TEST_EXPECT_PAGE", "UNKNOWN_TEST_EXPECT_ACTION", "UNSUPPORTED_TEST_EXPECT", "UNSUPPORTED_API_TARGET_TEST"},
	},
	"benchmark": {
		Keyword: "benchmark",
		Purpose: "Measures BlackLang source size, generated web output size, AI task scenarios, token estimates, long-running AI eval corpus metadata, eval history, web coverage, and tracked issue exports without modifying the project output directory.",
		Syntax:  "black benchmark [file] [--out <dir>] [--json|--ir] | black benchmark tasks [file] [--out <dir>] [--json|--ir] | black benchmark eval [file] [--out <dir>] [--json|--ir] | black benchmark eval-history [--history <file>] [--json|--ir] | black benchmark coverage [--json|--ir] | black benchmark issues [--json|--ir]",
		Example: `black benchmark --json
black benchmark examples/warehouse/app.black --out generated --ir
black benchmark tasks examples/warehouse/app.black --out generated --json
black benchmark eval examples/warehouse/app.black --out generated --json
black benchmark eval-history --json
black benchmark coverage --json
black benchmark issues --json`,
		AgentNotes: []string{
			"Benchmark loads, parses, and validates the .black project before measuring.",
			"Benchmark builds web output in a temporary directory, so it does not mutate the configured generated output directory.",
			"Source metrics include the primary .black file and the configured .blackthm theme file when one exists.",
			"Generated metrics count only files emitted by the BlackLang generator, not node_modules, dist, databases, generated Prisma client output, or package-lock files created by npm.",
			"Use --json for source/generated line counts in AI task reports and benchmark notes.",
			"Use black benchmark tasks --json for deterministic AI task scenarios and token estimate reports; estimates are planning signals, not billed-token measurements.",
			"Use black benchmark eval --json for the long-running AI eval corpus; it returns repeat counts, case prompts, evidence requirements, scoring criteria, and required validation commands without calling an AI model.",
			"Use black benchmark eval-history --json for evidence-backed eval result history from benchmarks/eval-history.blackdir. Local validation entries must not be presented as billed model benchmark scores.",
			"Use black benchmark coverage --json for the weighted web coverage matrix and remaining tracked issues.",
			"Use black benchmark issues --json for a compact tracked issue export with total/open/done counts and current open issue list.",
		},
		Errors: []string{"FILE_READ_ERROR", "BENCHMARK_TEMP_ERROR", "BENCHMARK_FILE_READ_ERROR", "EVAL_HISTORY_READ_ERROR", "INVALID_EVAL_HISTORY_MANIFEST"},
	},
	"coverage": {
		Keyword: "coverage",
		Purpose: "Reports the weighted BlackLang web coverage matrix, tracked issue list, compact issue export, and percentage milestones for the current draft.",
		Syntax:  "black benchmark coverage [--json|--ir] | black benchmark issues [--json|--ir]",
		Example: `black benchmark coverage --json
black benchmark coverage --ir
black benchmark issues --json`,
		AgentNotes: []string{
			"Use coverage before claiming the web target is complete.",
			"The areas array gives deterministic weights, scores, implemented capabilities, missing capabilities, and next actions.",
			"The issues array turns roadmap gaps into stable IDs that can be tracked in docs, GitHub issues, or worklogs.",
			"`black benchmark issues --json` returns the same tracked issue IDs in a smaller export focused on total/open/done counts and current open issue follow-up.",
			"CompletionPercent is a weighted progress signal, not a marketing claim.",
			"Use JSON for dashboards and AI planning; use IR for compact agent context.",
		},
		Errors: []string{},
	},
	"security": {
		Keyword: "security",
		Purpose: "Declares web security intent and describes generated secure defaults, source-security scanning, protected source encryption, and production packaging boundaries.",
		Syntax:  "security { cors { origins env <ENV_NAME> credentials true|false } } | black security scan [file] --json | black security encrypted-source --json | black security encrypt <file> [--out <file.black.enc>] [--key-env <ENV>] --json | black security decrypt <file.black.enc> --stdout --json",
		Example: `security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}

black security scan --json
black security encrypted-source --json
BLACKLANG_SOURCE_KEY=local-key black security encrypt app.black --out app.black.enc --json
BLACKLANG_SOURCE_KEY=local-key black security decrypt app.black.enc --stdout
generated/src/server.ts
app.disable("x-powered-by")
app.use(express.json({ limit: "100kb" }))`,
		AgentNotes: []string{
			"Run black security scan --json before packaging or deployment.",
			"Run black security encrypted-source --json to inspect protected source policy.",
			"Use black security encrypt to create .black.enc source with AES-256-GCM; key material comes only from the configured environment variable.",
			"Use black security decrypt --stdout only in a trusted developer or CI environment; the CLI does not write decrypted source files.",
			"Parse, validate, inspect, lint, benchmark, security scan, and build can read .black.enc source in memory when the header-declared key environment variable is set.",
			"Security scan reports likely hardcoded database URLs, private keys, API keys, tokens, secrets, and passwords.",
			"Generated apps write security/secrets.json, security:secrets:plan, and security:secrets:preflight for environment-backed secret/config readiness and provider CLI preflight without printing values.",
			"Generated Express servers add baseline security headers.",
			"Generated API servers apply a simple IP-based request rate limit.",
			"Generated CORS middleware is created when `security { cors { ... } }` exists.",
			"Generated ops endpoints are public runtime probes outside /api; use CORS and network policy around deployments that expose them.",
			"Do not edit generated server security manually; change the generator or future security syntax.",
		},
		Errors: []string{"INVALID_SECURITY_DECLARATION", "DUPLICATE_SECURITY", "UNCLOSED_SECURITY", "UNEXPECTED_SECURITY_TOKEN", "INVALID_CORS_DECLARATION", "DUPLICATE_CORS", "UNCLOSED_CORS", "UNEXPECTED_CORS_TOKEN", "INVALID_CORS_ORIGINS", "DUPLICATE_CORS_ORIGINS", "MISSING_CORS_ORIGINS", "INVALID_CORS_CREDENTIALS", "DUPLICATE_CORS_CREDENTIALS", "INVALID_ENV_NAME", "HARDCODED_DATABASE_URL", "HARDCODED_PRIVATE_KEY", "HARDCODED_TOKEN", "MISSING_SECURITY_SOURCE", "MISSING_ENCRYPTION_KEY", "INVALID_ENCRYPTED_SOURCE", "DECRYPTION_FAILED", "MISSING_DECRYPT_STDOUT", "ENCRYPTION_WRITE_ERROR", "ENCRYPTION_RANDOM_ERROR", "ENCRYPTION_KEY_ERROR", "ENCRYPTION_CIPHER_ERROR", "UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE"},
	},
	"package": {
		Keyword: "package",
		Purpose: "Creates deployable production artifacts without protected source files.",
		Syntax:  "black package --production --out <dir>",
		Example: `black build
black package --production --out artifacts/production`,
		AgentNotes: []string{
			"Run black build before packaging.",
			"Production package copies generated output and excludes .black and .black.enc source files.",
			"Production package excludes local secrets and local development files such as .env, dev.db, node_modules, and generated Prisma client output.",
			"Use generated artifacts for production servers instead of shipping protected .black source.",
		},
		Errors: []string{"MISSING_PACKAGE_MODE", "PACKAGE_PATH_ERROR", "PACKAGE_OUTPUT_CONFLICT", "PACKAGE_CLEAN_ERROR", "PACKAGE_CREATE_ERROR", "PACKAGE_COPY_ERROR"},
	},
	"accessibility": {
		Keyword: "accessibility",
		Purpose: "Runs read-only generated UI accessibility policy checks from .black source intent.",
		Syntax:  "black audit accessibility [file] [--json|--ir]",
		Example: `black audit accessibility --json
black audit accessibility examples/warehouse/app.black --ir`,
		AgentNotes: []string{
			"Run black audit accessibility --json after changing view composition, overlays, groups, triggers, forms, or page actions.",
			"The command parses and validates source first; source errors are returned in errors before accessibility findings.",
			"Overlay detail/form sections should declare title so generated dialogs have stable accessible names.",
			"View groups that wrap multiple sections should declare title so generated nested landmarks are easy to understand.",
			"Findings are policy diagnostics, not generated file edits; fix the .black source and rebuild.",
		},
		Errors: []string{"UNKNOWN_AUDIT_COMMAND", "FILE_READ_ERROR", "UNCLOSED_STRING", "UNEXPECTED_CHARACTER", "MISSING_APP", "ACCESSIBILITY_MISSING_OVERLAY_TITLE", "ACCESSIBILITY_MISSING_GROUP_TITLE"},
	},
	"audit": {
		Keyword: "audit",
		Purpose: "Describes generated audit log support for authenticated web apps.",
		Syntax:  "black build",
		Example: `generated/src/auth/AuditPage.tsx
GET /api/auth/audit
BlackAuditLog`,
		AgentNotes: []string{
			"Audit support is generated when auth and roles exist.",
			"Generated create, update, archive, restore, delete, bulk delete, custom action, register, and role update operations write audit records.",
			"The first declared role can view recent audit records in the generated Audit page.",
		},
		Errors: []string{},
	},
	"csrf": {
		Keyword: "csrf",
		Purpose: "Describes generated CSRF protection for cookie-authenticated web apps.",
		Syntax:  "black build",
		Example: `black_session cookie
black_csrf cookie
X-CSRF-Token header`,
		AgentNotes: []string{
			"Generated login and register responses set a readable black_csrf cookie next to the HttpOnly session cookie.",
			"Generated frontend write requests send the token through X-CSRF-Token.",
			"Generated API routes reject authenticated POST, PUT, PATCH, and DELETE requests when the CSRF cookie and header do not match.",
		},
		Errors: []string{},
	},
	"workflow": {
		Keyword: "workflow",
		Purpose: "Declares business state flow intent for one source entity.",
		Syntax:  "workflow <Name> { source <Entity>; states <state...>; transition <Name> { from <state>; to <state>; allow <Role...> } }",
		Example: `workflow OrderPreparation {
  source Order
  states draft, picking, verified, packaged, shipped

  transition ship {
    from packaged
    to shipped
    allow Admin
  }
}`,
		AgentNotes: []string{
			"Workflow is a top-level declaration.",
			"Draft v0.1 parses, validates, and exposes workflow intent in JSON and BlackIR.",
			"Workflow source entities must have a status text field.",
			"Generated authenticated apps expose POST /api/<pages>/:id/workflow/<transition> routes.",
			"Generated React pages show workflow buttons only when the row status matches the transition from state.",
			"Generated workflow routes check update permission, transition allow roles, current status, and write audit logs.",
			"Transition from/to values must exist in the workflow states list.",
			"Transition allow values must be existing roles or authenticated.",
		},
		Errors: []string{"INVALID_WORKFLOW_DECLARATION", "DUPLICATE_WORKFLOW", "UNKNOWN_WORKFLOW_SOURCE", "MISSING_WORKFLOW_STATUS_FIELD", "UNSUPPORTED_WORKFLOW_STATUS_FIELD_TYPE", "MISSING_WORKFLOW_STATES", "UNKNOWN_TRANSITION_FROM", "UNKNOWN_TRANSITION_TO", "UNKNOWN_WORKFLOW_ALLOW_ROLE"},
	},
	"state": {
		Keyword: "state",
		Purpose: "Declares client-side UI state intent for pages and components.",
		Syntax:  "state <Name> { <fieldName> <type|Entity[]>; modal <name> open|closed }",
		Example: `state OrdersPageState {
  selectedOrders Order[]
  activeFilter text
  modal createOrder closed
}`,
		AgentNotes: []string{
			"State is a top-level declaration.",
			"Draft v0.1 parses, validates, and exposes state intent in JSON and BlackIR.",
			"State fields may use primitive field types or existing entity names.",
			"Use Entity[] for list state such as selectedOrders Order[].",
			"State modal defaults must be open or closed.",
			"Generated React pages bind PageNameState or PageNamePageState declarations to useState hooks.",
			"Modal declarations generate open/close helpers and can control generated create form visibility.",
			"State declarations do not define computed expressions, arbitrary click handlers, loops, or browser-side BlackLang execution.",
		},
		Errors: []string{"INVALID_STATE_DECLARATION", "DUPLICATE_STATE", "DUPLICATE_STATE_FIELD", "UNSUPPORTED_STATE_FIELD_TYPE", "INVALID_STATE_MODAL", "DUPLICATE_STATE_MODAL", "UNSUPPORTED_STATE_MODAL_DEFAULT"},
	},
	"component": {
		Keyword: "component",
		Purpose: "Declares reusable UI component intent for generated field displays and page component sections.",
		Syntax:  "component <Name> { input <name> <type|Entity[]>; variant <name> when <condition> }",
		Example: `component StockBadge {
  input stock number

  variant low when stock < 10
  variant normal when stock >= 10
}`,
		AgentNotes: []string{
			"Component is a top-level declaration.",
			"Draft v0.1 parses, validates, and exposes component intent in JSON and BlackIR.",
			"Component inputs may use primitive field types or existing entity names.",
			"Use Entity[] for list inputs.",
			"Variant conditions are preserved as deterministic intent strings in v0.1.",
			"Generated apps create standalone React component files from component declarations.",
			"Simple input operator literal variant conditions can select generated CSS classes.",
			"Single-input components bind to matching table, detail, and form preview fields by field name and type.",
			"Page views can place a component as a generated section with `section Name component ComponentName bind selected|first|each` when scalar inputs match source fields.",
		},
		Errors: []string{"INVALID_COMPONENT_DECLARATION", "INVALID_COMPONENT_INPUT", "INVALID_COMPONENT_VARIANT", "DUPLICATE_COMPONENT", "DUPLICATE_COMPONENT_INPUT", "UNSUPPORTED_COMPONENT_INPUT_TYPE", "DUPLICATE_COMPONENT_VARIANT", "MISSING_COMPONENT_VARIANT_CONDITION", "INVALID_VIEW_COMPONENT_SECTION", "CONFLICTING_VIEW_COMPONENT_SECTION", "UNKNOWN_VIEW_COMPONENT", "MISSING_VIEW_COMPONENT_BIND", "UNSUPPORTED_VIEW_COMPONENT_BIND", "UNSUPPORTED_VIEW_COMPONENT_DISPLAY", "UNSUPPORTED_VIEW_COMPONENT_SIDE", "UNKNOWN_VIEW_COMPONENT_INPUT_FIELD", "VIEW_COMPONENT_INPUT_TYPE_MISMATCH", "UNSUPPORTED_VIEW_COMPONENT_INPUT"},
	},
}

func FindDoc(keyword string) (DocEntry, bool) {
	key := strings.ToLower(strings.TrimSpace(keyword))
	doc, ok := docs[key]
	return doc, ok
}

func AllDocs() []DocEntry {
	keys := make([]string, 0, len(docs))
	for key := range docs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]DocEntry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, docs[key])
	}
	return entries
}
