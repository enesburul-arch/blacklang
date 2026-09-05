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
			"The JSON changed field tells agents whether the file already matched the deterministic style.",
		},
		Errors: []string{"FILE_READ_ERROR", "FILE_WRITE_ERROR", "FORMAT_REQUIRED", "UNCLOSED_STRING", "UNEXPECTED_CHARACTER"},
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
		},
		Errors: []string{"FILE_READ_ERROR", "FORMAT_REQUIRED", "UNCLOSED_STRING", "UNEXPECTED_CHARACTER", "MISSING_APP", "UNKNOWN_TABLE_COLUMN", "HARDCODED_DATABASE_URL", "HARDCODED_TOKEN", "HARDCODED_PRIVATE_KEY"},
	},
	"inspect": {
		Keyword: "inspect",
		Purpose: "Prints project structure or a focused affected graph for AI agents before editing.",
		Syntax:  "black inspect [file] [--json|--ir] | black inspect [file] --affected <symbol> --json",
		Example: `black inspect examples/warehouse/app.black --json
black inspect examples/warehouse/app.black --ir
black inspect examples/warehouse/app.black --affected Product.stock --json`,
		AgentNotes: []string{
			"Use black inspect --json or --ir at project start to learn the current app structure.",
			"Use --affected before changing an entity, field, page, role, workflow, state, component, or api symbol.",
			"The affected JSON lists source symbols and generated files that should be validated after the edit.",
			"Unknown symbols return UNKNOWN_AFFECTED_SYMBOL instead of asking the model to guess.",
		},
		Errors: []string{"FILE_READ_ERROR", "MISSING_AFFECTED_SYMBOL", "UNKNOWN_AFFECTED_SYMBOL"},
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
		Errors: []string{"FILE_READ_ERROR", "UNKNOWN_TABLE_COLUMN", "UNKNOWN_AFFECTED_SYMBOL", "HARDCODED_TOKEN"},
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
			"If success is false, use the errors array before making source changes.",
		},
		Errors: []string{"UNKNOWN_AGENT_COMMAND", "FILE_READ_ERROR", "FORMAT_REQUIRED", "HARDCODED_TOKEN"},
	},
	"theme": {
		Keyword: "theme",
		Purpose: "Inspects .blackthm UI theme/profile files before CSS generation exists.",
		Syntax:  "black theme inspect [file] [--json|--ir]",
		Example: `black theme inspect --json
black theme inspect examples/warehouse/theme.blackthm --json
black theme inspect examples/warehouse/theme.blackthm --ir`,
		AgentNotes: []string{
			"Use .blackthm for deterministic UI tokens and mode slot profile metadata.",
			"Set theme = \"path/to/theme.blackthm\" in blacklang.toml so agents can discover it.",
			"Theme inspect returns profile.rules so agents know how compact inline UI values will be read.",
			"Hex colors should be quoted because # starts a comment outside strings.",
			"CSS generation from theme intent is planned for later Phase 20 steps.",
			"Treat .blackthm files as source assets, not generated output.",
		},
		Errors: []string{"FILE_READ_ERROR", "INVALID_THEME_DECLARATION", "MISSING_THEME_VERSION", "MISSING_UI_PROFILE", "INVALID_UI_MODE", "DUPLICATE_UI_MODE", "DUPLICATE_UI_SLOT"},
	},
	"ui-profile": {
		Keyword: "ui-profile",
		Purpose: "Documents compact positional UI mode slot rules used by .blackthm profiles.",
		Syntax:  "mode <name> <slot...> -> ui <mode> <values...> [| <mode> <values...>...]",
		Example: `profile UICompact {
  version 1
  mode box color width style pt pr pb pl radius place
  mode text color size weight align
}

form {
  fields email, password
  ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
}`,
		AgentNotes: []string{
			"Read profile.modes[].slots from black theme inspect --json before writing inline UI intent.",
			"Slots are positional and are read left to right.",
			"Trailing missing values use defaults in later CSS generation phases.",
			"Extra values are errors because they cannot map to a known slot.",
			"Each slot name may appear only once inside a mode.",
			"After a profile is locked, existing slots are immutable and new slots are append-only.",
		},
		Errors: []string{"INVALID_UI_MODE", "DUPLICATE_UI_MODE", "DUPLICATE_UI_SLOT"},
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
	"syntax": {
		Keyword: "syntax",
		Purpose: "Explains the minimal BlackLang v0.1 source structure.",
		Syntax:  "app Name; database { url env ENV_NAME }; entity Name { field type modifiers... }; page Name { source Entity ... }",
		Example: `app Warehouse
database {
  url env DATABASE_URL
}
entity Product {
  sku text required unique
}
page Products {
  source Product
}`,
		AgentNotes: []string{
			"Read blacklang.toml first to find source and output paths.",
			"Prefer changing .black source files instead of generated files.",
			"Run black validate --ir after edits.",
		},
		Errors: []string{"UNEXPECTED_TOP_LEVEL", "INVALID_ENTITY_DECLARATION", "INVALID_PAGE_DECLARATION"},
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
			"Draft v0.1 accepts one database block per project.",
			"Database url must reference an environment variable with `url env NAME`.",
			"Do not write real connection strings, passwords, or tokens directly in .black source.",
			"Generated production deployments should avoid shipping protected .black source files.",
		},
		Errors: []string{"INVALID_DATABASE_DECLARATION", "DUPLICATE_DATABASE", "INVALID_DATABASE_URL", "MISSING_DATABASE_URL", "INVALID_ENV_NAME"},
	},
	"entity": {
		Keyword: "entity",
		Purpose: "Declares stored application data and its fields.",
		Syntax:  "entity <PascalCaseName> { <fieldName> <fieldType> <modifiers...> }",
		Example: `entity Product {
  sku text required unique
  name text required label "Product Name" placeholder "Enter product name" help "Shown under the input"
  stock number default 0
}`,
		AgentNotes: []string{
			"Use entity for data that should be stored or shown in pages.",
			"Fields are referenced by table, form, and search blocks.",
			"Use label \"Text\" to control generated UI field labels.",
			"Use placeholder \"Text\" to control generated input hints.",
			"Use help \"Text\" to show persistent guidance under generated inputs.",
			"Use regex \"pattern\", url, and message \"Text\" for advanced validation intent.",
			"Use validate left <= right message \"Text\" inside an entity for cross-field validation.",
			"Use validate field required when otherField == value message \"Text\" for conditional required validation.",
		},
		Errors: []string{"DUPLICATE_ENTITY", "DUPLICATE_FIELD", "UNSUPPORTED_FIELD_TYPE", "UNSUPPORTED_FIELD_MODIFIER", "MISSING_LABEL_VALUE", "MISSING_PLACEHOLDER_VALUE", "MISSING_HELP_VALUE", "INVALID_REGEX_CONSTRAINT", "UNSUPPORTED_URL_CONSTRAINT", "MISSING_MESSAGE_VALUE", "INVALID_ENTITY_VALIDATION", "UNKNOWN_VALIDATION_FIELD", "UNSUPPORTED_VALIDATION_OPERATOR", "INCOMPATIBLE_VALIDATION_FIELDS", "MISSING_VALIDATION_CONDITION", "INVALID_VALIDATION_LITERAL"},
	},
	"page": {
		Keyword: "page",
		Purpose: "Declares a generated web screen bound to one source entity.",
		Syntax:  "page <PascalCaseName> { layout <LayoutName>; source <EntityName>; table {...}; form {...}; actions ... }",
		Example: `page Products {
  layout AdminLayout
  source Product
  actions create, edit, delete, archive, restore
}`,
		AgentNotes: []string{
			"Every page should have a source entity in v0.1.",
			"Use layout when the page should belong to an explicit generated app shell.",
			"Generated React pages are based on page blocks.",
		},
		Errors: []string{"DUPLICATE_PAGE", "MISSING_PAGE_SOURCE", "UNKNOWN_SOURCE_ENTITY", "UNKNOWN_PAGE_LAYOUT"},
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
			"Draft v0.1 parses, validates, stores, and checks basic single-role page and action access.",
			"Newly registered users receive the first declared role by default.",
			"When roles exist, draft v0.1 generates a basic Users role management page for the first declared role.",
			"Permission deny rules override allow rules.",
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
			"Field-level access rules come later.",
		},
		Errors: []string{"UNKNOWN_ACCESS_ROLE", "AUTH_REQUIRED_FOR_ACCESS"},
	},
	"api": {
		Keyword: "api",
		Purpose: "Declares explicit REST API contracts for generated OpenAPI output.",
		Syntax:  "api <Name> { method <GET|POST|PUT|PATCH|DELETE>; path \"/api/path/{id}\"; param id text; query limit integer; public|private; webhook }",
		Example: `api LowStockReport {
  method GET
  path "/api/reports/low-stock/{warehouseId}"
  param warehouseId text
  query limit integer
  private
}`,
		AgentNotes: []string{
			"API is a top-level declaration.",
			"Draft v0.1 treats explicit api blocks as contract-first declarations.",
			"Explicit api blocks are included in JSON, BlackIR, inspect output, and generated/openapi.json.",
			"Path parameters use `{name}` in path and must have matching `param name type` lines.",
			"Use public for unauthenticated contracts and private for authenticated contracts.",
			"Use webhook to mark inbound webhook endpoint contracts.",
			"Runtime route generation for explicit api blocks comes later.",
		},
		Errors: []string{"INVALID_API_DECLARATION", "MISSING_API_METHOD", "UNSUPPORTED_API_METHOD", "MISSING_API_PATH", "INVALID_API_PATH", "MISSING_API_PATH_PARAM", "UNUSED_API_PARAM", "DUPLICATE_API", "DUPLICATE_API_ROUTE", "UNSUPPORTED_API_QUERY_TYPE", "UNSUPPORTED_API_PARAM_TYPE"},
	},
	"table": {
		Keyword: "table",
		Purpose: "Defines list columns, searchable fields, field filters, default sort order, and pagination for a page.",
		Syntax:  "table { columns <field...>; search <field...>; filter <field...>; sort <field> <asc|desc>; paginate <number> }",
		Example: `table {
  columns sku, name, stock
  search sku, name
  filter stock
  sort stock desc
  paginate 25
}`,
		AgentNotes: []string{
			"Columns must exist on the page source entity.",
			"Search fields must be searchable types in v0.1.",
			"Filter fields must exist on the page source entity.",
			"Sort field must exist on the page source entity.",
			"Sort direction must be asc or desc.",
			"Paginate size must be a positive whole number.",
		},
		Errors: []string{"UNKNOWN_TABLE_COLUMN", "UNKNOWN_SEARCH_FIELD", "UNSEARCHABLE_FIELD_TYPE", "UNKNOWN_FILTER_FIELD", "UNKNOWN_SORT_FIELD", "UNSUPPORTED_SORT_DIRECTION", "INVALID_TABLE_PAGINATION", "UNSUPPORTED_PAGE_SIZE"},
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
		},
		Errors: []string{"UNKNOWN_FILTER_FIELD"},
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
		Purpose: "Defines generated input fields for create and edit UI.",
		Syntax:  "form { fields <field...> }",
		Example: `form {
  fields sku, name, stock
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
		},
		Errors: []string{"UNKNOWN_FORM_FIELD", "UNEXPECTED_FORM_TOKEN", "MISSING_CONSTRAINT_VALUE", "INVALID_NUMERIC_CONSTRAINT", "INVALID_LENGTH_CONSTRAINT", "INVALID_REGEX_CONSTRAINT", "UNSUPPORTED_URL_CONSTRAINT", "MISSING_MESSAGE_VALUE"},
	},
	"actions": {
		Keyword: "actions",
		Purpose: "Declares page behaviors that the generator should create.",
		Syntax:  "actions create, edit, delete, archive, restore",
		Example: "actions create, edit, delete, archive, restore",
		AgentNotes: []string{
			"v0.1 supports create, edit, delete, archive, and restore.",
			"Do not invent action names without adding validator and generator support.",
		},
		Errors: []string{"UNSUPPORTED_ACTION"},
	},
	"search": {
		Keyword: "search",
		Purpose: "Declares fields used for text search in generated list UI.",
		Syntax:  "search <field...>",
		Example: "search sku, name",
		AgentNotes: []string{
			"Search is currently defined inside table blocks.",
			"v0.1 supports text, email, and entity reference search fields.",
		},
		Errors: []string{"UNKNOWN_SEARCH_FIELD", "UNSEARCHABLE_FIELD_TYPE"},
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
GET /api/products
POST /api/products
GET /api/products/{id}`,
		AgentNotes: []string{
			"OpenAPI output is generated from page source entities and page actions.",
			"Read generated/openapi.json when integrating external clients or AI tools.",
			"Do not edit generated/openapi.json manually; change .black source or the generator.",
		},
		Errors: []string{},
	},
	"security": {
		Keyword: "security",
		Purpose: "Describes generated secure defaults, source-security scanning, and encrypted source planning.",
		Syntax:  "black security scan --json | black security encrypted-source --json",
		Example: `black security scan --json
black security encrypted-source --json
generated/src/server.ts
app.disable("x-powered-by")
app.use(express.json({ limit: "100kb" }))`,
		AgentNotes: []string{
			"Run black security scan --json before packaging or deployment.",
			"Run black security encrypted-source --json to inspect protected source policy.",
			"Security scan reports likely hardcoded database URLs, private keys, API keys, tokens, secrets, and passwords.",
			"Generated Express servers add baseline security headers.",
			"Generated API servers apply a simple IP-based request rate limit.",
			"Do not edit generated server security manually; change the generator or future security syntax.",
		},
		Errors: []string{"HARDCODED_DATABASE_URL", "HARDCODED_PRIVATE_KEY", "HARDCODED_TOKEN"},
	},
	"package": {
		Keyword: "package",
		Purpose: "Creates deployable production artifacts without protected source files.",
		Syntax:  "black package --production --out <dir>",
		Example: `black build
black package --production --out artifacts/production`,
		AgentNotes: []string{
			"Run black build before packaging.",
			"Production package copies generated output and excludes .black source files.",
			"Production package excludes local secrets and local development files such as .env, dev.db, node_modules, and generated Prisma client output.",
			"Use generated artifacts for production servers instead of shipping protected .black source.",
		},
		Errors: []string{"MISSING_PACKAGE_MODE", "PACKAGE_PATH_ERROR", "PACKAGE_OUTPUT_CONFLICT", "PACKAGE_CLEAN_ERROR", "PACKAGE_CREATE_ERROR", "PACKAGE_COPY_ERROR"},
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
			"Generated create, update, archive, restore, delete, bulk delete, register, and role update operations write audit records.",
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
		},
		Errors: []string{"INVALID_STATE_DECLARATION", "DUPLICATE_STATE", "DUPLICATE_STATE_FIELD", "UNSUPPORTED_STATE_FIELD_TYPE", "INVALID_STATE_MODAL", "DUPLICATE_STATE_MODAL", "UNSUPPORTED_STATE_MODAL_DEFAULT"},
	},
	"component": {
		Keyword: "component",
		Purpose: "Declares reusable UI component intent.",
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
		},
		Errors: []string{"INVALID_COMPONENT_DECLARATION", "INVALID_COMPONENT_INPUT", "INVALID_COMPONENT_VARIANT", "DUPLICATE_COMPONENT", "DUPLICATE_COMPONENT_INPUT", "UNSUPPORTED_COMPONENT_INPUT_TYPE", "DUPLICATE_COMPONENT_VARIANT", "MISSING_COMPONENT_VARIANT_CONDITION"},
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
