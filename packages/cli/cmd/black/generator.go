package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var supportedComparisonOperators = setOf(
	"==",
	"!=",
	"<",
	"<=",
	">",
	">=",
)

func BuildWeb(program Program, outDir string) ([]GeneratedFile, []Diagnostic) {
	return BuildWebWithTheme(program, outDir, nil)
}

func BuildWebWithTheme(program Program, outDir string, theme *ThemeDecl) ([]GeneratedFile, []Diagnostic) {
	generator := webGenerator{
		program: program,
		theme:   theme,
		outDir:  outDir,
		files:   []GeneratedFile{},
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, []Diagnostic{{
			Code:       "BUILD_OUTPUT_ERROR",
			Message:    err.Error(),
			Suggestion: "Choose a writable output directory with `--out <dir>`.",
		}}
	}

	generator.write("README.md", "documentation", generator.readme())
	generator.write(".env.example", "environment-example", generator.envExample())
	generator.write("package.json", "package", generator.packageJSON())
	generator.write(filepath.Join("security", "secrets.json"), "secret-manifest", generator.secretManifestJSON())
	generator.write(filepath.Join("scripts", "secrets-plan.mjs"), "secret-script", generator.secretsPlanMJS())
	generator.write(filepath.Join("scripts", "secrets-provider.mjs"), "secret-script", generator.secretsProviderMJS())
	if generator.hasOpsObserve() {
		generator.write(filepath.Join("ops", "observability.json"), "observability-manifest", generator.opsObservabilityManifestJSON())
	}
	if generator.deployTarget() == "docker" {
		generator.write(".dockerignore", "docker-ignore", generator.dockerignore())
		generator.write("Dockerfile", "dockerfile", generator.dockerfile())
		generator.write("docker-compose.yml", "docker-compose", generator.dockerComposeYAML())
		if generator.hasDeployPreviewLocal() {
			generator.write("docker-compose.preview.yml", "docker-compose-preview", generator.dockerComposePreviewYAML())
		}
		if generator.hasDeployManifest() {
			generator.write(filepath.Join("deploy", "manifest.json"), "deploy-manifest", generator.deployManifestJSON())
		}
		if generator.hasDeployRollback() {
			generator.write(filepath.Join("deploy", "rollback.json"), "deploy-rollback", generator.deployRollbackJSON())
			generator.write(filepath.Join("scripts", "rollback-plan.mjs"), "deploy-script", generator.rollbackPlanMJS())
		}
		if generator.hasDeployCloud() {
			generator.write(filepath.Join("deploy", "cloud.json"), "deploy-cloud", generator.deployCloudJSON())
			generator.write(filepath.Join("scripts", "cloud-plan.mjs"), "deploy-script", generator.cloudPlanMJS())
			generator.write(filepath.Join("scripts", "cloud-exec.mjs"), "deploy-script", generator.cloudExecMJS())
		}
	}
	generator.write("tsconfig.json", "typescript-config", generator.tsconfig())
	if !generator.isAPIOnlyTarget() {
		generator.write("index.html", "html-entry", generator.indexHTML())
		generator.write("vite.config.ts", "vite-config", generator.viteConfig())
	}
	generator.write("prisma.config.ts", "prisma-config", generator.prismaConfig())
	generator.write("openapi.json", "api-spec", generator.openapiJSON())
	generator.write(filepath.Join("prisma", "schema.prisma"), "database-schema", generator.prismaSchema())
	if len(program.Migrations) > 0 {
		generator.write(filepath.Join("migrations", "manifest.json"), "migration-manifest", generator.migrationManifestJSON())
		for index, migration := range program.Migrations {
			generator.write(filepath.Join("migrations", generator.migrationFileName(index, migration)), "migration-sql", generator.migrationSQL(index, migration))
		}
		generator.write(filepath.Join("src", "migrate.ts"), "migration-runner", generator.migrationRunnerTS())
	}
	if !generator.isAPIOnlyTarget() {
		generator.write(filepath.Join("src", "main.tsx"), "react-entry", generator.mainTSX())
		generator.write(filepath.Join("src", "App.tsx"), "react-app", generator.appTSX())
	}
	generator.write(filepath.Join("src", "db.ts"), "database-client", generator.dbTS())
	generator.write(filepath.Join("src", "setup-db.ts"), "database-setup", generator.setupDBTS())
	if generator.hasSeeds() {
		generator.write(filepath.Join("src", "seed.ts"), "database-seed", generator.seedTS())
	}
	if generator.hasJobs() {
		generator.write(filepath.Join("jobs", "manifest.json"), "job-manifest", generator.jobsManifestJSON())
		generator.write(filepath.Join("src", "worker.ts"), "worker", generator.workerTS())
	}
	if generator.hasServices() {
		generator.write(filepath.Join("services", "manifest.json"), "service-manifest", generator.serviceManifestJSON())
		for _, service := range program.Services {
			generator.write(filepath.Join("src", "services", kebabCase(service.Name)+".ts"), "service-module", generator.serviceModuleTS(service))
		}
	}
	generator.write(filepath.Join("src", "server.ts"), "api-server", generator.serverTS())
	if !generator.isAPIOnlyTarget() {
		generator.write(filepath.Join("src", "styles.css"), "stylesheet", generator.stylesCSS())
		generator.write(filepath.Join("src", "vite-env.d.ts"), "type-declarations", generator.viteEnv())
	}
	generator.write(filepath.Join("src", "types.ts"), "types", generator.types())
	if program.Auth != nil {
		generator.write(filepath.Join("src", "routes", "auth.ts"), "auth-route", generator.authRouteTS())
		if !generator.isAPIOnlyTarget() {
			generator.write(filepath.Join("src", "auth", "AuthPage.tsx"), "auth-page", generator.authPageTSX())
		}
		if !generator.isAPIOnlyTarget() && len(program.Roles) > 0 {
			generator.write(filepath.Join("src", "auth", "UsersPage.tsx"), "auth-users-page", generator.authUsersPageTSX())
			generator.write(filepath.Join("src", "auth", "AuditPage.tsx"), "auth-audit-page", generator.authAuditPageTSX())
		}
	}

	if !generator.isAPIOnlyTarget() {
		for _, component := range program.Components {
			generator.write(filepath.Join("src", "components", component.Name+".tsx"), "react-component", generator.componentTSX(component))
		}
	}

	writtenValidation := map[string]bool{}
	for _, page := range program.Pages {
		entity, ok := generator.findEntity(page.Source)
		if !ok {
			continue
		}
		name := generator.pageModuleName(page)
		generator.write(filepath.Join("src", "api", name+".ts"), "api-client", generator.apiClient(page, entity))
		generator.write(filepath.Join("src", "routes", name+".ts"), "api-route", generator.route(page, entity))
		if !writtenValidation[entity.Name] {
			generator.write(filepath.Join("src", "validation", strings.ToLower(entity.Name)+".ts"), "validation", generator.validation(entity))
			writtenValidation[entity.Name] = true
		}
		if !generator.isAPIOnlyTarget() {
			generator.write(filepath.Join("src", "pages", page.Name+"Page.tsx"), "react-page", generator.page(page, entity))
		}
	}
	generator.write(filepath.Join("src", "blacklang.contract.test.ts"), "test", generator.contractTestTS())
	generator.write(filepath.Join("src", "blacklang.api.test.ts"), "test", generator.apiSmokeTestTS())
	if !generator.isAPIOnlyTarget() {
		generator.write(filepath.Join("src", "blacklang.frontend.test.tsx"), "test", generator.frontendSmokeTestTSX())
	}
	if !generator.isAPIOnlyTarget() && generator.hasBrowserChecks() {
		generator.write(filepath.Join("src", "blacklang.browser.test.tsx"), "test", generator.browserCheckTestTSX())
		generator.write(filepath.Join("src", "blacklang.e2e.test.ts"), "test", generator.browserE2ETestTS())
		generator.write(filepath.Join("src", "blacklang.e2e.matrix.ts"), "test", generator.browserE2EMatrixTS())
		generator.write(filepath.Join("tests", "browser-matrix.json"), "test-manifest", generator.browserMatrixManifestJSON())
	}

	return generator.files, generator.diagnostics
}

type webGenerator struct {
	program     Program
	theme       *ThemeDecl
	outDir      string
	files       []GeneratedFile
	diagnostics []Diagnostic
}

type incomingRelation struct {
	entity EntityDecl
	field  FieldDecl
}

type generatedMigrationManifest struct {
	Version        string                   `json:"version"`
	TargetDatabase string                   `json:"targetDatabase"`
	Runner         generatedMigrationRunner `json:"runner"`
	Migrations     []generatedMigration     `json:"migrations"`
}

type generatedMigrationRunner struct {
	PlanCommand  string   `json:"planCommand"`
	ApplyCommand string   `json:"applyCommand"`
	Modes        []string `json:"modes"`
	Policy       string   `json:"policy"`
}

type generatedMigration struct {
	Name       string                        `json:"name"`
	File       string                        `json:"file"`
	Operations []generatedMigrationOperation `json:"operations"`
}

type generatedMigrationOperation struct {
	Kind      string `json:"kind"`
	Entity    string `json:"entity,omitempty"`
	From      string `json:"from"`
	To        string `json:"to"`
	OldColumn string `json:"oldColumn,omitempty"`
	NewColumn string `json:"newColumn,omitempty"`
}

func (g *webGenerator) write(relativePath string, kind string, content string) {
	fullPath := filepath.Join(g.outDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		g.addError(fullPath, err)
		return
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		g.addError(fullPath, err)
		return
	}
	g.files = append(g.files, GeneratedFile{
		Path: fullPath,
		Kind: kind,
	})
}

func (g *webGenerator) addError(path string, err error) {
	g.diagnostics = append(g.diagnostics, Diagnostic{
		File:       path,
		Code:       "BUILD_WRITE_ERROR",
		Message:    err.Error(),
		Suggestion: "Check output directory permissions.",
	})
}

func (g *webGenerator) findEntity(name string) (EntityDecl, bool) {
	for _, entity := range g.program.Entities {
		if entity.Name == name {
			return entity, true
		}
	}
	return EntityDecl{}, false
}

func (g *webGenerator) isRelationField(field FieldDecl) bool {
	for _, entity := range g.program.Entities {
		if entity.Name == field.Type {
			return true
		}
	}
	return false
}

func (g *webGenerator) sqliteColumnName(field FieldDecl) string {
	if g.isRelationField(field) {
		return relationIDFieldName(field)
	}
	return field.Name
}

func (g *webGenerator) sqliteFieldType(field FieldDecl) string {
	if g.isRelationField(field) {
		return "TEXT"
	}
	return sqliteType(field.Type)
}

func (g *webGenerator) incomingRelations(entityName string) []incomingRelation {
	relations := []incomingRelation{}
	for _, sourceEntity := range g.program.Entities {
		for _, field := range sourceEntity.Fields {
			if field.Type == entityName {
				relations = append(relations, incomingRelation{
					entity: sourceEntity,
					field:  field,
				})
			}
		}
	}
	return relations
}

func (g *webGenerator) relationFields(entity EntityDecl) []FieldDecl {
	fields := []FieldDecl{}
	for _, field := range entity.Fields {
		if g.isRelationField(field) {
			fields = append(fields, field)
		}
	}
	return fields
}

func (g *webGenerator) hasRelationFields(entity EntityDecl) bool {
	return len(g.relationFields(entity)) > 0
}

func (g *webGenerator) relationFieldsForLoad(entity EntityDecl, context string) []FieldDecl {
	fields := []FieldDecl{}
	for _, field := range g.relationFields(entity) {
		if relationLoadIncludesContext(field, context) {
			fields = append(fields, field)
		}
	}
	return fields
}

func (g *webGenerator) hasRelationFieldsForLoad(entity EntityDecl, context string) bool {
	return len(g.relationFieldsForLoad(entity, context)) > 0
}

func (g *webGenerator) prismaIncludeLine(entity EntityDecl, indent string, trailingComma ...bool) string {
	return g.prismaIncludeLineForFields(g.relationFields(entity), indent, trailingComma...)
}

func (g *webGenerator) prismaIncludeLineForLoad(entity EntityDecl, context string, indent string, trailingComma ...bool) string {
	return g.prismaIncludeLineForFields(g.relationFieldsForLoad(entity, context), indent, trailingComma...)
}

func (g *webGenerator) prismaIncludeLineForFields(relations []FieldDecl, indent string, trailingComma ...bool) string {
	withComma := true
	if len(trailingComma) > 0 {
		withComma = trailingComma[0]
	}
	parts := []string{}
	for _, field := range relations {
		parts = append(parts, fmt.Sprintf("%s: true", field.Name))
	}
	line := fmt.Sprintf("%sinclude: { %s }", indent, strings.Join(parts, ", "))
	if withComma {
		line += ","
	}
	return line + "\n"
}

func (g *webGenerator) generatedMigrations() []generatedMigration {
	migrations := []generatedMigration{}
	for index, migration := range g.program.Migrations {
		migrations = append(migrations, g.generatedMigration(index, migration))
	}
	return migrations
}

func (g *webGenerator) generatedMigration(index int, migration MigrationDecl) generatedMigration {
	return generatedMigration{
		Name:       migration.Name,
		File:       g.migrationFileName(index, migration),
		Operations: g.generatedMigrationOperations(migration),
	}
}

func (g *webGenerator) generatedMigrationOperations(migration MigrationDecl) []generatedMigrationOperation {
	operations := []generatedMigrationOperation{}
	for _, rename := range migration.Renames {
		if rename.Kind != "entity" {
			continue
		}
		operations = append(operations, generatedMigrationOperation{
			Kind: "entity",
			From: rename.From,
			To:   rename.To,
		})
	}
	for _, rename := range migration.Renames {
		if rename.Kind != "field" {
			continue
		}
		oldColumn, newColumn := g.migrationFieldColumns(rename)
		operations = append(operations, generatedMigrationOperation{
			Kind:      "field",
			Entity:    rename.Entity,
			From:      rename.From,
			To:        rename.To,
			OldColumn: oldColumn,
			NewColumn: newColumn,
		})
	}
	return operations
}

func (g *webGenerator) migrationFieldColumns(rename MigrationRenameDecl) (string, string) {
	oldColumn := rename.From
	newColumn := rename.To
	entity, ok := g.findEntity(rename.Entity)
	if !ok {
		return oldColumn, newColumn
	}
	field, ok := findField(entity, rename.To)
	if !ok {
		return oldColumn, newColumn
	}
	if g.isRelationField(field) {
		oldColumn = rename.From + "Id"
		newColumn = relationIDFieldName(field)
	}
	return oldColumn, newColumn
}

func (g *webGenerator) migrationManifestJSON() string {
	data, err := json.MarshalIndent(generatedMigrationManifest{
		Version:        version,
		TargetDatabase: g.targetDatabase(),
		Runner: generatedMigrationRunner{
			PlanCommand:  "npm run db:migrate:plan",
			ApplyCommand: "npm run db:migrate",
			Modes:        []string{"plan", "apply"},
			Policy:       "plan mode is read-only; apply mode runs only declared rename migrations before generated schema setup",
		},
		Migrations: g.generatedMigrations(),
	}, "", "  ")
	if err != nil {
		return "{}\n"
	}
	return string(data) + "\n"
}

func (g *webGenerator) migrationFileName(index int, migration MigrationDecl) string {
	return fmt.Sprintf("%03d_%s.sql", index+1, migrationSlug(migration.Name))
}

func migrationSlug(value string) string {
	var builder strings.Builder
	lastDash := false
	for _, char := range strings.ToLower(value) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			lastDash = false
			continue
		}
		if !lastDash && builder.Len() > 0 {
			builder.WriteString("-")
			lastDash = true
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "migration"
	}
	return slug
}

func (g *webGenerator) migrationSQL(index int, migration MigrationDecl) string {
	var builder strings.Builder
	builder.WriteString("-- Generated by BlackLang. Do not edit manually.\n")
	builder.WriteString(fmt.Sprintf("-- Migration: %s\n", migration.Name))
	builder.WriteString(fmt.Sprintf("-- Target database: %s\n", g.targetDatabase()))
	builder.WriteString("-- Applied by src/setup-db.ts before generated schema setup.\n\n")
	for _, operation := range g.generatedMigration(index, migration).Operations {
		switch operation.Kind {
		case "entity":
			builder.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;\n", sqlQuoteIdentifier(operation.From), sqlQuoteIdentifier(operation.To)))
		case "field":
			builder.WriteString(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s;\n", sqlQuoteIdentifier(operation.Entity), sqlQuoteIdentifier(operation.OldColumn), sqlQuoteIdentifier(operation.NewColumn)))
		}
	}
	return builder.String()
}

func (g *webGenerator) migrationRuntimeDataTS() string {
	return `type BlackMigrationOperation = {
  kind: "entity" | "field";
  entity?: string;
  from: string;
  to: string;
  oldColumn?: string;
  newColumn?: string;
};

type BlackMigration = {
  name: string;
  file: string;
  operations: BlackMigrationOperation[];
};

const blackMigrations: BlackMigration[] = ` + g.generatedMigrationsJSON() + `;

`
}

func (g *webGenerator) generatedMigrationsJSON() string {
	data, err := json.MarshalIndent(g.generatedMigrations(), "", "  ")
	if err != nil {
		return "[]"
	}
	return string(data)
}

func (g *webGenerator) sqliteMigrationRuntimeTS() string {
	if len(g.program.Migrations) == 0 {
		return ""
	}
	return g.migrationRuntimeDataTS() + `function quoteIdentifier(value: string): string {
  return "\"" + value.replace(/"/g, "\"\"") + "\"";
}

function ensureBlackMigrationTable(): void {
  db.exec("CREATE TABLE IF NOT EXISTS \"BlackMigration\" (\"id\" TEXT NOT NULL PRIMARY KEY, \"appliedAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);");
}

function sqliteMigrationApplied(name: string): boolean {
  const row = db.prepare("SELECT id FROM \"BlackMigration\" WHERE id = ?").get(name) as { id: string } | undefined;
  return Boolean(row);
}

function markSQLiteMigrationApplied(name: string): void {
  db.prepare("INSERT OR IGNORE INTO \"BlackMigration\" (id) VALUES (?)").run(name);
}

function sqliteTableExists(name: string): boolean {
  const row = db.prepare("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?").get(name) as { name: string } | undefined;
  return Boolean(row);
}

function sqliteColumnExists(table: string, column: string): boolean {
  const rows = db.prepare("PRAGMA table_info(" + quoteIdentifier(table) + ")").all() as Array<{ name: string }>;
  return rows.some((row) => row.name === column);
}

function applySQLiteEntityRename(migration: BlackMigration, operation: BlackMigrationOperation): void {
  const sourceExists = sqliteTableExists(operation.from);
  const targetExists = sqliteTableExists(operation.to);
  if (sourceExists && targetExists) {
    throw new Error("Migration " + migration.name + " cannot rename table " + operation.from + " to " + operation.to + " because both tables exist.");
  }
  if (sourceExists) {
    db.exec("ALTER TABLE " + quoteIdentifier(operation.from) + " RENAME TO " + quoteIdentifier(operation.to));
  }
}

function applySQLiteFieldRename(migration: BlackMigration, operation: BlackMigrationOperation): void {
  if (!operation.entity || !operation.oldColumn || !operation.newColumn) {
    throw new Error("Migration " + migration.name + " has an incomplete field rename operation.");
  }
  if (!sqliteTableExists(operation.entity)) {
    return;
  }
  const sourceExists = sqliteColumnExists(operation.entity, operation.oldColumn);
  const targetExists = sqliteColumnExists(operation.entity, operation.newColumn);
  if (sourceExists && targetExists) {
    throw new Error("Migration " + migration.name + " cannot rename column " + operation.entity + "." + operation.oldColumn + " to " + operation.newColumn + " because both columns exist.");
  }
  if (sourceExists) {
    db.exec("ALTER TABLE " + quoteIdentifier(operation.entity) + " RENAME COLUMN " + quoteIdentifier(operation.oldColumn) + " TO " + quoteIdentifier(operation.newColumn));
    return;
  }
  if (!targetExists) {
    throw new Error("Migration " + migration.name + " cannot find column " + operation.entity + "." + operation.oldColumn + " or " + operation.entity + "." + operation.newColumn + ".");
  }
}

function applyBlackMigrations(): void {
  ensureBlackMigrationTable();
  const transaction = db.transaction((migration: BlackMigration) => {
    if (sqliteMigrationApplied(migration.name)) {
      return;
    }
    for (const operation of migration.operations) {
      if (operation.kind === "entity") {
        applySQLiteEntityRename(migration, operation);
      } else {
        applySQLiteFieldRename(migration, operation);
      }
    }
    markSQLiteMigrationApplied(migration.name);
  });

  for (const migration of blackMigrations) {
    transaction(migration);
  }
}

applyBlackMigrations();

`
}

func (g *webGenerator) postgresMigrationRuntimeTS() string {
	if len(g.program.Migrations) == 0 {
		return ""
	}
	return g.migrationRuntimeDataTS() + `const { Client } = pg;
type PgClient = InstanceType<typeof Client>;

function quoteIdentifier(value: string): string {
  return "\"" + value.replace(/"/g, "\"\"") + "\"";
}

async function ensureBlackMigrationTable(client: PgClient): Promise<void> {
  await client.query("CREATE TABLE IF NOT EXISTS \"BlackMigration\" (\"id\" TEXT NOT NULL PRIMARY KEY, \"appliedAt\" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)");
}

async function postgresMigrationApplied(client: PgClient, name: string): Promise<boolean> {
  const result = await client.query<{ id: string }>("SELECT \"id\" FROM \"BlackMigration\" WHERE \"id\" = $1", [name]);
  return result.rows.length > 0;
}

async function markPostgresMigrationApplied(client: PgClient, name: string): Promise<void> {
  await client.query("INSERT INTO \"BlackMigration\" (\"id\") VALUES ($1) ON CONFLICT (\"id\") DO NOTHING", [name]);
}

async function postgresTableExists(client: PgClient, name: string): Promise<boolean> {
  const result = await client.query<{ exists: boolean }>(
    "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1) AS exists",
    [name]
  );
  return Boolean(result.rows[0]?.exists);
}

async function postgresColumnExists(client: PgClient, table: string, column: string): Promise<boolean> {
  const result = await client.query<{ exists: boolean }>(
    "SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2) AS exists",
    [table, column]
  );
  return Boolean(result.rows[0]?.exists);
}

async function applyPostgresEntityRename(client: PgClient, migration: BlackMigration, operation: BlackMigrationOperation): Promise<void> {
  const sourceExists = await postgresTableExists(client, operation.from);
  const targetExists = await postgresTableExists(client, operation.to);
  if (sourceExists && targetExists) {
    throw new Error("Migration " + migration.name + " cannot rename table " + operation.from + " to " + operation.to + " because both tables exist.");
  }
  if (sourceExists) {
    await client.query("ALTER TABLE " + quoteIdentifier(operation.from) + " RENAME TO " + quoteIdentifier(operation.to));
  }
}

async function applyPostgresFieldRename(client: PgClient, migration: BlackMigration, operation: BlackMigrationOperation): Promise<void> {
  if (!operation.entity || !operation.oldColumn || !operation.newColumn) {
    throw new Error("Migration " + migration.name + " has an incomplete field rename operation.");
  }
  if (!(await postgresTableExists(client, operation.entity))) {
    return;
  }
  const sourceExists = await postgresColumnExists(client, operation.entity, operation.oldColumn);
  const targetExists = await postgresColumnExists(client, operation.entity, operation.newColumn);
  if (sourceExists && targetExists) {
    throw new Error("Migration " + migration.name + " cannot rename column " + operation.entity + "." + operation.oldColumn + " to " + operation.newColumn + " because both columns exist.");
  }
  if (sourceExists) {
    await client.query("ALTER TABLE " + quoteIdentifier(operation.entity) + " RENAME COLUMN " + quoteIdentifier(operation.oldColumn) + " TO " + quoteIdentifier(operation.newColumn));
    return;
  }
  if (!targetExists) {
    throw new Error("Migration " + migration.name + " cannot find column " + operation.entity + "." + operation.oldColumn + " or " + operation.entity + "." + operation.newColumn + ".");
  }
}

async function applyBlackMigrations(): Promise<void> {
  const client = new Client({ connectionString: databaseUrl });
  await client.connect();
  try {
    await ensureBlackMigrationTable(client);
    for (const migration of blackMigrations) {
      if (await postgresMigrationApplied(client, migration.name)) {
        continue;
      }
      await client.query("BEGIN");
      try {
        for (const operation of migration.operations) {
          if (operation.kind === "entity") {
            await applyPostgresEntityRename(client, migration, operation);
          } else {
            await applyPostgresFieldRename(client, migration, operation);
          }
        }
        await markPostgresMigrationApplied(client, migration.name);
        await client.query("COMMIT");
      } catch (error) {
        await client.query("ROLLBACK");
        throw error;
      }
    }
  } finally {
    await client.end();
  }
}

`
}

func sqlQuoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func (g *webGenerator) readme() string {
	return fmt.Sprintf(`# %s

This folder was generated by BlackLang.

Do not edit generated files manually. Change the BlackLang source file and run black build again.

## Database

Copy .env.example to .env for local development.

~~~bash
npm run db:generate
npm run db:validate
npm run db:setup
npm run db:push
npm run api:dev
~~~

%s
%s
%s
%s
%s
## Generated Summary

- App: %s
- Target: %s
- Frontend: %s
- Backend: %s
- Database: %s
- Entities: %d
- Pages: %d
- Jobs: %d
`, g.program.App.Name, g.migrationReadmeSection(), g.browserTestReadmeSection(), g.deployReadmeSection(), g.jobsReadmeSection(), g.secretReadmeSection(), g.program.App.Name, g.targetName(), g.targetFrontend(), g.targetBackend(), g.targetDatabase(), len(g.program.Entities), len(g.program.Pages), len(g.program.Jobs))
}

func (g *webGenerator) envExample() string {
	var builder strings.Builder
	written := map[string]bool{}
	writeRaw := func(name string, value string) {
		if name == "" || written[name] {
			return
		}
		written[name] = true
		builder.WriteString(fmt.Sprintf("%s=%s\n", name, value))
	}
	writeQuoted := func(name string, value string) {
		writeRaw(name, fmt.Sprintf("%q", value))
	}

	writeRaw(g.deployPortEnv(), g.deployPortDefault())
	if g.hasDeployPreviewLocal() {
		writeRaw("BLACKLANG_PREVIEW_PORT", g.deployPreviewPortDefault())
	}
	writeQuoted(g.databaseEnvName(), g.defaultDatabaseURL())
	if g.isPostgres() {
		writeQuoted("POSTGRES_DB", "blacklang")
		writeQuoted("POSTGRES_USER", "blacklang")
		writeQuoted("POSTGRES_PASSWORD", "blacklang")
	}
	if g.isMySQL() {
		writeQuoted("MYSQL_DATABASE", "blacklang")
		writeQuoted("MYSQL_USER", "blacklang")
		writeQuoted("MYSQL_PASSWORD", "blacklang")
		writeQuoted("MYSQL_ROOT_PASSWORD", "blacklang_root")
	}
	if cors := g.corsConfig(); cors != nil && cors.Origins.Name != "" {
		writeQuoted(cors.Origins.Name, "http://localhost:5173")
	}
	if g.program.Deploy != nil && g.program.Deploy.Cloud != nil {
		writeRaw(g.program.Deploy.Cloud.App.Name, "")
		if g.program.Deploy.Cloud.Region.Name != "" {
			writeRaw(g.program.Deploy.Cloud.Region.Name, "")
		}
	}
	if g.program.Ops != nil && g.program.Ops.Observe != nil {
		writeRaw(g.program.Ops.Observe.Endpoint.Name, "")
	}
	if g.program.Deploy != nil {
		for _, env := range g.program.Deploy.Env {
			if value, ok := g.localEnvDefault(env.Name); ok {
				writeQuoted(env.Name, value)
				continue
			}
			writeRaw(env.Name, "")
		}
	}
	return builder.String()
}

func (g *webGenerator) localEnvDefault(name string) (string, bool) {
	if name == g.databaseEnvName() {
		return g.defaultDatabaseURL(), true
	}
	if g.isPostgres() {
		switch name {
		case "POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD":
			return "blacklang", true
		}
	}
	if g.isMySQL() {
		switch name {
		case "MYSQL_DATABASE", "MYSQL_USER", "MYSQL_PASSWORD":
			return "blacklang", true
		case "MYSQL_ROOT_PASSWORD":
			return "blacklang_root", true
		}
	}
	if cors := g.corsConfig(); cors != nil && cors.Origins.Name == name {
		return "http://localhost:5173", true
	}
	return "", false
}

func (g *webGenerator) corsConfig() *CORSDecl {
	if g.program.Security == nil {
		return nil
	}
	return g.program.Security.CORS
}

func (g *webGenerator) targetName() string {
	if g.program.Target != nil && g.program.Target.Name != "" {
		return g.program.Target.Name
	}
	return "web"
}

func (g *webGenerator) targetFrontend() string {
	if g.program.Target != nil && g.program.Target.Frontend != "" {
		return g.program.Target.Frontend
	}
	if g.isAPIOnlyTarget() {
		return "none"
	}
	return "react"
}

func (g *webGenerator) isAPIOnlyTarget() bool {
	return g.targetName() == "api"
}

func (g *webGenerator) targetBackend() string {
	if g.program.Target != nil && g.program.Target.Backend != "" {
		return g.program.Target.Backend
	}
	return "node"
}

func (g *webGenerator) targetDatabase() string {
	if g.program.Target != nil && g.program.Target.Database != "" {
		return g.program.Target.Database
	}
	return "sqlite"
}

func (g *webGenerator) databaseEnvName() string {
	if g.program.Database != nil && g.program.Database.URL.Name != "" {
		return g.program.Database.URL.Name
	}
	return "DATABASE_URL"
}

func (g *webGenerator) databaseEnvAccess() string {
	return "process.env." + g.databaseEnvName()
}

func (g *webGenerator) isPostgres() bool {
	return g.targetDatabase() == "postgres"
}

func (g *webGenerator) isMySQL() bool {
	return g.targetDatabase() == "mysql"
}

func (g *webGenerator) usesPrismaAuthRuntime() bool {
	return g.isPostgres() || g.isMySQL()
}

func (g *webGenerator) usesSQLComposeService() bool {
	return g.isPostgres() || g.isMySQL()
}

func (g *webGenerator) defaultDatabaseURL() string {
	if g.isPostgres() {
		return "postgresql://blacklang:blacklang@localhost:5432/blacklang"
	}
	if g.isMySQL() {
		return "mysql://blacklang:blacklang@localhost:3306/blacklang"
	}
	return "file:./dev.db"
}

func (g *webGenerator) composeDatabaseURL(databaseHost string) string {
	if g.isPostgres() {
		if databaseHost == "" {
			databaseHost = "postgres"
		}
		return fmt.Sprintf("postgresql://blacklang:blacklang@%s:5432/blacklang", databaseHost)
	}
	if g.isMySQL() {
		if databaseHost == "" {
			databaseHost = "mysql"
		}
		return fmt.Sprintf("mysql://blacklang:blacklang@%s:3306/blacklang", databaseHost)
	}
	return "file:/app/data/dev.db"
}

func (g *webGenerator) prismaProvider() string {
	if g.isPostgres() {
		return "postgresql"
	}
	if g.isMySQL() {
		return "mysql"
	}
	return "sqlite"
}

func (g *webGenerator) deployTarget() string {
	if g.program.Deploy == nil {
		return ""
	}
	return g.program.Deploy.Target
}

func (g *webGenerator) deployPortEnv() string {
	if g.program.Deploy != nil && g.program.Deploy.Port != nil && g.program.Deploy.Port.Env.Name != "" {
		return g.program.Deploy.Port.Env.Name
	}
	return "PORT"
}

func (g *webGenerator) deployPortDefault() string {
	if g.program.Deploy != nil && g.program.Deploy.Port != nil && g.program.Deploy.Port.Default != "" {
		return g.program.Deploy.Port.Default
	}
	return "3001"
}

func (g *webGenerator) deployEnvNames() []string {
	names := []string{g.deployPortEnv(), g.databaseEnvName()}
	if cors := g.corsConfig(); cors != nil && cors.Origins.Name != "" {
		names = append(names, cors.Origins.Name)
	}
	if g.program.Deploy != nil && g.program.Deploy.Cloud != nil {
		names = append(names, g.program.Deploy.Cloud.App.Name)
		if g.program.Deploy.Cloud.Region.Name != "" {
			names = append(names, g.program.Deploy.Cloud.Region.Name)
		}
	}
	if g.program.Ops != nil && g.program.Ops.Observe != nil && g.program.Ops.Observe.Endpoint.Name != "" {
		names = append(names, g.program.Ops.Observe.Endpoint.Name)
	}
	if g.program.Deploy != nil {
		for _, env := range g.program.Deploy.Env {
			names = append(names, env.Name)
		}
	}

	unique := []string{}
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		unique = append(unique, name)
	}
	return unique
}

func (g *webGenerator) composeEnvValue(name string, databaseHost string) string {
	if name == g.deployPortEnv() {
		return fmt.Sprintf("${%s:-%s}", name, g.deployPortDefault())
	}
	if name == g.databaseEnvName() {
		return fmt.Sprintf("${%s:-%s}", name, g.composeDatabaseURL(databaseHost))
	}
	if cors := g.corsConfig(); cors != nil && cors.Origins.Name == name {
		return fmt.Sprintf("${%s:-http://localhost:5173}", name)
	}
	return fmt.Sprintf("${%s}", name)
}

func (g *webGenerator) dockerignore() string {
	return `node_modules
dist
src/generated/prisma
dev.db
*.db
.env
npm-debug.log
`
}

func (g *webGenerator) dockerfile() string {
	return fmt.Sprintf(`# Generated by BlackLang. Do not edit manually.

FROM node:22-alpine

WORKDIR /app

COPY package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY . .
RUN npm run build

ENV NODE_ENV=production
ENV %s=%s

EXPOSE %s

%sCMD ["sh", "-c", "npm run db:setup && npm run start"]
`, g.deployPortEnv(), g.deployPortDefault(), g.deployPortDefault(), g.dockerfileHealthcheck())
}

func (g *webGenerator) dockerComposeYAML() string {
	serviceName := kebabCase(g.program.App.Name)
	if serviceName == "" {
		serviceName = "blacklang-app"
	}
	databaseServiceName := g.composeDatabaseServiceName(serviceName, false)
	volumeName := serviceName + "-data"

	var builder strings.Builder
	builder.WriteString("# Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("services:\n")
	builder.WriteString(fmt.Sprintf("  %s:\n", serviceName))
	builder.WriteString("    build: .\n")
	if g.usesSQLComposeService() {
		builder.WriteString("    depends_on:\n")
		builder.WriteString(fmt.Sprintf("      %s:\n", databaseServiceName))
		builder.WriteString("        condition: service_healthy\n")
	}
	builder.WriteString("    ports:\n")
	builder.WriteString(fmt.Sprintf("      - \"${%s:-%s}:${%s:-%s}\"\n", g.deployPortEnv(), g.deployPortDefault(), g.deployPortEnv(), g.deployPortDefault()))
	builder.WriteString("    environment:\n")
	builder.WriteString("      NODE_ENV: production\n")
	for _, name := range g.deployEnvNames() {
		builder.WriteString(fmt.Sprintf("      %s: \"%s\"\n", name, g.composeEnvValue(name, databaseServiceName)))
	}
	builder.WriteString(g.composeHealthcheck("    "))
	if !g.usesSQLComposeService() {
		builder.WriteString("    volumes:\n")
		builder.WriteString(fmt.Sprintf("      - %s:/app/data\n\n", volumeName))
		builder.WriteString("volumes:\n")
		builder.WriteString(fmt.Sprintf("  %s:\n", volumeName))
		return builder.String()
	}
	builder.WriteString("\n")
	builder.WriteString(g.sqlComposeServiceYAML(serviceName, databaseServiceName, false))
	builder.WriteString("volumes:\n")
	builder.WriteString(fmt.Sprintf("  %s-%s:\n", serviceName, g.targetDatabase()))
	return builder.String()
}

func (g *webGenerator) composeDatabaseServiceName(appServiceName string, preview bool) string {
	if !g.usesSQLComposeService() {
		return ""
	}
	serviceName := g.targetDatabase()
	if preview {
		serviceName += "-preview"
	}
	if serviceName != appServiceName {
		return serviceName
	}
	if preview {
		return g.targetDatabase() + "-database-preview"
	}
	return g.targetDatabase() + "-database"
}

func (g *webGenerator) sqlComposeServiceYAML(appServiceName string, databaseServiceName string, preview bool) string {
	if g.isPostgres() {
		return g.postgresComposeServiceYAML(appServiceName, databaseServiceName, preview)
	}
	if g.isMySQL() {
		return g.mysqlComposeServiceYAML(appServiceName, databaseServiceName, preview)
	}
	return ""
}

func (g *webGenerator) postgresComposeServiceYAML(appServiceName string, databaseServiceName string, preview bool) string {
	databaseName := "blacklang"
	if preview {
		databaseName = "blacklang_preview"
	}
	return fmt.Sprintf(`  %s:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: "${POSTGRES_DB:-%s}"
      POSTGRES_USER: "${POSTGRES_USER:-blacklang}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:-blacklang}"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-blacklang} -d ${POSTGRES_DB:-%s}"]
      interval: 5s
      timeout: 5s
      retries: 10
    volumes:
      - %s-postgres:/var/lib/postgresql/data

`, databaseServiceName, databaseName, databaseName, appServiceName)
}

func (g *webGenerator) mysqlComposeServiceYAML(appServiceName string, databaseServiceName string, preview bool) string {
	databaseName := "blacklang"
	if preview {
		databaseName = "blacklang_preview"
	}
	return fmt.Sprintf(`  %s:
    image: mysql:8.4
    environment:
      MYSQL_DATABASE: "${MYSQL_DATABASE:-%s}"
      MYSQL_USER: "${MYSQL_USER:-blacklang}"
      MYSQL_PASSWORD: "${MYSQL_PASSWORD:-blacklang}"
      MYSQL_ROOT_PASSWORD: "${MYSQL_ROOT_PASSWORD:-blacklang_root}"
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 -u$${MYSQL_USER:-blacklang} -p$${MYSQL_PASSWORD:-blacklang} --silent"]
      interval: 5s
      timeout: 5s
      retries: 10
    volumes:
      - %s-mysql:/var/lib/mysql

`, databaseServiceName, databaseName, appServiceName)
}

func (g *webGenerator) packageJSON() string {
	name := strings.ToLower(g.program.App.Name)
	setupScript := "tsx src/setup-db.ts"
	seedScript := ""
	if g.hasSeeds() {
		setupScript = "tsx src/setup-db.ts && npm run db:seed"
		seedScript = `
    "db:seed": "tsx src/seed.ts",`
	}
	devScript := "vite"
	buildScript := `node -e "require('fs').rmSync('dist', { recursive: true, force: true })" && npm run db:generate && tsc && vite build`
	testScript := "npm run db:generate && tsx src/blacklang.contract.test.ts && tsx src/blacklang.api.test.ts"
	e2eScripts := ""
	frontendDependencies := ""
	frontendDevDependencies := ""
	playwrightDevDependency := ""
	if !g.isAPIOnlyTarget() {
		testScript += " && tsx src/blacklang.frontend.test.tsx"
		frontendDependencies = `    "react": "latest",
    "react-dom": "latest",
`
		frontendDevDependencies = `    "@types/react": "latest",
    "@types/react-dom": "latest",
    "@vitejs/plugin-react": "latest",
    "vite": "latest",
`
	}
	if g.isAPIOnlyTarget() {
		devScript = "tsx src/server.ts"
		buildScript = "npm run db:generate && tsc"
	}
	if !g.isAPIOnlyTarget() && g.hasBrowserChecks() {
		testScript += " && tsx src/blacklang.browser.test.tsx"
		e2eScripts = `,
    "test:e2e": "npm run db:generate && tsx src/blacklang.e2e.test.ts",
    "test:e2e:plan": "tsx src/blacklang.e2e.matrix.ts --plan",
    "test:e2e:matrix": "npm run db:generate && tsx src/blacklang.e2e.matrix.ts --run",
    "test:all": "npm test && npm run test:e2e:matrix"`
		playwrightDevDependency = `
    "playwright-core": "latest",`
	}
	deployScripts := g.packageDeployScripts()
	jobScripts := g.packageJobScripts()
	migrationScripts := g.packageMigrationScripts()
	databaseDependencies := `    "@prisma/adapter-better-sqlite3": "7.10.0",
    "@prisma/client": "7.10.0",
    "better-sqlite3": "latest",`
	databaseDevDependencies := `    "@types/better-sqlite3": "latest",`
	if g.isPostgres() {
		databaseDependencies = `    "@prisma/adapter-pg": "7.10.0",
    "@prisma/client": "7.10.0",
    "pg": "latest",`
		databaseDevDependencies = `    "@types/pg": "latest",`
	}
	if g.isMySQL() {
		databaseDependencies = `    "@prisma/adapter-mariadb": "7.10.0",
    "@prisma/client": "7.10.0",`
		databaseDevDependencies = ""
	}
	devScriptJSON := strconv.Quote(devScript)
	buildScriptJSON := strconv.Quote(buildScript)
	testScriptJSON := strconv.Quote(testScript)
	setupScriptJSON := strconv.Quote(setupScript)
	return fmt.Sprintf(`{
  "name": "%s-generated",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": %s,
    "api:dev": "tsx src/server.ts",
    "start": "tsx src/server.ts",
    "build": %s,
    "test": %s%s%s%s,
    "security:secrets:plan": "node scripts/secrets-plan.mjs",
    "security:secrets:preflight": "node scripts/secrets-provider.mjs --preflight",
    "db:generate": "prisma generate",
    "db:validate": "prisma validate",%s
    "db:setup": %s,%s
    "db:push": "npm run db:setup",
    "db:push:native": "prisma db push"
  },
  "dependencies": {
%s
    "dotenv": "latest",
%s
    "express": "latest"
  },
  "devDependencies": {
    "@types/express": "latest",
%s
    "@types/node": "latest",
%s
    "prisma": "7.10.0",%s
    "tsx": "latest",
    "typescript": "latest"
  }
}
`, name, devScriptJSON, buildScriptJSON, testScriptJSON, e2eScripts, deployScripts, jobScripts, migrationScripts, setupScriptJSON, seedScript, databaseDependencies, frontendDependencies, databaseDevDependencies, frontendDevDependencies, playwrightDevDependency)
}

func (g *webGenerator) prismaConfig() string {
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { defineConfig } from "prisma/config";

export default defineConfig({
  schema: "prisma/schema.prisma",
  datasource: {
    url: %s ?? %q
  }
});
`, g.databaseEnvAccess(), g.defaultDatabaseURL())
}

func (g *webGenerator) viteConfig() string {
	return `// Generated by BlackLang. Do not edit manually.

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": "http://localhost:3001"
    }
  }
});
`
}

func (g *webGenerator) indexHTML() string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
`, g.program.App.Name)
}

func (g *webGenerator) tsconfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["DOM", "DOM.Iterable", "ES2020"],
    "allowJs": false,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx"
  },
  "include": ["src"]
}
`
}

func (g *webGenerator) mainTSX() string {
	return `// Generated by BlackLang. Do not edit manually.

import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./styles.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
`
}

func (g *webGenerator) authPageTSX() string {
	emailPlaceholder := "you@example.com"
	namePlaceholder := "Full name"
	if g.program.Auth != nil {
		for _, field := range g.program.Auth.User.Fields {
			if field.Name == "email" {
				if value := modifierValue(field, "placeholder"); value != "" {
					emailPlaceholder = value
				}
			}
			if field.Name == "name" {
				if value := modifierValue(field, "placeholder"); value != "" {
					namePlaceholder = value
				}
			}
		}
	}

	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import { useState } from "react";
import type { FormEvent } from "react";

type AuthPageProps = {
  appName: string;
  onAuthenticated: (user: { id: string; name: string; email: string; role: string; roles: string[] }) => void;
};

export function AuthPage({ appName, onAuthenticated }: AuthPageProps) {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const input = {
      name: String(form.get("name") ?? ""),
      email: String(form.get("email") ?? ""),
      password: String(form.get("password") ?? "")
    };

    setSaving(true);
    setError(null);
    try {
      const response = await fetch("/api/auth/" + mode, {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input)
      });
      if (!response.ok) {
        const body = await response.json().catch(() => ({ error: "Authentication failed" }));
        throw new Error(String(body.error ?? "Authentication failed"));
      }
      const body = await response.json();
      onAuthenticated(body.user);
    } catch (error) {
      setError(error instanceof Error ? error.message : "Authentication failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="auth-screen">
      <section className="auth-panel" aria-labelledby="auth-title">
        <div>
          <span className="eyebrow">{appName}</span>
          <h1 id="auth-title">{mode === "login" ? "Sign in" : "Create account"}</h1>
          <p className="muted">Generated from BlackLang auth intent with cookie session persistence.</p>
        </div>

        <div className="auth-tabs" role="tablist" aria-label="Authentication mode">
          <button className={mode === "login" ? "active" : "secondary"} type="button" onClick={() => setMode("login")}>Login</button>
          <button className={mode === "register" ? "active" : "secondary"} type="button" onClick={() => setMode("register")}>Register</button>
        </div>

        <form onSubmit={submit}>
          {error && <div className="error">{error}</div>}
          {mode === "register" && (
            <label>
              Name
              <input name="name" required placeholder=%q />
            </label>
          )}
          <label>
            Email
            <input name="email" required type="email" placeholder=%q />
          </label>
          <label>
            Password
            <input name="password" required minLength={8} type="password" placeholder="At least 8 characters" />
          </label>
          <button disabled={saving} type="submit">{saving ? "Please wait" : mode === "login" ? "Sign in" : "Create account"}</button>
        </form>
      </section>
    </main>
  );
}
`, namePlaceholder, emailPlaceholder)
}

func (g *webGenerator) authUsersPageTSX() string {
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import { useEffect, useState } from "react";

type User = {
  id: string;
  name: string;
  email: string;
  role: string;
  roles: string[];
  tenantId?: string;
};

const defaultRole = %q;
const roles = %s;
const tenantAdminEnabled = %t;

function csrfHeaders(): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const token = document.cookie.split("; ").find((item) => item.startsWith("black_csrf="))?.split("=")[1] ?? "";
  if (token) headers["X-CSRF-Token"] = decodeURIComponent(token);
  return headers;
}

function userRoles(user: User) {
  return user.roles.length > 0 ? user.roles : [user.role || defaultRole];
}

function toggleRole(user: User, role: string) {
  const current = userRoles(user);
  const next = current.includes(role) ? current.filter((item) => item !== role) : [...current, role];
  return next.length > 0 ? next : [defaultRole];
}

export function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [savingId, setSavingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    fetch("/api/auth/users", { credentials: "same-origin" })
      .then(async (response) => {
        if (!response.ok) {
          const body = await response.json().catch(() => ({ error: "Unable to load users" }));
          throw new Error(String(body.error ?? "Unable to load users"));
        }
        return response.json() as Promise<{ users: User[] }>;
      })
      .then((body) => {
        if (active) setUsers(body.users);
      })
      .catch((reason: unknown) => {
        if (active) setError(reason instanceof Error ? reason.message : "Unable to load users");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  async function updateRoles(userId: string, nextRoles: string[]) {
    setSavingId(userId);
    setError(null);
    try {
      const response = await fetch("/api/auth/users/" + userId + "/role", {
        method: "PUT",
        credentials: "same-origin",
        headers: csrfHeaders(),
        body: JSON.stringify({ roles: nextRoles })
      });
      if (!response.ok) {
        const body = await response.json().catch(() => ({ error: "Unable to update roles" }));
        throw new Error(String(body.error ?? "Unable to update roles"));
      }
      const body = await response.json() as { user: User };
      setUsers((current) => current.map((user) => user.id === userId ? body.user : user));
    } catch (reason: unknown) {
      setError(reason instanceof Error ? reason.message : "Unable to update roles");
    } finally {
      setSavingId(null);
    }
  }

  async function updateTenant(userId: string, nextTenantId: string) {
    if (!tenantAdminEnabled) return;
    const tenantId = nextTenantId.trim();
    if (!tenantId) {
      setError("Tenant id is required");
      return;
    }
    setSavingId(userId);
    setError(null);
    try {
      const response = await fetch("/api/auth/users/" + userId + "/tenant", {
        method: "PUT",
        credentials: "same-origin",
        headers: csrfHeaders(),
        body: JSON.stringify({ tenantId })
      });
      if (!response.ok) {
        const body = await response.json().catch(() => ({ error: "Unable to update tenant" }));
        throw new Error(String(body.error ?? "Unable to update tenant"));
      }
      const body = await response.json() as { user: User };
      setUsers((current) => current.map((user) => user.id === userId ? body.user : user));
    } catch (reason: unknown) {
      setError(reason instanceof Error ? reason.message : "Unable to update tenant");
    } finally {
      setSavingId(null);
    }
  }

  return (
    <main>
      <header>
        <h1>Users</h1>
        <span>{tenantAdminEnabled ? "Role and tenant management" : "Role management"}</span>
      </header>

      <section className="panel">
        <div className="toolbar">
          <div>
            <h2>Users</h2>
            <p className="muted">{tenantAdminEnabled ? "Generated from BlackLang role declarations and tenant policies. A user can have multiple roles and one tenant scope." : "Generated from BlackLang role declarations. A user can have multiple roles."}</p>
          </div>
        </div>
        {error && <div className="error" role="alert">{error}</div>}
        {loading && <div className="status">Loading users...</div>}
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Email</th>
              <th>Roles</th>
              {tenantAdminEnabled && <th>Tenant</th>}
            </tr>
          </thead>
          <tbody>
            {!loading && users.length === 0 && (
              <tr><td colSpan={tenantAdminEnabled ? 4 : 3}>No users yet.</td></tr>
            )}
            {users.map((user) => (
              <tr key={user.id}>
                <td>{user.name}</td>
                <td>{user.email}</td>
                <td>
                  <div className="role-list">
                    {roles.map((role) => (
                      <label key={role} className="role-check">
                        <input
                          checked={userRoles(user).includes(role)}
                          disabled={savingId === user.id}
                          type="checkbox"
                          onChange={() => updateRoles(user.id, toggleRole(user, role))}
                        />
                        <span>{role}</span>
                      </label>
                    ))}
                  </div>
                </td>
                {tenantAdminEnabled && (
                  <td>
                    <label className="tenant-editor">
                      <span className="field-note">Tenant ID</span>
                      <input
                        key={user.id + ":" + (user.tenantId ?? "default")}
                        defaultValue={user.tenantId ?? "default"}
                        disabled={savingId === user.id}
                        onBlur={(event) => {
                          const tenantId = event.currentTarget.value.trim();
                          if (tenantId && tenantId !== (user.tenantId ?? "default")) void updateTenant(user.id, tenantId);
                        }}
                      />
                    </label>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </main>
  );
}
`, g.defaultAuthRole(), tsStringArrayLiteral(g.roleNames()), g.hasTenantPolicies())
}

func (g *webGenerator) authAuditPageTSX() string {
	return `// Generated by BlackLang. Do not edit manually.

import { useEffect, useState } from "react";

type AuditLog = {
  id: string;
  actorUserId: string;
  actorRole: string;
  action: string;
  resource: string;
  resourceId: string;
  summary: string;
  createdAt: string;
};

export function AuditPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);
    fetch("/api/auth/audit", { credentials: "same-origin" })
      .then(async (response) => {
        if (!response.ok) {
          const body = await response.json().catch(() => ({ error: "Unable to load audit logs" }));
          throw new Error(String(body.error ?? "Unable to load audit logs"));
        }
        return response.json() as Promise<{ logs: AuditLog[] }>;
      })
      .then((body) => {
        if (active) setLogs(body.logs);
      })
      .catch((reason: unknown) => {
        if (active) setError(reason instanceof Error ? reason.message : "Unable to load audit logs");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  return (
    <main>
      <header>
        <h1>Audit</h1>
        <span>Security activity log</span>
      </header>

      <section className="panel">
        <div className="toolbar">
          <div>
            <h2>Audit Log</h2>
            <p className="muted">Generated from BlackLang auth, role, and action intent.</p>
          </div>
        </div>
        {error && <div className="error" role="alert">{error}</div>}
        {loading && <div className="status">Loading audit logs...</div>}
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Actor</th>
              <th>Action</th>
              <th>Resource</th>
              <th>Summary</th>
            </tr>
          </thead>
          <tbody>
            {!loading && logs.length === 0 && (
              <tr><td colSpan={5}>No audit logs yet.</td></tr>
            )}
            {logs.map((log) => (
              <tr key={log.id}>
                <td>{new Date(log.createdAt).toLocaleString()}</td>
                <td>{log.actorRole}</td>
                <td>{log.action}</td>
                <td>{log.resource}{log.resourceId ? " / " + log.resourceId : ""}</td>
                <td>{log.summary}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </main>
  );
}
`
}

func (g *webGenerator) appTSX() string {
	if len(g.program.Pages) == 0 {
		return `// Generated by BlackLang. Do not edit manually.

export function App() {
  return <main>No pages generated.</main>;
}
`
	}

	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	if g.program.Auth != nil {
		builder.WriteString("import { useEffect, useState } from \"react\";\n")
	} else {
		builder.WriteString("import { useState } from \"react\";\n")
	}
	if g.program.Auth != nil {
		builder.WriteString("import { AuthPage } from \"./auth/AuthPage\";\n")
		if len(g.program.Roles) > 0 {
			builder.WriteString("import { UsersPage } from \"./auth/UsersPage\";\n")
			builder.WriteString("import { AuditPage } from \"./auth/AuditPage\";\n")
		}
	}
	for _, page := range g.program.Pages {
		builder.WriteString(fmt.Sprintf("import { %sPage } from \"./pages/%sPage\";\n", page.Name, page.Name))
	}
	builder.WriteString("\n")
	if g.program.Auth != nil {
		builder.WriteString("type CurrentUser = { id: string; name: string; email: string; role: string; roles: string[] };\n\n")
		builder.WriteString("type RolePermission = { effect: string; action: string; resource: string; fields: string[] };\n\n")
		if g.hasCustomActions() {
			builder.WriteString("type CustomActionPermission = { name: string; allow: string[]; writes: string[] };\n\n")
		}
		builder.WriteString(fmt.Sprintf("const rolePermissions: Record<string, RolePermission[]> = %s;\n\n", g.rolePermissionsLiteral()))
		builder.WriteString("function permissionMatches(permission: RolePermission, action: string, resource: string, field?: string) {\n")
		builder.WriteString("  if (permission.action === \"all\") return true;\n")
		builder.WriteString("  const actionMatches = permission.action === action || permission.action === \"manage\";\n")
		builder.WriteString("  if (!actionMatches || permission.resource !== resource) return false;\n")
		builder.WriteString("  if (!field || permission.fields.length === 0) return true;\n")
		builder.WriteString("  return permission.fields.includes(field);\n")
		builder.WriteString("}\n\n")
		builder.WriteString("function permissionsForRoles(roles: string[]) {\n")
		builder.WriteString("  return roles.flatMap((role) => rolePermissions[role] ?? []);\n")
		builder.WriteString("}\n\n")
	}

	navigationPages := g.navigationPages()
	builder.WriteString("const pages = [\n")
	for _, page := range navigationPages {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("  { name: %q, label: %q, access: %s },\n", page.Name, g.uiLabelFallback("page."+page.Name, page.Name), tsStringArrayLiteral(page.Access)))
		} else {
			builder.WriteString(fmt.Sprintf("  { name: %q, access: %s },\n", page.Name, tsStringArrayLiteral(page.Access)))
		}
	}
	if g.program.Auth != nil && len(g.program.Roles) > 0 {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("  { name: \"Users\", label: %q, access: %s },\n", g.uiLabelFallback("page.Users", "Users"), tsStringArrayLiteral([]string{g.defaultAuthRole()})))
			builder.WriteString(fmt.Sprintf("  { name: \"Audit\", label: %q, access: %s },\n", g.uiLabelFallback("page.Audit", "Audit"), tsStringArrayLiteral([]string{g.defaultAuthRole()})))
		} else {
			builder.WriteString(fmt.Sprintf("  { name: \"Users\", access: %s },\n", tsStringArrayLiteral([]string{g.defaultAuthRole()})))
			builder.WriteString(fmt.Sprintf("  { name: \"Audit\", access: %s },\n", tsStringArrayLiteral([]string{g.defaultAuthRole()})))
		}
	}
	builder.WriteString("];\n\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("const locales = %s;\n", tsStringArrayLiteral(g.program.I18N.Locales)))
		builder.WriteString(fmt.Sprintf("const defaultLocale = %q;\n", g.program.I18N.Default))
		builder.WriteString("const rtlLocalePrefixes = [\"ar\", \"fa\", \"he\", \"ur\"];\n\n")
		builder.WriteString(g.localizedUILabelMap("uiLabels"))
		builder.WriteString("function localeDirection(locale: string) {\n")
		builder.WriteString("  const normalized = locale.toLowerCase();\n")
		builder.WriteString("  return rtlLocalePrefixes.some((prefix) => normalized === prefix || normalized.startsWith(prefix + \"-\")) ? \"rtl\" : \"ltr\";\n")
		builder.WriteString("}\n\n")
		builder.WriteString("function uiLabel(key: string, fallback: string, locale: string) {\n")
		builder.WriteString("  return uiLabels[key]?.[locale] ?? uiLabels[key]?.[defaultLocale] ?? fallback;\n")
		builder.WriteString("}\n\n")
	}
	builder.WriteString("export function App() {\n")
	builder.WriteString("  const [activePage, setActivePage] = useState(pages[0].name);\n")
	builder.WriteString("  const [navOpen, setNavOpen] = useState(false);\n")
	if g.hasRuntimeI18N() {
		builder.WriteString("  const [activeLocale, setActiveLocale] = useState(defaultLocale);\n")
		builder.WriteString("  const activeDirection = localeDirection(activeLocale);\n")
	}
	if g.program.Auth != nil {
		builder.WriteString("  const [authenticated, setAuthenticated] = useState<boolean | null>(null);\n")
		builder.WriteString("  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);\n")
		builder.WriteString("  const currentRoles = currentUser ? ((currentUser.roles?.length ?? 0) > 0 ? currentUser.roles : [currentUser.role]) : [];\n")
	}
	if g.program.Auth != nil {
		builder.WriteString("  const visiblePages = currentUser ? pages.filter((item) => item.access.length === 0 || item.access.includes(\"authenticated\") || currentRoles.some((role) => item.access.includes(role))) : pages;\n")
	} else {
		builder.WriteString("  const visiblePages = pages;\n")
	}
	builder.WriteString("  const page = visiblePages.find((item) => item.name === activePage) ?? visiblePages[0];\n\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("  const appTitle = %s;\n", g.uiLabelExpression("app.title", g.program.App.Name, "activeLocale")))
		builder.WriteString("  const pageLabel = page ? uiLabel(\"page.\" + page.name, page.label, activeLocale) : \"\";\n\n")
	}
	if g.program.Auth != nil {
		builder.WriteString("  useEffect(() => {\n")
		builder.WriteString("    fetch(\"/api/auth/me\", { credentials: \"same-origin\" })\n")
		builder.WriteString("      .then(async (response) => {\n")
		builder.WriteString("        if (!response.ok) {\n")
		builder.WriteString("          setCurrentUser(null);\n")
		builder.WriteString("          setAuthenticated(false);\n")
		builder.WriteString("          return;\n")
		builder.WriteString("        }\n")
		builder.WriteString("        const body = await response.json();\n")
		builder.WriteString("        setCurrentUser(body.user);\n")
		builder.WriteString("        setAuthenticated(true);\n")
		builder.WriteString("      })\n")
		builder.WriteString("      .catch(() => {\n")
		builder.WriteString("        setCurrentUser(null);\n")
		builder.WriteString("        setAuthenticated(false);\n")
		builder.WriteString("      });\n")
		builder.WriteString("  }, []);\n\n")
		builder.WriteString("  async function logout() {\n")
		builder.WriteString("    const csrfToken = document.cookie.split(\"; \").find((item) => item.startsWith(\"black_csrf=\"))?.split(\"=\")[1] ?? \"\";\n")
		builder.WriteString("    await fetch(\"/api/auth/logout\", { method: \"POST\", credentials: \"same-origin\", headers: csrfToken ? { \"X-CSRF-Token\": decodeURIComponent(csrfToken) } : {} });\n")
		builder.WriteString("    setCurrentUser(null);\n")
		builder.WriteString("    setAuthenticated(false);\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  if (authenticated === null) {\n")
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("    return <main className=\"auth-screen\"><section className=\"auth-panel\"><p className=\"muted\">{%s}</p></section></main>;\n", g.uiLabelExpression("app.checkingSession", "Checking session...", "activeLocale")))
		} else {
			builder.WriteString("    return <main className=\"auth-screen\"><section className=\"auth-panel\"><p className=\"muted\">Checking session...</p></section></main>;\n")
		}
		builder.WriteString("  }\n\n")
		builder.WriteString("  if (!authenticated) {\n")
		if g.hasRuntimeI18N() {
			builder.WriteString("    return <AuthPage appName={appTitle} onAuthenticated={(user) => { setCurrentUser(user); setAuthenticated(true); }} />;\n")
		} else {
			builder.WriteString("    return <AuthPage appName=\"")
			builder.WriteString(g.program.App.Name)
			builder.WriteString("\" onAuthenticated={(user) => { setCurrentUser(user); setAuthenticated(true); }} />;\n")
		}
		builder.WriteString("  }\n\n")
		builder.WriteString("  if (!page) {\n")
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("    return <main className=\"auth-screen\"><section className=\"auth-panel\"><p className=\"muted\">{%s}</p><button className=\"secondary\" type=\"button\" onClick={logout}>{%s}</button></section></main>;\n", g.uiLabelExpression("app.noPages", "No pages are available for this role.", "activeLocale"), g.uiLabelExpression("app.logout", "Logout", "activeLocale")))
		} else {
			builder.WriteString("    return <main className=\"auth-screen\"><section className=\"auth-panel\"><p className=\"muted\">No pages are available for this role.</p><button className=\"secondary\" type=\"button\" onClick={logout}>Logout</button></section></main>;\n")
		}
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  function navigateTo(pageName: string) {\n")
	builder.WriteString("    setActivePage(pageName);\n")
	builder.WriteString("    setNavOpen(false);\n")
	builder.WriteString("  }\n\n")
	if g.program.Auth != nil {
		builder.WriteString("  function canAccessAction(action: string, resource: string) {\n")
		builder.WriteString("    if (!currentUser) return false;\n")
		builder.WriteString("    const permissions = permissionsForRoles(currentRoles);\n")
		builder.WriteString("    const denied = permissions.some((permission) => permission.effect === \"deny\" && permission.fields.length === 0 && permissionMatches(permission, action, resource));\n")
		builder.WriteString("    if (denied) return false;\n")
		builder.WriteString("    return permissions.some((permission) => permission.effect === \"allow\" && permissionMatches(permission, action, resource));\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  function canAccessField(action: string, resource: string, field: string) {\n")
		builder.WriteString("    if (!currentUser) return false;\n")
		builder.WriteString("    const permissions = permissionsForRoles(currentRoles);\n")
		builder.WriteString("    const denied = permissions.some((permission) => permission.effect === \"deny\" && permissionMatches(permission, action, resource, field));\n")
		builder.WriteString("    if (denied) return false;\n")
		builder.WriteString("    return permissions.some((permission) => permission.effect === \"allow\" && permissionMatches(permission, action, resource, field));\n")
		builder.WriteString("  }\n\n")
		if g.hasCustomActions() {
			builder.WriteString("  function pagePermissions(resource: string, fields: string[], customActions: CustomActionPermission[] = []) {\n")
		} else {
			builder.WriteString("  function pagePermissions(resource: string, fields: string[]) {\n")
		}
		builder.WriteString("    const fieldAccess = Object.fromEntries(fields.map((field) => [field, canAccessField(\"read\", resource, field)]));\n")
		builder.WriteString("    const writeAccess = Object.fromEntries(fields.map((field) => [field, canAccessField(\"update\", resource, field)]));\n")
		if g.hasCustomActions() {
			builder.WriteString("    const customActionAccess = Object.fromEntries(customActions.map((action) => {\n")
			builder.WriteString("      const roleAllowed = action.allow.length === 0 || action.allow.includes(\"authenticated\") || currentRoles.some((role) => action.allow.includes(role));\n")
			builder.WriteString("      const fieldsAllowed = action.writes.every((field) => canAccessField(\"update\", resource, field));\n")
			builder.WriteString("      return [action.name, canAccessAction(\"update\", resource) && roleAllowed && fieldsAllowed];\n")
			builder.WriteString("    }));\n")
		}
		builder.WriteString("    return {\n")
		builder.WriteString("      read: canAccessAction(\"read\", resource),\n")
		builder.WriteString("      create: canAccessAction(\"create\", resource),\n")
		builder.WriteString("      update: canAccessAction(\"update\", resource),\n")
		builder.WriteString("      delete: canAccessAction(\"delete\", resource),\n")
		builder.WriteString("      fields: fieldAccess,\n")
		if g.hasCustomActions() {
			builder.WriteString("      writeFields: writeAccess,\n")
			builder.WriteString("      customActions: customActionAccess\n")
		} else {
			builder.WriteString("      writeFields: writeAccess\n")
		}
		builder.WriteString("    };\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  function renderPage() {\n")
	builder.WriteString("    switch (page.name) {\n")
	for _, page := range g.program.Pages {
		builder.WriteString(fmt.Sprintf("      case %q:\n", page.Name))
		if g.program.Auth != nil {
			if entity, ok := g.findEntity(page.Source); ok {
				builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo}%s permissions={%s} />;\n", page.Name, g.localePageProp(), g.pagePermissionsCall(page, entity)))
			}
		} else {
			builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo}%s />;\n", page.Name, g.localePageProp()))
		}
	}
	if g.program.Auth != nil && len(g.program.Roles) > 0 {
		builder.WriteString("      case \"Users\":\n")
		builder.WriteString("        return <UsersPage />;\n")
		builder.WriteString("      case \"Audit\":\n")
		builder.WriteString("        return <AuditPage />;\n")
	}
	builder.WriteString("      default:\n")
	if g.program.Auth != nil {
		if entity, ok := g.findEntity(g.program.Pages[0].Source); ok {
			builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo}%s permissions={%s} />;\n", g.program.Pages[0].Name, g.localePageProp(), g.pagePermissionsCall(g.program.Pages[0], entity)))
		}
	} else {
		builder.WriteString(fmt.Sprintf("        return <%sPage onNavigate={navigateTo}%s />;\n", g.program.Pages[0].Name, g.localePageProp()))
	}
	builder.WriteString("    }\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  return (\n")
	if g.hasRuntimeI18N() {
		builder.WriteString("    <div className={navOpen ? \"app-shell nav-open\" : \"app-shell\"} lang={activeLocale} dir={activeDirection}>\n")
	} else {
		builder.WriteString("    <div className={navOpen ? \"app-shell nav-open\" : \"app-shell\"}>\n")
	}
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("      {navOpen && <button className=\"nav-backdrop\" type=\"button\" aria-label={%s} onClick={() => setNavOpen(false)} />}\n", g.uiLabelExpression("app.closeNavigation", "Close navigation", "activeLocale")))
	} else {
		builder.WriteString("      {navOpen && <button className=\"nav-backdrop\" type=\"button\" aria-label=\"Close navigation\" onClick={() => setNavOpen(false)} />}\n")
	}
	builder.WriteString("      <aside className=\"app-sidebar\">\n")
	if g.hasRuntimeI18N() {
		builder.WriteString("        <div className=\"app-brand\">{appTitle}</div>\n")
		builder.WriteString(fmt.Sprintf("        <nav className=\"app-nav\" aria-label={%s}>\n", g.uiLabelExpression("app.primaryNavigation", "Primary navigation", "activeLocale")))
	} else {
		builder.WriteString(fmt.Sprintf("        <div className=\"app-brand\">%s</div>\n", g.program.App.Name))
		builder.WriteString("        <nav className=\"app-nav\" aria-label=\"Primary navigation\">\n")
	}
	builder.WriteString("          {visiblePages.map((item) => (\n")
	if g.hasRuntimeI18N() {
		builder.WriteString("            <button key={item.name} className={item.name === page.name ? \"active\" : \"secondary\"} type=\"button\" onClick={() => navigateTo(item.name)}>{uiLabel(\"page.\" + item.name, item.label, activeLocale)}</button>\n")
	} else {
		builder.WriteString("            <button key={item.name} className={item.name === page.name ? \"active\" : \"secondary\"} type=\"button\" onClick={() => navigateTo(item.name)}>{item.name}</button>\n")
	}
	builder.WriteString("          ))}\n")
	builder.WriteString("        </nav>\n")
	builder.WriteString("      </aside>\n")
	builder.WriteString("      <div className=\"app-workspace\">\n")
	builder.WriteString("        <div className=\"app-topbar\">\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("          <button className=\"menu-button secondary\" type=\"button\" aria-expanded={navOpen} onClick={() => setNavOpen(true)}>{%s}</button>\n", g.uiLabelExpression("app.menu", "Menu", "activeLocale")))
	} else {
		builder.WriteString("          <button className=\"menu-button secondary\" type=\"button\" aria-expanded={navOpen} onClick={() => setNavOpen(true)}>Menu</button>\n")
	}
	builder.WriteString("          <div>\n")
	if g.hasRuntimeI18N() {
		builder.WriteString("            <span className=\"breadcrumb\">{appTitle} / {pageLabel}</span>\n")
		builder.WriteString("            <h1>{pageLabel}</h1>\n")
	} else {
		builder.WriteString(fmt.Sprintf("            <span className=\"breadcrumb\">%s / {page.name}</span>\n", g.program.App.Name))
		builder.WriteString("            <h1>{page.name}</h1>\n")
	}
	builder.WriteString("          </div>\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("          <label className=\"locale-switcher\">{%s}\n", g.uiLabelExpression("app.language", "Language", "activeLocale")))
		builder.WriteString("            <select value={activeLocale} onChange={(event) => setActiveLocale(event.target.value)}>\n")
		builder.WriteString("              {locales.map((locale) => <option key={locale} value={locale}>{locale.toUpperCase()}</option>)}\n")
		builder.WriteString("            </select>\n")
		builder.WriteString("          </label>\n")
	}
	if g.program.Auth != nil {
		builder.WriteString("          <div className=\"user-menu\">\n")
		builder.WriteString("            {currentUser && <span className=\"role-badge\">{currentUser.name} / {currentRoles.join(\", \")}</span>}\n")
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("          <button className=\"secondary\" type=\"button\" onClick={logout}>{%s}</button>\n", g.uiLabelExpression("app.logout", "Logout", "activeLocale")))
		} else {
			builder.WriteString("          <button className=\"secondary\" type=\"button\" onClick={logout}>Logout</button>\n")
		}
		builder.WriteString("          </div>\n")
	}
	builder.WriteString("        </div>\n")
	builder.WriteString("        <div className=\"app-content\">{renderPage()}</div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </div>\n")
	builder.WriteString("  );\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) hasRuntimeI18N() bool {
	return g.program.I18N != nil && g.program.I18N.Default != "" && len(g.program.I18N.Locales) > 1
}

func (g *webGenerator) localePageProp() string {
	if !g.hasRuntimeI18N() {
		return ""
	}
	return " locale={activeLocale}"
}

func (g *webGenerator) localizedUILabelMap(name string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("const %s: Record<string, Record<string, string>> = {\n", name))
	entityIndex := programEntityIndex(g.program)
	for _, label := range g.program.Labels {
		if _, ok := isUILabelTarget(g.program, entityIndex, label.Target); !ok {
			continue
		}
		translations := translationValuesMap(label.Translations)
		if len(translations) == 0 {
			continue
		}
		builder.WriteString(fmt.Sprintf("  %q: %s,\n", label.Target, g.translationMapLiteral(translations)))
	}
	builder.WriteString("};\n\n")
	return builder.String()
}

func (g *webGenerator) uiLabelTranslations(target string) map[string]string {
	for _, label := range g.program.Labels {
		if label.Target != target {
			continue
		}
		return translationValuesMap(label.Translations)
	}
	return nil
}

func translationValuesMap(values []TranslationValue) map[string]string {
	translations := map[string]string{}
	for _, translation := range values {
		translations[translation.Locale] = translation.Text
	}
	return translations
}

func (g *webGenerator) uiLabelFallback(target string, fallback string) string {
	if g.program.I18N == nil || g.program.I18N.Default == "" {
		return fallback
	}
	if text := g.uiLabelTranslations(target)[g.program.I18N.Default]; text != "" {
		return text
	}
	return fallback
}

func (g *webGenerator) uiLabelExpression(target string, fallback string, locale string) string {
	fallback = g.uiLabelFallback(target, fallback)
	if !g.hasRuntimeI18N() {
		return fmt.Sprintf("%q", fallback)
	}
	return fmt.Sprintf("uiLabel(%q, %q, %s)", target, fallback, locale)
}

func (g *webGenerator) uiLabelJSX(target string, fallback string, locale string) string {
	if !g.hasRuntimeI18N() {
		return escapeJSXText(g.uiLabelFallback(target, fallback))
	}
	return fmt.Sprintf("{%s}", g.uiLabelExpression(target, fallback, locale))
}

func (g *webGenerator) serverTS() string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("import express from \"express\";\n")
	if g.hasOpsObserve() || g.hasTransactionalPrivateExplicitAPIHandlers() {
		builder.WriteString("import crypto from \"node:crypto\";\n")
	}
	if !g.isAPIOnlyTarget() {
		builder.WriteString("import path from \"node:path\";\n")
	}
	builder.WriteString("import { pathToFileURL } from \"node:url\";\n")
	if g.hasOpsReadiness() || g.hasExplicitAPIHandlers() {
		builder.WriteString("import { prisma } from \"./db\";\n")
	}
	if g.program.Auth != nil {
		authImports := []string{"authRouter", "requireAuth", "requireCsrf"}
		if g.hasRuntimePermissions() && g.hasPrivateExplicitAPIHandlers() {
			authImports = append(authImports, "canAccessField", "writeAuditLog")
		}
		builder.WriteString(fmt.Sprintf("import { %s } from \"./routes/auth\";\n", strings.Join(authImports, ", ")))
	}
	for _, page := range g.program.Pages {
		fileName := g.pageModuleName(page)
		identifier := lowerCamelCase(page.Source)
		builder.WriteString(fmt.Sprintf("import { %sRouter as %s } from \"./routes/%s\";\n", identifier, g.pageRouterName(page), fileName))
	}
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("const port = Number(process.env[%q] ?? %q);\n\n", g.deployPortEnv(), g.deployPortDefault()))
	builder.WriteString("const openAPISpec = ")
	builder.WriteString(strings.TrimSpace(g.openapiJSON()))
	builder.WriteString(" as const;\n\n")
	builder.WriteString(g.explicitAPIHelpers())
	builder.WriteString("export function createApp() {\n")
	builder.WriteString("const app = express();\n")
	builder.WriteString("const rateWindowMs = 60_000;\n")
	builder.WriteString("const rateLimit = 120;\n")
	builder.WriteString("const requestCounts = new Map<string, { count: number; resetAt: number }>();\n\n")
	builder.WriteString(g.opsServerState())
	builder.WriteString("app.disable(\"x-powered-by\");\n")
	builder.WriteString("app.use((_req, res, next) => {\n")
	builder.WriteString("  res.setHeader(\"X-Content-Type-Options\", \"nosniff\");\n")
	builder.WriteString("  res.setHeader(\"X-Frame-Options\", \"DENY\");\n")
	builder.WriteString("  res.setHeader(\"Referrer-Policy\", \"no-referrer\");\n")
	builder.WriteString("  res.setHeader(\"Permissions-Policy\", \"geolocation=(), microphone=(), camera=()\");\n")
	builder.WriteString("  next();\n")
	builder.WriteString("});\n")
	if cors := g.corsConfig(); cors != nil && cors.Origins.Name != "" {
		builder.WriteString(fmt.Sprintf("const corsOrigins = (process.env.%s ?? \"\").split(\",\").map((origin) => origin.trim()).filter(Boolean);\n\n", cors.Origins.Name))
		builder.WriteString("app.use((req, res, next) => {\n")
		builder.WriteString("  const origin = req.headers.origin;\n")
		builder.WriteString("  if (!origin) {\n")
		builder.WriteString("    next();\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n")
		builder.WriteString("  if (!corsOrigins.includes(origin)) {\n")
		builder.WriteString("    res.status(403).json({ error: \"CORS origin is not allowed\" });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n")
		builder.WriteString("  res.setHeader(\"Access-Control-Allow-Origin\", origin);\n")
		builder.WriteString("  res.setHeader(\"Vary\", \"Origin\");\n")
		allowedHeaders := "Content-Type, X-CSRF-Token"
		if g.hasOpsObserve() {
			allowedHeaders += ", traceparent"
		}
		builder.WriteString("  res.setHeader(\"Access-Control-Allow-Methods\", \"GET,POST,PUT,PATCH,DELETE,OPTIONS\");\n")
		builder.WriteString(fmt.Sprintf("  res.setHeader(\"Access-Control-Allow-Headers\", %s);\n", contractJSONString(allowedHeaders)))
		if g.hasOpsObserve() {
			builder.WriteString("  res.setHeader(\"Access-Control-Expose-Headers\", \"traceparent\");\n")
		}
		if cors.Credentials == "true" {
			builder.WriteString("  res.setHeader(\"Access-Control-Allow-Credentials\", \"true\");\n")
		}
		builder.WriteString("  if (req.method === \"OPTIONS\") {\n")
		builder.WriteString("    res.status(204).end();\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n")
		builder.WriteString("  next();\n")
		builder.WriteString("});\n")
	}
	builder.WriteString(g.opsServerMiddleware())
	builder.WriteString(g.opsServerRoutes())
	builder.WriteString("app.use((req, res, next) => {\n")
	builder.WriteString("  const now = Date.now();\n")
	builder.WriteString("  const key = req.ip ?? \"unknown\";\n")
	builder.WriteString("  const current = requestCounts.get(key);\n")
	builder.WriteString("  if (!current || current.resetAt <= now) {\n")
	builder.WriteString("    requestCounts.set(key, { count: 1, resetAt: now + rateWindowMs });\n")
	builder.WriteString("    next();\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	builder.WriteString("  if (current.count >= rateLimit) {\n")
	builder.WriteString("    res.status(429).json({ error: \"Too many requests\" });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	builder.WriteString("  current.count += 1;\n")
	builder.WriteString("  next();\n")
	builder.WriteString("});\n")
	if programHasMediaFields(g.program) {
		builder.WriteString("app.use(express.json({ limit: \"2mb\" }));\n")
	} else {
		builder.WriteString("app.use(express.json({ limit: \"100kb\" }));\n")
	}
	builder.WriteString("app.get(\"/openapi.json\", (_req, res) => {\n")
	builder.WriteString("  res.json(openAPISpec);\n")
	builder.WriteString("});\n")
	builder.WriteString(g.explicitAPIRoutes(true))
	if g.program.Auth != nil {
		builder.WriteString("app.use(\"/api/auth\", authRouter);\n")
		builder.WriteString("app.use(\"/api\", requireAuth);\n")
		builder.WriteString("app.use(\"/api\", requireCsrf);\n")
	}
	builder.WriteString(g.explicitAPIRoutes(false))
	for _, page := range g.program.Pages {
		builder.WriteString(fmt.Sprintf("app.use(\"/api\", %s);\n", g.pageRouterName(page)))
	}
	builder.WriteString("\n")
	builder.WriteString("app.use((req, res, next) => {\n")
	builder.WriteString("  if (req.path.startsWith(\"/api\")) {\n")
	builder.WriteString("    res.status(404).json({ error: \"Not found\" });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	builder.WriteString("  next();\n")
	builder.WriteString("});\n")
	if g.isAPIOnlyTarget() {
		builder.WriteString("app.use((_req, res) => {\n")
		builder.WriteString("  res.status(404).json({ error: \"Not found\" });\n")
		builder.WriteString("});\n")
	} else {
		builder.WriteString("app.use(express.static(path.join(process.cwd(), \"dist\")));\n")
		builder.WriteString("app.use((_req, res) => {\n")
		builder.WriteString("  res.sendFile(path.join(process.cwd(), \"dist\", \"index.html\"));\n")
		builder.WriteString("});\n")
	}
	builder.WriteString("return app;\n")
	builder.WriteString("}\n\n")
	builder.WriteString("if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {\n")
	builder.WriteString("  const app = createApp();\n")
	builder.WriteString("  app.listen(port, () => {\n")
	builder.WriteString("  console.log(`BlackLang API server running on http://localhost:${port}`);\n")
	builder.WriteString("  });\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) authRouteTS() string {
	if g.usesPrismaAuthRuntime() {
		return g.postgresAuthRouteTS()
	}
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import crypto from "node:crypto";
import express from "express";
import Database from "better-sqlite3";

const databaseUrl = %s ?? "file:./dev.db";
const filePath = databaseUrl.startsWith("file:") ? databaseUrl.slice(5) : databaseUrl;
const db = new Database(filePath);
const sessionCookie = "black_session";
const csrfCookie = "black_csrf";
const defaultRole = %q;
const defaultRoles = [defaultRole];
const defaultRolesJSON = %q;
const allowedRoles = %s;

type RolePermission = {
  effect: "allow" | "deny";
  action: string;
  resource: string;
  fields: string[];
};

const rolePermissions: Record<string, RolePermission[]> = %s;

type UserRow = {
  id: string;
  name: string;
  email: string;
  role: string;
  roles: string;
%s  passwordHash: string;
};

function hashPassword(password: string, salt = crypto.randomBytes(16).toString("hex")) {
  const hash = crypto.pbkdf2Sync(password, salt, 120_000, 32, "sha256").toString("hex");
  return salt + ":" + hash;
}

function verifyPassword(password: string, stored: string) {
  const [salt, originalHash] = stored.split(":");
  if (!salt || !originalHash) return false;
  const nextHash = hashPassword(password, salt).split(":")[1];
  return crypto.timingSafeEqual(Buffer.from(originalHash, "hex"), Buffer.from(nextHash, "hex"));
}

function readCookie(cookieHeader: string | undefined, name: string) {
  if (!cookieHeader) return "";
  const cookies = cookieHeader.split(";").map((item) => item.trim());
  const cookie = cookies.find((item) => item.startsWith(name + "="));
  return cookie ? decodeURIComponent(cookie.slice(name.length + 1)) : "";
}

function readSessionToken(cookieHeader: string | undefined) {
  return readCookie(cookieHeader, sessionCookie);
}

function setAuthCookies(res: express.Response, token: string, csrfToken: string) {
  res.cookie(sessionCookie, token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  });
  res.cookie(csrfCookie, csrfToken, {
    httpOnly: false,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  });
}

function isKnownRole(role: string) {
  return role === defaultRole || allowedRoles.includes(role);
}

function normalizeRoles(input: unknown) {
  const values = Array.isArray(input) ? input : typeof input === "string" ? [input] : [];
  const normalized = values.map((value) => String(value).trim()).filter((role) => isKnownRole(role));
  return Array.from(new Set(normalized));
}

function storedRoles(user: Pick<UserRow, "role" | "roles">) {
  try {
    const roles = normalizeRoles(JSON.parse(user.roles));
    if (roles.length > 0) return roles;
  } catch {}
  const fallback = normalizeRoles(user.role);
  return fallback.length > 0 ? fallback : defaultRoles;
}

function rolesJSON(roles: string[]) {
  const normalized = normalizeRoles(roles);
  return JSON.stringify(normalized.length > 0 ? normalized : defaultRoles);
}

function primaryRole(roles: string[]) {
  return roles[0] ?? defaultRole;
}

function publicUser(user: UserRow) {
  const roles = storedRoles(user);
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    role: primaryRole(roles),
    roles%s
  };
}

export function requireAuth(req: express.Request, res: express.Response, next: express.NextFunction) {
  const token = readSessionToken(req.headers.cookie);
  if (!token) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  const user = db.prepare("SELECT %s FROM \"BlackSession\" s JOIN \"BlackUser\" u ON u.id = s.userId WHERE s.id = ? AND s.expiresAt > datetime('now')").get(token) as UserRow | undefined;
  if (!user) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  (req as any).blackUser = publicUser(user);
  next();
}

export function requireCsrf(req: express.Request, res: express.Response, next: express.NextFunction) {
  if (["GET", "HEAD", "OPTIONS"].includes(req.method)) {
    next();
    return;
  }

  const headerToken = String(req.headers["x-csrf-token"] ?? "");
  const cookieToken = readCookie(req.headers.cookie, csrfCookie);
  if (headerToken && cookieToken && headerToken === cookieToken) {
    next();
    return;
  }

  res.status(403).json({ error: "Invalid CSRF token" });
}

export function requirePageAccess(allowedRoles: string[]) {
  return (req: express.Request, res: express.Response, next: express.NextFunction) => {
    const user = (req as any).blackUser as ReturnType<typeof publicUser> | undefined;
    if (!user) {
      res.status(401).json({ error: "Not authenticated" });
      return;
    }
    const roles = user.roles ?? [user.role];
    if (allowedRoles.includes("authenticated") || roles.some((role) => allowedRoles.includes(role))) {
      next();
      return;
    }
    res.status(403).json({ error: "Forbidden" });
  };
}

function permissionMatches(permission: RolePermission, action: string, resource: string, field?: string) {
  if (permission.action === "all") return true;
  const actionMatches = permission.action === action || permission.action === "manage";
  if (!actionMatches || permission.resource !== resource) return false;
  if (!field || permission.fields.length === 0) return true;
  return permission.fields.includes(field);
}

function permissionsForRoles(roles: string[]) {
  return roles.flatMap((role) => rolePermissions[role] ?? []);
}

export function canAccessAction(roles: string[], action: string, resource: string) {
  const permissions = permissionsForRoles(roles);
  const denied = permissions.some((permission) => permission.effect === "deny" && permission.fields.length === 0 && permissionMatches(permission, action, resource));
  if (denied) return false;
  return permissions.some((permission) => permission.effect === "allow" && permissionMatches(permission, action, resource));
}

export function canAccessField(roles: string[], action: string, resource: string, field: string) {
  const permissions = permissionsForRoles(roles);
  const denied = permissions.some((permission) => permission.effect === "deny" && permissionMatches(permission, action, resource, field));
  if (denied) return false;
  return permissions.some((permission) => permission.effect === "allow" && permissionMatches(permission, action, resource, field));
}

export function requirePermission(action: string, resource: string) {
  return (req: express.Request, res: express.Response, next: express.NextFunction) => {
    const user = (req as any).blackUser as ReturnType<typeof publicUser> | undefined;
    if (!user) {
      res.status(401).json({ error: "Not authenticated" });
      return;
    }
    if (canAccessAction(user.roles ?? [user.role], action, resource)) {
      next();
      return;
    }
    res.status(403).json({ error: "Forbidden" });
  };
}

export function filterWritableFields(roles: string[], action: string, resource: string, input: Record<string, unknown>) {
  const output: Record<string, unknown> = {};
  for (const [field, value] of Object.entries(input)) {
    if (canAccessField(roles, action, resource, field)) {
      output[field] = value;
    }
  }
  return output;
}

export function writeAuditLog(actor: ReturnType<typeof publicUser> | undefined, action: string, resource: string, resourceId: string, summary = "") {
  if (!actor) return;
  db.prepare("INSERT INTO \"BlackAuditLog\" (id, actorUserId, actorRole, action, resource, resourceId, summary) VALUES (?, ?, ?, ?, ?, ?, ?)").run(
    crypto.randomUUID(),
    actor.id,
    actor.role,
    action,
    resource,
    resourceId,
    summary
  );
}

export function closeAuthDatabase() {
  db.close();
}

export const authRouter = express.Router();

authRouter.post("/register", (req, res) => {
  const name = String(req.body?.name ?? "").trim();
  const email = String(req.body?.email ?? "").trim().toLowerCase();
  const password = String(req.body?.password ?? "");

  if (!name || !email || password.length < 8) {
    res.status(400).json({ error: "Name, email, and an 8 character password are required" });
    return;
  }

  const existing = db.prepare("SELECT id FROM \"BlackUser\" WHERE email = ?").get(email);
  if (existing) {
    res.status(409).json({ error: "Email is already registered" });
    return;
  }

  const user: UserRow = {
    id: crypto.randomUUID(),
    name,
    email,
    role: defaultRole,
    roles: defaultRolesJSON,
    passwordHash: hashPassword(password)%s
  };
  db.prepare("INSERT INTO \"BlackUser\" (%s) VALUES (%s)").run(%s);

  const token = crypto.randomUUID();
  const csrfToken = crypto.randomUUID();
  db.prepare("INSERT INTO \"BlackSession\" (id, userId, expiresAt) VALUES (?, ?, datetime('now', '+7 days'))").run(token, user.id);
  writeAuditLog(publicUser(user), "register", "BlackUser", user.id, "User registered");
  setAuthCookies(res, token, csrfToken);
  res.status(201).json({ user: publicUser(user) });
});

authRouter.post("/login", (req, res) => {
  const email = String(req.body?.email ?? "").trim().toLowerCase();
  const password = String(req.body?.password ?? "");
  const user = db.prepare("SELECT %s FROM \"BlackUser\" WHERE email = ?").get(email) as UserRow | undefined;

  if (!user || !verifyPassword(password, user.passwordHash)) {
    res.status(401).json({ error: "Invalid email or password" });
    return;
  }

  const token = crypto.randomUUID();
  const csrfToken = crypto.randomUUID();
  db.prepare("INSERT INTO \"BlackSession\" (id, userId, expiresAt) VALUES (?, ?, datetime('now', '+7 days'))").run(token, user.id);
  setAuthCookies(res, token, csrfToken);
  res.json({ user: publicUser(user) });
});

authRouter.post("/logout", requireCsrf, (req, res) => {
  const token = readSessionToken(req.headers.cookie);
  if (token) {
    db.prepare("DELETE FROM \"BlackSession\" WHERE id = ?").run(token);
  }
  res.clearCookie(sessionCookie, { path: "/" });
  res.clearCookie(csrfCookie, { path: "/" });
  res.status(204).send();
});

authRouter.get("/me", (req, res) => {
  const token = readSessionToken(req.headers.cookie);
  if (!token) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  const user = db.prepare("SELECT %s FROM \"BlackSession\" s JOIN \"BlackUser\" u ON u.id = s.userId WHERE s.id = ? AND s.expiresAt > datetime('now')").get(token) as UserRow | undefined;

  if (!user) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  res.json({ user: publicUser(user) });
});

authRouter.get("/users", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {
  const users = db.prepare("SELECT %s FROM \"BlackUser\" ORDER BY createdAt DESC").all() as UserRow[];
  res.json({ users: users.map(publicUser) });
});

authRouter.get("/audit", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {
  const logs = db.prepare("SELECT id, actorUserId, actorRole, action, resource, resourceId, summary, createdAt FROM \"BlackAuditLog\" ORDER BY createdAt DESC LIMIT 200").all();
  res.json({ logs });
});

authRouter.put("/users/:id/role", requireAuth, requireCsrf, requirePageAccess([defaultRole]), (req, res) => {
  const roles = normalizeRoles(req.body?.roles);
  if (roles.length === 0) {
    res.status(400).json({ error: "At least one supported role is required" });
    return;
  }
  const role = primaryRole(roles);
  const storedRoles = rolesJSON(roles);

  const result = db.prepare("UPDATE \"BlackUser\" SET role = ?, roles = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?").run(role, storedRoles, String(req.params.id));
  if (result.changes === 0) {
    res.status(404).json({ error: "User not found" });
    return;
  }

  const user = db.prepare("SELECT %s FROM \"BlackUser\" WHERE id = ?").get(String(req.params.id)) as UserRow;
  writeAuditLog((req as any).blackUser, "role.update", "BlackUser", user.id, "Roles changed to " + roles.join(", "));
  res.json({ user: publicUser(user) });
});
%s
`, g.databaseEnvAccess(), g.defaultAuthRole(), g.defaultAuthRolesJSON(), tsStringArrayLiteral(g.roleNames()), g.rolePermissionsLiteral(), g.authUserTenantFieldLine(), g.authPublicUserTenantLine(), g.sqliteAuthUserSelectColumns("u"), g.sqliteAuthUserObjectTenantLine(), g.sqliteAuthUserInsertColumns(), g.sqliteAuthUserInsertPlaceholders(), g.sqliteAuthUserInsertArgs(), g.sqliteAuthUserSelectColumns(""), g.sqliteAuthUserSelectColumns("u"), g.sqliteAuthUserSelectColumns(""), g.sqliteAuthUserSelectColumns(""), g.sqliteAuthTenantUpdateRouteTS())
}

func (g *webGenerator) sqliteAuthTenantUpdateRouteTS() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return fmt.Sprintf(`authRouter.put("/users/:id/tenant", requireAuth, requireCsrf, requirePageAccess([defaultRole]), (req, res) => {
  const tenantId = String(req.body?.tenantId ?? "").trim();
  if (!tenantId || tenantId.length > 64 || !/^[A-Za-z0-9_-]+$/.test(tenantId)) {
    res.status(400).json({ error: "A tenant id with letters, numbers, underscores, or dashes is required" });
    return;
  }

  const result = db.prepare("UPDATE \"BlackUser\" SET tenantId = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?").run(tenantId, String(req.params.id));
  if (result.changes === 0) {
    res.status(404).json({ error: "User not found" });
    return;
  }

  const user = db.prepare("SELECT %s FROM \"BlackUser\" WHERE id = ?").get(String(req.params.id)) as UserRow;
  writeAuditLog((req as any).blackUser, "tenant.update", "BlackUser", user.id, "Tenant changed to " + tenantId);
  res.json({ user: publicUser(user) });
});
`, g.sqliteAuthUserSelectColumns(""))
}

func (g *webGenerator) postgresAuthRouteTS() string {
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import crypto from "node:crypto";
import express from "express";
import { prisma } from "../db";

const sessionCookie = "black_session";
const csrfCookie = "black_csrf";
const defaultRole = %q;
const defaultRoles = [defaultRole];
const defaultRolesJSON = %q;
const allowedRoles = %s;

type RolePermission = {
  effect: "allow" | "deny";
  action: string;
  resource: string;
  fields: string[];
};

const rolePermissions: Record<string, RolePermission[]> = %s;

type UserRow = {
  id: string;
  name: string;
  email: string;
  role: string;
  roles: string;
%s  passwordHash: string;
};

function hashPassword(password: string, salt = crypto.randomBytes(16).toString("hex")) {
  const hash = crypto.pbkdf2Sync(password, salt, 120_000, 32, "sha256").toString("hex");
  return salt + ":" + hash;
}

function verifyPassword(password: string, stored: string) {
  const [salt, originalHash] = stored.split(":");
  if (!salt || !originalHash) return false;
  const nextHash = hashPassword(password, salt).split(":")[1];
  return crypto.timingSafeEqual(Buffer.from(originalHash, "hex"), Buffer.from(nextHash, "hex"));
}

function readCookie(cookieHeader: string | undefined, name: string) {
  if (!cookieHeader) return "";
  const cookies = cookieHeader.split(";").map((item) => item.trim());
  const cookie = cookies.find((item) => item.startsWith(name + "="));
  return cookie ? decodeURIComponent(cookie.slice(name.length + 1)) : "";
}

function readSessionToken(cookieHeader: string | undefined) {
  return readCookie(cookieHeader, sessionCookie);
}

function setAuthCookies(res: express.Response, token: string, csrfToken: string) {
  res.cookie(sessionCookie, token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  });
  res.cookie(csrfCookie, csrfToken, {
    httpOnly: false,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  });
}

function isKnownRole(role: string) {
  return role === defaultRole || allowedRoles.includes(role);
}

function normalizeRoles(input: unknown) {
  const values = Array.isArray(input) ? input : typeof input === "string" ? [input] : [];
  const normalized = values.map((value) => String(value).trim()).filter((role) => isKnownRole(role));
  return Array.from(new Set(normalized));
}

function storedRoles(user: Pick<UserRow, "role" | "roles">) {
  try {
    const roles = normalizeRoles(JSON.parse(user.roles));
    if (roles.length > 0) return roles;
  } catch {}
  const fallback = normalizeRoles(user.role);
  return fallback.length > 0 ? fallback : defaultRoles;
}

function rolesJSON(roles: string[]) {
  const normalized = normalizeRoles(roles);
  return JSON.stringify(normalized.length > 0 ? normalized : defaultRoles);
}

function primaryRole(roles: string[]) {
  return roles[0] ?? defaultRole;
}

function publicUser(user: UserRow) {
  const roles = storedRoles(user);
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    role: primaryRole(roles),
    roles%s
  };
}

function sessionExpiry() {
  return new Date(Date.now() + 7 * 24 * 60 * 60 * 1000);
}

function isUniqueConstraintError(error: unknown) {
  return typeof error === "object" && error !== null && "code" in error && (error as { code?: string }).code === "P2002";
}

async function userFromSession(token: string) {
  const session = await prisma.blackSession.findUnique({
    where: { id: token },
    include: { user: true }
  });
  if (!session || session.expiresAt <= new Date()) return undefined;
  return session.user as UserRow;
}

export async function requireAuth(req: express.Request, res: express.Response, next: express.NextFunction) {
  const token = readSessionToken(req.headers.cookie);
  if (!token) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  try {
    const user = await userFromSession(token);
    if (!user) {
      res.status(401).json({ error: "Not authenticated" });
      return;
    }
    (req as any).blackUser = publicUser(user);
    next();
  } catch {
    res.status(500).json({ error: "Authentication lookup failed" });
  }
}

export function requireCsrf(req: express.Request, res: express.Response, next: express.NextFunction) {
  if (["GET", "HEAD", "OPTIONS"].includes(req.method)) {
    next();
    return;
  }

  const headerToken = String(req.headers["x-csrf-token"] ?? "");
  const cookieToken = readCookie(req.headers.cookie, csrfCookie);
  if (headerToken && cookieToken && headerToken === cookieToken) {
    next();
    return;
  }

  res.status(403).json({ error: "Invalid CSRF token" });
}

export function requirePageAccess(allowedRoles: string[]) {
  return (req: express.Request, res: express.Response, next: express.NextFunction) => {
    const user = (req as any).blackUser as ReturnType<typeof publicUser> | undefined;
    if (!user) {
      res.status(401).json({ error: "Not authenticated" });
      return;
    }
    const roles = user.roles ?? [user.role];
    if (allowedRoles.includes("authenticated") || roles.some((role) => allowedRoles.includes(role))) {
      next();
      return;
    }
    res.status(403).json({ error: "Forbidden" });
  };
}

function permissionMatches(permission: RolePermission, action: string, resource: string, field?: string) {
  if (permission.action === "all") return true;
  const actionMatches = permission.action === action || permission.action === "manage";
  if (!actionMatches || permission.resource !== resource) return false;
  if (!field || permission.fields.length === 0) return true;
  return permission.fields.includes(field);
}

function permissionsForRoles(roles: string[]) {
  return roles.flatMap((role) => rolePermissions[role] ?? []);
}

export function canAccessAction(roles: string[], action: string, resource: string) {
  const permissions = permissionsForRoles(roles);
  const denied = permissions.some((permission) => permission.effect === "deny" && permission.fields.length === 0 && permissionMatches(permission, action, resource));
  if (denied) return false;
  return permissions.some((permission) => permission.effect === "allow" && permissionMatches(permission, action, resource));
}

export function canAccessField(roles: string[], action: string, resource: string, field: string) {
  const permissions = permissionsForRoles(roles);
  const denied = permissions.some((permission) => permission.effect === "deny" && permissionMatches(permission, action, resource, field));
  if (denied) return false;
  return permissions.some((permission) => permission.effect === "allow" && permissionMatches(permission, action, resource, field));
}

export function requirePermission(action: string, resource: string) {
  return (req: express.Request, res: express.Response, next: express.NextFunction) => {
    const user = (req as any).blackUser as ReturnType<typeof publicUser> | undefined;
    if (!user) {
      res.status(401).json({ error: "Not authenticated" });
      return;
    }
    if (canAccessAction(user.roles ?? [user.role], action, resource)) {
      next();
      return;
    }
    res.status(403).json({ error: "Forbidden" });
  };
}

export function filterWritableFields(roles: string[], action: string, resource: string, input: Record<string, unknown>) {
  const output: Record<string, unknown> = {};
  for (const [field, value] of Object.entries(input)) {
    if (canAccessField(roles, action, resource, field)) {
      output[field] = value;
    }
  }
  return output;
}

export function writeAuditLog(actor: ReturnType<typeof publicUser> | undefined, action: string, resource: string, resourceId: string, summary = "") {
  if (!actor) return;
  void prisma.blackAuditLog.create({
    data: {
      id: crypto.randomUUID(),
      actorUserId: actor.id,
      actorRole: actor.role,
      action,
      resource,
      resourceId,
      summary
    }
  }).catch(() => {});
}

export async function closeAuthDatabase() {
  await prisma.$disconnect();
}

export const authRouter = express.Router();

authRouter.post("/register", async (req, res) => {
  const name = String(req.body?.name ?? "").trim();
  const email = String(req.body?.email ?? "").trim().toLowerCase();
  const password = String(req.body?.password ?? "");

  if (!name || !email || password.length < 8) {
    res.status(400).json({ error: "Name, email, and an 8 character password are required" });
    return;
  }

  try {
    const existing = await prisma.blackUser.findUnique({ where: { email } });
    if (existing) {
      res.status(409).json({ error: "Email is already registered" });
      return;
    }

    const user = await prisma.blackUser.create({
      data: {
        id: crypto.randomUUID(),
        name,
        email,
        role: defaultRole,
        roles: defaultRolesJSON,
        passwordHash: hashPassword(password)%s
      }
    }) as UserRow;

    const token = crypto.randomUUID();
    const csrfToken = crypto.randomUUID();
    await prisma.blackSession.create({ data: { id: token, userId: user.id, expiresAt: sessionExpiry() } });
    writeAuditLog(publicUser(user), "register", "BlackUser", user.id, "User registered");
    setAuthCookies(res, token, csrfToken);
    res.status(201).json({ user: publicUser(user) });
  } catch (error) {
    if (isUniqueConstraintError(error)) {
      res.status(409).json({ error: "Email is already registered" });
      return;
    }
    res.status(500).json({ error: "Registration failed" });
  }
});

authRouter.post("/login", async (req, res) => {
  const email = String(req.body?.email ?? "").trim().toLowerCase();
  const password = String(req.body?.password ?? "");

  const user = await prisma.blackUser.findUnique({ where: { email } }) as UserRow | null;
  if (!user || !verifyPassword(password, user.passwordHash)) {
    res.status(401).json({ error: "Invalid email or password" });
    return;
  }

  const token = crypto.randomUUID();
  const csrfToken = crypto.randomUUID();
  await prisma.blackSession.create({ data: { id: token, userId: user.id, expiresAt: sessionExpiry() } });
  setAuthCookies(res, token, csrfToken);
  res.json({ user: publicUser(user) });
});

authRouter.post("/logout", requireCsrf, async (req, res) => {
  const token = readSessionToken(req.headers.cookie);
  if (token) {
    await prisma.blackSession.deleteMany({ where: { id: token } });
  }
  res.clearCookie(sessionCookie, { path: "/" });
  res.clearCookie(csrfCookie, { path: "/" });
  res.status(204).send();
});

authRouter.get("/me", async (req, res) => {
  const token = readSessionToken(req.headers.cookie);
  if (!token) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  const user = await userFromSession(token);
  if (!user) {
    res.status(401).json({ error: "Not authenticated" });
    return;
  }

  res.json({ user: publicUser(user) });
});

authRouter.get("/users", requireAuth, requirePageAccess([defaultRole]), async (_req, res) => {
  const users = await prisma.blackUser.findMany({ orderBy: { createdAt: "desc" } });
  res.json({ users: users.map(publicUser) });
});

authRouter.get("/audit", requireAuth, requirePageAccess([defaultRole]), async (_req, res) => {
  const logs = await prisma.blackAuditLog.findMany({ orderBy: { createdAt: "desc" }, take: 200 });
  res.json({ logs });
});

authRouter.put("/users/:id/role", requireAuth, requireCsrf, requirePageAccess([defaultRole]), async (req, res) => {
  const roles = normalizeRoles(req.body?.roles);
  if (roles.length === 0) {
    res.status(400).json({ error: "At least one supported role is required" });
    return;
  }
  const role = primaryRole(roles);
  const storedRoles = rolesJSON(roles);

  try {
    const user = await prisma.blackUser.update({
      where: { id: String(req.params.id) },
      data: { role, roles: storedRoles }
    });
    writeAuditLog((req as any).blackUser, "role.update", "BlackUser", user.id, "Roles changed to " + roles.join(", "));
    res.json({ user: publicUser(user) });
  } catch {
    res.status(404).json({ error: "User not found" });
  }
});
%s
`, g.defaultAuthRole(), g.defaultAuthRolesJSON(), tsStringArrayLiteral(g.roleNames()), g.rolePermissionsLiteral(), g.authUserTenantFieldLine(), g.authPublicUserTenantLine(), g.postgresAuthUserCreateTenantLine(), g.postgresAuthTenantUpdateRouteTS())
}

func (g *webGenerator) postgresAuthTenantUpdateRouteTS() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return `authRouter.put("/users/:id/tenant", requireAuth, requireCsrf, requirePageAccess([defaultRole]), async (req, res) => {
  const tenantId = String(req.body?.tenantId ?? "").trim();
  if (!tenantId || tenantId.length > 64 || !/^[A-Za-z0-9_-]+$/.test(tenantId)) {
    res.status(400).json({ error: "A tenant id with letters, numbers, underscores, or dashes is required" });
    return;
  }

  try {
    const user = await prisma.blackUser.update({
      where: { id: String(req.params.id) },
      data: { tenantId }
    }) as UserRow;
    writeAuditLog((req as any).blackUser, "tenant.update", "BlackUser", user.id, "Tenant changed to " + tenantId);
    res.json({ user: publicUser(user) });
  } catch {
    res.status(404).json({ error: "User not found" });
  }
});
`
}

func (g *webGenerator) dbTS() string {
	if g.isPostgres() {
		return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { PrismaPg } from "@prisma/adapter-pg";
import { PrismaClient } from "./generated/prisma/client";

const adapter = new PrismaPg({
  connectionString: %s ?? %q
});

export const prisma = new PrismaClient({ adapter });
`, g.databaseEnvAccess(), g.defaultDatabaseURL())
	}
	if g.isMySQL() {
		return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { PrismaMariaDb } from "@prisma/adapter-mariadb";
import { PrismaClient } from "./generated/prisma/client";

const adapter = new PrismaMariaDb(%s ?? %q);

export const prisma = new PrismaClient({ adapter });
`, g.databaseEnvAccess(), g.defaultDatabaseURL())
	}
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { PrismaBetterSqlite3 } from "@prisma/adapter-better-sqlite3";
import { PrismaClient } from "./generated/prisma/client";

const adapter = new PrismaBetterSqlite3({
  url: %s ?? "file:./dev.db"
});

export const prisma = new PrismaClient({ adapter });
`, g.databaseEnvAccess())
}

func (g *webGenerator) setupDBTS() string {
	if g.isPostgres() {
		return g.postgresSetupDBTS()
	}
	if g.isMySQL() {
		return g.mysqlSetupDBTS()
	}
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("import \"dotenv/config\";\n")
	builder.WriteString("import Database from \"better-sqlite3\";\n\n")
	builder.WriteString(fmt.Sprintf("const databaseUrl = %s ?? \"file:./dev.db\";\n", g.databaseEnvAccess()))
	builder.WriteString("const filePath = databaseUrl.startsWith(\"file:\") ? databaseUrl.slice(5) : databaseUrl;\n")
	builder.WriteString("const db = new Database(filePath);\n\n")
	builder.WriteString(g.sqliteMigrationRuntimeTS())
	if g.program.Auth != nil {
		builder.WriteString("db.exec(`\nCREATE TABLE IF NOT EXISTS \"BlackUser\" (\n")
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		builder.WriteString("  \"name\" TEXT NOT NULL,\n")
		builder.WriteString("  \"email\" TEXT NOT NULL UNIQUE,\n")
		builder.WriteString(fmt.Sprintf("  \"role\" TEXT NOT NULL DEFAULT %q,\n", g.defaultAuthRole()))
		builder.WriteString(g.sqliteAuthRolesColumnLine())
		builder.WriteString(g.sqliteAuthTenantColumnLine())
		builder.WriteString("  \"passwordHash\" TEXT NOT NULL,\n")
		builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
		builder.WriteString("  \"updatedAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		builder.WriteString(");\n`);\n\n")
		builder.WriteString("const blackUserColumns = db.prepare(`PRAGMA table_info(\"BlackUser\")`).all() as Array<{ name: string }>;\n")
		builder.WriteString("if (!blackUserColumns.some((column) => column.name === \"role\")) {\n")
		builder.WriteString(fmt.Sprintf("  db.prepare(`ALTER TABLE \"BlackUser\" ADD COLUMN \"role\" TEXT NOT NULL DEFAULT %q`).run();\n", g.defaultAuthRole()))
		builder.WriteString("}\n\n")
		builder.WriteString(g.sqliteAuthRolesMigration())
		builder.WriteString(g.sqliteAuthTenantMigration())
		builder.WriteString("db.exec(`\nCREATE TABLE IF NOT EXISTS \"BlackSession\" (\n")
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		builder.WriteString("  \"userId\" TEXT NOT NULL REFERENCES \"BlackUser\"(\"id\"),\n")
		builder.WriteString("  \"expiresAt\" DATETIME NOT NULL,\n")
		builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		builder.WriteString(");\n`);\n\n")
		builder.WriteString("db.exec(`\nCREATE TABLE IF NOT EXISTS \"BlackAuditLog\" (\n")
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		builder.WriteString("  \"actorUserId\" TEXT NOT NULL,\n")
		builder.WriteString("  \"actorRole\" TEXT NOT NULL,\n")
		builder.WriteString("  \"action\" TEXT NOT NULL,\n")
		builder.WriteString("  \"resource\" TEXT NOT NULL,\n")
		builder.WriteString("  \"resourceId\" TEXT NOT NULL,\n")
		builder.WriteString("  \"summary\" TEXT NOT NULL DEFAULT '',\n")
		builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		builder.WriteString(");\n`);\n\n")
	}
	for _, entity := range g.program.Entities {
		builder.WriteString(fmt.Sprintf("db.exec(`\nCREATE TABLE IF NOT EXISTS \"%s\" (\n", entity.Name))
		builder.WriteString("  \"id\" TEXT NOT NULL PRIMARY KEY,\n")
		for index, field := range entity.Fields {
			line := fmt.Sprintf("  \"%s\" %s%s", g.sqliteColumnName(field), g.sqliteFieldType(field), sqliteRequired(field))
			if defaultValue := modifierValue(field, "default"); defaultValue != "" {
				line += fmt.Sprintf(" DEFAULT %s", sqliteDefaultValue(field, defaultValue))
			}
			if hasModifier(field, "unique") {
				line += " UNIQUE"
			}
			if g.isRelationField(field) {
				line += fmt.Sprintf(" REFERENCES \"%s\"(\"id\")", field.Type)
			}
			line += ","
			builder.WriteString(line + "\n")
			if index == len(entity.Fields)-1 {
				builder.WriteString("  \"archivedAt\" DATETIME,\n")
				builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
				builder.WriteString("  \"updatedAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
			}
		}
		if len(entity.Fields) == 0 {
			builder.WriteString("  \"archivedAt\" DATETIME,\n")
			builder.WriteString("  \"createdAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,\n")
			builder.WriteString("  \"updatedAt\" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n")
		}
		builder.WriteString(");\n`);\n\n")
		for _, index := range entity.Indexes {
			line := g.sqliteIndexStatement(entity, index)
			if line == "" {
				continue
			}
			builder.WriteString("db.exec(`\n")
			builder.WriteString(line)
			builder.WriteString("\n`);\n\n")
		}
	}
	builder.WriteString("db.close();\n")
	builder.WriteString("console.log(`SQLite database ready at ${filePath}`);\n")
	return builder.String()
}

func (g *webGenerator) mysqlSetupDBTS() string {
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { spawnSync } from "node:child_process";

process.env.%s = process.env.%s ?? %q;

const command = process.platform === "win32" ? "npm.cmd" : "npm";
const result = spawnSync(command, ["run", "db:push:native"], {
  stdio: "inherit"
});

if (result.error) {
  console.error(result.error);
  process.exit(1);
}

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

console.log("MySQL database schema ready");
`, g.databaseEnvName(), g.databaseEnvName(), g.defaultDatabaseURL())
}

func (g *webGenerator) postgresSetupDBTS() string {
	if len(g.program.Migrations) > 0 {
		return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { spawnSync } from "node:child_process";
import pg from "pg";

process.env.%s = process.env.%s ?? %q;
const databaseUrl = process.env.%s ?? %q;

%sconst command = process.platform === "win32" ? "npm.cmd" : "npm";

async function main(): Promise<void> {
  await applyBlackMigrations();

  const result = spawnSync(command, ["run", "db:push:native"], {
    stdio: "inherit"
  });

  if (result.error) {
    console.error(result.error);
    process.exit(1);
  }

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }

  console.log("PostgreSQL database schema ready");
}

void main().catch((error) => {
  console.error(error);
  process.exit(1);
});
`, g.databaseEnvName(), g.databaseEnvName(), g.defaultDatabaseURL(), g.databaseEnvName(), g.defaultDatabaseURL(), g.postgresMigrationRuntimeTS())
	}
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { spawnSync } from "node:child_process";

process.env.%s = process.env.%s ?? %q;

const command = process.platform === "win32" ? "npm.cmd" : "npm";
const result = spawnSync(command, ["run", "db:push:native"], {
  stdio: "inherit"
});

if (result.error) {
  console.error(result.error);
  process.exit(1);
}

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

console.log("PostgreSQL database schema ready");
`, g.databaseEnvName(), g.databaseEnvName(), g.defaultDatabaseURL())
}

func (g *webGenerator) stylesCSS() string {
	base := `:root {
  color: #172026;
  background: #f6f8fa;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

body {
  margin: 0;
}

.app-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  min-height: 100vh;
}

.app-sidebar {
  background: #172026;
  color: #ffffff;
  padding: 20px;
}

.nav-backdrop {
  display: none;
}

.app-brand {
  font-size: 18px;
  font-weight: 750;
  margin-bottom: 20px;
}

.app-nav {
  display: grid;
  gap: 8px;
}

.app-nav button {
  justify-content: flex-start;
  text-align: left;
  width: 100%;
}

.app-nav button.active {
  background: #ffffff;
  border-color: #ffffff;
  color: #172026;
}

.app-nav button.secondary {
  background: transparent;
  border-color: #3d4b59;
  color: #d8dee4;
}

.app-workspace {
  min-width: 0;
}

.app-topbar {
  align-items: center;
  background: #ffffff;
  border-bottom: 1px solid #d8dee4;
  display: flex;
  justify-content: space-between;
  min-height: 72px;
  padding: 16px 24px;
}

.menu-button {
  display: none;
}

.breadcrumb {
  color: #57606a;
  display: block;
  font-size: 13px;
  font-weight: 650;
  margin-bottom: 4px;
}

.locale-switcher {
  align-items: center;
  color: #57606a;
  display: flex;
  font-size: 13px;
  gap: 8px;
}

.locale-switcher select {
  min-width: 84px;
}

.app-content main {
  max-width: 1080px;
  margin: 0 auto;
  padding: 24px 20px 32px;
}

.page-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-view .panel {
  margin-bottom: 0;
}

.view-group {
  min-width: 0;
}

.view-group-title {
  color: #17202a;
  font-size: 18px;
  margin: 0;
}

main > header {
  display: none;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

h1,
h2 {
  margin: 0;
}

.panel {
  background: #ffffff;
  border: 1px solid #d8dee4;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.panel-header {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-overlay {
  background: rgba(23, 32, 38, 0.42);
  bottom: 0;
  display: flex;
  left: 0;
  overflow: auto;
  padding: 24px;
  position: fixed;
  right: 0;
  top: 0;
  z-index: 40;
}

.section-overlay .panel {
  box-shadow: 0 24px 60px rgba(23, 32, 38, 0.18);
  margin-bottom: 0;
  max-height: calc(100vh - 48px);
  overflow: auto;
}

.section-overlay-modal {
  align-items: center;
  justify-content: center;
}

.section-overlay-modal .panel {
  width: min(720px, 100%);
}

.section-overlay-drawer {
  align-items: stretch;
}

.section-overlay-drawer-right {
  justify-content: flex-end;
}

.section-overlay-drawer-left {
  justify-content: flex-start;
}

.section-overlay-drawer .panel {
  min-height: calc(100vh - 48px);
  width: min(560px, 100%);
}

.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.query-summary {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  margin-bottom: 16px;
}

.query-summary-card {
  background: #f6f8fa;
  border: 1px solid #d8dee4;
  border-radius: 8px;
  padding: 12px;
}

.query-summary-card span {
  color: #57606a;
  display: block;
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 6px;
  text-transform: uppercase;
}

.query-summary-card strong {
  color: #172026;
  font-size: 24px;
  line-height: 1.2;
}

.view-tabs {
  border-bottom: 1px solid #d8dee4;
  display: flex;
  gap: 8px;
  margin-bottom: 0;
}

.view-tabs button {
  background: transparent;
  border-bottom-color: transparent;
  border-radius: 6px 6px 0 0;
}

.view-tabs button.active {
  background: #ffffff;
  border-color: #d8dee4;
  border-bottom-color: #ffffff;
  color: #172026;
}

.view-tab-panel {
  display: contents;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

label {
  display: grid;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
}

input,
select {
  border: 1px solid #c9d1d9;
  border-radius: 6px;
  font: inherit;
  padding: 9px 10px;
}

input[type="checkbox"] {
  height: 16px;
  width: 16px;
}

.role-list {
  display: grid;
  gap: 6px;
}

.role-check {
  align-items: center;
  display: inline-flex;
  gap: 6px;
  grid-template-columns: none;
  font-weight: 500;
}

.tenant-editor {
  display: grid;
  gap: 4px;
  min-width: 160px;
}

.inline-control {
  align-items: center;
  display: inline-flex;
  gap: 6px;
  grid-template-columns: none;
}

.field-note {
  color: #57606a;
  font-size: 12px;
  font-weight: 500;
}

.field-error {
  color: #a40e26;
  font-size: 12px;
  font-weight: 650;
}

.field-preview {
  display: flex;
  margin-top: 6px;
}

.media-thumb {
  border: 1px solid #d8dee4;
  border-radius: 8px;
  height: 56px;
  object-fit: cover;
  width: 56px;
}

.media-link {
  color: #0969da;
  font-size: 13px;
  font-weight: 650;
}

.relation-status {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
}

.pagination {
  align-items: center;
  justify-content: flex-end;
}

.column-controls {
  align-items: center;
  flex-wrap: wrap;
}

.filter-controls {
  align-items: end;
  flex-wrap: wrap;
}

.filter-controls label {
  min-width: 180px;
}

.auth-screen {
  align-items: center;
  background: #f6f8fa;
  display: flex;
  min-height: 100vh;
  padding: 24px;
}

.auth-panel {
  background: #ffffff;
  border: 1px solid #d8dee4;
  border-radius: 8px;
  box-shadow: 0 16px 36px rgba(23, 32, 38, 0.08);
  display: grid;
  gap: 18px;
  margin: 0 auto;
  max-width: 420px;
  padding: 24px;
  width: 100%;
}

.auth-panel form {
  display: grid;
  gap: 14px;
}

.auth-tabs {
  display: grid;
  gap: 8px;
  grid-template-columns: 1fr 1fr;
}

.user-menu {
  align-items: center;
  display: flex;
  gap: 10px;
}

.role-badge {
  background: #eef2f6;
  border: 1px solid #d8dee4;
  border-radius: 999px;
  color: #172026;
  font-size: 13px;
  font-weight: 650;
  padding: 6px 10px;
}

.eyebrow {
  color: #57606a;
  display: block;
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 6px;
  text-transform: uppercase;
}

button {
  border: 1px solid #1f6feb;
  border-radius: 6px;
  background: #1f6feb;
  color: #ffffff;
  cursor: pointer;
  font: inherit;
  font-weight: 650;
  padding: 9px 12px;
}

button.secondary {
  border-color: #c9d1d9;
  background: #ffffff;
  color: #172026;
}

button.danger {
  border-color: #cf222e;
  background: #cf222e;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.68;
}

.status,
.error {
  border-radius: 6px;
  margin-bottom: 12px;
  padding: 10px 12px;
}

.status {
  background: #e7f5ff;
  color: #0b4678;
}

.error {
  background: #fff1f3;
  color: #a40e26;
}

.muted {
  color: #57606a;
}

.detail-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  margin: 16px 0 0;
}

.detail-grid div {
  border: 1px solid #d8dee4;
  border-radius: 6px;
  padding: 10px;
}

.detail-grid dt {
  color: #57606a;
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 4px;
  text-transform: uppercase;
}

.detail-grid dd {
  margin: 0;
  overflow-wrap: anywhere;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  border-bottom: 1px solid #d8dee4;
  padding: 10px;
  text-align: left;
}

th {
  color: #57606a;
  font-size: 13px;
}

.select-cell {
  width: 40px;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.component-section-body {
  align-items: center;
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.component-section-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.component-section-list-item {
  display: inline-flex;
}

.black-component {
  border: 1px solid #d8dee4;
  border-radius: 999px;
  display: inline-flex;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  min-width: 34px;
  padding: 5px 8px;
}

.black-component-stock-badge-low {
  background: #fff1f3;
  border-color: #ffccd5;
  color: #a40e26;
}

.black-component-stock-badge-normal {
  background: #e6fcf5;
  border-color: #b7ebd8;
  color: #087f5b;
}

@media (max-width: 760px) {
  .app-shell {
    display: block;
  }

  .app-sidebar {
    bottom: 0;
    left: 0;
    max-width: 280px;
    position: fixed;
    top: 0;
    transform: translateX(-100%);
    transition: transform 160ms ease;
    width: 78vw;
    z-index: 20;
  }

  .app-shell.nav-open .app-sidebar {
    transform: translateX(0);
  }

  .nav-backdrop {
    background: rgba(23, 32, 38, 0.44);
    border: 0;
    bottom: 0;
    display: block;
    left: 0;
    padding: 0;
    position: fixed;
    right: 0;
    top: 0;
    z-index: 10;
  }

  .menu-button {
    display: inline-flex;
  }

  .app-topbar {
    gap: 12px;
    justify-content: flex-start;
    padding: 14px 16px;
  }

  .app-content main {
    padding: 16px;
  }
}
`
	blocks := []string{strings.TrimRight(base, "\n")}
	if pageView := g.pageViewCSS(); pageView != "" {
		blocks = append(blocks, pageView)
	}
	if inlineUI := g.inlineUICSS(); inlineUI != "" {
		blocks = append(blocks, inlineUI)
	}
	return strings.Join(blocks, "\n\n") + "\n"
}

type uiCSSRule struct {
	selector     string
	declarations []string
}

func (g *webGenerator) pageViewCSS() string {
	rules := []uiCSSRule{}
	mediaBlocks := []string{}
	for _, page := range g.program.Pages {
		if page.View == nil {
			continue
		}
		if len(page.View.Order) > 0 {
			for index, section := range pageViewOrder(page) {
				if !pageViewSectionSupported(page, section) {
					continue
				}
				rules = append(rules, uiCSSRule{
					selector:     fmt.Sprintf(".page-view-%s .bl-view-section-%s", kebabCase(page.Name), kebabCase(section)),
					declarations: []string{fmt.Sprintf("order: %d;", index+1)},
				})
			}
		}
		for _, group := range page.View.Groups {
			if order, ok := viewGroupOrder(page, group); ok {
				rules = append(rules, uiCSSRule{
					selector:     fmt.Sprintf(".page-view-%s .bl-view-group-%s", kebabCase(page.Name), kebabCase(group.Name)),
					declarations: []string{fmt.Sprintf("order: %d;", order)},
				})
			}
			rules = append(rules, pageViewGroupRule(page, group))
			if group.Span > 0 {
				rules = append(rules, uiCSSRule{
					selector:     fmt.Sprintf(".page-view-%s .bl-view-group-%s", kebabCase(page.Name), kebabCase(group.Name)),
					declarations: []string{fmt.Sprintf("grid-column: span %d;", group.Span)},
				})
			}
			if mediaBlock := pageViewGroupMediaBlock(page, group); mediaBlock != "" {
				mediaBlocks = append(mediaBlocks, mediaBlock)
			}
		}
		if page.View.Compose != nil {
			rules = append(rules, pageViewComposeRule(page))
			mediaBlocks = append(mediaBlocks, pageViewComposeMediaBlocks(page)...)
		}
		for _, section := range page.View.Sections {
			if !pageViewSectionSupported(page, section.Name) || section.Span <= 0 {
				continue
			}
			rules = append(rules, uiCSSRule{
				selector:     fmt.Sprintf(".page-view-%s .bl-view-section-%s", kebabCase(page.Name), kebabCase(section.Name)),
				declarations: []string{fmt.Sprintf("grid-column: span %d;", section.Span)},
			})
		}
	}

	if len(rules) == 0 && len(mediaBlocks) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("/* Generated from BlackLang page view order and composition. */\n")
	for _, rule := range rules {
		builder.WriteString(rule.selector)
		builder.WriteString(" {\n")
		for _, declaration := range rule.declarations {
			builder.WriteString("  ")
			builder.WriteString(declaration)
			builder.WriteString("\n")
		}
		builder.WriteString("}\n\n")
	}
	for _, mediaBlock := range mediaBlocks {
		builder.WriteString(mediaBlock)
		builder.WriteString("\n\n")
	}
	return strings.TrimRight(builder.String(), "\n")
}

func pageViewComposeRule(page PageDecl) uiCSSRule {
	compose := page.View.Compose
	selector := fmt.Sprintf(".page-view-%s.bl-view-compose-%s", kebabCase(page.Name), compose.Mode)
	declarations := []string{fmt.Sprintf("gap: %s;", viewGapValue(compose.Gap))}
	if compose.Mode == "grid" {
		declarations = append([]string{"display: grid;", fmt.Sprintf("grid-template-columns: repeat(%d, minmax(0, 1fr));", viewComposeColumns(*compose)), "align-items: start;"}, declarations...)
	}
	return uiCSSRule{selector: selector, declarations: declarations}
}

func pageViewComposeMediaBlocks(page PageDecl) []string {
	compose := page.View.Compose
	if compose.Mode != "grid" {
		return nil
	}
	breakpoint := compose.StackAt
	if breakpoint == "" {
		breakpoint = "md"
	}
	if breakpoint == "none" {
		return nil
	}

	columns := viewComposeColumns(*compose)
	selector := fmt.Sprintf(".page-view-%s.bl-view-compose-grid", kebabCase(page.Name))
	blocks := []string{}
	currentColumns := columns
	for _, step := range viewResponsiveSteps(columns, breakpoint) {
		if step.Columns == currentColumns && !step.Stack {
			continue
		}
		maxWidth := viewBreakpointMaxWidth(step.Name)
		if maxWidth == "" {
			continue
		}
		blocks = append(blocks, pageViewComposeMediaBlock(page, selector, maxWidth, step.Columns, step.Stack))
		currentColumns = step.Columns
	}
	return blocks
}

type viewResponsiveStep struct {
	Name    string
	Columns int
	Stack   bool
}

func viewResponsiveSteps(columns int, stackAt string) []viewResponsiveStep {
	names := viewBreakpointNamesThrough(stackAt)
	steps := []viewResponsiveStep{}
	for _, name := range names {
		stack := name == stackAt
		targetColumns := viewBreakpointColumns(name, columns)
		if stack {
			targetColumns = 1
		}
		steps = append(steps, viewResponsiveStep{Name: name, Columns: targetColumns, Stack: stack})
	}
	return steps
}

func viewBreakpointNamesThrough(stackAt string) []string {
	switch stackAt {
	case "lg":
		return []string{"lg"}
	case "sm":
		return []string{"lg", "md", "sm"}
	default:
		return []string{"lg", "md"}
	}
}

func viewBreakpointColumns(name string, columns int) int {
	switch name {
	case "lg":
		if columns > 3 {
			return 3
		}
	case "md":
		if columns > 2 {
			return 2
		}
	case "sm":
		return 1
	}
	return columns
}

func pageViewComposeMediaBlock(page PageDecl, selector string, maxWidth string, columns int, stack bool) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("@media (max-width: %s) {\n", maxWidth))
	builder.WriteString(fmt.Sprintf("  %s {\n", selector))
	if columns == 1 {
		builder.WriteString("    grid-template-columns: 1fr;\n")
	} else {
		builder.WriteString(fmt.Sprintf("    grid-template-columns: repeat(%d, minmax(0, 1fr));\n", columns))
	}
	builder.WriteString("  }\n")
	if stack {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("  %s .panel {\n", selector))
		builder.WriteString("    grid-column: 1 / -1;\n")
		builder.WriteString("  }\n")
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("  %s .view-group {\n", selector))
		builder.WriteString("    grid-column: 1 / -1;\n")
		builder.WriteString("  }\n")
	} else {
		for _, section := range page.View.Sections {
			if section.Span <= columns || !pageViewSectionSupported(page, section.Name) {
				continue
			}
			builder.WriteString("\n")
			builder.WriteString(fmt.Sprintf("  .page-view-%s .bl-view-section-%s {\n", kebabCase(page.Name), kebabCase(section.Name)))
			builder.WriteString(fmt.Sprintf("    grid-column: span %d;\n", columns))
			builder.WriteString("  }\n")
		}
		for _, group := range page.View.Groups {
			if group.Span <= columns {
				continue
			}
			builder.WriteString("\n")
			builder.WriteString(fmt.Sprintf("  .page-view-%s .bl-view-group-%s {\n", kebabCase(page.Name), kebabCase(group.Name)))
			builder.WriteString(fmt.Sprintf("    grid-column: span %d;\n", columns))
			builder.WriteString("  }\n")
		}
	}
	builder.WriteString("}")
	return builder.String()
}

func pageViewGroupRule(page PageDecl, group ViewGroupDecl) uiCSSRule {
	selector := fmt.Sprintf(".page-view-%s .bl-view-group-%s", kebabCase(page.Name), kebabCase(group.Name))
	mode := viewGroupComposeMode(group)
	declarations := []string{fmt.Sprintf("gap: %s;", viewGapValue(""))}
	if group.Compose != nil && group.Compose.Gap != "" {
		declarations = []string{fmt.Sprintf("gap: %s;", viewGapValue(group.Compose.Gap))}
	}
	if mode == "grid" {
		columns := 2
		if group.Compose != nil && group.Compose.Columns > 0 {
			columns = group.Compose.Columns
		}
		declarations = append([]string{"display: grid;", fmt.Sprintf("grid-template-columns: repeat(%d, minmax(0, 1fr));", columns), "align-items: start;"}, declarations...)
	} else {
		declarations = append([]string{"display: flex;", "flex-direction: column;"}, declarations...)
	}
	return uiCSSRule{selector: selector, declarations: declarations}
}

func pageViewGroupMediaBlock(page PageDecl, group ViewGroupDecl) string {
	if viewGroupComposeMode(group) != "grid" {
		return ""
	}
	columns := 2
	if group.Compose != nil && group.Compose.Columns > 0 {
		columns = group.Compose.Columns
	}
	if columns <= 1 {
		return ""
	}
	selector := fmt.Sprintf(".page-view-%s .bl-view-group-%s", kebabCase(page.Name), kebabCase(group.Name))
	var builder strings.Builder
	builder.WriteString("@media (max-width: 768px) {\n")
	builder.WriteString(fmt.Sprintf("  %s {\n", selector))
	builder.WriteString("    grid-template-columns: 1fr;\n")
	builder.WriteString("  }\n\n")
	builder.WriteString(fmt.Sprintf("  %s .panel {\n", selector))
	builder.WriteString("    grid-column: 1 / -1;\n")
	builder.WriteString("  }\n")
	builder.WriteString("}")
	return builder.String()
}

func viewGroupComposeMode(group ViewGroupDecl) string {
	if group.Compose != nil && group.Compose.Mode != "" {
		return group.Compose.Mode
	}
	return "stack"
}

func viewComposeColumns(compose ViewComposeDecl) int {
	if compose.Columns > 0 {
		return compose.Columns
	}
	return 2
}

func viewGapValue(gap string) string {
	switch gap {
	case "sm":
		return "8px"
	case "lg":
		return "24px"
	default:
		return "16px"
	}
}

func viewBreakpointMaxWidth(stackAt string) string {
	switch stackAt {
	case "sm":
		return "640px"
	case "lg":
		return "1024px"
	default:
		return "768px"
	}
}

func pageViewOrder(page PageDecl) []string {
	defaultOrder := []string{"table", "detail", "form"}
	if page.View == nil || len(page.View.Order) == 0 {
		return append(defaultOrder, pageViewComponentSectionNames(page)...)
	}

	seen := map[string]bool{}
	order := []string{}
	for _, section := range page.View.Order {
		if seen[section] || !pageViewSectionSupported(page, section) {
			continue
		}
		order = append(order, section)
		seen[section] = true
	}
	for _, section := range defaultOrder {
		if !seen[section] {
			order = append(order, section)
			seen[section] = true
		}
	}
	for _, section := range pageViewComponentSectionNames(page) {
		if !seen[section] {
			order = append(order, section)
			seen[section] = true
		}
	}
	return order
}

func pageViewComponentSectionNames(page PageDecl) []string {
	sections := []string{}
	for _, section := range pageViewComponentSections(page) {
		sections = append(sections, section.Name)
	}
	return sections
}

func (g *webGenerator) inlineUICSS() string {
	rules := []uiCSSRule{}
	seenFieldRules := map[string]bool{}

	for _, page := range g.program.Pages {
		if len(page.Table.UI) > 0 {
			rules = append(rules, g.inlineUIRulesFor("."+g.tableUIClass(page), "table", page.Table.UI)...)
		}
		if len(page.Form.UI) > 0 {
			rules = append(rules, g.inlineUIRulesFor("."+g.formUIClass(page), "form", page.Form.UI)...)
		}
		for _, actionUI := range page.ActionUI {
			if len(actionUI.UI) == 0 {
				continue
			}
			rules = append(rules, g.inlineUIRulesFor("."+g.actionUIClass(page, actionUI.Action), "button", actionUI.UI)...)
		}

		entity, ok := g.findEntity(page.Source)
		if !ok {
			continue
		}
		for _, fieldName := range page.Form.Fields {
			field, ok := findField(entity, fieldName)
			if !ok || len(field.UI) == 0 {
				continue
			}
			className := g.fieldUIClass(entity, field)
			if seenFieldRules[className] {
				continue
			}
			seenFieldRules[className] = true
			rules = append(rules, g.inlineUIRulesFor("."+className, "field", field.UI)...)
		}
	}

	if len(rules) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("/* Generated from BlackLang inline UI intent. */\n")
	for _, rule := range rules {
		if len(rule.declarations) == 0 {
			continue
		}
		builder.WriteString(rule.selector)
		builder.WriteString(" {\n")
		for _, declaration := range rule.declarations {
			builder.WriteString("  ")
			builder.WriteString(declaration)
			builder.WriteString("\n")
		}
		builder.WriteString("}\n\n")
	}
	return strings.TrimRight(builder.String(), "\n") + "\n"
}

func (g *webGenerator) inlineUIRulesFor(selector string, target string, intents []UIIntent) []uiCSSRule {
	rules := []uiCSSRule{}
	for _, intent := range intents {
		values := g.uiSlotValues(intent)
		switch intent.Mode {
		case "box":
			rules = append(rules, uiCSSRule{selector: selector, declarations: boxUIDeclarations(values)})
		case "text":
			textDeclarations := textUIDeclarations(values)
			rules = append(rules, uiCSSRule{selector: selector, declarations: textDeclarations})
			switch target {
			case "form":
				rules = append(rules, uiCSSRule{selector: selector + " label, " + selector + " h2", declarations: textDeclarations})
			case "table":
				rules = append(rules, uiCSSRule{selector: selector + " th, " + selector + " td", declarations: textDeclarations})
			}
		case "table":
			rules = append(rules, tableUIRules(selector, values)...)
		case "button":
			buttonSelector := selector
			if target == "form" {
				buttonSelector += " button"
			}
			rules = append(rules, uiCSSRule{selector: buttonSelector, declarations: buttonUIDeclarations(values)})
		}
	}
	return rules
}

func (g *webGenerator) uiSlotValues(intent UIIntent) map[string]string {
	slots := g.uiSlotsForMode(intent.Mode)
	values := map[string]string{}
	for index, slot := range slots {
		if index >= len(intent.Values) {
			break
		}
		if value := uiValue(intent.Values, index); value != "" {
			values[slot] = value
		}
	}
	return values
}

func (g *webGenerator) uiSlotsForMode(mode string) []string {
	if g.theme != nil {
		for _, declaredMode := range g.theme.Profile.Modes {
			if declaredMode.Name == mode && len(declaredMode.Slots) > 0 {
				return declaredMode.Slots
			}
		}
	}
	if group, ok := standardUIModeGroup(mode); ok {
		return group.DefaultSlots
	}
	return []string{}
}

func boxUIDeclarations(values map[string]string) []string {
	declarations := []string{}
	color := uiSlotValue(values, "color")
	width := uiSlotValue(values, "width")
	style := uiSlotValue(values, "style")
	pt := uiSlotValue(values, "pt")
	pr := uiSlotValue(values, "pr")
	pb := uiSlotValue(values, "pb")
	pl := uiSlotValue(values, "pl")
	radius := uiSlotValue(values, "radius")
	place := uiSlotValue(values, "place")

	if color != "" {
		declarations = append(declarations, "border-color: "+cssColor(color)+";")
	}
	if width != "" {
		declarations = append(declarations, "border-width: "+cssLength(width)+";")
	}
	if style != "" {
		declarations = append(declarations, "border-style: "+style+";")
	}
	if pt != "" || pr != "" || pb != "" || pl != "" {
		declarations = append(declarations, "padding: "+cssPadding(pt, pr, pb, pl)+";")
	}
	if radius != "" {
		declarations = append(declarations, "border-radius: "+cssRadius(radius)+";")
	}
	if place != "" {
		declarations = append(declarations, "text-align: "+place+";")
	}
	return declarations
}

func textUIDeclarations(values map[string]string) []string {
	declarations := []string{}
	if color := uiSlotValue(values, "color"); color != "" {
		declarations = append(declarations, "color: "+cssColor(color)+";")
	}
	if size := uiSlotValue(values, "size"); size != "" {
		declarations = append(declarations, "font-size: "+cssLength(size)+";")
	}
	if weight := uiSlotValue(values, "weight"); weight != "" {
		declarations = append(declarations, "font-weight: "+cssFontWeight(weight)+";")
	}
	if align := uiSlotValue(values, "align"); align != "" {
		declarations = append(declarations, "text-align: "+align+";")
	}
	return declarations
}

func tableUIRules(selector string, values map[string]string) []uiCSSRule {
	color := cssColor(uiSlotValue(values, "color"))
	width := cssLength(uiSlotValue(values, "width"))
	style := uiSlotValue(values, "style")
	density := strings.ToLower(uiSlotValue(values, "density"))
	zebra := strings.ToLower(uiSlotValue(values, "zebra"))

	if color == "" {
		color = "#d8dee4"
	}
	if width == "" {
		width = "1px"
	}
	if style == "" {
		style = "solid"
	}

	cellPadding := "10px"
	switch density {
	case "tight":
		cellPadding = "6px"
	case "compact":
		cellPadding = "8px"
	case "comfortable", "relaxed":
		cellPadding = "14px"
	}

	rules := []uiCSSRule{
		{
			selector: selector + " table",
			declarations: []string{
				"border-collapse: separate;",
				"border-spacing: 0;",
				fmt.Sprintf("border: %s %s %s;", width, style, color),
				"border-radius: 6px;",
				"overflow: hidden;",
			},
		},
		{
			selector: selector + " th, " + selector + " td",
			declarations: []string{
				fmt.Sprintf("border-bottom: %s %s %s;", width, style, color),
				"padding: " + cellPadding + ";",
			},
		},
	}

	if zebra == "true" || zebra == "yes" || zebra == "on" || zebra == "zebra" {
		rules = append(rules, uiCSSRule{
			selector: selector + " tbody tr:nth-child(even)",
			declarations: []string{
				"background: rgba(37, 99, 235, 0.045);",
			},
		})
	}
	return rules
}

func buttonUIDeclarations(values map[string]string) []string {
	bg := cssColor(uiSlotValue(values, "bg", "background"))
	color := cssColor(uiSlotValue(values, "color"))
	radius := cssRadius(uiSlotValue(values, "radius"))
	size := strings.ToLower(uiSlotValue(values, "size"))
	variant := strings.ToLower(uiSlotValue(values, "variant"))

	if bg == "" {
		bg = "#1f6feb"
	}
	if color == "" {
		color = "#ffffff"
	}
	if radius == "" {
		radius = "6px"
	}

	padding := "9px 12px"
	fontSize := ""
	switch size {
	case "xs":
		padding = "5px 8px"
		fontSize = "12px"
	case "sm":
		padding = "7px 10px"
		fontSize = "13px"
	case "lg":
		padding = "11px 16px"
		fontSize = "15px"
	case "xl":
		padding = "13px 18px"
		fontSize = "16px"
	}

	declarations := []string{
		"border-color: " + bg + ";",
		"border-radius: " + radius + ";",
		"padding: " + padding + ";",
	}
	if fontSize != "" {
		declarations = append(declarations, "font-size: "+fontSize+";")
	}

	switch variant {
	case "outline":
		declarations = append(declarations,
			"background: transparent;",
			"color: "+bg+";",
		)
	case "ghost":
		declarations = append(declarations,
			"background: transparent;",
			"border-color: transparent;",
			"color: "+bg+";",
		)
	default:
		declarations = append(declarations,
			"background: "+bg+";",
			"color: "+color+";",
		)
	}
	return declarations
}

func uiSlotValue(values map[string]string, names ...string) string {
	for _, name := range names {
		if value := values[name]; value != "" {
			return value
		}
	}
	return ""
}

func uiValue(values []string, index int) string {
	if index >= len(values) {
		return ""
	}
	value := strings.TrimSpace(values[index])
	if value == "" || value == "_" || value == "-" || strings.EqualFold(value, "default") {
		return ""
	}
	return value
}

func cssColor(value string) string {
	switch strings.ToLower(value) {
	case "":
		return ""
	case "primary":
		return "#2563eb"
	case "surface", "white":
		return "#ffffff"
	case "text", "black":
		return "#172026"
	case "border":
		return "#d8dee4"
	case "muted":
		return "#57606a"
	case "danger":
		return "#cf222e"
	case "success":
		return "#087f5b"
	default:
		return value
	}
}

func cssLength(value string) string {
	switch strings.ToLower(value) {
	case "":
		return ""
	case "xs":
		return "4px"
	case "sm":
		return "8px"
	case "md":
		return "12px"
	case "lg":
		return "16px"
	case "xl":
		return "24px"
	}
	if isNumericLiteral(value) {
		return value + "px"
	}
	return value
}

func cssRadius(value string) string {
	switch strings.ToLower(value) {
	case "":
		return ""
	case "none":
		return "0"
	case "sm":
		return "4px"
	case "md":
		return "6px"
	case "lg":
		return "8px"
	case "xl":
		return "12px"
	case "full", "pill":
		return "999px"
	default:
		return cssLength(value)
	}
}

func cssPadding(top string, right string, bottom string, left string) string {
	values := []string{cssLength(top), cssLength(right), cssLength(bottom), cssLength(left)}
	for index, value := range values {
		if value == "" {
			values[index] = "0"
		}
	}
	return strings.Join(values, " ")
}

func cssFontWeight(value string) string {
	switch strings.ToLower(value) {
	case "thin":
		return "100"
	case "light":
		return "300"
	case "regular", "normal":
		return "400"
	case "medium":
		return "500"
	case "semibold":
		return "600"
	case "bold":
		return "700"
	case "black":
		return "900"
	default:
		return value
	}
}

func (g *webGenerator) viteEnv() string {
	return `/// <reference types="vite/client" />
`
}

func (g *webGenerator) componentTSX(component ComponentDecl) string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString(fmt.Sprintf("type %sProps = {\n", component.Name))
	for _, input := range component.Inputs {
		builder.WriteString(fmt.Sprintf("  %s: %s;\n", input.Name, g.componentInputTSType(input)))
	}
	builder.WriteString("};\n\n")
	builder.WriteString(fmt.Sprintf("export function %s(props: %sProps) {\n", component.Name, component.Name))
	builder.WriteString("  const variant = resolveVariant(props);\n\n")
	builder.WriteString(fmt.Sprintf("  return <span className={\"black-component black-component-%s black-component-%s-\" + variant}>{String(%s)}</span>;\n", kebabCase(component.Name), kebabCase(component.Name), g.componentDisplayExpression(component)))
	builder.WriteString("}\n\n")
	builder.WriteString(fmt.Sprintf("function resolveVariant(props: %sProps) {\n", component.Name))
	for _, variant := range component.Variants {
		if condition := componentVariantConditionExpression(variant.Condition, component.Inputs); condition != "" {
			builder.WriteString(fmt.Sprintf("  if (%s) return %q;\n", condition, variant.Name))
		}
	}
	builder.WriteString("  return \"default\";\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) componentInputTSType(input ComponentInput) string {
	inputType := tsType(input.Type)
	if _, ok := g.findEntity(input.Type); ok {
		inputType = input.Type
	}
	if input.List {
		return inputType + "[]"
	}
	return inputType
}

func (g *webGenerator) componentDisplayExpression(component ComponentDecl) string {
	if len(component.Inputs) == 0 {
		return "variant"
	}
	for _, input := range component.Inputs {
		if !input.List && input.Type != "boolean" {
			return "props." + input.Name
		}
	}
	return "variant"
}

func (g *webGenerator) pageDOMSectionMarkup(page PageDecl, entity EntityDecl, sectionName string, state StateDecl, hasState bool) string {
	if section, ok := pageViewComponentSection(page, sectionName); ok {
		return g.componentViewSectionMarkup(page, entity, section)
	}

	switch sectionName {
	case "table":
		return g.pageTableSectionMarkup(page, entity, state, hasState)
	case "detail":
		return g.pageDetailSectionMarkup(page, entity)
	case "form":
		return g.pageFormSectionMarkup(page, entity, state, hasState)
	default:
		return ""
	}
}

func (g *webGenerator) pageTableSectionMarkup(page PageDecl, entity EntityDecl, state StateDecl, hasState bool) string {
	var builder strings.Builder
	builder.WriteString(g.viewGroupOpen(page, "table"))
	builder.WriteString(g.viewSectionConditionalOpen(page, "table"))
	builder.WriteString(fmt.Sprintf("      <section%s>\n", g.tablePanelAttributes(page)))
	builder.WriteString("        <div className=\"toolbar\">\n")
	builder.WriteString("          <input\n")
	builder.WriteString("        value={search}\n")
	builder.WriteString("        onChange={(event) => setSearch(event.target.value)}\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("        placeholder={%s}\n", g.uiLabelExpression("table.search", "Search", "locale")))
	} else {
		builder.WriteString("        placeholder=\"Search\"\n")
	}
	builder.WriteString("      />\n\n")
	if hasAction(page, "archive") || hasAction(page, "restore") {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("          <label className=\"inline-control\"><input checked={showArchived} type=\"checkbox\" onChange={(event) => setShowArchived(event.target.checked)} /> {%s}</label>\n", g.uiLabelExpression("table.showArchived", "Show archived", "locale")))
		} else {
			builder.WriteString("          <label className=\"inline-control\"><input checked={showArchived} type=\"checkbox\" onChange={(event) => setShowArchived(event.target.checked)} /> Show archived</label>\n")
		}
	}
	if hasAction(page, "delete") {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("          {canDelete && <button%s type=\"button\" disabled={saving || selectedIds.length === 0} onClick={bulkDeleteSelected}>{%s} ({selectedIds.length})</button>}\n", g.actionButtonAttributesWithSuffix(page, "delete", "bulk", "danger"), g.uiLabelExpression("action.deleteSelected", "Delete Selected", "locale")))
		} else {
			builder.WriteString(fmt.Sprintf("          {canDelete && <button%s type=\"button\" disabled={saving || selectedIds.length === 0} onClick={bulkDeleteSelected}>Delete Selected ({selectedIds.length})</button>}\n", g.actionButtonAttributesWithSuffix(page, "delete", "bulk", "danger")))
		}
	}
	if g.createStartButtonEnabled(page) {
		builder.WriteString(g.openFormPanelButton(page, entity))
	} else if hasState && hasAction(page, "create") {
		builder.WriteString(g.openCreateModalButton(page, state, entity))
	}
	builder.WriteString("        </div>\n")
	builder.WriteString(g.querySummaryMarkup(page))
	if len(page.Table.Filters) > 0 {
		builder.WriteString("        <div className=\"toolbar filter-controls\">\n")
		for _, fieldName := range page.Table.Filters {
			field, ok := findField(entity, fieldName)
			if !ok {
				continue
			}
			label := g.fieldLabel(entity, field)
			if g.hasRuntimeI18N() {
				builder.WriteString(fmt.Sprintf("          <label>\n            %s {%s}\n            <input type=\"text\" value={filters.%s} onChange={(event) => updateFilter(%q, event.target.value)} placeholder={%s + \" \" + %s} />\n          </label>\n", g.fieldLabelJSX(entity, field), g.uiLabelExpression("table.filter", "Filter", "locale"), field.Name, field.Name, g.uiLabelExpression("table.filter", "Filter", "locale"), g.fieldLabelExpression(entity, field)))
			} else {
				builder.WriteString(fmt.Sprintf("          <label>\n            %s Filter\n            <input type=\"text\" value={filters.%s} onChange={(event) => updateFilter(%q, event.target.value)} placeholder=\"Filter %s\" />\n          </label>\n", g.fieldLabelJSX(entity, field), field.Name, field.Name, label))
			}
		}
		builder.WriteString("        </div>\n")
	}
	builder.WriteString("        <div className=\"toolbar column-controls\">\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("          <span className=\"muted\">{%s}</span>\n", g.uiLabelExpression("table.columns", "Columns", "locale")))
	} else {
		builder.WriteString("          <span className=\"muted\">Columns</span>\n")
	}
	for _, column := range page.Table.Columns {
		field, ok := findField(entity, column)
		if ok {
			builder.WriteString(fmt.Sprintf("          {permissions.fields.%s !== false && <label className=\"inline-control\"><input checked={visibleColumns.%s} type=\"checkbox\" onChange={() => toggleColumn(%q)} /> %s</label>}\n", field.Name, field.Name, field.Name, g.fieldLabelJSX(entity, field)))
			continue
		}
		computed, ok := findComputedField(entity, column)
		if ok {
			builder.WriteString(fmt.Sprintf("          {%s && <label className=\"inline-control\"><input checked={visibleColumns.%s} type=\"checkbox\" onChange={() => toggleColumn(%q)} /> %s</label>}\n", g.computedFieldPermissionExpression(computed), computed.Name, computed.Name, g.computedFieldLabelJSX(entity, computed)))
		}
	}
	builder.WriteString("        </div>\n")
	builder.WriteString("        {error && <div className=\"error\" role=\"alert\">{error}</div>}\n")
	if len(customActionsForPage(g.program, page)) > 0 {
		builder.WriteString("        {notice && <div className=\"status\" role=\"status\">{notice}</div>}\n")
	}
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("        {loading && <div className=\"status\">{%s}</div>}\n", g.uiLabelExpression("status.loadingRecords", "Loading records...", "locale")))
	} else {
		builder.WriteString("        {loading && <div className=\"status\">Loading records...</div>}\n")
	}
	builder.WriteString(g.paginationToolbar(page.Table.Paginate, "        "))
	builder.WriteString("        <table>\n")
	builder.WriteString("        <thead>\n")
	builder.WriteString("          <tr>\n")
	if hasAction(page, "delete") {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("            {canDelete && <th className=\"select-cell\"><input aria-label={%s} checked={allVisibleSelected} type=\"checkbox\" onChange={toggleVisibleSelection} /></th>}\n", g.uiLabelExpression("table.selectVisibleRecords", "Select visible records", "locale")))
		} else {
			builder.WriteString("            {canDelete && <th className=\"select-cell\"><input aria-label=\"Select visible records\" checked={allVisibleSelected} type=\"checkbox\" onChange={toggleVisibleSelection} /></th>}\n")
		}
	}
	for _, column := range page.Table.Columns {
		field, ok := findField(entity, column)
		if ok {
			builder.WriteString(fmt.Sprintf("            {visibleColumns.%s && permissions.fields.%s !== false && <th>%s</th>}\n", field.Name, field.Name, g.fieldLabelJSX(entity, field)))
			continue
		}
		computed, ok := findComputedField(entity, column)
		if ok {
			builder.WriteString(fmt.Sprintf("            {visibleColumns.%s && %s && <th>%s</th>}\n", computed.Name, g.computedFieldPermissionExpression(computed), g.computedFieldLabelJSX(entity, computed)))
		}
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("            <th>{%s}</th>\n", g.uiLabelExpression("table.status", "Status", "locale")))
		} else {
			builder.WriteString("            <th>Status</th>\n")
		}
	}
	if g.hasRowActions(page, entity) {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("            <th>{%s}</th>\n", g.uiLabelExpression("table.actions", "Actions", "locale")))
		} else {
			builder.WriteString("            <th>Actions</th>\n")
		}
	}
	builder.WriteString("          </tr>\n")
	builder.WriteString("        </thead>\n")
	builder.WriteString("        <tbody>\n")
	builder.WriteString("          {!loading && visibleItems.length === 0 && (\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("            <tr><td colSpan={tableColspan}>{%s}</td></tr>\n", g.uiLabelExpression("table.empty", "No "+strings.ToLower(entity.Name)+" records yet.", "locale")))
	} else {
		builder.WriteString(fmt.Sprintf("            <tr><td colSpan={tableColspan}>No %s records yet.</td></tr>\n", strings.ToLower(entity.Name)))
	}
	builder.WriteString("          )}\n")
	builder.WriteString("          {paginatedItems.map((item) => (\n")
	builder.WriteString("            <tr key={item.id}>\n")
	if hasAction(page, "delete") {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("              {canDelete && <td className=\"select-cell\"><input aria-label={%s} checked={selectedIds.includes(item.id)} type=\"checkbox\" onChange={() => toggleItemSelection(item.id)} /></td>}\n", g.uiLabelExpression("table.selectRecord", "Select record", "locale")))
		} else {
			builder.WriteString("              {canDelete && <td className=\"select-cell\"><input aria-label=\"Select record\" checked={selectedIds.includes(item.id)} type=\"checkbox\" onChange={() => toggleItemSelection(item.id)} /></td>}\n")
		}
	}
	for _, column := range page.Table.Columns {
		field, ok := findField(entity, column)
		if ok {
			builder.WriteString(fmt.Sprintf("              {visibleColumns.%s && permissions.fields.%s !== false && <td>{%s}</td>}\n", field.Name, field.Name, g.itemRenderExpression("item", field)))
			continue
		}
		computed, ok := findComputedField(entity, column)
		if ok {
			builder.WriteString(fmt.Sprintf("              {visibleColumns.%s && %s && <td>{%s}</td>}\n", computed.Name, g.computedFieldPermissionExpression(computed), g.computedFieldDisplayExpression("item", computed)))
		}
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("              <td>{item.archivedAt ? %s : %s}</td>\n", g.uiLabelExpression("status.archived", "Archived", "locale"), g.uiLabelExpression("status.active", "Active", "locale")))
		} else {
			builder.WriteString("              <td>{item.archivedAt ? \"Archived\" : \"Active\"}</td>\n")
		}
	}
	if g.hasRowActions(page, entity) {
		builder.WriteString("              <td>\n")
		builder.WriteString("                <div className=\"actions\">\n")
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("                  <button className=\"secondary\" type=\"button\" onClick={() => viewItem(item.id)}>{%s}</button>\n", g.uiLabelExpression("action.view", "View", "locale")))
		} else {
			builder.WriteString("                  <button className=\"secondary\" type=\"button\" onClick={() => viewItem(item.id)}>View</button>\n")
		}
		if hasAction(page, "archive") {
			if g.hasRuntimeI18N() {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && !item.archivedAt && <button%s type=\"button\" onClick={() => archiveItem(item)}>{%s}</button>}\n", g.actionRowButtonAttributes(page, "archive", "secondary"), g.uiLabelExpression("action.archive", "Archive", "locale")))
			} else {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && !item.archivedAt && <button%s type=\"button\" onClick={() => archiveItem(item)}>Archive</button>}\n", g.actionRowButtonAttributes(page, "archive", "secondary")))
			}
		}
		if hasAction(page, "restore") {
			if g.hasRuntimeI18N() {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && item.archivedAt && <button%s type=\"button\" onClick={() => restoreItem(item)}>{%s}</button>}\n", g.actionRowButtonAttributes(page, "restore", "secondary"), g.uiLabelExpression("action.restore", "Restore", "locale")))
			} else {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && item.archivedAt && <button%s type=\"button\" onClick={() => restoreItem(item)}>Restore</button>}\n", g.actionRowButtonAttributes(page, "restore", "secondary")))
			}
		}
		if hasAction(page, "edit") {
			if g.hasRuntimeI18N() {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && <button%s type=\"button\" onClick={() => editItem(item)}>{%s}</button>}\n", g.actionRowButtonAttributes(page, "edit", "secondary"), g.uiLabelExpression("action.edit", "Edit", "locale")))
			} else {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && <button%s type=\"button\" onClick={() => editItem(item)}>Edit</button>}\n", g.actionRowButtonAttributes(page, "edit", "secondary")))
			}
		}
		if hasAction(page, "delete") {
			if g.hasRuntimeI18N() {
				builder.WriteString(fmt.Sprintf("                  {canDelete && <button%s type=\"button\" onClick={() => deleteItem(item.id)}>{%s}</button>}\n", g.actionRowButtonAttributes(page, "delete", "danger"), g.uiLabelExpression("action.delete", "Delete", "locale")))
			} else {
				builder.WriteString(fmt.Sprintf("                  {canDelete && <button%s type=\"button\" onClick={() => deleteItem(item.id)}>Delete</button>}\n", g.actionRowButtonAttributes(page, "delete", "danger")))
			}
		}
		builder.WriteString(g.customActionPageButtons(page, entity))
		builder.WriteString(g.workflowPageActionButtons(entity))
		builder.WriteString("                </div>\n")
		builder.WriteString("              </td>\n")
	}
	builder.WriteString("            </tr>\n")
	builder.WriteString("          ))}\n")
	builder.WriteString("        </tbody>\n")
	builder.WriteString("      </table>\n\n")
	builder.WriteString(g.paginationToolbar(page.Table.Paginate, "        "))
	builder.WriteString("      </section>\n")
	builder.WriteString(g.viewSectionConditionalClose(page, "table"))
	builder.WriteString(g.viewGroupClose(page, "table"))
	builder.WriteString("\n")
	return builder.String()
}

func (g *webGenerator) pageDetailSectionMarkup(page PageDecl, entity EntityDecl) string {
	var builder strings.Builder
	builder.WriteString(g.viewGroupOpen(page, "detail"))
	builder.WriteString(g.viewSectionConditionalOpen(page, "detail"))
	builder.WriteString(g.detailSectionOpen(page, entity))
	builder.WriteString(g.detailSectionHeader(page, entity))
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("        {reading && <div className=\"status\">{%s}</div>}\n", g.uiLabelExpression("status.loadingDetails", "Loading details...", "locale")))
		builder.WriteString(fmt.Sprintf("        {!reading && !selectedItem && <p className=\"muted\">{%s}</p>}\n", g.uiLabelExpression("status.selectRecord", "Select a record to view details.", "locale")))
	} else {
		builder.WriteString("        {reading && <div className=\"status\">Loading details...</div>}\n")
		builder.WriteString("        {!reading && !selectedItem && <p className=\"muted\">Select a record to view details.</p>}\n")
	}
	builder.WriteString("        {selectedItem && (\n")
	builder.WriteString("          <dl className=\"detail-grid\">\n")
	builder.WriteString("            <div><dt>ID</dt><dd>{selectedItem.id}</dd></div>\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("            <div><dt>{%s}</dt><dd>{selectedItem.archivedAt ? %s : %s}</dd></div>\n", g.uiLabelExpression("table.status", "Status", "locale"), g.uiLabelExpression("status.archived", "Archived", "locale"), g.uiLabelExpression("status.active", "Active", "locale")))
	} else {
		builder.WriteString("            <div><dt>Status</dt><dd>{selectedItem.archivedAt ? \"Archived\" : \"Active\"}</dd></div>\n")
	}
	for _, field := range entity.Fields {
		builder.WriteString(fmt.Sprintf("            {permissions.fields.%s !== false && <div><dt>%s</dt><dd>{%s}</dd></div>}\n", field.Name, g.fieldLabelJSX(entity, field), g.itemRenderExpression("selectedItem", field)))
	}
	for _, computed := range entity.ComputedFields {
		builder.WriteString(fmt.Sprintf("            {%s && <div><dt>%s</dt><dd>{%s}</dd></div>}\n", g.computedFieldPermissionExpression(computed), g.computedFieldLabelJSX(entity, computed), g.computedFieldDisplayExpression("selectedItem", computed)))
	}
	builder.WriteString("          </dl>\n")
	builder.WriteString("        )}\n")
	builder.WriteString(g.detailSectionClose(page))
	builder.WriteString(g.customActionPanels(page, entity))
	builder.WriteString(g.viewSectionConditionalClose(page, "detail"))
	builder.WriteString(g.viewGroupClose(page, "detail"))
	builder.WriteString("\n")
	return builder.String()
}

func (g *webGenerator) pageFormSectionMarkup(page PageDecl, entity EntityDecl, state StateDecl, hasState bool) string {
	var builder strings.Builder
	builder.WriteString(g.viewGroupOpen(page, "form"))
	builder.WriteString(g.viewSectionConditionalOpen(page, "form"))
	if hasAction(page, "create") || hasAction(page, "edit") {
		builder.WriteString(fmt.Sprintf("      {%s && (\n", g.formVisibleExpression(page, entity, state, hasState)))
		builder.WriteString(g.formSectionOpen(page))
		builder.WriteString(g.formSectionHeader(page, entity))
		builder.WriteString("        {missingRequiredRelations && (\n")
		builder.WriteString("          <div className=\"status relation-status\">\n")
		builder.WriteString(fmt.Sprintf("            <span>%s</span>\n", g.missingRequiredRelationsMessage(page, entity)))
		builder.WriteString(g.relationNavigationButtons(page, entity))
		builder.WriteString("          </div>\n")
		builder.WriteString("        )}\n")
		builder.WriteString("        <form noValidate onSubmit={saveItem}>\n")
		builder.WriteString("          <div className=\"form-grid\">\n")
		for _, fieldName := range page.Form.Fields {
			field, ok := findField(entity, fieldName)
			if !ok {
				continue
			}
			if g.isRelationField(field) {
				optionsName := relationOptionsStateName(field)
				label := g.fieldLabel(entity, field)
				builder.WriteString(fmt.Sprintf("            {permissions.fields.%s !== false && (editingId ? permissions.writeFields.%s !== false : canCreate) && <label%s>\n              %s\n              <select%s disabled={%s.length === 0} value={form.%s} onChange={(event) => updateField(\"%s\", event.target.value)}>\n                <option value=\"\">%s</option>\n                {%s.map((option) => (\n                  <option key={option.id} value={option.id}>{%s}</option>\n                ))}\n              </select>\n              {visibleFormErrors.%s && <span className=\"field-error\">{visibleFormErrors.%s}</span>}\n              %s{%s.length === 0 && <span className=\"field-note\">Create a %s record before selecting %s.</span>}\n            </label>}\n", field.Name, field.Name, g.fieldUIClassAttribute(entity, field), g.fieldLabelJSX(entity, field), selectAttributes(field), optionsName, field.Name, field.Name, g.fieldPlaceholderJSX(entity, field, "Select "+label), optionsName, g.relationOptionLabelExpression("option", field.Type), field.Name, field.Name, g.fieldHelpElement(entity, field), optionsName, field.Type, label))
				continue
			}
			if mediaField(field) {
				builder.WriteString(fmt.Sprintf("            {permissions.fields.%s !== false && (editingId ? permissions.writeFields.%s !== false : canCreate) && <label%s>\n              %s\n              <input%s onChange={(event) => void updateFileField(\"%s\", event.currentTarget.files?.[0] ?? null)} />\n              {form.%s && <span className=\"field-preview\">{renderMediaField(form.%s, %q, %q)}</span>}\n              {visibleFormErrors.%s && <span className=\"field-error\">{visibleFormErrors.%s}</span>}\n              %s</label>}\n", field.Name, field.Name, g.fieldUIClassAttribute(entity, field), g.fieldLabelJSX(entity, field), inputAttributes(field), field.Name, field.Name, field.Name, field.Type, fieldLabel(field), field.Name, field.Name, g.fieldHelpElement(entity, field)))
				continue
			}
			builder.WriteString(fmt.Sprintf("            {permissions.fields.%s !== false && (editingId ? permissions.writeFields.%s !== false : canCreate) && <label%s>\n              %s\n              <input%s%s value={form.%s} onChange={(event) => updateField(\"%s\", event.target.value)} />\n              {visibleFormErrors.%s && <span className=\"field-error\">{visibleFormErrors.%s}</span>}\n              %s%s</label>}\n", field.Name, field.Name, g.fieldUIClassAttribute(entity, field), g.fieldLabelJSX(entity, field), inputAttributes(field), g.inputPlaceholderAttribute(entity, field), field.Name, field.Name, field.Name, field.Name, g.formComponentPreview(field), g.fieldHelpElement(entity, field)))
		}
		builder.WriteString("          </div>\n")
		builder.WriteString("          <div className=\"toolbar\">\n")
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("            <button%s type=\"submit\" disabled={saving || missingRequiredRelations}>{saving ? %s : editingId ? %s : %s}</button>\n", g.formSubmitAttributes(page), g.uiLabelExpression("action.saving", "Saving...", "locale"), g.uiLabelExpression("action.saveChanges", "Save Changes", "locale"), g.uiLabelExpression("action.create", "Create", "locale")))
			builder.WriteString(fmt.Sprintf("            {%s && <button className=\"secondary\" type=\"button\" onClick={resetForm}>{%s}</button>}\n", g.cancelVisibleExpression(page, entity, state, hasState), g.uiLabelExpression("action.cancel", "Cancel", "locale")))
		} else {
			builder.WriteString(fmt.Sprintf("            <button%s type=\"submit\" disabled={saving || missingRequiredRelations}>{saving ? \"Saving...\" : editingId ? \"Save Changes\" : \"Create\"}</button>\n", g.formSubmitAttributes(page)))
			builder.WriteString(fmt.Sprintf("            {%s && <button className=\"secondary\" type=\"button\" onClick={resetForm}>Cancel</button>}\n", g.cancelVisibleExpression(page, entity, state, hasState)))
		}
		builder.WriteString("          </div>\n")
		builder.WriteString("        </form>\n")
		builder.WriteString(g.formSectionClose(page))
		builder.WriteString("      )}\n")
	}
	builder.WriteString(g.viewSectionConditionalClose(page, "form"))
	builder.WriteString(g.viewGroupClose(page, "form"))
	return builder.String()
}

func (g *webGenerator) componentsForPage(page PageDecl, entity EntityDecl) []ComponentDecl {
	components := []ComponentDecl{}
	seen := map[string]bool{}
	fieldNames := append([]string{}, page.Table.Columns...)
	fieldNames = append(fieldNames, page.Form.Fields...)
	for _, fieldName := range fieldNames {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		component, ok := g.componentForField(field)
		if !ok || seen[component.Name] {
			continue
		}
		seen[component.Name] = true
		components = append(components, component)
	}
	for _, section := range pageViewComponentSections(page) {
		component, ok := g.componentByName(section.Component)
		if !ok || seen[component.Name] {
			continue
		}
		seen[component.Name] = true
		components = append(components, component)
	}
	return components
}

func (g *webGenerator) componentByName(name string) (ComponentDecl, bool) {
	for _, component := range g.program.Components {
		if component.Name == name {
			return component, true
		}
	}
	return ComponentDecl{}, false
}

func (g *webGenerator) componentForField(field FieldDecl) (ComponentDecl, bool) {
	for _, component := range g.program.Components {
		if len(component.Inputs) != 1 {
			continue
		}
		input := component.Inputs[0]
		if input.Name == field.Name && input.Type == field.Type && !input.List {
			return component, true
		}
	}
	return ComponentDecl{}, false
}

func (g *webGenerator) componentViewSectionState(page PageDecl) string {
	sections := pageViewComponentSections(page)
	if len(sections) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, section := range sections {
		switch section.Bind {
		case "selected":
			builder.WriteString(fmt.Sprintf("  const %s = selectedItem;\n", viewComponentSectionItemName(section)))
		case "first":
			builder.WriteString(fmt.Sprintf("  const %s = items[0] ?? null;\n", viewComponentSectionItemName(section)))
		case "each":
			builder.WriteString(fmt.Sprintf("  const %s = items;\n", viewComponentSectionItemsName(section)))
		default:
			builder.WriteString(fmt.Sprintf("  const %s = null;\n", viewComponentSectionItemName(section)))
		}
	}
	builder.WriteString("\n")
	return builder.String()
}

func (g *webGenerator) componentViewSectionMarkup(page PageDecl, entity EntityDecl, section ViewSectionDecl) string {
	component, ok := g.componentByName(section.Component)
	if !ok {
		return ""
	}
	itemName := viewComponentSectionItemName(section)
	titleText := section.Title
	if titleText == "" {
		titleText = identifierLabel(section.Name)
	}
	permissionExpression := g.componentViewSectionPermissionExpression(entity, component)
	if section.Bind == "each" {
		return g.componentCollectionViewSectionMarkup(page, entity, section, component, titleText, permissionExpression)
	}
	emptyMessage := "Select a record to render " + component.Name + "."
	if section.Bind == "first" {
		emptyMessage = "Load a record to render " + component.Name + "."
	}
	props := g.componentPropsExpression(entity, itemName, component)
	if props != "" {
		props = " " + props
	}

	var builder strings.Builder
	builder.WriteString(g.viewGroupOpen(page, section.Name))
	builder.WriteString(g.viewSectionConditionalOpen(page, section.Name))
	builder.WriteString(fmt.Sprintf("      <section%s>\n", g.componentSectionPanelAttributes(page, section)))
	builder.WriteString(fmt.Sprintf("        <h2>%s</h2>\n", escapeJSXText(titleText)))
	builder.WriteString(fmt.Sprintf("        {%s ? (\n", permissionExpression))
	builder.WriteString(fmt.Sprintf("          %s ? (\n", itemName))
	builder.WriteString("            <div className=\"component-section-body\">\n")
	builder.WriteString(fmt.Sprintf("              <%s%s />\n", component.Name, props))
	builder.WriteString("            </div>\n")
	builder.WriteString("          ) : (\n")
	builder.WriteString(fmt.Sprintf("            <p className=\"muted\">%s</p>\n", escapeJSXText(emptyMessage)))
	builder.WriteString("          )\n")
	builder.WriteString("        ) : (\n")
	builder.WriteString("          <p className=\"muted\">Component hidden by field permissions.</p>\n")
	builder.WriteString("        )}\n")
	builder.WriteString("      </section>\n")
	builder.WriteString(g.viewSectionConditionalClose(page, section.Name))
	builder.WriteString(g.viewGroupClose(page, section.Name))
	return builder.String()
}

func (g *webGenerator) componentCollectionViewSectionMarkup(page PageDecl, entity EntityDecl, section ViewSectionDecl, component ComponentDecl, titleText string, permissionExpression string) string {
	itemsName := viewComponentSectionItemsName(section)
	props := g.componentPropsExpression(entity, "item", component)
	if props != "" {
		props = " " + props
	}
	emptyMessage := "Load records to render " + component.Name + "."

	var builder strings.Builder
	builder.WriteString(g.viewGroupOpen(page, section.Name))
	builder.WriteString(g.viewSectionConditionalOpen(page, section.Name))
	builder.WriteString(fmt.Sprintf("      <section%s>\n", g.componentSectionPanelAttributes(page, section)))
	builder.WriteString(fmt.Sprintf("        <h2>%s</h2>\n", escapeJSXText(titleText)))
	builder.WriteString(fmt.Sprintf("        {%s ? (\n", permissionExpression))
	builder.WriteString(fmt.Sprintf("          %s.length > 0 ? (\n", itemsName))
	builder.WriteString(fmt.Sprintf("            <div className=\"component-section-list\" role=\"list\" aria-label=%q>\n", titleText))
	builder.WriteString(fmt.Sprintf("              {%s.map((item) => (\n", itemsName))
	builder.WriteString("                <div className=\"component-section-list-item\" role=\"listitem\" key={String(item.id)}>\n")
	builder.WriteString(fmt.Sprintf("                  <%s%s />\n", component.Name, props))
	builder.WriteString("                </div>\n")
	builder.WriteString("              ))}\n")
	builder.WriteString("            </div>\n")
	builder.WriteString("          ) : (\n")
	builder.WriteString(fmt.Sprintf("            <p className=\"muted\">%s</p>\n", escapeJSXText(emptyMessage)))
	builder.WriteString("          )\n")
	builder.WriteString("        ) : (\n")
	builder.WriteString("          <p className=\"muted\">Component hidden by field permissions.</p>\n")
	builder.WriteString("        )}\n")
	builder.WriteString("      </section>\n")
	builder.WriteString(g.viewSectionConditionalClose(page, section.Name))
	builder.WriteString(g.viewGroupClose(page, section.Name))
	return builder.String()
}

func (g *webGenerator) componentSectionPanelAttributes(page PageDecl, section ViewSectionDecl) string {
	classes := []string{
		"panel",
		"bl-view-section-" + kebabCase(section.Name),
		"bl-view-component-section",
		"bl-view-component-" + kebabCase(section.Component),
	}
	classes = append(classes, g.viewSectionClasses(page, section.Name)...)
	return classNameAttribute(classes...)
}

func (g *webGenerator) componentPropsExpression(entity EntityDecl, itemName string, component ComponentDecl) string {
	props := []string{}
	for _, input := range component.Inputs {
		props = append(props, fmt.Sprintf("%s={%s}", input.Name, g.componentPropExpressionForItem(entity, itemName, input)))
	}
	return strings.Join(props, " ")
}

func (g *webGenerator) componentPropExpressionForItem(entity EntityDecl, itemName string, input ComponentInput) string {
	if computed, ok := findComputedField(entity, input.Name); ok {
		value := fmt.Sprintf("%s(%s)", computedFieldFunctionName(computed), itemName)
		switch input.Type {
		case "number", "integer", "decimal", "money":
			return fmt.Sprintf("Number(%s ?? 0)", value)
		case "boolean":
			return fmt.Sprintf("Boolean(%s)", value)
		default:
			return fmt.Sprintf("String(%s ?? \"\")", value)
		}
	}
	return componentPropExpression(itemName, input)
}

func (g *webGenerator) componentViewSectionPermissionExpression(entity EntityDecl, component ComponentDecl) string {
	conditions := []string{}
	seen := map[string]bool{}
	for _, input := range component.Inputs {
		if _, ok := findField(entity, input.Name); ok {
			condition := fmt.Sprintf("permissions.fields.%s !== false", input.Name)
			if !seen[condition] {
				conditions = append(conditions, condition)
				seen[condition] = true
			}
			continue
		}
		if computed, ok := findComputedField(entity, input.Name); ok {
			for _, condition := range strings.Split(g.computedFieldPermissionExpression(computed), " && ") {
				if condition == "" || seen[condition] {
					continue
				}
				conditions = append(conditions, condition)
				seen[condition] = true
			}
		}
	}
	if len(conditions) == 0 {
		return "true"
	}
	return strings.Join(conditions, " && ")
}

func viewComponentSectionItemName(section ViewSectionDecl) string {
	return tsSafeIdentifier(lowerCamelCase(section.Name)) + "ComponentItem"
}

func viewComponentSectionItemsName(section ViewSectionDecl) string {
	return tsSafeIdentifier(lowerCamelCase(section.Name)) + "ComponentItems"
}

func tsSafeIdentifier(value string) string {
	var builder strings.Builder
	for index, char := range value {
		valid := (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_'
		if valid {
			if index == 0 && char >= '0' && char <= '9' {
				builder.WriteByte('_')
			}
			builder.WriteRune(char)
			continue
		}
		builder.WriteByte('_')
	}
	if builder.Len() == 0 {
		return "section"
	}
	return builder.String()
}

func (g *webGenerator) itemRenderExpression(itemName string, field FieldDecl) string {
	component, ok := g.componentForField(field)
	if !ok {
		if mediaField(field) {
			return fmt.Sprintf("renderMediaField(%s.%s, %q, %q)", itemName, field.Name, field.Type, fieldLabel(field))
		}
		return g.itemDisplayExpression(itemName, field)
	}
	input := component.Inputs[0]
	return fmt.Sprintf("<%s %s={%s} />", component.Name, input.Name, componentPropExpression(itemName, input))
}

func entityHasMediaFields(entity EntityDecl) bool {
	for _, field := range entity.Fields {
		if mediaField(field) {
			return true
		}
	}
	return false
}

func programHasMediaFields(program Program) bool {
	for _, entity := range program.Entities {
		if entityHasMediaFields(entity) {
			return true
		}
	}
	return false
}

func pageFormHasMediaFields(page PageDecl, entity EntityDecl) bool {
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if ok && mediaField(field) {
			return true
		}
	}
	return false
}

func mediaField(field FieldDecl) bool {
	return field.Type == "file" || field.Type == "image"
}

func mediaValidationExpression(valueExpression string, fieldType string) string {
	if fieldType == "image" {
		return fmt.Sprintf("(%s.startsWith(\"data:image/\") || %s.startsWith(\"https://\") || %s.startsWith(\"http://\"))", valueExpression, valueExpression, valueExpression)
	}
	return fmt.Sprintf("(%s.startsWith(\"data:\") || %s.startsWith(\"https://\") || %s.startsWith(\"http://\"))", valueExpression, valueExpression, valueExpression)
}

func (g *webGenerator) mediaFieldHelpers(entity EntityDecl) string {
	if !entityHasMediaFields(entity) {
		return ""
	}
	return `function renderMediaField(value: unknown, kind: "file" | "image", label: string) {
  const source = String(value ?? "");
  if (!source) return "";
  if (kind === "image") {
    return <img className="media-thumb" src={source} alt={label} loading="lazy" />;
  }
  return <a className="media-link" href={source} target="_blank" rel="noreferrer">Open {label}</a>;
}

`
}

func (g *webGenerator) fileInputHelpers(page PageDecl, entity EntityDecl) string {
	if !pageFormHasMediaFields(page, entity) {
		return ""
	}
	return `  function readFileAsDataURL(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onerror = () => reject(new Error("Unable to read selected file"));
      reader.onload = () => resolve(String(reader.result ?? ""));
      reader.readAsDataURL(file);
    });
  }

  async function updateFileField(field: string, file: File | null) {
    setTouchedFields((current) => ({ ...current, [field]: true }));
    if (!file) {
      setForm((current) => ({ ...current, [field]: "" }));
      return;
    }
    try {
      const value = await readFileAsDataURL(file);
      setForm((current) => ({ ...current, [field]: value }));
    } catch (reason: unknown) {
      setError(reason instanceof Error ? reason.message : "Unable to read selected file");
    }
  }

`
}

func (g *webGenerator) computedFieldHelpers(entity EntityDecl) string {
	var builder strings.Builder
	for _, field := range entity.ComputedFields {
		builder.WriteString(fmt.Sprintf("function %s(item: %s) {\n", computedFieldFunctionName(field), entity.Name))
		builder.WriteString(fmt.Sprintf("  const value = %s;\n", computedFieldExpression("item", field)))
		builder.WriteString("  return Number.isFinite(value) ? value : null;\n")
		builder.WriteString("}\n\n")
	}
	if g.hasRuntimeI18N() {
		builder.WriteString("function formatComputedValue(value: number | null, kind: string, locale: string) {\n")
		builder.WriteString("  return value === null ? \"\" : formatFieldValue(value, kind, locale);\n")
		builder.WriteString("}\n\n")
	} else {
		builder.WriteString("function formatComputedValue(value: number | null) {\n")
		builder.WriteString("  return value === null ? \"\" : String(value);\n")
		builder.WriteString("}\n\n")
	}
	return builder.String()
}

func computedFieldFunctionName(field ComputedFieldDecl) string {
	return "compute" + title(field.Name)
}

func computedFieldExpression(itemName string, field ComputedFieldDecl) string {
	if field.Expression.Tree != nil {
		return renderCoreExpressionJS(*field.Expression.Tree, func(expression ExpressionDecl) string {
			return computedFieldOperandRawExpression(itemName, expression)
		}, true)
	}
	left := computedFieldOperandExpression(itemName, field.Expression.Left)
	right := computedFieldOperandExpression(itemName, field.Expression.Right)
	return fmt.Sprintf("%s %s %s", left, field.Expression.Operator, right)
}

func computedFieldOperandRawExpression(itemName string, expression ExpressionDecl) string {
	if expression.ValueKind == "number" {
		return expression.Value
	}
	return fmt.Sprintf("%s.%s", itemName, expression.Value)
}

func computedFieldOperandExpression(itemName string, operand string) string {
	if isNumericLiteral(operand) {
		return operand
	}
	return fmt.Sprintf("Number(%s.%s ?? 0)", itemName, operand)
}

func (g *webGenerator) formComponentPreview(field FieldDecl) string {
	component, ok := g.componentForField(field)
	if !ok {
		return ""
	}
	input := component.Inputs[0]
	return fmt.Sprintf("<span className=\"field-preview\"><%s %s={%s} /></span>\n              ", component.Name, input.Name, formComponentPropExpression(field, input))
}

func componentPropExpression(itemName string, input ComponentInput) string {
	switch input.Type {
	case "number", "integer", "decimal", "money":
		return fmt.Sprintf("Number(%s.%s ?? 0)", itemName, input.Name)
	case "boolean":
		return fmt.Sprintf("Boolean(%s.%s)", itemName, input.Name)
	default:
		return fmt.Sprintf("String(%s.%s ?? \"\")", itemName, input.Name)
	}
}

func formComponentPropExpression(field FieldDecl, input ComponentInput) string {
	switch input.Type {
	case "number", "integer", "decimal", "money":
		return fmt.Sprintf("Number(form.%s || 0)", field.Name)
	case "boolean":
		return fmt.Sprintf("form.%s === \"true\"", field.Name)
	default:
		return fmt.Sprintf("String(form.%s || \"\")", field.Name)
	}
}

func componentVariantConditionExpression(condition string, inputs []ComponentInput) string {
	parts := strings.Fields(condition)
	if len(parts) != 3 {
		return ""
	}
	input, ok := componentInputByName(inputs, parts[0])
	if !ok || input.List {
		return ""
	}
	operator := parts[1]
	if !supportedComparisonOperators[operator] {
		return ""
	}
	right := parts[2]
	switch input.Type {
	case "number", "integer", "decimal", "money":
		if !isNumericLiteral(right) {
			return ""
		}
		return fmt.Sprintf("Number(props.%s) %s %s", input.Name, operator, right)
	case "boolean":
		if right != "true" && right != "false" {
			return ""
		}
		return fmt.Sprintf("props.%s %s %s", input.Name, operator, right)
	default:
		if operator != "==" && operator != "!=" {
			return ""
		}
		return fmt.Sprintf("String(props.%s) %s %q", input.Name, operator, right)
	}
}

func (g *webGenerator) openapiJSON() string {
	components := map[string]any{"schemas": g.openapiSchemas()}
	if g.program.Auth != nil {
		components["securitySchemes"] = map[string]any{
			"cookieAuth": map[string]any{
				"type": "apiKey",
				"in":   "cookie",
				"name": "black_session",
			},
		}
	}
	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   g.program.App.Name + " API",
			"version": "0.1.0",
		},
		"paths":      g.openapiPaths(),
		"components": components,
	}
	if g.hasServices() {
		spec["tags"] = g.openapiServiceTags()
		spec["x-blacklang-services"] = g.openapiServicesMetadata()
	}
	if observability := g.openAPIObservabilityExtension(); observability != nil {
		spec["x-blacklang-observability"] = observability
	}
	if g.hasJobs() {
		spec["x-blacklang-jobs"] = g.openapiJobsMetadata()
	}
	content, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "{\n  \"openapi\": \"3.1.0\"\n}\n"
	}
	return string(content) + "\n"
}

func (g *webGenerator) openapiPaths() map[string]any {
	paths := map[string]any{}
	g.addOpenAPIOpsPaths(paths)
	if g.program.Auth != nil {
		paths["/api/auth/register"] = map[string]any{
			"post": map[string]any{
				"summary":     "Register a user",
				"requestBody": openapiRequestBody("#/components/schemas/AuthRegisterInput"),
				"responses":   openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
			},
		}
		paths["/api/auth/login"] = map[string]any{
			"post": map[string]any{
				"summary":     "Login a user",
				"requestBody": openapiRequestBody("#/components/schemas/AuthLoginInput"),
				"responses":   openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
			},
		}
		paths["/api/auth/logout"] = map[string]any{
			"post": map[string]any{
				"summary":   "Logout current user",
				"responses": openapiJSONResponses(map[string]any{"type": "object"}),
			},
		}
		paths["/api/auth/me"] = map[string]any{
			"get": map[string]any{
				"summary":   "Read current user session",
				"responses": openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
			},
		}
		if len(g.program.Roles) > 0 {
			paths["/api/auth/users"] = map[string]any{
				"get": map[string]any{
					"summary":  "List authenticated users",
					"security": []any{map[string]any{"cookieAuth": []any{}}},
					"responses": openapiJSONResponses(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"users": map[string]any{
								"type":  "array",
								"items": map[string]any{"$ref": "#/components/schemas/AuthUser"},
							},
						},
						"required": []string{"users"},
					}),
				},
			}
			paths["/api/auth/audit"] = map[string]any{
				"get": map[string]any{
					"summary":  "List audit log entries",
					"security": []any{map[string]any{"cookieAuth": []any{}}},
					"responses": openapiJSONResponses(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"logs": map[string]any{
								"type":  "array",
								"items": map[string]any{"$ref": "#/components/schemas/AuditLog"},
							},
						},
						"required": []string{"logs"},
					}),
				},
			}
			paths["/api/auth/users/{id}/role"] = map[string]any{
				"put": map[string]any{
					"summary":     "Update authenticated user role",
					"security":    []any{map[string]any{"cookieAuth": []any{}}},
					"parameters":  []any{openapiIDParameter()},
					"requestBody": openapiRequestBody("#/components/schemas/AuthRoleUpdateInput"),
					"responses":   openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
				},
			}
			if g.hasTenantPolicies() {
				paths["/api/auth/users/{id}/tenant"] = map[string]any{
					"put": map[string]any{
						"summary":     "Update authenticated user tenant",
						"security":    []any{map[string]any{"cookieAuth": []any{}}},
						"parameters":  []any{openapiIDParameter()},
						"requestBody": openapiRequestBody("#/components/schemas/AuthTenantUpdateInput"),
						"responses":   openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/AuthUserResponse"}),
					},
				}
			}
		}
	}
	for _, page := range g.program.Pages {
		entity, ok := g.findEntity(page.Source)
		if !ok {
			continue
		}
		pathName := "/api/" + strings.ToLower(page.Name)
		if query, ok := findQuery(g.program, page.Query); ok {
			paths[pathName+"/query"] = map[string]any{"get": g.openapiQueryOperation(page, entity, query)}
			if len(query.Aggregates) > 0 {
				paths[pathName+"/query/summary"] = map[string]any{"get": g.openapiQuerySummaryOperation(page, query)}
			}
		}
		entityRef := "#/components/schemas/" + entity.Name
		inputRef := "#/components/schemas/" + entity.Name + "Input"
		collectionOperations := map[string]any{
			"get": g.withOpenAPIRelationLoad(entity, "list", map[string]any{
				"summary": "List " + entity.Name + " records",
				"parameters": []any{
					map[string]any{
						"name": "archived",
						"in":   "query",
						"schema": map[string]any{
							"type": "string",
							"enum": []string{"all"},
						},
					},
				},
				"responses": openapiJSONResponses(map[string]any{
					"type":  "array",
					"items": map[string]any{"$ref": entityRef},
				}),
			}),
		}
		if hasAction(page, "create") {
			collectionOperations["post"] = g.withOpenAPIRelationLoad(entity, "mutation", map[string]any{
				"summary":     "Create a " + entity.Name + " record",
				"requestBody": openapiRequestBody(inputRef),
				"responses":   openapiJSONResponses(map[string]any{"$ref": entityRef}),
			})
		}
		if hasAction(page, "delete") {
			collectionOperations["delete"] = map[string]any{
				"summary":     "Bulk delete " + entity.Name + " records",
				"requestBody": openapiArrayRequestBody("ids", "string"),
				"responses":   openapiJSONResponses(map[string]any{"type": "object"}),
			}
		}
		paths[pathName] = collectionOperations

		itemPath := pathName + "/{id}"
		itemOperations := map[string]any{
			"get": g.withOpenAPIRelationLoad(entity, "detail", map[string]any{
				"summary":    "Read a " + entity.Name + " record",
				"parameters": []any{openapiIDParameter()},
				"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
			}),
		}
		if hasAction(page, "edit") {
			itemOperations["put"] = g.withOpenAPIRelationLoad(entity, "mutation", map[string]any{
				"summary":     "Update a " + entity.Name + " record",
				"parameters":  []any{openapiIDParameter()},
				"requestBody": openapiRequestBody(inputRef),
				"responses":   openapiJSONResponses(map[string]any{"$ref": entityRef}),
			})
		}
		if hasAction(page, "delete") {
			itemOperations["delete"] = map[string]any{
				"summary":    "Delete a " + entity.Name + " record",
				"parameters": []any{openapiIDParameter()},
				"responses":  openapiJSONResponses(map[string]any{"type": "object"}),
			}
		}
		paths[itemPath] = itemOperations

		if hasAction(page, "archive") {
			paths[itemPath+"/archive"] = map[string]any{
				"patch": g.withOpenAPIRelationLoad(entity, "mutation", map[string]any{
					"summary":    "Archive a " + entity.Name + " record",
					"parameters": []any{openapiIDParameter()},
					"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
				}),
			}
		}
		if hasAction(page, "restore") {
			paths[itemPath+"/restore"] = map[string]any{
				"patch": g.withOpenAPIRelationLoad(entity, "mutation", map[string]any{
					"summary":    "Restore a " + entity.Name + " record",
					"parameters": []any{openapiIDParameter()},
					"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
				}),
			}
		}
		for _, action := range customActionsForPage(g.program, page) {
			paths[itemPath+"/actions/"+strings.ToLower(action.Name)] = map[string]any{
				"post": g.openapiActionOperation(page, entity, action),
			}
		}
		if g.hasRuntimePermissions() {
			for _, workflow := range g.workflowsForEntity(entity.Name) {
				for _, transition := range workflow.Transitions {
					paths[itemPath+"/workflow/"+transition.Name] = map[string]any{
						"post": g.withOpenAPIRelationLoad(entity, "mutation", map[string]any{
							"summary":    "Run " + workflow.Name + " transition " + transition.Name,
							"security":   []any{map[string]any{"cookieAuth": []any{}}},
							"parameters": []any{openapiIDParameter()},
							"responses":  openapiJSONResponses(map[string]any{"$ref": entityRef}),
						}),
					}
				}
			}
		}
	}
	for _, api := range g.program.APIs {
		method := strings.ToLower(api.Method)
		if method == "" || api.Path == "" {
			continue
		}
		operations, ok := paths[api.Path].(map[string]any)
		if !ok {
			operations = map[string]any{}
		}
		operation := map[string]any{
			"summary":             api.Name,
			"x-blacklang-api":     api.Name,
			"x-blacklang-access":  explicitAPIAccess(api),
			"x-blacklang-runtime": "declared",
			"responses":           openapiJSONResponses(map[string]any{"type": "object"}),
		}
		if api.Webhook {
			operation["x-blacklang-webhook"] = true
		}
		if api.Update != nil {
			operation["x-blacklang-handler"] = "update"
			operation["x-blacklang-update"] = openapiAPIUpdateMetadata(*api.Update)
		}
		if transactionName, transactional := g.explicitAPITransactionName(api); transactional {
			operation["x-blacklang-transaction"] = true
			if transactionName != "" {
				operation["x-blacklang-transaction-name"] = transactionName
			}
		}
		if serviceName, ok := g.explicitAPIServiceName(api); ok {
			operation["tags"] = []string{serviceName}
			operation["x-blacklang-service"] = serviceName
		}
		if explicitAPIAccess(api) == "private" && g.program.Auth != nil {
			operation["security"] = []any{map[string]any{"cookieAuth": []any{}}}
		}
		parameters := []any{}
		for _, param := range api.Params {
			parameters = append(parameters, openapiExplicitParameter(param, "path", true))
		}
		for _, query := range api.Queries {
			parameters = append(parameters, openapiExplicitParameter(query, "query", false))
		}
		if len(parameters) > 0 {
			operation["parameters"] = parameters
		}
		if method == "post" || method == "put" || method == "patch" {
			operation["requestBody"] = openapiRequestBody("#/components/schemas/" + api.Name + "Input")
		}
		operations[method] = operation
		paths[api.Path] = operations
	}
	return paths
}

func (g *webGenerator) openapiSchemas() map[string]any {
	schemas := map[string]any{}
	if g.program.Auth != nil {
		schemas["AuthRegisterInput"] = openapiObjectSchema(map[string]any{
			"name":     map[string]any{"type": "string"},
			"email":    map[string]any{"type": "string", "format": "email"},
			"password": map[string]any{"type": "string", "minLength": 8},
		}, []string{"name", "email", "password"})
		schemas["AuthLoginInput"] = openapiObjectSchema(map[string]any{
			"email":    map[string]any{"type": "string", "format": "email"},
			"password": map[string]any{"type": "string", "minLength": 8},
		}, []string{"email", "password"})
		authUserProperties := map[string]any{
			"id":    map[string]any{"type": "string"},
			"name":  map[string]any{"type": "string"},
			"email": map[string]any{"type": "string", "format": "email"},
			"role":  map[string]any{"type": "string"},
			"roles": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
					"enum": g.roleNames(),
				},
				"minItems":    1,
				"uniqueItems": true,
			},
		}
		authUserRequired := []string{"id", "name", "email", "role", "roles"}
		if g.hasTenantPolicies() {
			authUserProperties["tenantId"] = map[string]any{"type": "string"}
			authUserRequired = append(authUserRequired, "tenantId")
		}
		schemas["AuthUser"] = openapiObjectSchema(authUserProperties, authUserRequired)
		if len(g.program.Roles) > 0 {
			schemas["AuthRoleUpdateInput"] = openapiObjectSchema(map[string]any{
				"roles": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
						"enum": g.roleNames(),
					},
					"minItems":    1,
					"uniqueItems": true,
				},
			}, []string{"roles"})
			if g.hasTenantPolicies() {
				schemas["AuthTenantUpdateInput"] = openapiObjectSchema(map[string]any{
					"tenantId": map[string]any{
						"type":      "string",
						"minLength": 1,
						"maxLength": 64,
						"pattern":   "^[A-Za-z0-9_-]+$",
					},
				}, []string{"tenantId"})
			}
			schemas["AuditLog"] = openapiObjectSchema(map[string]any{
				"id":          map[string]any{"type": "string"},
				"actorUserId": map[string]any{"type": "string"},
				"actorRole":   map[string]any{"type": "string"},
				"action":      map[string]any{"type": "string"},
				"resource":    map[string]any{"type": "string"},
				"resourceId":  map[string]any{"type": "string"},
				"summary":     map[string]any{"type": "string"},
				"createdAt":   map[string]any{"type": "string"},
			}, []string{"id", "actorUserId", "actorRole", "action", "resource", "resourceId", "summary", "createdAt"})
		}
		schemas["AuthUserResponse"] = openapiObjectSchema(map[string]any{
			"user": map[string]any{"$ref": "#/components/schemas/AuthUser"},
		}, []string{"user"})
	}
	for _, api := range g.program.APIs {
		method := strings.ToUpper(api.Method)
		if method == "POST" || method == "PUT" || method == "PATCH" {
			if len(api.Body) > 0 {
				schemas[api.Name+"Input"] = openapiAPIBodySchema(api)
			} else {
				schemas[api.Name+"Input"] = map[string]any{
					"type":                 "object",
					"additionalProperties": true,
				}
			}
		}
	}
	for _, action := range g.program.Actions {
		schemas[action.Name+"Input"] = openapiActionInputSchema(action)
		schemas[action.Name+"Response"] = openapiObjectSchema(map[string]any{
			"item":    map[string]any{"$ref": "#/components/schemas/" + action.Source},
			"message": map[string]any{"type": "string"},
		}, []string{"item", "message"})
	}
	for _, entity := range g.program.Entities {
		properties := map[string]any{
			"id":         map[string]any{"type": "string"},
			"archivedAt": map[string]any{"type": []string{"string", "null"}},
		}
		inputProperties := map[string]any{}
		required := []string{"id"}
		inputRequired := []string{}
		policyFields := entityPolicyFieldMap(entity)
		for _, field := range entity.Fields {
			name := field.Name
			if g.isRelationField(field) {
				name = relationIDFieldName(field)
			}
			schema := openapiFieldSchema(field, g.isRelationField(field))
			properties[name] = schema
			if _, policyField := policyFields[field.Name]; !policyField {
				inputProperties[name] = schema
			}
			if hasModifier(field, "required") {
				required = append(required, name)
				if _, policyField := policyFields[field.Name]; !policyField {
					inputRequired = append(inputRequired, name)
				}
			}
		}
		schemas[entity.Name] = openapiObjectSchema(properties, required)
		schemas[entity.Name+"Input"] = openapiObjectSchema(inputProperties, inputRequired)
	}
	return schemas
}

func openapiFieldSchema(field FieldDecl, relation bool) map[string]any {
	if relation {
		return map[string]any{"type": "string"}
	}
	switch field.Type {
	case "number", "integer":
		return map[string]any{"type": "integer"}
	case "decimal", "money":
		return map[string]any{"type": "number"}
	case "boolean":
		return map[string]any{"type": "boolean"}
	case "date", "datetime":
		return map[string]any{"type": "string", "format": "date-time"}
	case "email":
		return map[string]any{"type": "string", "format": "email"}
	case "file":
		return map[string]any{"type": "string", "format": "uri", "x-blacklang-media": "file", "x-blacklang-encoding": "data-url"}
	case "image":
		return map[string]any{"type": "string", "format": "uri", "x-blacklang-media": "image", "x-blacklang-encoding": "data-url", "contentMediaType": "image/*"}
	default:
		return map[string]any{"type": "string"}
	}
}

func openapiAPIUpdateMetadata(update APIUpdateDecl) map[string]any {
	sets := []any{}
	for _, field := range apiUpdateSetFieldNames(update) {
		sets = append(sets, field)
	}
	return map[string]any{
		"source": update.Source,
		"where": map[string]any{
			"field":    update.Where.Field,
			"operator": update.Where.Operator,
			"value":    update.Where.Value,
		},
		"sets": sets,
	}
}

func openapiAPIBodySchema(api APIDecl) map[string]any {
	properties := map[string]any{}
	required := []string{}
	for _, field := range api.Body {
		properties[field.Name] = openapiAPIBodyFieldSchema(field)
		if hasModifier(field, "required") {
			required = append(required, field.Name)
		}
	}
	return openapiObjectSchema(properties, required)
}

func openapiAPIBodyFieldSchema(field FieldDecl) map[string]any {
	schema := openapiAPIParamSchema(field.Type)
	if minValue := modifierValue(field, "min"); minValue != "" {
		schema["minimum"] = numericSchemaValue(minValue)
	}
	if maxValue := modifierValue(field, "max"); maxValue != "" {
		schema["maximum"] = numericSchemaValue(maxValue)
	}
	if minLength, maxLength, ok := fieldLengthBounds(field); ok {
		schema["minLength"] = minLength
		schema["maxLength"] = maxLength
	}
	if pattern := modifierValue(field, "regex"); pattern != "" {
		schema["pattern"] = pattern
	}
	if hasModifier(field, "url") {
		schema["format"] = "uri"
	}
	return schema
}

func numericSchemaValue(value string) any {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}
	return parsed
}

func openapiObjectSchema(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func openapiRequestBody(schemaRef string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": schemaRef},
			},
		},
	}
}

func openapiArrayRequestBody(name string, itemType string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						name: map[string]any{
							"type":  "array",
							"items": map[string]any{"type": itemType},
						},
					},
					"required": []string{name},
				},
			},
		},
	}
}

func openapiIDParameter() map[string]any {
	return map[string]any{
		"name":     "id",
		"in":       "path",
		"required": true,
		"schema":   map[string]any{"type": "string"},
	}
}

func openapiExplicitParameter(param APIParamDecl, location string, required bool) map[string]any {
	return map[string]any{
		"name":     param.Name,
		"in":       location,
		"required": required,
		"schema":   openapiAPIParamSchema(param.Type),
	}
}

func openapiAPIParamSchema(paramType string) map[string]any {
	switch paramType {
	case "number", "integer":
		return map[string]any{"type": "integer"}
	case "decimal", "money":
		return map[string]any{"type": "number"}
	case "boolean":
		return map[string]any{"type": "boolean"}
	case "date":
		return map[string]any{"type": "string", "format": "date"}
	case "datetime":
		return map[string]any{"type": "string", "format": "date-time"}
	case "email":
		return map[string]any{"type": "string", "format": "email"}
	case "file":
		return map[string]any{"type": "string", "format": "uri", "x-blacklang-media": "file", "x-blacklang-encoding": "data-url"}
	case "image":
		return map[string]any{"type": "string", "format": "uri", "x-blacklang-media": "image", "x-blacklang-encoding": "data-url", "contentMediaType": "image/*"}
	default:
		return map[string]any{"type": "string"}
	}
}

func explicitAPIAccess(api APIDecl) string {
	if api.Access != "" {
		return api.Access
	}
	return "private"
}

func openapiJSONResponses(schema map[string]any) map[string]any {
	return map[string]any{
		"200": map[string]any{
			"description": "OK",
			"content": map[string]any{
				"application/json": map[string]any{"schema": schema},
			},
		},
		"400": map[string]any{"description": "Validation error"},
		"404": map[string]any{"description": "Not found"},
	}
}

func (g *webGenerator) prismaSchema() string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("generator client {\n")
	builder.WriteString("  provider = \"prisma-client\"\n")
	builder.WriteString("  output   = \"../src/generated/prisma\"\n")
	builder.WriteString("}\n\n")
	builder.WriteString("datasource db {\n")
	builder.WriteString(fmt.Sprintf("  provider = %q\n", g.prismaProvider()))
	builder.WriteString("}\n\n")
	if g.program.Auth != nil && g.usesPrismaAuthRuntime() {
		builder.WriteString(g.authPrismaModels())
	}
	if len(g.program.Migrations) > 0 {
		builder.WriteString(g.blackMigrationPrismaModel())
	}
	for _, entity := range g.program.Entities {
		builder.WriteString(fmt.Sprintf("model %s {\n", entity.Name))
		builder.WriteString("  id String @id @default(cuid())\n")
		for _, field := range entity.Fields {
			if g.isRelationField(field) {
				builder.WriteString(fmt.Sprintf("  %s String%s%s\n", relationIDFieldName(field), prismaRelationIDOptional(field), prismaAttributes(field)))
				builder.WriteString(fmt.Sprintf("  %s %s%s @relation(%q, fields: [%s], references: [id])\n", field.Name, field.Type, prismaOptional(field), relationName(entity, field), relationIDFieldName(field)))
				continue
			}
			builder.WriteString(fmt.Sprintf("  %s %s%s%s\n", field.Name, prismaType(field.Type), prismaOptional(field), prismaAttributes(field)))
		}
		for _, relation := range g.incomingRelations(entity.Name) {
			builder.WriteString(fmt.Sprintf("  %s %s[] @relation(%q)\n", relationBackFieldName(relation.entity, relation.field), relation.entity.Name, relationName(relation.entity, relation.field)))
		}
		builder.WriteString("  archivedAt DateTime?\n")
		builder.WriteString("  createdAt DateTime @default(now())\n")
		builder.WriteString("  updatedAt DateTime @updatedAt\n")
		for _, index := range entity.Indexes {
			line := g.prismaIndexLine(entity, index)
			if line != "" {
				builder.WriteString(line)
			}
		}
		builder.WriteString("}\n\n")
	}
	return builder.String()
}

func (g *webGenerator) authPrismaModels() string {
	return fmt.Sprintf(`model BlackUser {
  id String @id @default(cuid())
  name String
  email String @unique
  role String @default(%q)
  roles String @default(%q)
%s  passwordHash String
  sessions BlackSession[]
  auditLogs BlackAuditLog[]
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
}

model BlackSession {
  id String @id @default(cuid())
  userId String
  user BlackUser @relation(fields: [userId], references: [id], onDelete: Cascade)
  expiresAt DateTime
  createdAt DateTime @default(now())
  @@index([userId], map: "BlackSession_userId_idx")
}

model BlackAuditLog {
  id String @id @default(cuid())
  actorUserId String
  actorUser BlackUser @relation(fields: [actorUserId], references: [id], onDelete: Cascade)
  actorRole String
  action String
  resource String
  resourceId String
  summary String @default("")
  createdAt DateTime @default(now())
  @@index([actorUserId], map: "BlackAuditLog_actorUserId_idx")
}

`, g.defaultAuthRole(), g.defaultAuthRolesJSON(), g.authPrismaTenantLine())
}

func (g *webGenerator) blackMigrationPrismaModel() string {
	return `model BlackMigration {
  id String @id
  appliedAt DateTime @default(now())
}

`
}

func (g *webGenerator) types() string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	for _, entity := range g.program.Entities {
		builder.WriteString(fmt.Sprintf("export type %s = {\n", entity.Name))
		builder.WriteString("  id: string;\n")
		for _, field := range entity.Fields {
			if g.isRelationField(field) {
				optional := ""
				if !hasModifier(field, "required") {
					optional = "?"
				}
				builder.WriteString(fmt.Sprintf("  %s%s: string;\n", relationIDFieldName(field), optional))
				builder.WriteString(fmt.Sprintf("  %s?: %s | null;\n", field.Name, field.Type))
				continue
			}
			optional := ""
			if hasModifier(field, "optional") || (!hasModifier(field, "required") && !hasModifier(field, "default")) {
				optional = "?"
			}
			builder.WriteString(fmt.Sprintf("  %s%s: %s;\n", field.Name, optional, tsType(field.Type)))
		}
		builder.WriteString("  archivedAt?: string | null;\n")
		builder.WriteString("};\n\n")
	}
	builder.WriteString(g.actionInputTypes())
	return builder.String()
}

func (g *webGenerator) route(page PageDecl, entity EntityDecl) string {
	fileName := strings.ToLower(entity.Name)
	identifier := lowerCamelCase(entity.Name)
	path := strings.ToLower(page.Name)
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("import express from \"express\";\n")
	builder.WriteString("import { prisma } from \"../db\";\n")
	if g.routeUsesTransactionalCustomActionAudit(page) {
		builder.WriteString("import crypto from \"node:crypto\";\n")
	}
	if g.hasRuntimePermissions() {
		builder.WriteString("import { canAccessField, filterWritableFields, requirePageAccess, requirePermission, writeAuditLog } from \"./auth\";\n")
	} else if g.program.Auth != nil && len(page.Access) > 0 {
		builder.WriteString("import { requirePageAccess } from \"./auth\";\n")
	}
	builder.WriteString(fmt.Sprintf("import { %s } from \"../validation/%s\";\n\n", g.validationImportNames(entity), fileName))
	builder.WriteString(fmt.Sprintf("export const %sRouter = express.Router();\n\n", identifier))
	builder.WriteString(fmt.Sprintf("const %sModel = prisma.%s;\n\n", identifier, identifier))
	if g.program.Auth != nil && len(page.Access) > 0 {
		builder.WriteString(fmt.Sprintf("%sRouter.use(\"/%s\", requirePageAccess(%s));\n\n", identifier, path, tsStringArrayLiteral(page.Access)))
	}
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("function sanitize%s(item: any, roles: string[]) {\n", entity.Name))
		builder.WriteString("  if (!item) return item;\n")
		builder.WriteString("  const output = { ...item };\n")
		for _, field := range entity.Fields {
			builder.WriteString(fmt.Sprintf("  if (!canAccessField(roles, \"read\", %q, %q)) delete output.%s;\n", entity.Name, field.Name, g.sqliteColumnName(field)))
			if g.isRelationField(field) {
				builder.WriteString(fmt.Sprintf("  if (!canAccessField(roles, \"read\", %q, %q)) delete output.%s;\n", entity.Name, field.Name, field.Name))
			}
		}
		builder.WriteString("  return output;\n")
		builder.WriteString("}\n\n")
	}
	if g.routeNeedsCurrentUser(entity) {
		builder.WriteString("function currentUser(req: express.Request) {\n")
		builder.WriteString(fmt.Sprintf("  return (req as any).blackUser as %s | undefined;\n", g.currentUserTypeLiteral()))
		builder.WriteString("}\n\n")
	}
	if g.hasRuntimePermissions() {
		builder.WriteString("function currentRoles(req: express.Request) {\n")
		builder.WriteString("  const user = currentUser(req);\n")
		builder.WriteString("  if (!user) return [];\n")
		builder.WriteString("  const roles = user.roles ?? [user.role];\n")
		builder.WriteString("  return roles.length > 0 ? roles : [user.role];\n")
		builder.WriteString("}\n\n")
	}
	builder.WriteString(g.rowPolicyHelpers(entity))
	builder.WriteString(g.routeRelationLoadHelpers(entity))
	builder.WriteString(g.queryRoute(page, entity))
	builder.WriteString(fmt.Sprintf("%sRouter.get(\"/%s\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("read", entity.Name)))
	builder.WriteString("  const includeArchived = req.query.archived === \"all\";\n")
	builder.WriteString(fmt.Sprintf("  const items = await %sModel.findMany({\n", identifier))
	builder.WriteString(fmt.Sprintf("    where: %s,\n", g.rowPolicyWhere(entity, "includeArchived ? {} : { archivedAt: null }")))
	builder.WriteString("    orderBy: { createdAt: \"desc\" }\n")
	builder.WriteString("  });\n")
	builder.WriteString(g.relationLoadAttachStatement(entity, "list", "items", "  "))
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  res.json(items.map((item) => sanitize%s(item, currentRoles(req))));\n", entity.Name))
	} else {
		builder.WriteString("  res.json(items);\n")
	}
	builder.WriteString("});\n\n")
	builder.WriteString(fmt.Sprintf("%sRouter.get(\"/%s/:id\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("read", entity.Name)))
	findMethod := "findUnique"
	if g.hasEntityPolicies(entity) {
		findMethod = "findFirst"
	}
	builder.WriteString(fmt.Sprintf("  const item = await %sModel.%s({\n", identifier, findMethod))
	builder.WriteString(fmt.Sprintf("    where: %s\n", g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")))
	builder.WriteString("  });\n")
	builder.WriteString("  if (!item) {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n\n")
	builder.WriteString(g.relationLoadAttachStatement(entity, "detail", "[item]", "  "))
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  res.json(sanitize%s(item, currentRoles(req)));\n", entity.Name))
	} else {
		builder.WriteString("  res.json(item);\n")
	}
	builder.WriteString("});\n\n")
	if hasAction(page, "create") {
		builder.WriteString(fmt.Sprintf("%sRouter.post(\"/%s\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("create", entity.Name)))
		builder.WriteString(fmt.Sprintf("  const validation = validate%sInput(%s);\n", entity.Name, g.rowPolicyInputExpression(entity, "req.body as Record<string, unknown>")))
		builder.WriteString("  if (!validation.valid) {\n")
		builder.WriteString("    res.status(400).json({ error: validation.errors });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n\n")
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("  const writableValue = filterWritableFields(currentRoles(req), \"create\", %q, validation.value as Record<string, unknown>);\n", entity.Name))
			if g.hasEntityPolicies(entity) {
				builder.WriteString("  const writableCreateValue = stripRowPolicyFields(writableValue);\n")
			}
			if g.hasEntityPolicies(entity) {
				builder.WriteString("  if (Object.keys(writableCreateValue).length === 0) {\n")
			} else {
				builder.WriteString("  if (Object.keys(writableValue).length === 0) {\n")
			}
			builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n\n")
			if g.hasEntityPolicies(entity) {
				builder.WriteString("  Object.assign(writableValue, applyRowPolicyInput(req, {}));\n\n")
			}
		}
		builder.WriteString(fmt.Sprintf("  const item = await %sModel.create({\n", identifier))
		if g.hasRuntimePermissions() {
			builder.WriteString("    data: writableValue as any")
		} else {
			builder.WriteString(fmt.Sprintf("    data: %s as any", g.rowPolicyInputExpression(entity, "validation.value as Record<string, unknown>")))
		}
		builder.WriteString("\n")
		builder.WriteString("  });\n\n")
		builder.WriteString(g.relationLoadAttachStatement(entity, "mutation", "[item]", "  "))
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("  writeAuditLog(currentUser(req), \"create\", %q, item.id, %q);\n", entity.Name, entity.Name+" record created"))
			builder.WriteString(fmt.Sprintf("  res.status(201).json(sanitize%s(item, currentRoles(req)));\n", entity.Name))
		} else {
			builder.WriteString("  res.status(201).json(item);\n")
		}
		builder.WriteString("});\n\n")
	}
	if hasAction(page, "edit") {
		builder.WriteString(fmt.Sprintf("%sRouter.put(\"/%s/:id\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("update", entity.Name)))
		builder.WriteString(fmt.Sprintf("  const validation = validate%sInput(%s);\n", entity.Name, g.rowPolicyInputExpression(entity, "req.body as Record<string, unknown>")))
		builder.WriteString("  if (!validation.valid) {\n")
		builder.WriteString("    res.status(400).json({ error: validation.errors });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n\n")
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("  const writableValue = filterWritableFields(currentRoles(req), \"update\", %q, validation.value as Record<string, unknown>);\n", entity.Name))
			if g.hasEntityPolicies(entity) {
				builder.WriteString("  const writableUpdateValue = stripRowPolicyFields(writableValue);\n")
			}
			if g.hasEntityPolicies(entity) {
				builder.WriteString("  if (Object.keys(writableUpdateValue).length === 0) {\n")
			} else {
				builder.WriteString("  if (Object.keys(writableValue).length === 0) {\n")
			}
			builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n\n")
		}
		builder.WriteString("  try {\n")
		if g.hasEntityPolicies(entity) {
			builder.WriteString(fmt.Sprintf("    const existing = await %sModel.findFirst({ where: %s });\n", identifier, g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")))
			builder.WriteString("    if (!existing) {\n")
			builder.WriteString(fmt.Sprintf("      res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
			builder.WriteString("      return;\n")
			builder.WriteString("    }\n\n")
		}
		builder.WriteString(fmt.Sprintf("    const item = await %sModel.update({\n", identifier))
		if g.hasEntityPolicies(entity) {
			builder.WriteString("      where: { id: existing.id },\n")
		} else {
			builder.WriteString("      where: { id: String(req.params.id) },\n")
		}
		if g.hasRuntimePermissions() {
			if g.hasEntityPolicies(entity) {
				builder.WriteString("      data: writableUpdateValue as any")
			} else {
				builder.WriteString("      data: writableValue as any")
			}
		} else {
			builder.WriteString(fmt.Sprintf("      data: %s as any", g.mutableInputExpression(entity, "validation.value as Record<string, unknown>")))
		}
		builder.WriteString("\n")
		builder.WriteString("    });\n")
		builder.WriteString(g.relationLoadAttachStatement(entity, "mutation", "[item]", "    "))
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"update\", %q, item.id, %q);\n", entity.Name, entity.Name+" record updated"))
			builder.WriteString(fmt.Sprintf("    res.json(sanitize%s(item, currentRoles(req)));\n", entity.Name))
		} else {
			builder.WriteString("    res.json(item);\n")
		}
		builder.WriteString("  } catch {\n")
		builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
		builder.WriteString("  }\n")
		builder.WriteString("});\n\n")
	}
	if hasAction(page, "archive") {
		builder.WriteString(fmt.Sprintf("%sRouter.patch(\"/%s/:id/archive\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("update", entity.Name)))
		builder.WriteString("  try {\n")
		if g.hasEntityPolicies(entity) {
			builder.WriteString(fmt.Sprintf("    const existing = await %sModel.findFirst({ where: %s });\n", identifier, g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")))
			builder.WriteString("    if (!existing) {\n")
			builder.WriteString(fmt.Sprintf("      res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
			builder.WriteString("      return;\n")
			builder.WriteString("    }\n\n")
		}
		builder.WriteString(fmt.Sprintf("    const item = await %sModel.update({\n", identifier))
		if g.hasEntityPolicies(entity) {
			builder.WriteString("      where: { id: existing.id },\n")
		} else {
			builder.WriteString("      where: { id: String(req.params.id) },\n")
		}
		builder.WriteString("      data: { archivedAt: new Date() }")
		builder.WriteString("\n")
		builder.WriteString("    });\n")
		builder.WriteString(g.relationLoadAttachStatement(entity, "mutation", "[item]", "    "))
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"archive\", %q, item.id, %q);\n", entity.Name, entity.Name+" record archived"))
			builder.WriteString(fmt.Sprintf("    res.json(sanitize%s(item, currentRoles(req)));\n", entity.Name))
		} else {
			builder.WriteString("    res.json(item);\n")
		}
		builder.WriteString("  } catch {\n")
		builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
		builder.WriteString("  }\n")
		builder.WriteString("});\n\n")
	}
	if hasAction(page, "restore") {
		builder.WriteString(fmt.Sprintf("%sRouter.patch(\"/%s/:id/restore\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("update", entity.Name)))
		builder.WriteString("  try {\n")
		if g.hasEntityPolicies(entity) {
			builder.WriteString(fmt.Sprintf("    const existing = await %sModel.findFirst({ where: %s });\n", identifier, g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")))
			builder.WriteString("    if (!existing) {\n")
			builder.WriteString(fmt.Sprintf("      res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
			builder.WriteString("      return;\n")
			builder.WriteString("    }\n\n")
		}
		builder.WriteString(fmt.Sprintf("    const item = await %sModel.update({\n", identifier))
		if g.hasEntityPolicies(entity) {
			builder.WriteString("      where: { id: existing.id },\n")
		} else {
			builder.WriteString("      where: { id: String(req.params.id) },\n")
		}
		builder.WriteString("      data: { archivedAt: null }")
		builder.WriteString("\n")
		builder.WriteString("    });\n")
		builder.WriteString(g.relationLoadAttachStatement(entity, "mutation", "[item]", "    "))
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"restore\", %q, item.id, %q);\n", entity.Name, entity.Name+" record restored"))
			builder.WriteString(fmt.Sprintf("    res.json(sanitize%s(item, currentRoles(req)));\n", entity.Name))
		} else {
			builder.WriteString("    res.json(item);\n")
		}
		builder.WriteString("  } catch {\n")
		builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
		builder.WriteString("  }\n")
		builder.WriteString("});\n\n")
	}
	builder.WriteString(g.customActionRoutes(page, entity))
	builder.WriteString(g.workflowTransitionRoutes(page, entity))
	if hasAction(page, "delete") {
		builder.WriteString(fmt.Sprintf("%sRouter.delete(\"/%s\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("delete", entity.Name)))
		builder.WriteString("  const ids = Array.isArray(req.body?.ids)\n")
		builder.WriteString("    ? req.body.ids.filter((id: unknown) => typeof id === \"string\")\n")
		builder.WriteString("    : [];\n\n")
		builder.WriteString("  if (ids.length === 0) {\n")
		builder.WriteString("    res.status(400).json({ error: \"ids are required\" });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n\n")
		builder.WriteString(fmt.Sprintf("  const result = await %sModel.deleteMany({\n", identifier))
		builder.WriteString(fmt.Sprintf("    where: %s\n", g.rowPolicyWhere(entity, "{ id: { in: ids } }")))
		builder.WriteString("  });\n\n")
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("  writeAuditLog(currentUser(req), \"bulkDelete\", %q, ids.join(\",\"), String(result.count) + \" %s records deleted\");\n", entity.Name, strings.ToLower(entity.Name)))
		}
		builder.WriteString("  res.json({ deleted: result.count });\n")
		builder.WriteString("});\n\n")
		builder.WriteString(fmt.Sprintf("%sRouter.delete(\"/%s/:id\", %sasync (req, res) => {\n", identifier, path, g.permissionMiddleware("delete", entity.Name)))
		builder.WriteString("  try {\n")
		if g.hasEntityPolicies(entity) {
			builder.WriteString(fmt.Sprintf("    const result = await %sModel.deleteMany({ where: %s });\n", identifier, g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")))
			builder.WriteString("    if (result.count === 0) {\n")
			builder.WriteString(fmt.Sprintf("      res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
			builder.WriteString("      return;\n")
			builder.WriteString("    }\n")
		} else {
			builder.WriteString(fmt.Sprintf("    await %sModel.delete({\n", identifier))
			builder.WriteString("      where: { id: String(req.params.id) }\n")
			builder.WriteString("    });\n")
		}
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("    writeAuditLog(currentUser(req), \"delete\", %q, String(req.params.id), %q);\n", entity.Name, entity.Name+" record deleted"))
		}
		builder.WriteString("    res.status(204).send();\n")
		builder.WriteString("  } catch {\n")
		builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
		builder.WriteString("  }\n")
		builder.WriteString("});\n")
	}
	return builder.String()
}

func (g *webGenerator) apiClient(page PageDecl, entity EntityDecl) string {
	identifier := lowerCamelCase(entity.Name)
	path := strings.ToLower(page.Name)
	return fmt.Sprintf(`// Generated by BlackLang. Do not edit manually.

import type { %s } from "../types";

export type %sInput = Omit<%s, %s>;

const endpoint = "/api/%s";

function csrfHeaders(): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const token = document.cookie.split("; ").find((item) => item.startsWith("black_csrf="))?.split("=")[1] ?? "";
  if (token) headers["X-CSRF-Token"] = decodeURIComponent(token);
  return headers;
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: "same-origin",
    headers: csrfHeaders(),
    ...options
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || "Request failed");
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export const %sApi = {
%s
  list: (includeArchived = false) =>
    request<%s[]>(includeArchived ? endpoint + "?archived=all" : endpoint),
  get: (id: string) => request<%s>(endpoint + "/" + id),
  create: (input: %sInput) =>
    request<%s>(endpoint, {
      method: "POST",
      body: JSON.stringify(input)
    }),
  update: (id: string, input: %sInput) =>
    request<%s>(endpoint + "/" + id, {
      method: "PUT",
      body: JSON.stringify(input)
    }),
  bulkDelete: (ids: string[]) =>
    request<{ deleted: number }>(endpoint, {
      method: "DELETE",
      body: JSON.stringify({ ids })
    }),
  archive: (id: string) =>
    request<%s>(endpoint + "/" + id + "/archive", {
      method: "PATCH"
    }),
  restore: (id: string) =>
    request<%s>(endpoint + "/" + id + "/restore", {
      method: "PATCH"
    }),
  delete: (id: string) =>
    request<void>(endpoint + "/" + id, {
      method: "DELETE"
    })%s
};
`, g.apiClientTypeImportList(page, entity), entity.Name, entity.Name, g.apiClientInputOmit(entity), path, identifier, g.queryClientMethod(page, entity), entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, entity.Name, g.customActionClientMethods(page, entity)+g.workflowClientMethods(entity))
}

func (g *webGenerator) workflowTransitionRoutes(page PageDecl, entity EntityDecl) string {
	if !g.hasRuntimePermissions() {
		return ""
	}
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	identifier := lowerCamelCase(entity.Name)
	path := strings.ToLower(page.Name)
	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			builder.WriteString(fmt.Sprintf("%sRouter.post(\"/%s/:id/workflow/%s\", %sasync (req, res) => {\n", identifier, path, transition.Name, g.permissionMiddleware("update", entity.Name)))
			if len(transition.Allow) > 0 {
				builder.WriteString("  const user = currentUser(req);\n")
				if containsString(transition.Allow, "authenticated") {
					builder.WriteString("  if (!user) {\n")
				} else {
					builder.WriteString(fmt.Sprintf("  if (!user || !(user.roles ?? [user.role]).some((role) => %s.includes(role))) {\n", tsStringArrayLiteral(transition.Allow)))
				}
				builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
				builder.WriteString("    return;\n")
				builder.WriteString("  }\n\n")
			}
			findMethod := "findUnique"
			whereExpression := "{ id: String(req.params.id) }"
			if g.hasEntityPolicies(entity) {
				findMethod = "findFirst"
				whereExpression = g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")
			}
			builder.WriteString(fmt.Sprintf("  const existing = await %sModel.%s({ where: %s });\n", identifier, findMethod, whereExpression))
			builder.WriteString("  if (!existing) {\n")
			builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n\n")
			builder.WriteString(fmt.Sprintf("  if (String(existing.status ?? \"\") !== %q) {\n", transition.From))
			builder.WriteString(fmt.Sprintf("    res.status(409).json({ error: %q });\n", fmt.Sprintf("Transition %s requires status %s", transition.Name, transition.From)))
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n\n")
			builder.WriteString(fmt.Sprintf("  const item = await %sModel.update({\n", identifier))
			if g.hasEntityPolicies(entity) {
				builder.WriteString("    where: { id: existing.id },\n")
			} else {
				builder.WriteString("    where: { id: String(req.params.id) },\n")
			}
			builder.WriteString(fmt.Sprintf("    data: { status: %q }", transition.To))
			builder.WriteString("\n")
			builder.WriteString("  });\n\n")
			builder.WriteString(g.relationLoadAttachStatement(entity, "mutation", "[item]", "  "))
			builder.WriteString(fmt.Sprintf("  writeAuditLog(currentUser(req), %q, %q, item.id, %q);\n", "workflow."+transition.Name, entity.Name, fmt.Sprintf("%s: %s -> %s", workflow.Name, transition.From, transition.To)))
			builder.WriteString(fmt.Sprintf("  res.json(sanitize%s(item, currentRoles(req)));\n", entity.Name))
			builder.WriteString("});\n\n")
		}
	}
	return builder.String()
}

func (g *webGenerator) workflowClientMethods(entity EntityDecl) string {
	if !g.hasRuntimePermissions() {
		return ""
	}
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			builder.WriteString(fmt.Sprintf(",\n  transition%s: (id: string) =>\n    request<%s>(endpoint + \"/\" + id + \"/workflow/%s\", {\n      method: \"POST\"\n    })", title(transition.Name), entity.Name, transition.Name))
		}
	}
	return builder.String()
}

func (g *webGenerator) workflowsForEntity(entityName string) []WorkflowDecl {
	workflows := []WorkflowDecl{}
	for _, workflow := range g.program.Workflows {
		if workflow.Source == entityName {
			workflows = append(workflows, workflow)
		}
	}
	return workflows
}

func (g *webGenerator) workflowPageActionFunctions(page PageDecl, entity EntityDecl) string {
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	identifier := lowerCamelCase(entity.Name)
	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			functionName := workflowActionFunctionName(transition)
			builder.WriteString(fmt.Sprintf("  async function %s(item: %s) {\n", functionName, entity.Name))
			builder.WriteString(fmt.Sprintf("    if (!canUpdate || item.archivedAt || String(item.status ?? \"\") !== %q) return;\n", transition.From))
			builder.WriteString("    setSaving(true);\n")
			builder.WriteString("    setError(null);\n")
			builder.WriteString("    try {\n")
			builder.WriteString(fmt.Sprintf("      const updated = await %sApi.transition%s(item.id);\n", identifier, title(transition.Name)))
			builder.WriteString(queryPageMutation(page, "      setItems((current) => current.map((existing) => existing.id === item.id ? updated : existing));\n"))
			builder.WriteString("      setSelectedItem((current) => current?.id === item.id ? updated : current);\n")
			builder.WriteString("    } catch (reason: unknown) {\n")
			builder.WriteString(fmt.Sprintf("      setError(reason instanceof Error ? reason.message : %q);\n", "Unable to run transition "+transition.Name))
			builder.WriteString("    } finally {\n")
			builder.WriteString("      setSaving(false);\n")
			builder.WriteString("    }\n")
			builder.WriteString("  }\n\n")
		}
	}
	return builder.String()
}

func (g *webGenerator) workflowPageActionButtons(entity EntityDecl) string {
	workflows := g.workflowsForEntity(entity.Name)
	if len(workflows) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, workflow := range workflows {
		for _, transition := range workflow.Transitions {
			if g.hasRuntimeI18N() {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && !item.archivedAt && String(item.status ?? \"\") === %q && <button className=\"secondary\" type=\"button\" disabled={saving} onClick={() => %s(item)}>{%s}</button>}\n", transition.From, workflowActionFunctionName(transition), g.uiLabelExpression("action."+transition.Name, identifierLabel(transition.Name), "locale")))
			} else {
				builder.WriteString(fmt.Sprintf("                  {canUpdate && !item.archivedAt && String(item.status ?? \"\") === %q && <button className=\"secondary\" type=\"button\" disabled={saving} onClick={() => %s(item)}>%s</button>}\n", transition.From, workflowActionFunctionName(transition), identifierLabel(transition.Name)))
			}
		}
	}
	return builder.String()
}

func (g *webGenerator) validation(entity EntityDecl) string {
	var builder strings.Builder
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("type ValidationResult<T> =\n")
	builder.WriteString("  | { valid: true; value: T }\n")
	builder.WriteString("  | { valid: false; errors: string[] };\n\n")
	builder.WriteString(fmt.Sprintf("export function validate%sInput(input: any): ValidationResult<Record<string, unknown>> {\n", entity.Name))
	builder.WriteString("  const errors: string[] = [];\n")
	builder.WriteString("  const value: any = {};\n\n")
	for _, field := range entity.Fields {
		builder.WriteString(g.validationFieldBlock(field))
	}
	builder.WriteString(g.entityValidationBlocks(entity))
	builder.WriteString("\n  if (errors.length > 0) {\n")
	builder.WriteString("    return { valid: false, errors };\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  return { valid: true, value };\n")
	builder.WriteString("}\n")
	builder.WriteString(g.actionValidationFunctions(entity))
	return builder.String()
}

func (g *webGenerator) validationFieldBlock(field FieldDecl) string {
	if !g.isRelationField(field) {
		return validationFieldBlock(field)
	}

	var builder strings.Builder
	name := field.Name
	valueName := relationIDFieldName(field)
	builder.WriteString(fmt.Sprintf("  if (input.%s === undefined || input.%s === null || input.%s === \"\") {\n", valueName, valueName, valueName))
	if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.push(%q);\n", fieldValidationMessage(field, name+" is required")))
	} else {
		builder.WriteString(fmt.Sprintf("    value.%s = undefined;\n", valueName))
	}
	builder.WriteString("  } else {\n")
	builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", valueName, valueName))
	builder.WriteString("  }\n\n")
	return builder.String()
}

func (g *webGenerator) page(page PageDecl, entity EntityDecl) string {
	var builder strings.Builder
	entityAPI := lowerCamelCase(entity.Name)
	builder.WriteString("// Generated by BlackLang. Do not edit manually.\n\n")
	if page.Query != "" {
		builder.WriteString("import { useEffect, useMemo, useRef, useState } from \"react\";\n")
	} else {
		builder.WriteString("import { useEffect, useMemo, useState } from \"react\";\n")
	}
	builder.WriteString("import type { FormEvent } from \"react\";\n")
	builder.WriteString(fmt.Sprintf("import { %sApi } from \"../api/%s\";\n", entityAPI, g.pageModuleName(page)))
	importedRelationAPIs := map[string]bool{}
	for _, field := range g.relationFields(entity) {
		if importedRelationAPIs[field.Type] || field.Type == entity.Name {
			continue
		}
		builder.WriteString(fmt.Sprintf("import { %sApi } from \"../api/%s\";\n", lowerCamelCase(field.Type), strings.ToLower(field.Type)))
		importedRelationAPIs[field.Type] = true
	}
	for _, component := range g.componentsForPage(page, entity) {
		builder.WriteString(fmt.Sprintf("import { %s } from \"../components/%s\";\n", component.Name, component.Name))
	}
	builder.WriteString(fmt.Sprintf("import type { %s } from \"../types\";\n\n", g.typeImportList(page, entity)))
	builder.WriteString("type FormState = Record<string, string>;\n\n")
	builder.WriteString("type FormErrors = Record<string, string>;\n\n")
	if g.pageQueryHasAggregates(page) {
		builder.WriteString("type QuerySummary = Record<string, number | null>;\n\n")
	}
	builder.WriteString("type PagePermissions = {\n")
	builder.WriteString("  read: boolean;\n")
	builder.WriteString("  create: boolean;\n")
	builder.WriteString("  update: boolean;\n")
	builder.WriteString("  delete: boolean;\n")
	builder.WriteString("  fields: Record<string, boolean>;\n")
	builder.WriteString("  writeFields: Record<string, boolean>;\n")
	if g.hasCustomActions() {
		builder.WriteString("  customActions: Record<string, boolean>;\n")
	}
	builder.WriteString("};\n\n")
	if g.hasCustomActions() {
		builder.WriteString("const defaultPermissions: PagePermissions = { read: true, create: true, update: true, delete: true, fields: {}, writeFields: {}, customActions: {} };\n\n")
	} else {
		builder.WriteString("const defaultPermissions: PagePermissions = { read: true, create: true, update: true, delete: true, fields: {}, writeFields: {} };\n\n")
	}
	builder.WriteString("type PageProps = {\n")
	builder.WriteString("  onNavigate?: (page: string) => void;\n")
	builder.WriteString("  permissions?: PagePermissions;\n")
	if g.hasRuntimeI18N() {
		builder.WriteString("  locale?: string;\n")
	}
	builder.WriteString("};\n\n")
	builder.WriteString(fmt.Sprintf("const emptyForm: FormState = %s;\n\n", formStateLiteral(page.Form.Fields, entity)))
	if g.hasRuntimeI18N() {
		builder.WriteString(g.pageI18NHelpers(entity))
	}
	builder.WriteString(g.querySummaryHelpers(page))
	builder.WriteString(g.formValidationFunction(page, entity))
	builder.WriteString(g.customActionFormValidationFunctions(page))
	if len(entity.ComputedFields) > 0 {
		builder.WriteString(g.computedFieldHelpers(entity))
	}
	builder.WriteString(g.mediaFieldHelpers(entity))
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("export function %sPage({ onNavigate, permissions = defaultPermissions, locale = %q }: PageProps) {\n", page.Name, g.program.I18N.Default))
	} else {
		builder.WriteString(fmt.Sprintf("export function %sPage({ onNavigate, permissions = defaultPermissions }: PageProps) {\n", page.Name))
	}
	builder.WriteString(fmt.Sprintf("  const [items, setItems] = useState<%s[]>([]);\n", entity.Name))
	builder.WriteString("  const [search, setSearch] = useState(\"\");\n\n")
	builder.WriteString(g.viewTabsStateHook(page))
	if g.formSectionUsesOverlay(page) {
		builder.WriteString("  const [formPanelOpen, setFormPanelOpen] = useState(false);\n\n")
	}
	if len(page.Table.Filters) > 0 {
		builder.WriteString(fmt.Sprintf("  const [filters, setFilters] = useState<Record<string, string>>(%s);\n", tableFiltersLiteral(page.Table.Filters)))
	}
	builder.WriteString("  const [showArchived, setShowArchived] = useState(false);\n")
	if page.Query != "" {
		builder.WriteString("  const [queryRevision, setQueryRevision] = useState(0);\n")
		builder.WriteString("  const queryRequestVersion = useRef(0);\n")
	}
	if g.pageQueryHasAggregates(page) {
		builder.WriteString(fmt.Sprintf("  const [querySummary, setQuerySummary] = useState<QuerySummary>(%s);\n", g.querySummaryInitialState(page)))
	}
	builder.WriteString(fmt.Sprintf("  const [visibleColumns, setVisibleColumns] = useState<Record<string, boolean>>(%s);\n", columnVisibilityLiteral(page.Table.Columns)))
	if page.Table.Paginate > 0 {
		builder.WriteString("  const [currentPage, setCurrentPage] = useState(1);\n")
	}
	builder.WriteString("  const [form, setForm] = useState<FormState>(emptyForm);\n")
	builder.WriteString("  const [touchedFields, setTouchedFields] = useState<Record<string, boolean>>({});\n")
	builder.WriteString("  const [submitted, setSubmitted] = useState(false);\n")
	builder.WriteString("  const [editingId, setEditingId] = useState<string | null>(null);\n\n")
	builder.WriteString(fmt.Sprintf("  const [selectedItem, setSelectedItem] = useState<%s | null>(null);\n", entity.Name))
	for _, field := range g.relationFields(entity) {
		builder.WriteString(fmt.Sprintf("  const [%s, set%s] = useState<%s[]>([]);\n", relationOptionsStateName(field), title(relationOptionsStateName(field)), field.Type))
	}
	state, hasState := g.stateForPage(page)
	if hasState {
		builder.WriteString(g.pageStateHooks(state))
	}
	builder.WriteString("  const [selectedIds, setSelectedIds] = useState<string[]>([]);\n")
	builder.WriteString("  const [loading, setLoading] = useState(true);\n")
	builder.WriteString("  const [reading, setReading] = useState(false);\n")
	builder.WriteString("  const [saving, setSaving] = useState(false);\n")
	builder.WriteString("  const [error, setError] = useState<string | null>(null);\n\n")
	if len(customActionsForPage(g.program, page)) > 0 {
		builder.WriteString("  const [notice, setNotice] = useState<string | null>(null);\n")
	}
	builder.WriteString(g.customActionPageState(page))
	builder.WriteString("  useEffect(() => {\n")
	builder.WriteString("    let active = true;\n")
	activeRequest := "active"
	if page.Query != "" {
		builder.WriteString("    const requestVersion = ++queryRequestVersion.current;\n")
		builder.WriteString("    setItems([]);\n")
		builder.WriteString("    setSelectedIds([]);\n")
		if g.pageQueryHasAggregates(page) {
			builder.WriteString(fmt.Sprintf("    setQuerySummary(%s);\n", g.querySummaryInitialState(page)))
		}
		activeRequest = "active && requestVersion === queryRequestVersion.current"
	}
	builder.WriteString("    setLoading(true);\n")
	builder.WriteString("    setError(null);\n")
	listMethod := "list"
	if page.Query != "" {
		listMethod = "queryList"
	}
	builder.WriteString(fmt.Sprintf("    %sApi.%s(showArchived)\n", entityAPI, listMethod))
	builder.WriteString("      .then((records) => {\n")
	builder.WriteString(fmt.Sprintf("        if (%s) setItems(records);\n", activeRequest))
	builder.WriteString("      })\n")
	builder.WriteString("      .catch((reason: unknown) => {\n")
	builder.WriteString(fmt.Sprintf("        if (%s) setError(reason instanceof Error ? reason.message : \"Unable to load records\");\n", activeRequest))
	builder.WriteString("      })\n")
	builder.WriteString("      .finally(() => {\n")
	builder.WriteString(fmt.Sprintf("        if (%s) setLoading(false);\n", activeRequest))
	builder.WriteString("      });\n\n")
	builder.WriteString(g.querySummaryLoadStatement(page, entityAPI, activeRequest))
	builder.WriteString("    return () => {\n")
	builder.WriteString("      active = false;\n")
	builder.WriteString("    };\n")
	if page.Query != "" {
		builder.WriteString("  }, [showArchived, queryRevision]);\n\n")
		builder.WriteString("  function refreshQuery() {\n")
		builder.WriteString("    queryRequestVersion.current += 1;\n")
		builder.WriteString("    setLoading(true);\n")
		builder.WriteString("    setItems([]);\n")
		builder.WriteString("    setSelectedIds([]);\n")
		if g.pageQueryHasAggregates(page) {
			builder.WriteString(fmt.Sprintf("    setQuerySummary(%s);\n", g.querySummaryInitialState(page)))
		}
		builder.WriteString("    setQueryRevision((current) => current + 1);\n")
		builder.WriteString("  }\n\n")
	} else {
		builder.WriteString("  }, [showArchived]);\n\n")
	}
	for _, field := range g.relationFields(entity) {
		builder.WriteString("  useEffect(() => {\n")
		builder.WriteString("    let active = true;\n")
		builder.WriteString(fmt.Sprintf("    %sApi.list()\n", lowerCamelCase(field.Type)))
		builder.WriteString("      .then((records) => {\n")
		builder.WriteString(fmt.Sprintf("        if (active) set%s(records);\n", title(relationOptionsStateName(field))))
		builder.WriteString("      })\n")
		builder.WriteString("      .catch((reason: unknown) => {\n")
		builder.WriteString("        if (active) setError(reason instanceof Error ? reason.message : \"Unable to load relation options\");\n")
		builder.WriteString("      });\n\n")
		builder.WriteString("    return () => {\n")
		builder.WriteString("      active = false;\n")
		builder.WriteString("    };\n")
		builder.WriteString("  }, []);\n\n")
	}
	if page.Table.Paginate > 0 {
		builder.WriteString("  useEffect(() => {\n")
		builder.WriteString("    setCurrentPage(1);\n")
		builder.WriteString(fmt.Sprintf("  }, [%s]);\n\n", g.tableStateDependencyList(page, "search", "showArchived")))
	}
	builder.WriteString("  const visibleItems = useMemo(() => {\n")
	builder.WriteString("    const query = search.trim().toLowerCase();\n")
	builder.WriteString("    const searchedItems = query ? items.filter((item) =>\n")
	builder.WriteString(g.searchExpression(page.Table.Search, entity))
	builder.WriteString("    ) : items;\n")
	builder.WriteString(g.filterExpression(page.Table.Filters, entity))
	builder.WriteString(g.sortExpression(page.Table.Sort, entity))
	builder.WriteString(fmt.Sprintf("  }, [%s]);\n\n", g.tableStateDependencyList(page, "items", "search")))
	if page.Table.Paginate > 0 {
		builder.WriteString(fmt.Sprintf("  const pageSize = %d;\n", page.Table.Paginate))
		builder.WriteString("  const totalPages = Math.max(1, Math.ceil(visibleItems.length / pageSize));\n")
		builder.WriteString("  const safeCurrentPage = Math.min(currentPage, totalPages);\n")
		builder.WriteString("  const paginatedItems = useMemo(() => {\n")
		builder.WriteString("    const start = (safeCurrentPage - 1) * pageSize;\n")
		builder.WriteString("    return visibleItems.slice(start, start + pageSize);\n")
		builder.WriteString("  }, [visibleItems, safeCurrentPage]);\n\n")
	} else {
		builder.WriteString("  const paginatedItems = visibleItems;\n\n")
	}
	builder.WriteString("  const visibleItemIds = useMemo(() => paginatedItems.map((item) => item.id), [paginatedItems]);\n")
	builder.WriteString("  const allVisibleSelected = visibleItemIds.length > 0 && visibleItemIds.every((id) => selectedIds.includes(id));\n\n")
	builder.WriteString(fmt.Sprintf("  const visibleColumnCount = %s;\n", g.visibleColumnCountExpression(page, entity)))
	builder.WriteString(fmt.Sprintf("  const tableColspan = visibleColumnCount + %d;\n\n", g.staticTableExtraColumns(page, entity)))
	builder.WriteString("  const canCreate = permissions.create;\n")
	builder.WriteString("  const canUpdate = permissions.update;\n")
	builder.WriteString("  const canDelete = permissions.delete;\n")
	builder.WriteString("  const canSave = editingId ? canUpdate : canCreate;\n\n")
	builder.WriteString(g.componentViewSectionState(page))
	builder.WriteString(g.customActionPageDerivedState(page))
	if g.hasRuntimeI18N() {
		builder.WriteString("  const formErrors = useMemo(() => validateForm(form, locale), [form, locale]);\n")
	} else {
		builder.WriteString("  const formErrors = useMemo(() => validateForm(form), [form]);\n")
	}
	builder.WriteString("  const visibleFormErrors = useMemo(() => Object.fromEntries(Object.entries(formErrors).filter(([field]) => touchedFields[field] || submitted)), [formErrors, touchedFields, submitted]);\n\n")
	builder.WriteString(fmt.Sprintf("  const missingRequiredRelations = %s;\n\n", g.missingRequiredRelationsExpression(page, entity)))
	builder.WriteString("  function updateField(field: string, value: string) {\n")
	builder.WriteString("    setForm((current) => ({ ...current, [field]: value }));\n")
	builder.WriteString("    setTouchedFields((current) => ({ ...current, [field]: true }));\n")
	builder.WriteString("  }\n\n")
	builder.WriteString(g.fileInputHelpers(page, entity))
	builder.WriteString("  function toggleItemSelection(id: string) {\n")
	builder.WriteString("    setSelectedIds((current) => current.includes(id) ? current.filter((value) => value !== id) : [...current, id]);\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  function toggleVisibleSelection() {\n")
	builder.WriteString("    if (allVisibleSelected) {\n")
	builder.WriteString("      setSelectedIds((current) => current.filter((id) => !visibleItemIds.includes(id)));\n")
	builder.WriteString("      return;\n")
	builder.WriteString("    }\n\n")
	builder.WriteString("    setSelectedIds((current) => Array.from(new Set([...current, ...visibleItemIds])));\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  function toggleColumn(column: string) {\n")
	builder.WriteString("    setVisibleColumns((current) => ({ ...current, [column]: !current[column] }));\n")
	builder.WriteString("  }\n\n")
	if len(page.Table.Filters) > 0 {
		builder.WriteString("  function updateFilter(field: string, value: string) {\n")
		builder.WriteString("    setFilters((current) => ({ ...current, [field]: value }));\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  function resetForm() {\n")
	builder.WriteString("    setForm(emptyForm);\n")
	builder.WriteString("    setTouchedFields({});\n")
	builder.WriteString("    setSubmitted(false);\n")
	builder.WriteString("    setEditingId(null);\n")
	if g.formSectionUsesOverlay(page) {
		builder.WriteString("    setFormPanelOpen(false);\n")
	}
	if hasState {
		builder.WriteString(g.resetPageModals(state))
	}
	builder.WriteString(g.viewTriggerEffects(page, "close", "", entity, "    "))
	builder.WriteString("  }\n\n")
	if hasState {
		builder.WriteString(g.pageModalHelpers(state))
	}
	if g.createStartButtonEnabled(page) {
		builder.WriteString("  function startCreateItem() {\n")
		builder.WriteString("    resetForm();\n")
		if g.formSectionUsesOverlay(page) {
			builder.WriteString("    setFormPanelOpen(true);\n")
		}
		builder.WriteString(g.viewTriggerEffects(page, "createStart", "", entity, "    "))
		builder.WriteString("  }\n\n")
	}
	if g.detailSectionUsesOverlay(page) {
		builder.WriteString("  function closeDetailPanel() {\n")
		builder.WriteString("    setSelectedItem(null);\n")
		builder.WriteString(g.viewTriggerEffects(page, "close", "", entity, "    "))
		builder.WriteString("  }\n\n")
	}
	builder.WriteString("  async function viewItem(id: string) {\n")
	builder.WriteString("    setReading(true);\n")
	builder.WriteString("    setError(null);\n")
	builder.WriteString("    try {\n")
	builder.WriteString(fmt.Sprintf("      const item = await %sApi.get(id);\n", entityAPI))
	builder.WriteString("      setSelectedItem(item);\n")
	builder.WriteString(g.viewTriggerEffects(page, "rowSelect", "item", entity, "      "))
	builder.WriteString("    } catch (reason: unknown) {\n")
	builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to load record details\");\n")
	builder.WriteString("    } finally {\n")
	builder.WriteString("      setReading(false);\n")
	builder.WriteString("    }\n")
	builder.WriteString("  }\n\n")
	if hasAction(page, "create") || hasAction(page, "edit") {
		builder.WriteString("  async function saveItem(event: FormEvent<HTMLFormElement>) {\n")
		builder.WriteString("    event.preventDefault();\n")
		builder.WriteString("    if (!canSave) return;\n")
		builder.WriteString("    setSubmitted(true);\n")
		builder.WriteString("    setError(null);\n")
		if g.hasRuntimeI18N() {
			builder.WriteString("    const nextErrors = validateForm(form, locale);\n")
		} else {
			builder.WriteString("    const nextErrors = validateForm(form);\n")
		}
		builder.WriteString("    if (Object.keys(nextErrors).length > 0) return;\n\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    const input = {\n")
		for _, fieldName := range page.Form.Fields {
			field, ok := findField(entity, fieldName)
			if !ok {
				continue
			}
			inputName := field.Name
			inputValue := formValueExpression(field)
			if g.isRelationField(field) {
				inputName = relationIDFieldName(field)
				inputValue = fmt.Sprintf("form.%s", field.Name)
			}
			builder.WriteString(fmt.Sprintf("      %s: %s,\n", inputName, inputValue))
		}
		builder.WriteString("    };\n\n")
		builder.WriteString("    try {\n")
		builder.WriteString("      if (editingId) {\n")
		builder.WriteString(fmt.Sprintf("        const saved = await %sApi.update(editingId, input);\n", entityAPI))
		builder.WriteString(queryPageMutation(page, "        setItems((current) => current.map((existing) => existing.id === editingId ? saved : existing));\n"))
		builder.WriteString("        setSelectedItem((current) => current?.id === saved.id ? saved : current);\n")
		builder.WriteString("        resetForm();\n")
		builder.WriteString(g.viewTriggerEffects(page, "saveSuccess", "saved", entity, "        "))
		builder.WriteString("      } else {\n")
		builder.WriteString(fmt.Sprintf("        const saved = await %sApi.create(input);\n", entityAPI))
		builder.WriteString(queryPageMutation(page, "        setItems((current) => [...current, saved]);\n"))
		builder.WriteString("        resetForm();\n")
		builder.WriteString(g.viewTriggerEffects(page, "saveSuccess", "saved", entity, "        "))
		builder.WriteString("      }\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to save record\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "edit") {
		builder.WriteString(fmt.Sprintf("  function editItem(item: %s) {\n", entity.Name))
		builder.WriteString("    if (!canUpdate) return;\n")
		if hasState {
			builder.WriteString(g.closePageModals(state))
		}
		if g.formSectionUsesOverlay(page) {
			builder.WriteString("    setFormPanelOpen(true);\n")
		}
		builder.WriteString(g.viewTriggerEffects(page, "editStart", "item", entity, "    "))
		builder.WriteString("    setEditingId(item.id);\n")
		builder.WriteString("    setTouchedFields({});\n")
		builder.WriteString("    setSubmitted(false);\n")
		builder.WriteString(g.setFormFromItemStatements(page, entity, "item", "    "))
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "delete") {
		builder.WriteString("  async function deleteItem(id: string) {\n")
		builder.WriteString("    if (!canDelete) return;\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      await %sApi.delete(id);\n", entityAPI))
		builder.WriteString(queryPageMutation(page, "      setItems((current) => current.filter((item) => item.id !== id));\n"))
		builder.WriteString("      setSelectedIds((current) => current.filter((value) => value !== id));\n")
		builder.WriteString("      setSelectedItem((current) => current?.id === id ? null : current);\n")
		builder.WriteString("      if (editingId === id) resetForm();\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to delete record\");\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "delete") {
		builder.WriteString("  async function bulkDeleteSelected() {\n")
		builder.WriteString("    if (!canDelete) return;\n")
		builder.WriteString("    if (selectedIds.length === 0) return;\n")
		builder.WriteString("    const ids = [...selectedIds];\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      await %sApi.bulkDelete(ids);\n", entityAPI))
		builder.WriteString(queryPageMutation(page, "      setItems((current) => current.filter((item) => !ids.includes(item.id)));\n"))
		builder.WriteString("      setSelectedIds((current) => current.filter((id) => !ids.includes(id)));\n")
		builder.WriteString("      setSelectedItem((current) => current && ids.includes(current.id) ? null : current);\n")
		builder.WriteString("      if (editingId && ids.includes(editingId)) resetForm();\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to delete selected records\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "archive") {
		builder.WriteString(fmt.Sprintf("  async function archiveItem(item: %s) {\n", entity.Name))
		builder.WriteString("    if (!canUpdate) return;\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      const archived = await %sApi.archive(item.id);\n", entityAPI))
		builder.WriteString(queryPageMutation(page, "      if (showArchived) {\n        setItems((current) => current.map((existing) => existing.id === item.id ? archived : existing));\n      } else {\n        setItems((current) => current.filter((existing) => existing.id !== item.id));\n      }\n"))
		builder.WriteString("      setSelectedIds((current) => current.filter((id) => id !== item.id));\n")
		builder.WriteString("      setSelectedItem((current) => current?.id === item.id ? archived : current);\n")
		builder.WriteString("      if (editingId === item.id) resetForm();\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to archive record\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	if hasAction(page, "restore") {
		builder.WriteString(fmt.Sprintf("  async function restoreItem(item: %s) {\n", entity.Name))
		builder.WriteString("    if (!canUpdate) return;\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      const restored = await %sApi.restore(item.id);\n", entityAPI))
		builder.WriteString(queryPageMutation(page, "      setItems((current) => current.map((existing) => existing.id === item.id ? restored : existing));\n"))
		builder.WriteString("      setSelectedItem((current) => current?.id === item.id ? restored : current);\n")
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString("      setError(reason instanceof Error ? reason.message : \"Unable to restore record\");\n")
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	builder.WriteString(g.customActionPageFunctions(page, entity))
	builder.WriteString(g.workflowPageActionFunctions(page, entity))
	builder.WriteString("  return (\n")
	builder.WriteString(fmt.Sprintf("    <main className=\"%s\">\n", g.pageViewClassName(page)))
	builder.WriteString("      <header>\n")
	if g.hasRuntimeI18N() {
		builder.WriteString(fmt.Sprintf("        <h1>{%s}</h1>\n", g.uiLabelExpression("page."+page.Name, page.Name, "locale")))
		builder.WriteString(fmt.Sprintf("        <span>{%s} {%s}: %s</span>\n", g.uiLabelExpression("app.title", g.program.App.Name, "locale"), g.uiLabelExpression("app.source", "source", "locale"), entity.Name))
	} else {
		builder.WriteString(fmt.Sprintf("        <h1>%s</h1>\n", page.Name))
		builder.WriteString(fmt.Sprintf("        <span>%s source: %s</span>\n", g.program.App.Name, entity.Name))
	}
	builder.WriteString("      </header>\n\n")
	builder.WriteString(g.viewTabsMarkup(page))
	for _, sectionName := range pageViewOrder(page) {
		builder.WriteString(g.pageDOMSectionMarkup(page, entity, sectionName, state, hasState))
	}
	builder.WriteString("    </main>\n")
	builder.WriteString("  );\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (g *webGenerator) searchExpression(fields []string, entity EntityDecl) string {
	if len(fields) == 0 {
		return "      false\n"
	}
	lines := []string{}
	for _, fieldName := range fields {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("      %s.toLowerCase().includes(query)", g.itemDisplayExpression("item", field)))
	}
	if len(lines) == 0 {
		return "      false\n"
	}
	return strings.Join(lines, " ||\n") + "\n"
}

func (g *webGenerator) sortExpression(sort SortDecl, entity EntityDecl) string {
	if sort.Field == "" {
		return "    return filteredItems;\n"
	}
	field, ok := findField(entity, sort.Field)
	if !ok {
		return "    return filteredItems;\n"
	}
	direction := 1
	if sort.Direction == "desc" {
		direction = -1
	}
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return fmt.Sprintf("    return [...filteredItems].sort((left, right) => (Number(left.%s ?? 0) - Number(right.%s ?? 0)) * %d);\n", field.Name, field.Name, direction)
	default:
		return fmt.Sprintf("    return [...filteredItems].sort((left, right) => %s.localeCompare(%s) * %d);\n", g.itemDisplayExpression("left", field), g.itemDisplayExpression("right", field), direction)
	}
}

func (g *webGenerator) filterExpression(filters []string, entity EntityDecl) string {
	if len(filters) == 0 {
		return "    const filteredItems = searchedItems;\n"
	}
	lines := []string{}
	for _, fieldName := range filters {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("      (filters.%s.trim() === \"\" || %s.toLowerCase().includes(filters.%s.trim().toLowerCase()))", field.Name, g.itemDisplayExpression("item", field), field.Name))
	}
	if len(lines) == 0 {
		return "    const filteredItems = searchedItems;\n"
	}
	return "    const filteredItems = searchedItems.filter((item) =>\n" + strings.Join(lines, " &&\n") + "\n    );\n"
}

func (g *webGenerator) paginationToolbar(pageSize int, indent string) string {
	if pageSize <= 0 {
		return ""
	}
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("%s<div className=\"toolbar pagination\">\n%s  <button className=\"secondary\" type=\"button\" disabled={safeCurrentPage <= 1} onClick={() => setCurrentPage((page) => Math.max(1, page - 1))}>{%s}</button>\n%s  <span className=\"muted\">{%s} {safeCurrentPage} {%s} {totalPages}</span>\n%s  <button className=\"secondary\" type=\"button\" disabled={safeCurrentPage >= totalPages} onClick={() => setCurrentPage((page) => Math.min(totalPages, page + 1))}>{%s}</button>\n%s</div>\n", indent, indent, g.uiLabelExpression("table.previous", "Previous", "locale"), indent, g.uiLabelExpression("table.page", "Page", "locale"), g.uiLabelExpression("table.of", "of", "locale"), indent, g.uiLabelExpression("table.next", "Next", "locale"), indent)
	}
	return fmt.Sprintf("%s<div className=\"toolbar pagination\">\n%s  <button className=\"secondary\" type=\"button\" disabled={safeCurrentPage <= 1} onClick={() => setCurrentPage((page) => Math.max(1, page - 1))}>Previous</button>\n%s  <span className=\"muted\">Page {safeCurrentPage} of {totalPages}</span>\n%s  <button className=\"secondary\" type=\"button\" disabled={safeCurrentPage >= totalPages} onClick={() => setCurrentPage((page) => Math.min(totalPages, page + 1))}>Next</button>\n%s</div>\n", indent, indent, indent, indent, indent)
}

func (g *webGenerator) formValidationFunction(page PageDecl, entity EntityDecl) string {
	var builder strings.Builder
	if g.hasRuntimeI18N() {
		builder.WriteString("function validateForm(form: FormState, locale: string): FormErrors {\n")
	} else {
		builder.WriteString("function validateForm(form: FormState): FormErrors {\n")
	}
	builder.WriteString("  const errors: FormErrors = {};\n\n")
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		builder.WriteString(g.formValidationFieldBlock(entity, field))
	}
	builder.WriteString(g.formEntityValidationBlocks(entity, page.Form.Fields))
	builder.WriteString("  return errors;\n")
	builder.WriteString("}\n\n")
	return builder.String()
}

func (g *webGenerator) formValidationFieldBlock(entity EntityDecl, field FieldDecl) string {
	var builder strings.Builder
	name := field.Name
	label := g.fieldLabel(entity, field)
	if g.isRelationField(field) {
		if hasModifier(field, "required") {
			builder.WriteString(fmt.Sprintf("  if (form.%s.trim() === \"\") {\n", name))
			builder.WriteString(fmt.Sprintf("    errors.%s = %s;\n", name, g.fieldValidationMessageExpression(entity, field, label+" is required")))
			builder.WriteString("  }\n\n")
		}
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("  if (form.%s.trim() === \"\") {\n", name))
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		builder.WriteString("    // Empty input will use the field default on save.\n")
	} else if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.%s = %s;\n", name, g.fieldValidationMessageExpression(entity, field, label+" is required")))
	}
	builder.WriteString("  } else {\n")
	switch field.Type {
	case "number", "integer":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(form.%s);\n", name))
		builder.WriteString(fmt.Sprintf("    if (!Number.isInteger(parsed) || parsed < -2147483648 || parsed > 2147483647) errors.%s = %s;\n", name, g.fieldValidationMessageExpression(entity, field, label+" must be a whole number")))
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed < %s) errors.%s = %s;\n", minValue, name, g.fieldValidationMessageExpression(entity, field, label+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed > %s) errors.%s = %s;\n", maxValue, name, g.fieldValidationMessageExpression(entity, field, label+" must be at most "+maxValue)))
		}
	case "decimal", "money":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(form.%s);\n", name))
		builder.WriteString(fmt.Sprintf("    if (!Number.isFinite(parsed)) errors.%s = %s;\n", name, g.fieldValidationMessageExpression(entity, field, label+" must be a number")))
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed < %s) errors.%s = %s;\n", minValue, name, g.fieldValidationMessageExpression(entity, field, label+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed > %s) errors.%s = %s;\n", maxValue, name, g.fieldValidationMessageExpression(entity, field, label+" must be at most "+maxValue)))
		}
	case "email":
		builder.WriteString(fmt.Sprintf("    if (!form.%s.includes(\"@\")) errors.%s = %s;\n", name, name, g.fieldValidationMessageExpression(entity, field, label+" must be an email")))
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    else if (form.%s.length < %d || form.%s.length > %d) errors.%s = %s;\n", name, minLength, name, maxLength, name, g.fieldValidationMessageExpression(entity, field, fmt.Sprintf("%s length must be between %d and %d", label, minLength, maxLength))))
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!errors.%s && !(new RegExp(%q)).test(form.%s)) errors.%s = %s;\n", name, pattern, name, name, g.fieldValidationMessageExpression(entity, field, label+" has an invalid format")))
		}
	case "file", "image":
		builder.WriteString(fmt.Sprintf("    if (!%s) errors.%s = %s;\n", mediaValidationExpression("form."+name, field.Type), name, g.fieldValidationMessageExpression(entity, field, label+" must be a valid "+field.Type)))
	default:
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    if (form.%s.length < %d || form.%s.length > %d) errors.%s = %s;\n", name, minLength, name, maxLength, name, g.fieldValidationMessageExpression(entity, field, fmt.Sprintf("%s length must be between %d and %d", label, minLength, maxLength))))
		}
		if hasModifier(field, "url") {
			builder.WriteString(fmt.Sprintf("    if (!errors.%s) {\n", name))
			builder.WriteString(fmt.Sprintf("      try { new URL(form.%s); } catch { errors.%s = %s; }\n", name, name, g.fieldValidationMessageExpression(entity, field, label+" must be a valid URL")))
			builder.WriteString("    }\n")
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!errors.%s && !(new RegExp(%q)).test(form.%s)) errors.%s = %s;\n", name, pattern, name, name, g.fieldValidationMessageExpression(entity, field, label+" has an invalid format")))
		}
	}
	builder.WriteString("  }\n\n")
	return builder.String()
}

func emptyStateColspan(page PageDecl) int {
	colspan := len(page.Table.Columns)
	if len(page.Actions) > 0 {
		colspan++
	}
	if hasAction(page, "delete") {
		colspan++
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		colspan++
	}
	return colspan
}

func (g *webGenerator) hasRowActions(page PageDecl, entity EntityDecl) bool {
	return len(page.Actions) > 0 || len(g.workflowsForEntity(entity.Name)) > 0
}

func staticTableExtraColumns(page PageDecl) int {
	count := 0
	if len(page.Actions) > 0 {
		count++
	}
	if hasAction(page, "delete") {
		count++
	}
	if hasAction(page, "archive") || hasAction(page, "restore") {
		count++
	}
	return count
}

func (g *webGenerator) staticTableExtraColumns(page PageDecl, entity EntityDecl) int {
	count := staticTableExtraColumns(page)
	if len(page.Actions) == 0 && len(g.workflowsForEntity(entity.Name)) > 0 {
		count++
	}
	return count
}

func relationIDFieldName(field FieldDecl) string {
	return field.Name + "Id"
}

func relationOptionsStateName(field FieldDecl) string {
	return field.Name + "Options"
}

func relationName(entity EntityDecl, field FieldDecl) string {
	return entity.Name + "_" + field.Name
}

func relationBackFieldName(entity EntityDecl, field FieldDecl) string {
	return strings.ToLower(entity.Name[:1]) + entity.Name[1:] + title(field.Name) + "Items"
}

func (g *webGenerator) typeImportList(page PageDecl, entity EntityDecl) string {
	names := []string{entity.Name}
	seen := map[string]bool{entity.Name: true}
	for _, field := range g.relationFields(entity) {
		if seen[field.Type] {
			continue
		}
		seen[field.Type] = true
		names = append(names, field.Type)
	}
	if state, ok := g.stateForPage(page); ok {
		for _, field := range state.Fields {
			if _, exists := g.findEntity(field.Type); !exists || seen[field.Type] {
				continue
			}
			seen[field.Type] = true
			names = append(names, field.Type)
		}
	}
	return strings.Join(names, ", ")
}

func (g *webGenerator) stateForPage(page PageDecl) (StateDecl, bool) {
	expectedNames := map[string]bool{
		page.Name + "State":     true,
		page.Name + "PageState": true,
	}
	for _, state := range g.program.States {
		if expectedNames[state.Name] {
			return state, true
		}
	}
	return StateDecl{}, false
}

func (g *webGenerator) pageStateHooks(state StateDecl) string {
	var builder strings.Builder
	for _, field := range state.Fields {
		builder.WriteString(fmt.Sprintf("  const [%s, set%s] = useState<%s>(%s);\n", field.Name, title(field.Name), g.stateFieldTSType(field), stateFieldDefaultValue(field)))
	}
	for _, modal := range state.Modals {
		builder.WriteString(fmt.Sprintf("  const [%s, set%s] = useState(%t);\n", modalStateName(modal), title(modalStateName(modal)), modal.Default == "open"))
	}
	return builder.String()
}

func (g *webGenerator) stateFieldTSType(field StateField) string {
	fieldType := tsType(field.Type)
	if _, ok := g.findEntity(field.Type); ok {
		fieldType = field.Type
	}
	if field.List {
		return fieldType + "[]"
	}
	return fieldType
}

func stateFieldDefaultValue(field StateField) string {
	if field.List {
		return "[]"
	}
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return "0"
	case "boolean":
		return "false"
	default:
		return "\"\""
	}
}

func (g *webGenerator) pageModalHelpers(state StateDecl) string {
	var builder strings.Builder
	for _, modal := range state.Modals {
		stateName := modalStateName(modal)
		builder.WriteString(fmt.Sprintf("  function open%s() {\n", title(modal.Name)))
		builder.WriteString(fmt.Sprintf("    set%s(true);\n", title(stateName)))
		builder.WriteString("  }\n\n")
		builder.WriteString(fmt.Sprintf("  function close%s() {\n", title(modal.Name)))
		builder.WriteString(fmt.Sprintf("    set%s(false);\n", title(stateName)))
		builder.WriteString("  }\n\n")
	}
	return builder.String()
}

func (g *webGenerator) resetPageModals(state StateDecl) string {
	var builder strings.Builder
	for _, modal := range state.Modals {
		builder.WriteString(fmt.Sprintf("    close%s();\n", title(modal.Name)))
	}
	return builder.String()
}

func (g *webGenerator) closePageModals(state StateDecl) string {
	var builder strings.Builder
	for _, modal := range state.Modals {
		builder.WriteString(fmt.Sprintf("    close%s();\n", title(modal.Name)))
	}
	return builder.String()
}

func (g *webGenerator) openCreateModalButton(page PageDecl, state StateDecl, entity EntityDecl) string {
	for _, modal := range state.Modals {
		if modal.Name == "create"+entity.Name {
			if g.hasRuntimeI18N() {
				return fmt.Sprintf("          {canCreate && !editingId && !%s && <button%s type=\"button\" onClick={open%s}>{%s}</button>}\n", modalStateName(modal), g.actionButtonAttributesWithSuffix(page, "create", "open", "secondary"), title(modal.Name), g.uiLabelExpression("action.new."+entity.Name, "New "+entity.Name, "locale"))
			}
			return fmt.Sprintf("          {canCreate && !editingId && !%s && <button%s type=\"button\" onClick={open%s}>New %s</button>}\n", modalStateName(modal), g.actionButtonAttributesWithSuffix(page, "create", "open", "secondary"), title(modal.Name), entity.Name)
		}
	}
	return ""
}

func (g *webGenerator) openFormPanelButton(page PageDecl, entity EntityDecl) string {
	guard := "!editingId"
	if g.formSectionUsesOverlay(page) {
		guard = "!editingId && !formPanelOpen"
	}
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("          {canCreate && %s && <button%s type=\"button\" onClick={startCreateItem}>{%s}</button>}\n", guard, g.actionButtonAttributesWithSuffix(page, "create", "open", "secondary"), g.uiLabelExpression("action.new."+entity.Name, "New "+entity.Name, "locale"))
	}
	return fmt.Sprintf("          {canCreate && %s && <button%s type=\"button\" onClick={startCreateItem}>New %s</button>}\n", guard, g.actionButtonAttributesWithSuffix(page, "create", "open", "secondary"), entity.Name)
}

func (g *webGenerator) tablePanelAttributes(page PageDecl) string {
	classes := []string{"panel", "bl-view-section-table"}
	classes = append(classes, g.viewSectionClasses(page, "table")...)
	classes = append(classes, identityCSSClasses(page.Table.Identity)...)
	classes = append(classes, g.tableUIClass(page))
	return identityIDAttribute(page.Table.Identity) + classNameAttribute(classes...)
}

func (g *webGenerator) tableUIClass(page PageDecl) string {
	if len(page.Table.UI) == 0 {
		return ""
	}
	return "bl-ui-table-" + kebabCase(page.Name)
}

func (g *webGenerator) formPanelAttributes(page PageDecl) string {
	classes := []string{"panel", "bl-view-section-form"}
	classes = append(classes, g.viewSectionClasses(page, "form")...)
	classes = append(classes, identityCSSClasses(page.Form.Identity)...)
	classes = append(classes, g.formUIClass(page))
	return identityIDAttribute(page.Form.Identity) + classNameAttribute(classes...)
}

func (g *webGenerator) detailPanelAttributes(page PageDecl) string {
	classes := []string{"panel", "bl-view-section-detail"}
	classes = append(classes, g.viewSectionClasses(page, "detail")...)
	return classNameAttribute(classes...)
}

func (g *webGenerator) detailSectionOpen(page PageDecl, entity EntityDecl) string {
	if !g.detailSectionUsesOverlay(page) {
		return fmt.Sprintf("      <section%s>\n", g.detailPanelAttributes(page))
	}
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("      {(selectedItem || reading) && (\n      <div className=%q role=\"dialog\" aria-modal=\"true\" aria-label={%s}>\n        <section%s>\n", g.sectionOverlayClassName(page, "detail"), g.uiLabelExpression("action.view."+entity.Name, g.viewSectionTitle(page, "detail", entity.Name+" Details"), "locale"), g.detailPanelAttributes(page))
	}
	return fmt.Sprintf("      {(selectedItem || reading) && (\n      <div className=%q role=\"dialog\" aria-modal=\"true\" aria-label=%q>\n        <section%s>\n", g.sectionOverlayClassName(page, "detail"), g.viewSectionTitle(page, "detail", entity.Name+" Details"), g.detailPanelAttributes(page))
}

func (g *webGenerator) detailSectionHeader(page PageDecl, entity EntityDecl) string {
	titleExpression := fmt.Sprintf("%q", g.viewSectionTitle(page, "detail", entity.Name+" Details"))
	if g.hasRuntimeI18N() {
		titleExpression = g.uiLabelExpression("action.view."+entity.Name, g.viewSectionTitle(page, "detail", entity.Name+" Details"), "locale")
	}
	if !g.detailSectionUsesOverlay(page) {
		return fmt.Sprintf("        <h2>{%s}</h2>\n", titleExpression)
	}
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("          <div className=\"panel-header\">\n            <h2>{%s}</h2>\n            <button className=\"secondary\" type=\"button\" onClick={closeDetailPanel}>{%s}</button>\n          </div>\n", titleExpression, g.uiLabelExpression("action.close", "Close", "locale"))
	}
	return fmt.Sprintf("          <div className=\"panel-header\">\n            <h2>{%s}</h2>\n            <button className=\"secondary\" type=\"button\" onClick={closeDetailPanel}>Close</button>\n          </div>\n", titleExpression)
}

func (g *webGenerator) detailSectionClose(page PageDecl) string {
	if !g.detailSectionUsesOverlay(page) {
		return "      </section>\n"
	}
	return "        </section>\n      </div>\n      )}\n"
}

func (g *webGenerator) formSectionOpen(page PageDecl) string {
	if !g.formSectionUsesOverlay(page) {
		return fmt.Sprintf("      <section%s>\n", g.formPanelAttributes(page))
	}
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("      <div className=%q role=\"dialog\" aria-modal=\"true\" aria-label={%s}>\n        <section%s>\n", g.sectionOverlayClassName(page, "form"), g.uiLabelExpression("app.recordForm", g.viewSectionTitle(page, "form", "Record form"), "locale"), g.formPanelAttributes(page))
	}
	return fmt.Sprintf("      <div className=%q role=\"dialog\" aria-modal=\"true\" aria-label=%q>\n        <section%s>\n", g.sectionOverlayClassName(page, "form"), g.viewSectionTitle(page, "form", "Record form"), g.formPanelAttributes(page))
}

func (g *webGenerator) formSectionHeader(page PageDecl, entity EntityDecl) string {
	titleExpression := g.formSectionTitleExpression(page, entity)
	if !g.formSectionUsesOverlay(page) {
		return fmt.Sprintf("        <h2>{%s}</h2>\n", titleExpression)
	}
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("          <div className=\"panel-header\">\n            <h2>{%s}</h2>\n            <button className=\"secondary\" type=\"button\" onClick={resetForm}>{%s}</button>\n          </div>\n", titleExpression, g.uiLabelExpression("action.close", "Close", "locale"))
	}
	return fmt.Sprintf("          <div className=\"panel-header\">\n            <h2>{%s}</h2>\n            <button className=\"secondary\" type=\"button\" onClick={resetForm}>Close</button>\n          </div>\n", titleExpression)
}

func (g *webGenerator) formSectionClose(page PageDecl) string {
	if !g.formSectionUsesOverlay(page) {
		return "      </section>\n"
	}
	return "        </section>\n      </div>\n"
}

func (g *webGenerator) pageViewClassName(page PageDecl) string {
	classes := []string{"page-view", "page-view-" + kebabCase(page.Name)}
	if page.View != nil && page.View.Compose != nil && supportedViewComposeModes[page.View.Compose.Mode] {
		classes = append(classes, "bl-view-compose-"+page.View.Compose.Mode)
	}
	if g.pageUsesSectionOverlay(page) {
		classes = append(classes, "bl-view-has-overlay")
	}
	if page.View != nil && len(page.View.Triggers) > 0 {
		classes = append(classes, "bl-view-has-triggers")
	}
	return strings.Join(classes, " ")
}

func (g *webGenerator) viewGroupOpen(page PageDecl, sectionName string) string {
	group, ok := pageViewGroupForSection(page, sectionName)
	if !ok || !viewGroupStartsAtSection(page, group, sectionName) {
		return ""
	}
	label := viewGroupLabel(group)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("      <section className=%q aria-label=%q>\n", g.viewGroupClassName(group), label))
	if group.Title != "" {
		builder.WriteString(fmt.Sprintf("        <h2 className=\"view-group-title\">%s</h2>\n", escapeJSXText(group.Title)))
	}
	return builder.String()
}

func (g *webGenerator) viewGroupClose(page PageDecl, sectionName string) string {
	group, ok := pageViewGroupForSection(page, sectionName)
	if !ok || !viewGroupEndsAtSection(page, group, sectionName) {
		return ""
	}
	return "      </section>\n"
}

func (g *webGenerator) viewGroupClassName(group ViewGroupDecl) string {
	classes := []string{"view-group", "bl-view-group-" + kebabCase(group.Name), "bl-view-group-compose-" + viewGroupComposeMode(group)}
	if group.Span > 0 {
		classes = append(classes, fmt.Sprintf("bl-view-span-%d", group.Span))
	}
	return strings.Join(classes, " ")
}

func (g *webGenerator) viewSectionClasses(page PageDecl, sectionName string) []string {
	if page.View == nil {
		return nil
	}
	classes := []string{}
	for _, section := range page.View.Sections {
		if section.Name != sectionName {
			continue
		}
		if section.Span > 0 {
			classes = append(classes, fmt.Sprintf("bl-view-span-%d", section.Span))
		}
		if section.Display != "" {
			classes = append(classes, "bl-view-display-"+section.Display)
		}
		if section.Side != "" {
			classes = append(classes, "bl-view-side-"+section.Side)
		}
		return classes
	}
	return nil
}

func (g *webGenerator) pageUsesSectionOverlay(page PageDecl) bool {
	return g.detailSectionUsesOverlay(page) || g.formSectionUsesOverlay(page)
}

func (g *webGenerator) detailSectionUsesOverlay(page PageDecl) bool {
	return g.viewSectionUsesOverlay(page, "detail")
}

func (g *webGenerator) formSectionUsesOverlay(page PageDecl) bool {
	return g.viewSectionUsesOverlay(page, "form")
}

func (g *webGenerator) viewSectionUsesOverlay(page PageDecl, sectionName string) bool {
	display := g.viewSectionDisplay(page, sectionName)
	return display == "modal" || display == "drawer"
}

func (g *webGenerator) viewSectionDisplay(page PageDecl, sectionName string) string {
	section, ok := pageViewSection(page, sectionName)
	if !ok || section.Display == "" {
		return "inline"
	}
	return section.Display
}

func (g *webGenerator) viewSectionSide(page PageDecl, sectionName string) string {
	section, ok := pageViewSection(page, sectionName)
	if !ok || section.Side == "" {
		return "right"
	}
	return section.Side
}

func (g *webGenerator) viewSectionTitle(page PageDecl, sectionName string, fallback string) string {
	section, ok := pageViewSection(page, sectionName)
	if !ok || section.Title == "" {
		return fallback
	}
	return section.Title
}

func (g *webGenerator) createStartButtonEnabled(page PageDecl) bool {
	return hasAction(page, "create") && (g.formSectionUsesOverlay(page) || pageHasViewTrigger(page, "form", "createStart"))
}

func (g *webGenerator) viewTriggerEffects(page PageDecl, event string, itemExpression string, entity EntityDecl, indent string) string {
	if page.View == nil || len(page.View.Triggers) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, trigger := range page.View.Triggers {
		if trigger.Event != event {
			continue
		}
		switch trigger.Section {
		case "table":
			builder.WriteString(g.activateViewSectionStatement(page, "table", indent))
		case "detail":
			if event == "saveSuccess" && itemExpression != "" {
				builder.WriteString(fmt.Sprintf("%ssetSelectedItem(%s);\n", indent, itemExpression))
			}
			builder.WriteString(g.activateViewSectionStatement(page, "detail", indent))
		case "form":
			if event == "rowSelect" && itemExpression != "" {
				builder.WriteString(fmt.Sprintf("%sif (canUpdate) {\n", indent))
				builder.WriteString(g.activateViewSectionStatement(page, "form", indent+"  "))
				if g.formSectionUsesOverlay(page) {
					builder.WriteString(fmt.Sprintf("%s  setFormPanelOpen(true);\n", indent))
				}
				builder.WriteString(fmt.Sprintf("%s  setEditingId(%s.id);\n", indent, itemExpression))
				builder.WriteString(fmt.Sprintf("%s  setTouchedFields({});\n", indent))
				builder.WriteString(fmt.Sprintf("%s  setSubmitted(false);\n", indent))
				builder.WriteString(g.setFormFromItemStatements(page, entity, itemExpression, indent+"  "))
				builder.WriteString(fmt.Sprintf("%s}\n", indent))
				continue
			}
			builder.WriteString(g.activateViewSectionStatement(page, "form", indent))
		}
	}
	return builder.String()
}

func (g *webGenerator) activateViewSectionStatement(page PageDecl, sectionName string, indent string) string {
	tab, ok := pageViewTabForSection(page, sectionName)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%ssetActiveViewTab(%q);\n", indent, tab.Name)
}

func (g *webGenerator) setFormFromItemStatements(page PageDecl, entity EntityDecl, itemExpression string, indent string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%ssetForm({\n", indent))
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		itemValue := fmt.Sprintf("%s.%s", itemExpression, field.Name)
		if g.isRelationField(field) {
			itemValue = fmt.Sprintf("%s.%s", itemExpression, relationIDFieldName(field))
		}
		builder.WriteString(fmt.Sprintf("%s  %s: String(%s ?? \"\"),\n", indent, fieldName, itemValue))
	}
	builder.WriteString(fmt.Sprintf("%s});\n", indent))
	return builder.String()
}

func pageViewSection(page PageDecl, sectionName string) (ViewSectionDecl, bool) {
	if page.View == nil {
		return ViewSectionDecl{}, false
	}
	for _, section := range page.View.Sections {
		if section.Name == sectionName {
			return section, true
		}
	}
	return ViewSectionDecl{}, false
}

func pageHasViewTrigger(page PageDecl, sectionName string, event string) bool {
	if page.View == nil {
		return false
	}
	for _, trigger := range page.View.Triggers {
		if trigger.Section == sectionName && trigger.Event == event {
			return true
		}
	}
	return false
}

func pageViewGroupForSection(page PageDecl, sectionName string) (ViewGroupDecl, bool) {
	if page.View == nil {
		return ViewGroupDecl{}, false
	}
	for _, group := range page.View.Groups {
		for _, groupedSection := range group.Sections {
			if groupedSection == sectionName {
				return group, true
			}
		}
	}
	return ViewGroupDecl{}, false
}

func viewGroupStartsAtSection(page PageDecl, group ViewGroupDecl, sectionName string) bool {
	first, ok := viewGroupBoundarySection(page, group, true)
	return ok && first == sectionName
}

func viewGroupEndsAtSection(page PageDecl, group ViewGroupDecl, sectionName string) bool {
	last, ok := viewGroupBoundarySection(page, group, false)
	return ok && last == sectionName
}

func viewGroupBoundarySection(page PageDecl, group ViewGroupDecl, first bool) (string, bool) {
	grouped := map[string]bool{}
	for _, sectionName := range group.Sections {
		grouped[sectionName] = true
	}
	sections := append([]string{}, pageViewOrder(page)...)
	if !first {
		for left, right := 0, len(sections)-1; left < right; left, right = left+1, right-1 {
			sections[left], sections[right] = sections[right], sections[left]
		}
	}
	for _, sectionName := range sections {
		if grouped[sectionName] {
			return sectionName, true
		}
	}
	return "", false
}

func viewGroupOrder(page PageDecl, group ViewGroupDecl) (int, bool) {
	orderIndex := map[string]int{}
	for index, sectionName := range pageViewOrder(page) {
		orderIndex[sectionName] = index + 1
	}
	order := 0
	for _, sectionName := range group.Sections {
		index, ok := orderIndex[sectionName]
		if !ok {
			continue
		}
		if order == 0 || index < order {
			order = index
		}
	}
	return order, order > 0
}

func viewGroupLabel(group ViewGroupDecl) string {
	if group.Title != "" {
		return group.Title
	}
	return identifierLabel(group.Name)
}

func (g *webGenerator) sectionOverlayClassName(page PageDecl, sectionName string) string {
	display := g.viewSectionDisplay(page, sectionName)
	classes := []string{"section-overlay", "section-overlay-" + display}
	if display == "drawer" {
		classes = append(classes, "section-overlay-drawer-"+g.viewSectionSide(page, sectionName))
	}
	return strings.Join(classes, " ")
}

func (g *webGenerator) formSectionTitleExpression(page PageDecl, entity EntityDecl) string {
	if title := g.viewSectionTitle(page, "form", ""); title != "" {
		return fmt.Sprintf("%q", title)
	}
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("editingId ? %s : %s", g.uiLabelExpression("action.edit."+entity.Name, "Edit "+entity.Name, "locale"), g.uiLabelExpression("action.create."+entity.Name, "Create "+entity.Name, "locale"))
	}
	return fmt.Sprintf("editingId ? %q : %q", "Edit "+entity.Name, "Create "+entity.Name)
}

func pageViewTabs(page PageDecl) []ViewTabDecl {
	if page.View == nil || page.View.Compose == nil || page.View.Compose.Mode != "tabs" || len(page.View.Tabs) == 0 {
		return nil
	}
	return page.View.Tabs
}

func pageUsesViewTabs(page PageDecl) bool {
	return len(pageViewTabs(page)) > 0
}

func pageViewTabForSection(page PageDecl, sectionName string) (ViewTabDecl, bool) {
	for _, tab := range pageViewTabs(page) {
		for _, section := range tab.Sections {
			if section == sectionName {
				return tab, true
			}
		}
	}
	return ViewTabDecl{}, false
}

func (g *webGenerator) viewTabsStateHook(page PageDecl) string {
	tabs := pageViewTabs(page)
	if len(tabs) == 0 {
		return ""
	}
	return fmt.Sprintf("  const [activeViewTab, setActiveViewTab] = useState(%q);\n\n", tabs[0].Name)
}

func (g *webGenerator) viewTabsMarkup(page PageDecl) string {
	tabs := pageViewTabs(page)
	if len(tabs) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("      <nav className=\"view-tabs\" role=\"tablist\" aria-label=%q>\n", page.Name+" view sections"))
	for _, tab := range tabs {
		label := identifierLabel(tab.Name)
		builder.WriteString(fmt.Sprintf("        <button className={activeViewTab === %q ? \"active\" : \"\"} type=\"button\" role=\"tab\" aria-selected={activeViewTab === %q} onClick={() => setActiveViewTab(%q)}>%s</button>\n", tab.Name, tab.Name, tab.Name, label))
	}
	builder.WriteString("      </nav>\n\n")
	return builder.String()
}

func (g *webGenerator) viewSectionConditionalOpen(page PageDecl, sectionName string) string {
	tab, ok := pageViewTabForSection(page, sectionName)
	if !ok {
		return ""
	}
	return fmt.Sprintf("      {activeViewTab === %q && (\n        <div className=\"view-tab-panel\" role=\"tabpanel\" aria-label=%q>\n", tab.Name, identifierLabel(tab.Name)+" tab")
}

func (g *webGenerator) viewSectionConditionalClose(page PageDecl, sectionName string) string {
	if _, ok := pageViewTabForSection(page, sectionName); !ok {
		return ""
	}
	return "        </div>\n      )}\n"
}

func (g *webGenerator) formUIClass(page PageDecl) string {
	if len(page.Form.UI) == 0 {
		return ""
	}
	return "bl-ui-form-" + kebabCase(page.Name)
}

func (g *webGenerator) fieldUIClass(entity EntityDecl, field FieldDecl) string {
	return "bl-ui-field-" + kebabCase(entity.Name) + "-" + kebabCase(field.Name)
}

func (g *webGenerator) fieldUIClassAttribute(entity EntityDecl, field FieldDecl) string {
	if len(field.UI) == 0 {
		return ""
	}
	return classNameAttribute(g.fieldUIClass(entity, field))
}

func (g *webGenerator) actionUIClass(page PageDecl, action string) string {
	for _, actionUI := range page.ActionUI {
		if actionUI.Action == action && len(actionUI.UI) > 0 {
			return "bl-ui-action-" + kebabCase(page.Name) + "-" + kebabCase(action)
		}
	}
	return ""
}

func (g *webGenerator) actionIdentity(page PageDecl, action string) *UIIdentity {
	for _, actionUI := range page.ActionUI {
		if actionUI.Action == action {
			return actionUI.Identity
		}
	}
	return nil
}

func (g *webGenerator) actionButtonClassAttribute(page PageDecl, action string, baseClasses ...string) string {
	classes := append([]string{}, baseClasses...)
	classes = append(classes, identityCSSClasses(g.actionIdentity(page, action))...)
	classes = append(classes, g.actionUIClass(page, action))
	return classNameAttribute(classes...)
}

func (g *webGenerator) actionButtonAttributes(page PageDecl, action string, baseClasses ...string) string {
	return g.actionButtonIDAttribute(page, action) + g.actionButtonClassAttribute(page, action, baseClasses...)
}

func (g *webGenerator) actionButtonAttributesWithSuffix(page PageDecl, action string, suffix string, baseClasses ...string) string {
	return g.actionButtonIDSuffixAttribute(page, action, suffix) + g.actionButtonClassAttribute(page, action, baseClasses...)
}

func (g *webGenerator) actionRowButtonAttributes(page PageDecl, action string, baseClasses ...string) string {
	return g.actionRowButtonIDAttribute(page, action) + g.actionButtonClassAttribute(page, action, baseClasses...)
}

func (g *webGenerator) actionButtonIDAttribute(page PageDecl, action string) string {
	id := g.actionHTMLID(page, action)
	if id == "" {
		return ""
	}
	return idAttribute(id)
}

func (g *webGenerator) actionButtonIDSuffixAttribute(page PageDecl, action string, suffix string) string {
	id := g.actionHTMLID(page, action)
	if id == "" {
		return ""
	}
	if suffix != "" {
		id += "-" + kebabCase(suffix)
	}
	return idAttribute(id)
}

func (g *webGenerator) actionRowButtonIDAttribute(page PageDecl, action string) string {
	id := g.actionHTMLID(page, action)
	if id == "" {
		return ""
	}
	return fmt.Sprintf(" id={%q + item.id}", id+"-item-")
}

func (g *webGenerator) actionHTMLID(page PageDecl, action string) string {
	identity := g.actionIdentity(page, action)
	if identity == nil || identity.ID == "" {
		return ""
	}
	return kebabCase(identity.ID)
}

func identityIDAttribute(identity *UIIdentity) string {
	if identity == nil || identity.ID == "" {
		return ""
	}
	return idAttribute(identity.ID)
}

func identityCSSClasses(identity *UIIdentity) []string {
	if identity == nil || len(identity.Classes) == 0 {
		return nil
	}
	classes := []string{}
	for _, className := range identity.Classes {
		classes = append(classes, kebabCase(className))
	}
	return classes
}

func (g *webGenerator) formSubmitClassAttribute(page PageDecl) string {
	createClass := g.actionButtonClassString(page, "create")
	editClass := g.actionButtonClassString(page, "edit")
	if createClass != "" && editClass != "" {
		return fmt.Sprintf(" className={editingId ? %q : %q}", editClass, createClass)
	}
	if createClass != "" {
		return fmt.Sprintf(" className={!editingId ? %q : undefined}", createClass)
	}
	if editClass != "" {
		return fmt.Sprintf(" className={editingId ? %q : undefined}", editClass)
	}
	return ""
}

func (g *webGenerator) formSubmitAttributes(page PageDecl) string {
	return g.formSubmitIDAttribute(page) + g.formSubmitClassAttribute(page)
}

func (g *webGenerator) formSubmitIDAttribute(page PageDecl) string {
	createID := g.actionHTMLIDSuffix(page, "create", "submit")
	editID := g.actionHTMLIDSuffix(page, "edit", "submit")
	if createID != "" && editID != "" {
		return fmt.Sprintf(" id={editingId ? %q : %q}", editID, createID)
	}
	if createID != "" {
		return fmt.Sprintf(" id={!editingId ? %q : undefined}", createID)
	}
	if editID != "" {
		return fmt.Sprintf(" id={editingId ? %q : undefined}", editID)
	}
	return ""
}

func (g *webGenerator) actionHTMLIDSuffix(page PageDecl, action string, suffix string) string {
	id := g.actionHTMLID(page, action)
	if id == "" {
		return ""
	}
	if suffix == "" {
		return id
	}
	return id + "-" + kebabCase(suffix)
}

func (g *webGenerator) actionButtonClassString(page PageDecl, action string) string {
	classes := identityCSSClasses(g.actionIdentity(page, action))
	classes = append(classes, g.actionUIClass(page, action))
	return joinCSSClasses(classes...)
}

func idAttribute(id string) string {
	id = kebabCase(strings.TrimSpace(id))
	if id == "" {
		return ""
	}
	return fmt.Sprintf(" id=%q", id)
}

func classNameAttribute(classes ...string) string {
	joined := joinCSSClasses(classes...)
	if joined == "" {
		return ""
	}
	return fmt.Sprintf(" className=%q", joined)
}

func joinCSSClasses(classes ...string) string {
	filtered := []string{}
	for _, className := range classes {
		className = strings.TrimSpace(className)
		if className == "" {
			continue
		}
		filtered = append(filtered, className)
	}
	return strings.Join(filtered, " ")
}

func (g *webGenerator) formVisibleExpression(page PageDecl, entity EntityDecl, state StateDecl, hasState bool) string {
	if g.formSectionUsesOverlay(page) {
		if hasAction(page, "create") && hasAction(page, "edit") {
			return "((canUpdate && editingId) || (canCreate && formPanelOpen))"
		}
		if hasAction(page, "create") {
			return "(canCreate && formPanelOpen)"
		}
		if hasAction(page, "edit") {
			return "(canUpdate && editingId)"
		}
		return "false"
	}
	base := "(canCreate || canUpdate)"
	if !hasState {
		return base
	}
	for _, modal := range state.Modals {
		if modal.Name == "create"+entity.Name && hasAction(page, "create") {
			return fmt.Sprintf("((canUpdate && editingId) || (canCreate && %s))", modalStateName(modal))
		}
	}
	return base
}

func (g *webGenerator) cancelVisibleExpression(page PageDecl, entity EntityDecl, state StateDecl, hasState bool) string {
	if g.formSectionUsesOverlay(page) {
		return "(editingId || formPanelOpen)"
	}
	if !hasState {
		return "editingId"
	}
	for _, modal := range state.Modals {
		if modal.Name == "create"+entity.Name {
			return "(editingId || " + modalStateName(modal) + ")"
		}
	}
	return "editingId"
}

func modalStateName(modal StateModal) string {
	return modal.Name + "Open"
}

func (g *webGenerator) missingRequiredRelationsExpression(page PageDecl, entity EntityDecl) string {
	parts := []string{}
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok || !g.isRelationField(field) || !hasModifier(field, "required") {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s.length === 0", relationOptionsStateName(field)))
	}
	if len(parts) == 0 {
		return "false"
	}
	return strings.Join(parts, " || ")
}

func (g *webGenerator) missingRequiredRelationsMessage(page PageDecl, entity EntityDecl) string {
	targets := []string{}
	seen := map[string]bool{}
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok || !g.isRelationField(field) || !hasModifier(field, "required") || seen[field.Type] {
			continue
		}
		seen[field.Type] = true
		targets = append(targets, field.Type)
	}
	if len(targets) == 0 {
		return ""
	}
	if len(targets) == 1 {
		return fmt.Sprintf("Create a %s record before creating %s.", targets[0], entity.Name)
	}
	return fmt.Sprintf("Create required related records (%s) before creating %s.", strings.Join(targets, ", "), entity.Name)
}

func (g *webGenerator) relationNavigationButtons(page PageDecl, entity EntityDecl) string {
	buttons := []string{}
	seen := map[string]bool{}
	for _, fieldName := range page.Form.Fields {
		field, ok := findField(entity, fieldName)
		if !ok || !g.isRelationField(field) || !hasModifier(field, "required") || seen[field.Type] {
			continue
		}
		targetPage, ok := g.pageForEntity(field.Type)
		if !ok {
			continue
		}
		seen[field.Type] = true
		buttons = append(buttons, fmt.Sprintf("              <button className=\"secondary\" type=\"button\" onClick={() => onNavigate(%q)}>Open %s</button>", targetPage.Name, targetPage.Name))
	}
	if len(buttons) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("            {onNavigate && (\n")
	builder.WriteString("              <div className=\"actions\">\n")
	for _, button := range buttons {
		builder.WriteString(button)
		builder.WriteString("\n")
	}
	builder.WriteString("              </div>\n")
	builder.WriteString("            )}\n")
	return builder.String()
}

func (g *webGenerator) pageForEntity(entityName string) (PageDecl, bool) {
	for _, page := range g.program.Pages {
		if page.Source == entityName && page.Query == "" {
			return page, true
		}
	}
	for _, page := range g.program.Pages {
		if page.Source == entityName {
			return page, true
		}
	}
	return PageDecl{}, false
}

func (g *webGenerator) navigationPages() []PageDecl {
	if len(g.program.Layouts) == 0 || len(g.program.Layouts[0].Sidebar.Items) == 0 {
		return g.program.Pages
	}

	pageIndex := map[string]PageDecl{}
	for _, page := range g.program.Pages {
		pageIndex[page.Name] = page
	}

	ordered := []PageDecl{}
	seen := map[string]bool{}
	for _, item := range g.program.Layouts[0].Sidebar.Items {
		page, ok := pageIndex[item]
		if !ok || seen[item] {
			continue
		}
		seen[item] = true
		ordered = append(ordered, page)
	}
	for _, page := range g.program.Pages {
		if seen[page.Name] {
			continue
		}
		ordered = append(ordered, page)
	}
	if len(ordered) == 0 {
		return g.program.Pages
	}
	return ordered
}

func (g *webGenerator) itemDisplayExpression(itemName string, field FieldDecl) string {
	if !g.isRelationField(field) {
		if g.hasRuntimeI18N() && runtimeFormattedFieldType(field.Type) {
			return fmt.Sprintf("formatFieldValue(%s.%s, %q, locale)", itemName, field.Name, field.Type)
		}
		return fmt.Sprintf("String(%s.%s ?? \"\")", itemName, field.Name)
	}
	target, ok := g.findEntity(field.Type)
	if !ok {
		return fmt.Sprintf("String(%s.%s ?? \"\")", itemName, relationIDFieldName(field))
	}
	labelField := relationLabelField(target)
	return fmt.Sprintf("String(%s.%s?.%s ?? %s.%s ?? \"\")", itemName, field.Name, labelField, itemName, relationIDFieldName(field))
}

func (g *webGenerator) computedFieldDisplayExpression(itemName string, field ComputedFieldDecl) string {
	expression := fmt.Sprintf("%s(%s)", computedFieldFunctionName(field), itemName)
	if g.hasRuntimeI18N() {
		return fmt.Sprintf("formatComputedValue(%s, %q, locale)", expression, field.Type)
	}
	return fmt.Sprintf("formatComputedValue(%s)", expression)
}

func runtimeFormattedFieldType(fieldType string) bool {
	switch fieldType {
	case "number", "integer", "decimal", "money", "date", "datetime":
		return true
	default:
		return false
	}
}

func (g *webGenerator) relationOptionLabelExpression(optionName string, entityName string) string {
	target, ok := g.findEntity(entityName)
	if !ok {
		return fmt.Sprintf("String(%s.id)", optionName)
	}
	labelField := relationLabelField(target)
	return fmt.Sprintf("String(%s.%s ?? %s.id)", optionName, labelField, optionName)
}

func relationLabelField(entity EntityDecl) string {
	for _, preferred := range []string{"name", "title", "email", "sku"} {
		for _, field := range entity.Fields {
			if field.Name == preferred {
				return field.Name
			}
		}
	}
	if len(entity.Fields) > 0 {
		return entity.Fields[0].Name
	}
	return "id"
}

func tsType(fieldType string) string {
	switch fieldType {
	case "number", "integer", "decimal", "money":
		return "number"
	case "boolean":
		return "boolean"
	case "date", "datetime":
		return "string"
	default:
		return "string"
	}
}

func prismaType(fieldType string) string {
	switch fieldType {
	case "number", "integer":
		return "Int"
	case "decimal", "money":
		return "Float"
	case "boolean":
		return "Boolean"
	case "date", "datetime":
		return "DateTime"
	default:
		return "String"
	}
}

func sqliteType(fieldType string) string {
	switch fieldType {
	case "number", "integer":
		return "INTEGER"
	case "decimal", "money":
		return "REAL"
	case "boolean":
		return "INTEGER"
	case "date", "datetime":
		return "DATETIME"
	default:
		return "TEXT"
	}
}

func sqliteRequired(field FieldDecl) string {
	if hasModifier(field, "required") || hasModifier(field, "default") {
		return " NOT NULL"
	}
	return ""
}

func sqliteDefaultValue(field FieldDecl, value string) string {
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return value
	case "boolean":
		if value == "true" {
			return "1"
		}
		return "0"
	default:
		return "'" + strings.ReplaceAll(value, "'", "''") + "'"
	}
}

func prismaOptional(field FieldDecl) string {
	if hasModifier(field, "required") || hasModifier(field, "default") {
		return ""
	}
	return "?"
}

func prismaRelationIDOptional(field FieldDecl) string {
	if hasModifier(field, "required") {
		return ""
	}
	return "?"
}

func prismaAttributes(field FieldDecl) string {
	attributes := []string{}
	if hasModifier(field, "unique") {
		attributes = append(attributes, "@unique")
	}
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		attributes = append(attributes, fmt.Sprintf("@default(%s)", prismaDefaultValue(field, defaultValue)))
	}
	if len(attributes) == 0 {
		return ""
	}
	return " " + strings.Join(attributes, " ")
}

func (g *webGenerator) prismaIndexLine(entity EntityDecl, index EntityIndexDecl) string {
	fields := g.prismaIndexFields(entity, index)
	if len(fields) == 0 {
		return ""
	}
	return fmt.Sprintf("  @@index([%s], map: %q)\n", strings.Join(fields, ", "), entityIndexName(entity, index))
}

func (g *webGenerator) prismaIndexFields(entity EntityDecl, index EntityIndexDecl) []string {
	fields := []string{}
	for _, fieldName := range index.Fields {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		if g.isRelationField(field) {
			fields = append(fields, relationIDFieldName(field))
			continue
		}
		fields = append(fields, field.Name)
	}
	return fields
}

func (g *webGenerator) sqliteIndexStatement(entity EntityDecl, index EntityIndexDecl) string {
	fields := g.sqliteIndexFields(entity, index)
	if len(fields) == 0 {
		return ""
	}
	quotedFields := []string{}
	for _, field := range fields {
		quotedFields = append(quotedFields, fmt.Sprintf("%q", field))
	}
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %q ON %q (%s);", entityIndexName(entity, index), entity.Name, strings.Join(quotedFields, ", "))
}

func (g *webGenerator) sqliteIndexFields(entity EntityDecl, index EntityIndexDecl) []string {
	fields := []string{}
	for _, fieldName := range index.Fields {
		field, ok := findField(entity, fieldName)
		if !ok {
			continue
		}
		fields = append(fields, g.sqliteColumnName(field))
	}
	return fields
}

func entityIndexName(entity EntityDecl, index EntityIndexDecl) string {
	parts := []string{entity.Name}
	parts = append(parts, index.Fields...)
	parts = append(parts, "idx")
	return strings.Join(parts, "_")
}

func prismaDefaultValue(field FieldDecl, value string) string {
	switch field.Type {
	case "text", "email", "date", "datetime":
		return fmt.Sprintf("%q", value)
	default:
		return value
	}
}

func hasModifier(field FieldDecl, name string) bool {
	for _, modifier := range field.Modifiers {
		if modifier.Name == name {
			return true
		}
	}
	return false
}

func hasAction(page PageDecl, name string) bool {
	for _, action := range page.Actions {
		if action == name {
			return true
		}
	}
	return false
}

func (g *webGenerator) defaultAuthRole() string {
	if len(g.program.Roles) > 0 {
		return g.program.Roles[0].Name
	}
	return "authenticated"
}

func (g *webGenerator) roleNames() []string {
	names := []string{}
	for _, role := range g.program.Roles {
		names = append(names, role.Name)
	}
	return names
}

func (g *webGenerator) hasRuntimePermissions() bool {
	return g.program.Auth != nil && len(g.program.Roles) > 0
}

func (g *webGenerator) permissionMiddleware(action string, resource string) string {
	if !g.hasRuntimePermissions() {
		return ""
	}
	return fmt.Sprintf("requirePermission(%q, %q), ", action, resource)
}

func (g *webGenerator) rolePermissionsLiteral() string {
	type generatedPermission struct {
		Effect   string   `json:"effect"`
		Action   string   `json:"action"`
		Resource string   `json:"resource"`
		Fields   []string `json:"fields"`
	}
	permissions := map[string][]generatedPermission{}
	for _, role := range g.program.Roles {
		for _, permission := range role.Permissions {
			fields := permission.Fields
			if fields == nil {
				fields = []string{}
			}
			permissions[role.Name] = append(permissions[role.Name], generatedPermission{
				Effect:   permission.Effect,
				Action:   permission.Action,
				Resource: permission.Resource,
				Fields:   fields,
			})
		}
	}
	content, err := json.Marshal(permissions)
	if err != nil {
		return "{}"
	}
	return string(content)
}

func tsStringArrayLiteral(values []string) string {
	parts := []string{}
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%q", value))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func fieldNames(entity EntityDecl) []string {
	names := []string{}
	for _, field := range entity.Fields {
		names = append(names, field.Name)
	}
	return names
}

func findField(entity EntityDecl, name string) (FieldDecl, bool) {
	for _, field := range entity.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return FieldDecl{}, false
}

func findComputedField(entity EntityDecl, name string) (ComputedFieldDecl, bool) {
	for _, field := range entity.ComputedFields {
		if field.Name == name {
			return field, true
		}
	}
	return ComputedFieldDecl{}, false
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func componentInputByName(inputs []ComponentInput, name string) (ComponentInput, bool) {
	for _, input := range inputs {
		if input.Name == name {
			return input, true
		}
	}
	return ComponentInput{}, false
}

func isNumericLiteral(value string) bool {
	if value == "" {
		return false
	}
	dotSeen := false
	for index, char := range value {
		if char == '-' && index == 0 {
			continue
		}
		if char == '.' && !dotSeen {
			dotSeen = true
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	return value != "-" && value != "." && value != "-."
}

func kebabCase(value string) string {
	if value == "" {
		return value
	}
	var builder strings.Builder
	for index, char := range value {
		if index > 0 && char >= 'A' && char <= 'Z' {
			builder.WriteByte('-')
		}
		builder.WriteRune(char)
	}
	return strings.ToLower(builder.String())
}

func lowerCamelCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func workflowActionFunctionName(transition TransitionDecl) string {
	return "run" + title(transition.Name) + "Workflow"
}

func identifierLabel(value string) string {
	if value == "" {
		return value
	}
	words := []rune{}
	for index, char := range value {
		if index > 0 && char >= 'A' && char <= 'Z' {
			words = append(words, ' ')
		}
		words = append(words, char)
	}
	return title(string(words))
}

func escapeJSXText(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"{", "&#123;",
		"}", "&#125;",
	)
	return replacer.Replace(value)
}

func formStateLiteral(fields []string, entity EntityDecl) string {
	if len(fields) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, fieldName := range fields {
		value := ""
		if field, ok := findField(entity, fieldName); ok {
			value = modifierValue(field, "default")
		}
		parts = append(parts, fmt.Sprintf("%s: %q", fieldName, value))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func columnVisibilityLiteral(columns []string) string {
	if len(columns) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, column := range columns {
		parts = append(parts, fmt.Sprintf("%s: true", column))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func (g *webGenerator) visibleColumnCountExpression(page PageDecl, entity EntityDecl) string {
	conditions := []string{}
	for _, column := range page.Table.Columns {
		if field, ok := findField(entity, column); ok {
			conditions = append(conditions, fmt.Sprintf("visibleColumns.%s && permissions.fields.%s !== false", field.Name, field.Name))
			continue
		}
		if computed, ok := findComputedField(entity, column); ok {
			conditions = append(conditions, fmt.Sprintf("visibleColumns.%s && %s", computed.Name, g.computedFieldPermissionExpression(computed)))
		}
	}
	if len(conditions) == 0 {
		return "0"
	}
	return "[" + strings.Join(conditions, ", ") + "].filter(Boolean).length"
}

func tableFiltersLiteral(filters []string) string {
	if len(filters) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, filter := range filters {
		parts = append(parts, fmt.Sprintf("%s: %q", filter, ""))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func (g *webGenerator) tableStateDependencyList(page PageDecl, dependencies ...string) string {
	values := append([]string{}, dependencies...)
	if len(page.Table.Filters) > 0 {
		values = append(values, "filters")
	}
	if g.hasRuntimeI18N() {
		values = append(values, "locale")
	}
	return strings.Join(values, ", ")
}

func formValueExpression(field FieldDecl) string {
	switch field.Type {
	case "number", "integer", "decimal", "money":
		if defaultValue := modifierValue(field, "default"); defaultValue != "" {
			return fmt.Sprintf("form.%s === \"\" ? %s : Number(form.%s)", field.Name, defaultValue, field.Name)
		}
		if hasModifier(field, "required") {
			return fmt.Sprintf("Number(form.%s)", field.Name)
		}
		return fmt.Sprintf("form.%s === \"\" ? undefined : Number(form.%s)", field.Name, field.Name)
	case "boolean":
		return fmt.Sprintf("form.%s === \"true\"", field.Name)
	default:
		return fmt.Sprintf("form.%s", field.Name)
	}
}

func fieldLabel(field FieldDecl) string {
	if label := modifierValue(field, "label"); label != "" {
		return label
	}
	return title(field.Name)
}

func (g *webGenerator) fieldLabel(entity EntityDecl, field FieldDecl) string {
	if label := g.defaultFieldLabelTranslation(entity.Name, field.Name); label != "" {
		return label
	}
	return fieldLabel(field)
}

func (g *webGenerator) computedFieldLabel(entity EntityDecl, field ComputedFieldDecl) string {
	if label := g.defaultFieldLabelTranslation(entity.Name, field.Name); label != "" {
		return label
	}
	return computedFieldFallbackLabel(field)
}

func computedFieldFallbackLabel(field ComputedFieldDecl) string {
	if label := computedModifierValue(field, "label"); label != "" {
		return label
	}
	return title(field.Name)
}

func (g *webGenerator) pageI18NHelpers(entity EntityDecl) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("const defaultLocale = %q;\n\n", g.program.I18N.Default))
	builder.WriteString(g.localizedFieldTextMap("fieldLabels", entity, "label", true))
	builder.WriteString(g.localizedFieldTextMap("fieldPlaceholders", entity, "placeholder", false))
	builder.WriteString(g.localizedFieldTextMap("fieldHelpTexts", entity, "help", false))
	builder.WriteString(g.localizedFieldTextMap("fieldMessages", entity, "message", false))
	builder.WriteString(g.localizedUILabelMap("uiLabels"))
	builder.WriteString("function fieldText(texts: Record<string, Record<string, string>>, field: string, fallback: string, locale: string) {\n")
	builder.WriteString("  return texts[field]?.[locale] ?? texts[field]?.[defaultLocale] ?? fallback;\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function uiLabel(key: string, fallback: string, locale: string) {\n")
	builder.WriteString("  return uiLabels[key]?.[locale] ?? uiLabels[key]?.[defaultLocale] ?? fallback;\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function fieldLabel(field: string, fallback: string, locale: string) {\n")
	builder.WriteString("  return fieldText(fieldLabels, field, fallback, locale);\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function fieldPlaceholder(field: string, fallback: string, locale: string) {\n")
	builder.WriteString("  return fieldText(fieldPlaceholders, field, fallback, locale);\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function fieldHelpText(field: string, fallback: string, locale: string) {\n")
	builder.WriteString("  return fieldText(fieldHelpTexts, field, fallback, locale);\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function fieldMessage(field: string, fallback: string, locale: string) {\n")
	builder.WriteString("  return fieldText(fieldMessages, field, fallback, locale);\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function formatFieldValue(value: unknown, kind: string, locale: string) {\n")
	builder.WriteString("  if (value === null || value === undefined || value === \"\") return \"\";\n")
	builder.WriteString("  if ([\"number\", \"integer\", \"decimal\"].includes(kind)) return formatNumberValue(value, locale);\n")
	builder.WriteString("  if (kind === \"money\") return formatNumberValue(value, locale, { style: \"currency\", currency: localeCurrency(locale) });\n")
	builder.WriteString("  if (kind === \"date\") return formatDateValue(value, locale, { dateStyle: \"medium\" });\n")
	builder.WriteString("  if (kind === \"datetime\") return formatDateValue(value, locale, { dateStyle: \"medium\", timeStyle: \"short\" });\n")
	builder.WriteString("  return String(value);\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function formatNumberValue(value: unknown, locale: string, options?: Intl.NumberFormatOptions) {\n")
	builder.WriteString("  const number = typeof value === \"number\" ? value : Number(value);\n")
	builder.WriteString("  return Number.isFinite(number) ? new Intl.NumberFormat(locale, options).format(number) : String(value);\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function formatDateValue(value: unknown, locale: string, options: Intl.DateTimeFormatOptions) {\n")
	builder.WriteString("  const date = new Date(String(value));\n")
	builder.WriteString("  return Number.isNaN(date.getTime()) ? String(value) : new Intl.DateTimeFormat(locale, options).format(date);\n")
	builder.WriteString("}\n\n")
	builder.WriteString("function localeCurrency(locale: string) {\n")
	builder.WriteString("  const [language, region = \"\"] = locale.toLowerCase().split(\"-\");\n")
	builder.WriteString("  const currencies: Record<string, string> = { tr: \"TRY\", en: \"USD\", us: \"USD\", gb: \"GBP\", de: \"EUR\", fr: \"EUR\", es: \"EUR\", it: \"EUR\", nl: \"EUR\", pt: \"EUR\", ar: \"AED\", ae: \"AED\", sa: \"SAR\", fa: \"IRR\", ir: \"IRR\", he: \"ILS\", il: \"ILS\", ur: \"PKR\", pk: \"PKR\" };\n")
	builder.WriteString("  return currencies[region] ?? currencies[language] ?? \"USD\";\n")
	builder.WriteString("}\n\n")
	return builder.String()
}

func (g *webGenerator) localizedFieldTextMap(name string, entity EntityDecl, kind string, includeComputed bool) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("const %s: Record<string, Record<string, string>> = {\n", name))
	for _, field := range entity.Fields {
		translations := g.fieldTextTranslations(kind, entity.Name, field.Name)
		if len(translations) == 0 {
			continue
		}
		builder.WriteString(fmt.Sprintf("  %q: %s,\n", field.Name, g.translationMapLiteral(translations)))
	}
	if includeComputed {
		for _, field := range entity.ComputedFields {
			translations := g.fieldTextTranslations(kind, entity.Name, field.Name)
			if len(translations) == 0 {
				continue
			}
			builder.WriteString(fmt.Sprintf("  %q: %s,\n", field.Name, g.translationMapLiteral(translations)))
		}
	}
	builder.WriteString("};\n\n")
	return builder.String()
}

func (g *webGenerator) translationMapLiteral(translations map[string]string) string {
	parts := []string{}
	for _, locale := range g.program.I18N.Locales {
		if text, ok := translations[locale]; ok {
			parts = append(parts, fmt.Sprintf("%q: %q", locale, text))
		}
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func (g *webGenerator) fieldLabelTranslations(entityName string, fieldName string) map[string]string {
	target := entityName + "." + fieldName
	for _, label := range g.program.Labels {
		if label.Target != target {
			continue
		}
		translations := map[string]string{}
		for _, translation := range label.Translations {
			translations[translation.Locale] = translation.Text
		}
		return translations
	}
	return nil
}

func (g *webGenerator) fieldTextTranslations(kind string, entityName string, fieldName string) map[string]string {
	if kind == "label" {
		return g.fieldLabelTranslations(entityName, fieldName)
	}
	switch kind {
	case "placeholder":
		return fieldTextBlockTranslations(g.program.Placeholders, entityName, fieldName)
	case "help":
		return fieldTextBlockTranslations(g.program.HelpTexts, entityName, fieldName)
	case "message":
		return fieldTextBlockTranslations(g.program.Messages, entityName, fieldName)
	default:
		return nil
	}
}

func fieldTextBlockTranslations(blocks []FieldTextTranslationDecl, entityName string, fieldName string) map[string]string {
	target := entityName + "." + fieldName
	for _, block := range blocks {
		if block.Target != target {
			continue
		}
		translations := map[string]string{}
		for _, translation := range block.Translations {
			translations[translation.Locale] = translation.Text
		}
		return translations
	}
	return nil
}

func (g *webGenerator) fieldLabelJSX(entity EntityDecl, field FieldDecl) string {
	if !g.hasRuntimeI18N() {
		return g.fieldLabel(entity, field)
	}
	return fmt.Sprintf("{fieldLabel(%q, %q, locale)}", field.Name, fieldLabel(field))
}

func (g *webGenerator) fieldLabelExpression(entity EntityDecl, field FieldDecl) string {
	if !g.hasRuntimeI18N() {
		return fmt.Sprintf("%q", g.fieldLabel(entity, field))
	}
	return fmt.Sprintf("fieldLabel(%q, %q, locale)", field.Name, fieldLabel(field))
}

func (g *webGenerator) computedFieldLabelJSX(entity EntityDecl, field ComputedFieldDecl) string {
	if !g.hasRuntimeI18N() {
		return g.computedFieldLabel(entity, field)
	}
	return fmt.Sprintf("{fieldLabel(%q, %q, locale)}", field.Name, computedFieldFallbackLabel(field))
}

func (g *webGenerator) computedFieldPermissionExpression(field ComputedFieldDecl) string {
	references := computedFieldReferences(field)
	if len(references) == 0 {
		return "true"
	}
	conditions := []string{}
	for _, reference := range references {
		conditions = append(conditions, fmt.Sprintf("permissions.fields.%s !== false", reference))
	}
	return strings.Join(conditions, " && ")
}

func computedFieldReferences(field ComputedFieldDecl) []string {
	references := []string{}
	seen := map[string]bool{}
	if field.Expression.Tree != nil {
		for _, operand := range expressionOperands(*field.Expression.Tree) {
			if operand.ValueKind != "reference" || operand.Value == "" || seen[operand.Value] {
				continue
			}
			seen[operand.Value] = true
			references = append(references, operand.Value)
		}
		return references
	}
	for _, operand := range []string{field.Expression.Left, field.Expression.Right} {
		if operand == "" || isNumericLiteral(operand) || seen[operand] {
			continue
		}
		seen[operand] = true
		references = append(references, operand)
	}
	return references
}

func (g *webGenerator) defaultFieldLabelTranslation(entityName string, fieldName string) string {
	return g.defaultFieldTextTranslation("label", entityName, fieldName)
}

func (g *webGenerator) defaultFieldTextTranslation(kind string, entityName string, fieldName string) string {
	if g.program.I18N == nil || g.program.I18N.Default == "" {
		return ""
	}
	translations := g.fieldTextTranslations(kind, entityName, fieldName)
	return translations[g.program.I18N.Default]
}

func (g *webGenerator) fieldPlaceholderJSX(entity EntityDecl, field FieldDecl, fallback string) string {
	text := fieldPlaceholder(field, fallback)
	if !g.hasRuntimeI18N() {
		return escapeJSXText(text)
	}
	if text == "" && len(g.fieldTextTranslations("placeholder", entity.Name, field.Name)) == 0 {
		return ""
	}
	return fmt.Sprintf("{fieldPlaceholder(%q, %q, locale)}", field.Name, text)
}

func (g *webGenerator) inputPlaceholderAttribute(entity EntityDecl, field FieldDecl) string {
	fallback := modifierValue(field, "placeholder")
	if !g.hasRuntimeI18N() {
		if fallback == "" {
			return ""
		}
		return fmt.Sprintf(" placeholder=%q", fallback)
	}
	if fallback == "" && len(g.fieldTextTranslations("placeholder", entity.Name, field.Name)) == 0 {
		return ""
	}
	return fmt.Sprintf(" placeholder={fieldPlaceholder(%q, %q, locale)}", field.Name, fallback)
}

func (g *webGenerator) fieldHelpElement(entity EntityDecl, field FieldDecl) string {
	fallback := modifierValue(field, "help")
	if !g.hasRuntimeI18N() {
		if fallback == "" {
			return ""
		}
		return fmt.Sprintf("<span className=\"field-note\">%s</span>\n              ", escapeJSXText(fallback))
	}
	if fallback == "" && len(g.fieldTextTranslations("help", entity.Name, field.Name)) == 0 {
		return ""
	}
	return fmt.Sprintf("{fieldHelpText(%q, %q, locale) && <span className=\"field-note\">{fieldHelpText(%q, %q, locale)}</span>}\n              ", field.Name, fallback, field.Name, fallback)
}

func (g *webGenerator) fieldValidationMessageExpression(entity EntityDecl, field FieldDecl, fallback string) string {
	message := fieldValidationMessage(field, fallback)
	if !g.hasRuntimeI18N() {
		return fmt.Sprintf("%q", message)
	}
	return fmt.Sprintf("fieldMessage(%q, %q, locale)", field.Name, message)
}

func fieldPlaceholder(field FieldDecl, fallback string) string {
	if placeholder := modifierValue(field, "placeholder"); placeholder != "" {
		return placeholder
	}
	return fallback
}

func placeholderAttribute(field FieldDecl) string {
	if placeholder := modifierValue(field, "placeholder"); placeholder != "" {
		return fmt.Sprintf(" placeholder=%q", placeholder)
	}
	return ""
}

func helpElement(field FieldDecl) string {
	if help := modifierValue(field, "help"); help != "" {
		return fmt.Sprintf("<span className=\"field-note\">%s</span>\n              ", help)
	}
	return ""
}

func inputAttributes(field FieldDecl) string {
	attributes := []string{}
	switch field.Type {
	case "number", "integer", "decimal", "money":
		attributes = append(attributes, " type=\"number\"")
		if minValue := modifierValue(field, "min"); minValue != "" {
			attributes = append(attributes, fmt.Sprintf(" min=%q", minValue))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			attributes = append(attributes, fmt.Sprintf(" max=%q", maxValue))
		}
	case "email":
		attributes = append(attributes, " type=\"email\"")
	case "date":
		attributes = append(attributes, " type=\"date\"")
	case "datetime":
		attributes = append(attributes, " type=\"datetime-local\"")
	case "file", "image":
		attributes = append(attributes, " type=\"file\"")
		accept := modifierValue(field, "accept")
		if accept == "" && field.Type == "image" {
			accept = "image/*"
		}
		if accept != "" {
			attributes = append(attributes, fmt.Sprintf(" accept=%q", accept))
		}
	default:
		if hasModifier(field, "url") {
			attributes = append(attributes, " type=\"url\"")
		} else {
			attributes = append(attributes, " type=\"text\"")
		}
	}
	if hasModifier(field, "required") {
		attributes = append(attributes, " required")
	}
	if minLength, maxLength, ok := fieldLengthBounds(field); ok {
		attributes = append(attributes, fmt.Sprintf(" minLength={%d}", minLength))
		attributes = append(attributes, fmt.Sprintf(" maxLength={%d}", maxLength))
	}
	if pattern := modifierValue(field, "regex"); pattern != "" {
		attributes = append(attributes, fmt.Sprintf(" pattern=%q", pattern))
	}
	return strings.Join(attributes, "")
}

func fieldLengthBounds(field FieldDecl) (int, int, bool) {
	value := modifierValue(field, "length")
	if value == "" {
		return 0, 0, false
	}
	return parseLengthConstraint(value)
}

func selectAttributes(field FieldDecl) string {
	if hasModifier(field, "required") {
		return " required"
	}
	return ""
}

func validationFieldBlock(field FieldDecl) string {
	var builder strings.Builder
	name := field.Name
	builder.WriteString(fmt.Sprintf("  if (input.%s === undefined || input.%s === null || input.%s === \"\") {\n", name, name, name))
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		builder.WriteString(fmt.Sprintf("    value.%s = %s;\n", name, typedLiteral(field, defaultValue)))
	} else if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.push(%q);\n", fieldValidationMessage(field, name+" is required")))
	} else {
		builder.WriteString(fmt.Sprintf("    value.%s = undefined;\n", name))
	}
	builder.WriteString("  } else {\n")
	switch field.Type {
	case "number", "integer":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(input.%s);\n", name))
		builder.WriteString("    if (!Number.isInteger(parsed) || parsed < -2147483648 || parsed > 2147483647) {\n")
		builder.WriteString(fmt.Sprintf("      errors.push(%q);\n", fieldValidationMessage(field, name+" must be a whole number")))
		builder.WriteString("    } else {\n")
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed < %s) errors.push(%q);\n", minValue, fieldValidationMessage(field, name+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed > %s) errors.push(%q);\n", maxValue, fieldValidationMessage(field, name+" must be at most "+maxValue)))
		}
		builder.WriteString(fmt.Sprintf("      value.%s = parsed;\n", name))
		builder.WriteString("    }\n")
	case "decimal", "money":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(input.%s);\n", name))
		builder.WriteString("    if (!Number.isFinite(parsed)) {\n")
		builder.WriteString(fmt.Sprintf("      errors.push(%q);\n", fieldValidationMessage(field, name+" must be a number")))
		builder.WriteString("    } else {\n")
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed < %s) errors.push(%q);\n", minValue, fieldValidationMessage(field, name+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed > %s) errors.push(%q);\n", maxValue, fieldValidationMessage(field, name+" must be at most "+maxValue)))
		}
		builder.WriteString(fmt.Sprintf("      value.%s = parsed;\n", name))
		builder.WriteString("    }\n")
	case "boolean":
		builder.WriteString(fmt.Sprintf("    value.%s = Boolean(input.%s);\n", name, name))
	case "email":
		builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", name, name))
		builder.WriteString(fmt.Sprintf("    if (!value.%s.includes(\"@\")) errors.push(%q);\n", name, fieldValidationMessage(field, name+" must be an email")))
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    if (value.%s.length < %d || value.%s.length > %d) errors.push(%q);\n", name, minLength, name, maxLength, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", name, minLength, maxLength))))
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!(new RegExp(%q)).test(value.%s)) errors.push(%q);\n", pattern, name, fieldValidationMessage(field, name+" has an invalid format")))
		}
	case "file", "image":
		builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", name, name))
		builder.WriteString(fmt.Sprintf("    if (!%s) errors.push(%q);\n", mediaValidationExpression("value."+name, field.Type), fieldValidationMessage(field, name+" must be a valid "+field.Type)))
	default:
		builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", name, name))
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			builder.WriteString(fmt.Sprintf("    if (value.%s.length < %d || value.%s.length > %d) errors.push(%q);\n", name, minLength, name, maxLength, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", name, minLength, maxLength))))
		}
		if hasModifier(field, "url") {
			builder.WriteString(fmt.Sprintf("    try { new URL(value.%s); } catch { errors.push(%q); }\n", name, fieldValidationMessage(field, name+" must be a valid URL")))
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			builder.WriteString(fmt.Sprintf("    if (!(new RegExp(%q)).test(value.%s)) errors.push(%q);\n", pattern, name, fieldValidationMessage(field, name+" has an invalid format")))
		}
	}
	builder.WriteString("  }\n\n")
	return builder.String()
}

func (g *webGenerator) entityValidationBlocks(entity EntityDecl) string {
	if len(entity.Validations) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, validation := range entity.Validations {
		if validation.Required && validation.When != nil {
			condition := g.validationConditionExpression(entity, *validation.When, "value")
			if condition == "" {
				continue
			}
			builder.WriteString(fmt.Sprintf("  if ((%s) && (value.%s === undefined || value.%s === null || String(value.%s).trim() === \"\")) errors.push(%q);\n", condition, validation.Left, validation.Left, validation.Left, entityValidationMessage(entity, validation)))
			continue
		}
		expression := g.entityValidationExpression(entity, validation, "value")
		if expression == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("  if (!(%s)) errors.push(%q);\n", expression, entityValidationMessage(entity, validation)))
	}
	if builder.Len() > 0 {
		builder.WriteString("\n")
	}
	return builder.String()
}

func (g *webGenerator) formEntityValidationBlocks(entity EntityDecl, formFields []string) string {
	if len(entity.Validations) == 0 {
		return ""
	}
	formFieldIndex := map[string]bool{}
	for _, field := range formFields {
		formFieldIndex[field] = true
	}
	var builder strings.Builder
	for _, validation := range entity.Validations {
		if validation.Required && validation.When != nil {
			if !formFieldIndex[validation.Left] || !formFieldIndex[validation.When.Left] {
				continue
			}
			condition := g.validationConditionExpression(entity, *validation.When, "form")
			if condition == "" {
				continue
			}
			builder.WriteString(fmt.Sprintf("  if ((%s) && form.%s.trim() === \"\") {\n", condition, validation.Left))
			builder.WriteString(fmt.Sprintf("    errors.%s = %q;\n", validation.Left, entityValidationMessage(entity, validation)))
			builder.WriteString("  }\n\n")
			continue
		}
		if !formFieldIndex[validation.Left] || !formFieldIndex[validation.Right] {
			continue
		}
		expression := g.entityValidationExpression(entity, validation, "form")
		if expression == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("  if (form.%s.trim() !== \"\" && form.%s.trim() !== \"\" && !(%s)) {\n", validation.Left, validation.Right, expression))
		builder.WriteString(fmt.Sprintf("    errors.%s = %q;\n", validation.Left, entityValidationMessage(entity, validation)))
		builder.WriteString("  }\n\n")
	}
	return builder.String()
}

func (g *webGenerator) validationConditionExpression(entity EntityDecl, condition ValidationConditionDecl, objectName string) string {
	left, leftOK := findField(entity, condition.Left)
	if !leftOK {
		return ""
	}
	if right, rightOK := findField(entity, condition.Right); rightOK {
		if numberLikeField(left) && numberLikeField(right) {
			return fmt.Sprintf("Number(%s.%s) %s Number(%s.%s)", objectName, condition.Left, condition.Operator, objectName, condition.Right)
		}
		if condition.Operator == "==" || condition.Operator == "!=" {
			return fmt.Sprintf("String(%s.%s) %s String(%s.%s)", objectName, condition.Left, condition.Operator, objectName, condition.Right)
		}
		return ""
	}
	if numberLikeField(left) {
		return fmt.Sprintf("Number(%s.%s) %s %s", objectName, condition.Left, condition.Operator, condition.Right)
	}
	if condition.Operator == "==" || condition.Operator == "!=" {
		return fmt.Sprintf("String(%s.%s) %s %q", objectName, condition.Left, condition.Operator, condition.Right)
	}
	return ""
}

func (g *webGenerator) entityValidationExpression(entity EntityDecl, validation EntityValidationDecl, objectName string) string {
	left, leftOK := findField(entity, validation.Left)
	right, rightOK := findField(entity, validation.Right)
	if !leftOK || !rightOK {
		return ""
	}
	if numberLikeField(left) && numberLikeField(right) {
		return fmt.Sprintf("Number(%s.%s) %s Number(%s.%s)", objectName, validation.Left, validation.Operator, objectName, validation.Right)
	}
	if validation.Operator == "==" || validation.Operator == "!=" {
		return fmt.Sprintf("String(%s.%s) %s String(%s.%s)", objectName, validation.Left, validation.Operator, objectName, validation.Right)
	}
	return ""
}

func entityValidationMessage(entity EntityDecl, validation EntityValidationDecl) string {
	if validation.Message != "" {
		return validation.Message
	}
	if validation.Required && validation.When != nil {
		return fmt.Sprintf("%s.%s is required when %s.%s %s %s", entity.Name, validation.Left, entity.Name, validation.When.Left, validation.When.Operator, validation.When.Right)
	}
	return fmt.Sprintf("%s.%s must be %s %s.%s", entity.Name, validation.Left, validation.Operator, entity.Name, validation.Right)
}

func typedLiteral(field FieldDecl, value string) string {
	switch field.Type {
	case "number", "integer", "decimal", "money":
		return value
	case "boolean":
		if value == "true" {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%q", value)
	}
}

func modifierValue(field FieldDecl, name string) string {
	for _, modifier := range field.Modifiers {
		if modifier.Name == name {
			return modifier.Value
		}
	}
	return ""
}

func computedModifierValue(field ComputedFieldDecl, name string) string {
	for _, modifier := range field.Modifiers {
		if modifier.Name == name {
			return modifier.Value
		}
	}
	return ""
}

func fieldValidationMessage(field FieldDecl, fallback string) string {
	if message := modifierValue(field, "message"); message != "" {
		return message
	}
	return fallback
}

func title(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
