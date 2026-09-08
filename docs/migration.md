# Schema Migration Plan, Rename Runtime, And Runner

`black migrate plan` compares an old `.black` source file with a new `.black` source file and reports generated database shape changes before deployment.

The CLI command is read-only. It does not connect to a database. First-class rename declarations in the new source are applied later by generated migration runner scripts and by generated `db:setup` / `db:push` before the current schema is created or pushed.

```bash
black migrate plan app-v1.black app-v2.black --json
black migrate plan app-v1.black app-v2.black --ir
```

The JSON output includes:

- `success`: false only when the old or new source cannot be read, parsed, or validated.
- `safe`: true when every detected schema change is safe.
- `destructive`: true when at least one change can remove stored rows, stored values, or require type conversion.
- `summary`: counts of added/removed entities, fields, indexes, and risk groups.
- `changes`: deterministic source-level schema changes.
- `steps`: deployment review steps for the detected risk profile.

Risk values:

- `safe`: preserves stored data shape, such as adding an optional/defaulted column, adding an index, or applying an explicit rename.
- `manual`: needs explicit data checks or backfills, such as adding a required field without a default or adding a unique constraint.
- `destructive`: can remove or rewrite stored data, such as dropping a table, dropping a column, changing generated column name, or changing generated type.

Field and entity renames are not guessed. Declare them in the new/current source:

```black
migration RenameProductName {
  rename entity ProductItem to Product
  rename field Product.title to name
}
```

The `to` entity or field must exist in the current source. The `from` entity or field must not remain in the current source. Generated setup applies entity renames before field renames inside one migration.

Generated apps with migration blocks also expose online migration runner commands:

```bash
npm run db:migrate:plan
npm run db:migrate
npm run db:setup
```

`db:migrate:plan` is read-only. It reports JSON with `databaseReachable`, `ready`, `summary`, per-migration `applied`/`pending` state, operation checks, `applied`, `skipped`, and redacted `errors`. SQLite plan mode does not create a missing database file. PostgreSQL plan mode reports an unreachable database without printing the connection string.

`db:migrate` explicitly applies only declared rename migrations and records them in the `BlackMigration` ledger. Run `db:setup` after `db:migrate` so the normal generated setup and seed path runs. `db:setup` still applies declared rename migrations as part of the deterministic local setup path.

Build output for migration blocks includes:

- `migrations/manifest.json`
- `migrations/*.sql`
- `src/migrate.ts`
- `src/setup-db.ts`
- `prisma/schema.prisma`
- `package.json` scripts `db:migrate:plan` and `db:migrate`

Run this before deploying entity field, relation, uniqueness, default, required, or index changes to an existing database.
