package main

import (
	"fmt"
	"strings"
)

type migrationColumn struct {
	Field          string
	Column         string
	SourceType     string
	PrismaType     string
	SQLiteType     string
	Required       bool
	Unique         bool
	Default        string
	HasDefault     bool
	Relation       bool
	RelationTarget string
}

func MigratePlan(args []string) SchemaMigrationPlanResult {
	files := nonOptionArgs(args)
	if len(files) < 2 {
		diagnostic := Diagnostic{
			Code:       "MISSING_SCHEMA_MIGRATION_FILES",
			Message:    "Schema migration plan requires an old .black file and a new .black file.",
			Suggestion: "Use `black migrate plan old.black new.black --json`.",
		}
		return SchemaMigrationPlanResult{
			Success:        false,
			Command:        "migrate plan",
			Version:        version,
			Safe:           false,
			Destructive:    false,
			Changes:        []SchemaMigrationChange{},
			Steps:          []SchemaMigrationStep{},
			GeneratedFiles: migrationGeneratedFiles(),
			AgentNotes:     migrationAgentNotes(),
			Errors:         []Diagnostic{diagnostic},
		}
	}
	return AnalyzeSchemaMigrationPlan(files[0], files[1])
}

func AnalyzeSchemaMigrationPlan(oldFile string, newFile string) SchemaMigrationPlanResult {
	oldProgram, oldDiagnostics := loadMigrationProgram(oldFile)
	newProgram, newDiagnostics := loadMigrationProgram(newFile)
	result := SchemaMigrationPlanResult{
		Success:        false,
		Command:        "migrate plan",
		Version:        version,
		OldFile:        oldFile,
		NewFile:        newFile,
		Safe:           false,
		Destructive:    false,
		Summary:        SchemaMigrationSummary{OldApp: oldProgram.App.Name, NewApp: newProgram.App.Name},
		Changes:        []SchemaMigrationChange{},
		Steps:          []SchemaMigrationStep{},
		GeneratedFiles: migrationGeneratedFiles(),
		AgentNotes:     migrationAgentNotes(),
		Errors:         append(append([]Diagnostic{}, oldDiagnostics...), newDiagnostics...),
	}
	if len(result.Errors) > 0 {
		return result
	}

	result.Changes = compareSchemaPrograms(oldProgram, newProgram)
	result.Summary = summarizeSchemaMigration(oldProgram, newProgram, result.Changes)
	result.Destructive = result.Summary.DestructiveChanges > 0
	result.Safe = result.Summary.ManualChanges == 0 && result.Summary.DestructiveChanges == 0
	result.Success = true
	result.Steps = schemaMigrationSteps(result)
	return result
}

func loadMigrationProgram(file string) (Program, []Diagnostic) {
	source, readDiagnostics := ReadBlackSource(file)
	if len(readDiagnostics) > 0 {
		return Program{}, readDiagnostics
	}
	program, parseDiagnostics := Parse(file, source)
	validateDiagnostics := Validate(program)
	diagnostics := append([]Diagnostic{}, parseDiagnostics...)
	diagnostics = append(diagnostics, validateDiagnostics...)
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	return program, diagnostics
}

func compareSchemaPrograms(oldProgram Program, newProgram Program) []SchemaMigrationChange {
	changes := []SchemaMigrationChange{}
	oldEntities := migrationEntityIndex(oldProgram)
	newEntities := migrationEntityIndex(newProgram)
	entityRenames := migrationEntityRenameMap(newProgram)
	processedOldEntities := map[string]bool{}
	processedNewEntities := map[string]bool{}

	for _, migration := range newProgram.Migrations {
		for _, rename := range migration.Renames {
			if rename.Kind != "entity" {
				continue
			}
			oldEntity, oldOK := oldEntities[rename.From]
			newEntity, newOK := newEntities[rename.To]
			if !oldOK || !newOK {
				continue
			}
			changes = append(changes, SchemaMigrationChange{
				Type:     "rename-table",
				Risk:     "safe",
				Entity:   newEntity.Name,
				OldValue: oldEntity.Name,
				NewValue: newEntity.Name,
				Reason:   "Explicit migration rename preserves stored rows when the old table exists and the new table does not.",
			})
			processedOldEntities[oldEntity.Name] = true
			processedNewEntities[newEntity.Name] = true
			changes = append(changes, compareEntityColumns(oldProgram, newProgram, oldEntity, newEntity, migrationFieldRenamesForEntity(newProgram, newEntity.Name), entityRenames)...)
			changes = append(changes, compareEntityIndexes(oldEntity, newEntity, migrationFieldRenameMap(newProgram, newEntity.Name))...)
		}
	}

	for _, newEntity := range newProgram.Entities {
		if processedNewEntities[newEntity.Name] {
			continue
		}
		oldEntity, ok := oldEntities[newEntity.Name]
		if !ok {
			continue
		}
		processedOldEntities[oldEntity.Name] = true
		processedNewEntities[newEntity.Name] = true
		changes = append(changes, compareEntityColumns(oldProgram, newProgram, oldEntity, newEntity, migrationFieldRenamesForEntity(newProgram, newEntity.Name), entityRenames)...)
		changes = append(changes, compareEntityIndexes(oldEntity, newEntity, migrationFieldRenameMap(newProgram, newEntity.Name))...)
	}

	for _, entity := range newProgram.Entities {
		if processedNewEntities[entity.Name] {
			continue
		}
		if _, ok := oldEntities[entity.Name]; !ok {
			changes = append(changes, SchemaMigrationChange{
				Type:   "create-table",
				Risk:   "safe",
				Entity: entity.Name,
				Reason: "New entity creates a new generated database table.",
			})
		}
	}

	for _, entity := range oldProgram.Entities {
		if processedOldEntities[entity.Name] {
			continue
		}
		if _, ok := newEntities[entity.Name]; !ok {
			changes = append(changes, SchemaMigrationChange{
				Type:   "drop-table",
				Risk:   "destructive",
				Entity: entity.Name,
				Reason: "Removed entity drops a generated database table and stored rows unless data is preserved manually.",
			})
		}
	}

	return changes
}

func compareEntityColumns(oldProgram Program, newProgram Program, oldEntity EntityDecl, newEntity EntityDecl, fieldRenames []MigrationRenameDecl, entityRenames map[string]string) []SchemaMigrationChange {
	changes := []SchemaMigrationChange{}
	oldFields := fieldIndex(oldEntity)
	newFields := fieldIndex(newEntity)
	processedOldFields := map[string]bool{}
	processedNewFields := map[string]bool{}

	for _, newField := range newEntity.Fields {
		oldField, ok := oldFields[newField.Name]
		if !ok {
			continue
		}
		oldColumn := schemaMigrationColumn(oldProgram, oldField)
		newColumn := schemaMigrationColumn(newProgram, newField)
		changes = append(changes, compareMigrationColumn(newEntity.Name, oldColumn, newColumn, entityRenames)...)
		processedOldFields[oldField.Name] = true
		processedNewFields[newField.Name] = true
	}

	for _, rename := range fieldRenames {
		oldField, oldOK := oldFields[rename.From]
		newField, newOK := newFields[rename.To]
		if !oldOK || !newOK {
			continue
		}
		oldColumn := schemaMigrationColumn(oldProgram, oldField)
		newColumn := schemaMigrationColumn(newProgram, newField)
		changes = append(changes, SchemaMigrationChange{
			Type:     "rename-column",
			Risk:     "safe",
			Entity:   newEntity.Name,
			Field:    newField.Name,
			Column:   newColumn.Column,
			OldType:  oldColumn.SourceType,
			NewType:  newColumn.SourceType,
			OldValue: oldColumn.Column,
			NewValue: newColumn.Column,
			Reason:   "Explicit migration rename preserves stored values when the old column exists and the new column does not.",
		})
		alignedOldColumn := oldColumn
		alignedOldColumn.Field = newColumn.Field
		alignedOldColumn.Column = newColumn.Column
		changes = append(changes, compareMigrationColumn(newEntity.Name, alignedOldColumn, newColumn, entityRenames)...)
		processedOldFields[oldField.Name] = true
		processedNewFields[newField.Name] = true
	}

	for _, field := range newEntity.Fields {
		if processedNewFields[field.Name] {
			continue
		}
		if _, ok := oldFields[field.Name]; ok {
			continue
		}
		column := schemaMigrationColumn(newProgram, field)
		risk := "safe"
		reason := "New optional or defaulted field can be added without losing existing data."
		if column.Required && !column.HasDefault {
			risk = "manual"
			reason = "New required field needs a value for existing rows before the column can be enforced."
		}
		changes = append(changes, SchemaMigrationChange{
			Type:     "add-column",
			Risk:     risk,
			Entity:   newEntity.Name,
			Field:    field.Name,
			Column:   column.Column,
			NewType:  column.SourceType,
			NewValue: columnDefaultValue(column),
			Reason:   reason,
		})
	}

	for _, field := range oldEntity.Fields {
		if processedOldFields[field.Name] {
			continue
		}
		if _, ok := newFields[field.Name]; ok {
			continue
		}
		column := schemaMigrationColumn(oldProgram, field)
		changes = append(changes, SchemaMigrationChange{
			Type:    "drop-column",
			Risk:    "destructive",
			Entity:  oldEntity.Name,
			Field:   field.Name,
			Column:  column.Column,
			OldType: column.SourceType,
			Reason:  "Removed field drops a generated database column and stored values unless data is preserved manually.",
		})
	}

	return changes
}

func compareMigrationColumn(entity string, oldColumn migrationColumn, newColumn migrationColumn, entityRenames map[string]string) []SchemaMigrationChange {
	changes := []SchemaMigrationChange{}
	if oldColumn.Column != newColumn.Column {
		changes = append(changes, SchemaMigrationChange{
			Type:     "change-column-name",
			Risk:     "destructive",
			Entity:   entity,
			Field:    newColumn.Field,
			Column:   newColumn.Column,
			OldValue: oldColumn.Column,
			NewValue: newColumn.Column,
			Reason:   "Generated column name changed; treat it as a manual data move before deployment.",
		})
	}
	sourceTypeChanged := oldColumn.SourceType != newColumn.SourceType
	if oldColumn.Relation && newColumn.Relation && migrationRenamedEntityTarget(oldColumn.RelationTarget, newColumn.RelationTarget, entityRenames) {
		sourceTypeChanged = false
	}
	if sourceTypeChanged || oldColumn.PrismaType != newColumn.PrismaType || oldColumn.SQLiteType != newColumn.SQLiteType {
		changes = append(changes, SchemaMigrationChange{
			Type:     "change-column-type",
			Risk:     "destructive",
			Entity:   entity,
			Field:    newColumn.Field,
			Column:   newColumn.Column,
			OldType:  oldColumn.SourceType,
			NewType:  newColumn.SourceType,
			OldValue: oldColumn.SQLiteType,
			NewValue: newColumn.SQLiteType,
			Reason:   "Generated database type changed; existing data may need conversion before applying the new schema.",
		})
	}
	if oldColumn.Relation && newColumn.Relation && oldColumn.RelationTarget != newColumn.RelationTarget && !migrationRenamedEntityTarget(oldColumn.RelationTarget, newColumn.RelationTarget, entityRenames) {
		changes = append(changes, SchemaMigrationChange{
			Type:     "change-relation-target",
			Risk:     "manual",
			Entity:   entity,
			Field:    newColumn.Field,
			Column:   newColumn.Column,
			OldType:  oldColumn.RelationTarget,
			NewType:  newColumn.RelationTarget,
			OldValue: oldColumn.RelationTarget,
			NewValue: newColumn.RelationTarget,
			Reason:   "Relation target changed; verify existing foreign key values before applying the schema.",
		})
	}
	if !oldColumn.Required && newColumn.Required {
		changes = append(changes, SchemaMigrationChange{
			Type:    "make-column-required",
			Risk:    "manual",
			Entity:  entity,
			Field:   newColumn.Field,
			Column:  newColumn.Column,
			NewType: newColumn.SourceType,
			Reason:  "Existing rows must be backfilled before this column can be required.",
		})
	}
	if oldColumn.Required && !newColumn.Required {
		changes = append(changes, SchemaMigrationChange{
			Type:    "make-column-optional",
			Risk:    "safe",
			Entity:  entity,
			Field:   newColumn.Field,
			Column:  newColumn.Column,
			OldType: oldColumn.SourceType,
			NewType: newColumn.SourceType,
			Reason:  "Column nullability is relaxed.",
		})
	}
	if !oldColumn.Unique && newColumn.Unique {
		changes = append(changes, SchemaMigrationChange{
			Type:   "add-unique",
			Risk:   "manual",
			Entity: entity,
			Field:  newColumn.Field,
			Column: newColumn.Column,
			Reason: "Existing rows must be checked for duplicates before adding a unique constraint.",
		})
	}
	if oldColumn.Unique && !newColumn.Unique {
		changes = append(changes, SchemaMigrationChange{
			Type:   "drop-unique",
			Risk:   "safe",
			Entity: entity,
			Field:  newColumn.Field,
			Column: newColumn.Column,
			Reason: "Unique constraint is removed; stored data is preserved.",
		})
	}
	if !oldColumn.HasDefault && newColumn.HasDefault {
		changes = append(changes, SchemaMigrationChange{
			Type:     "add-default",
			Risk:     "safe",
			Entity:   entity,
			Field:    newColumn.Field,
			Column:   newColumn.Column,
			NewValue: newColumn.Default,
			Reason:   "Default value affects future inserts and does not rewrite existing rows by itself.",
		})
	}
	if oldColumn.HasDefault && newColumn.HasDefault && oldColumn.Default != newColumn.Default {
		changes = append(changes, SchemaMigrationChange{
			Type:     "change-default",
			Risk:     "safe",
			Entity:   entity,
			Field:    newColumn.Field,
			Column:   newColumn.Column,
			OldValue: oldColumn.Default,
			NewValue: newColumn.Default,
			Reason:   "Default value change affects future inserts.",
		})
	}
	if oldColumn.HasDefault && !newColumn.HasDefault {
		changes = append(changes, SchemaMigrationChange{
			Type:     "drop-default",
			Risk:     "safe",
			Entity:   entity,
			Field:    newColumn.Field,
			Column:   newColumn.Column,
			OldValue: oldColumn.Default,
			Reason:   "Default value is removed; stored data is preserved.",
		})
	}
	return changes
}

func compareEntityIndexes(oldEntity EntityDecl, newEntity EntityDecl, fieldRenames map[string]string) []SchemaMigrationChange {
	changes := []SchemaMigrationChange{}
	oldIndexes := migrationIndexMap(oldEntity, fieldRenames)
	newIndexes := migrationIndexMap(newEntity)

	for _, index := range newEntity.Indexes {
		signature := migrationIndexSignature(index, nil)
		if _, ok := oldIndexes[signature]; ok {
			continue
		}
		changes = append(changes, SchemaMigrationChange{
			Type:        "add-index",
			Risk:        "safe",
			Entity:      newEntity.Name,
			Index:       entityIndexName(newEntity, index),
			IndexFields: append([]string{}, index.Fields...),
			Reason:      "New index improves query/read performance without changing stored row shape.",
		})
	}

	for _, index := range oldEntity.Indexes {
		signature := migrationIndexSignature(index, fieldRenames)
		if _, ok := newIndexes[signature]; ok {
			continue
		}
		changes = append(changes, SchemaMigrationChange{
			Type:        "drop-index",
			Risk:        "safe",
			Entity:      oldEntity.Name,
			Index:       entityIndexName(oldEntity, index),
			IndexFields: append([]string{}, index.Fields...),
			Reason:      "Removed index changes performance characteristics but preserves stored row data.",
		})
	}

	return changes
}

func migrationEntityIndex(program Program) map[string]EntityDecl {
	entities := map[string]EntityDecl{}
	for _, entity := range program.Entities {
		entities[entity.Name] = entity
	}
	return entities
}

func migrationFieldRenamesForEntity(program Program, entityName string) []MigrationRenameDecl {
	renames := []MigrationRenameDecl{}
	for _, migration := range program.Migrations {
		for _, rename := range migration.Renames {
			if rename.Kind == "field" && rename.Entity == entityName {
				renames = append(renames, rename)
			}
		}
	}
	return renames
}

func migrationFieldRenameMap(program Program, entityName string) map[string]string {
	renames := map[string]string{}
	for _, rename := range migrationFieldRenamesForEntity(program, entityName) {
		renames[rename.From] = rename.To
	}
	return renames
}

func migrationEntityRenameMap(program Program) map[string]string {
	renames := map[string]string{}
	for _, migration := range program.Migrations {
		for _, rename := range migration.Renames {
			if rename.Kind == "entity" {
				renames[rename.From] = rename.To
			}
		}
	}
	return renames
}

func migrationRenamedEntityTarget(oldTarget string, newTarget string, entityRenames map[string]string) bool {
	if entityRenames == nil {
		return false
	}
	return entityRenames[oldTarget] == newTarget
}

func schemaMigrationColumn(program Program, field FieldDecl) migrationColumn {
	isRelation := migrationIsRelation(program, field)
	column := migrationColumn{
		Field:      field.Name,
		Column:     field.Name,
		SourceType: field.Type,
		PrismaType: prismaType(field.Type),
		SQLiteType: sqliteType(field.Type),
		Required:   hasModifier(field, "required") || hasModifier(field, "default"),
		Unique:     hasModifier(field, "unique"),
	}
	if value, ok := migrationModifierValue(field, "default"); ok {
		column.HasDefault = true
		column.Default = value
	}
	if isRelation {
		column.Column = relationIDFieldName(field)
		column.PrismaType = "String"
		column.SQLiteType = "TEXT"
		column.Required = hasModifier(field, "required")
		column.Relation = true
		column.RelationTarget = field.Type
	}
	return column
}

func migrationIsRelation(program Program, field FieldDecl) bool {
	for _, entity := range program.Entities {
		if entity.Name == field.Type {
			return true
		}
	}
	return false
}

func migrationModifierValue(field FieldDecl, name string) (string, bool) {
	for _, modifier := range field.Modifiers {
		if modifier.Name == name {
			return modifier.Value, true
		}
	}
	return "", false
}

func columnDefaultValue(column migrationColumn) string {
	if !column.HasDefault {
		return ""
	}
	return column.Default
}

func migrationIndexMap(entity EntityDecl, renameMaps ...map[string]string) map[string]EntityIndexDecl {
	indexes := map[string]EntityIndexDecl{}
	renames := map[string]string{}
	if len(renameMaps) > 0 && renameMaps[0] != nil {
		renames = renameMaps[0]
	}
	for _, index := range entity.Indexes {
		indexes[migrationIndexSignature(index, renames)] = index
	}
	return indexes
}

func migrationIndexSignature(index EntityIndexDecl, renames map[string]string) string {
	fields := append([]string{}, index.Fields...)
	for i, field := range fields {
		if renamed, ok := renames[field]; ok {
			fields[i] = renamed
		}
	}
	return strings.Join(fields, ",")
}

func summarizeSchemaMigration(oldProgram Program, newProgram Program, changes []SchemaMigrationChange) SchemaMigrationSummary {
	summary := SchemaMigrationSummary{OldApp: oldProgram.App.Name, NewApp: newProgram.App.Name}
	for _, change := range changes {
		switch change.Type {
		case "create-table":
			summary.EntitiesAdded++
		case "drop-table":
			summary.EntitiesRemoved++
		case "add-column":
			summary.FieldsAdded++
		case "drop-column":
			summary.FieldsRemoved++
		case "add-index":
			summary.IndexesAdded++
		case "drop-index":
			summary.IndexesRemoved++
		case "rename-table", "rename-column":
			summary.Renames++
		default:
			summary.FieldChanges++
		}

		switch change.Risk {
		case "destructive":
			summary.DestructiveChanges++
		case "manual":
			summary.ManualChanges++
		default:
			summary.SafeChanges++
		}
	}
	return summary
}

func schemaMigrationSteps(result SchemaMigrationPlanResult) []SchemaMigrationStep {
	if len(result.Changes) == 0 {
		return []SchemaMigrationStep{{
			Order:  1,
			Action: "No schema migration is needed; rebuild generated output from the new source.",
			Reason: "The old and new entity schemas have the same generated database shape.",
		}}
	}

	steps := []SchemaMigrationStep{{
		Order:  1,
		Action: "Review changes[].risk before applying the new generated schema.",
		Reason: "The plan separates safe, manual, and destructive schema changes for CI and AI agents.",
	}}
	if result.Summary.ManualChanges > 0 {
		steps = append(steps, SchemaMigrationStep{
			Order:  len(steps) + 1,
			Action: "Resolve manual changes with explicit data checks or backfills.",
			Reason: "Manual changes can fail on existing data even when they are not destructive by themselves.",
		})
	}
	if result.Destructive {
		steps = append(steps, SchemaMigrationStep{
			Order:  len(steps) + 1,
			Action: "Back up or export affected tables/columns before deployment.",
			Reason: "Destructive changes can remove stored rows, stored values, or require type conversion.",
		})
	}
	steps = append(steps,
		SchemaMigrationStep{
			Order:  len(steps) + 1,
			Action: "Run black build on the new source after the plan is accepted.",
			Reason: "The generated Prisma schema, setup runtime, and migration files are derived from the new .black source.",
		},
		SchemaMigrationStep{
			Order:  len(steps) + 2,
			Action: "Run the generated db:setup/db:push path in the target environment.",
			Reason: "First-class rename migrations are applied by the generated setup runtime before the current schema is created or pushed.",
		},
		SchemaMigrationStep{
			Order:  len(steps) + 3,
			Action: "Run black validate --json, generated npm run build, and generated npm test.",
			Reason: "Validation and generated tests confirm the new source and generated app still agree.",
		},
	)
	return steps
}

func migrationGeneratedFiles() []AffectedItem {
	return []AffectedItem{
		{Name: "prisma/schema.prisma", Reason: "Generated schema receives table, column, relation, constraint, and index shape."},
		{Name: "src/setup-db.ts", Reason: "Generated setup applies explicit rename migrations and then prepares the target database shape."},
		{Name: "migrations/manifest.json", Reason: "Generated migration manifest records first-class migration declarations for AI agents and review."},
		{Name: "migrations/*.sql", Reason: "Generated SQL files show deterministic rename operations for the selected database target."},
	}
}

func migrationAgentNotes() []string {
	return []string{
		"Run migrate plan before deploying entity field, relation, uniqueness, default, required, or index changes to an existing database.",
		"Risk values are deterministic: safe changes preserve stored data shape, manual changes need data checks/backfills, destructive changes can remove or rewrite stored data.",
		"Field and entity renames are never guessed; declare them with migration rename syntax in the new .black source.",
		"`black migrate plan` is read-only; generated db:migrate:plan inspects the target database without applying changes, db:migrate applies only declared renames, and db:setup/db:push runs the normal generated setup path.",
	}
}

func FormatSchemaMigrationPlanIR(result SchemaMigrationPlanResult) string {
	var builder strings.Builder
	status := "failed"
	if result.Success {
		status = "ok"
	}
	builder.WriteString("blackir 0.1\n")
	builder.WriteString(fmt.Sprintf("migrate plan %s\n", status))
	if result.OldFile != "" {
		builder.WriteString(fmt.Sprintf("old %s\n", result.OldFile))
	}
	if result.NewFile != "" {
		builder.WriteString(fmt.Sprintf("new %s\n", result.NewFile))
	}
	builder.WriteString(fmt.Sprintf("safe %t\n", result.Safe))
	builder.WriteString(fmt.Sprintf("destructive %t\n", result.Destructive))
	builder.WriteString(fmt.Sprintf("changes %d\n", len(result.Changes)))
	for _, change := range result.Changes {
		target := change.Entity
		if change.Field != "" {
			target += "." + change.Field
		}
		if change.Index != "" {
			target += "." + change.Index
		}
		builder.WriteString(fmt.Sprintf("  %s risk %s target %s", change.Type, change.Risk, target))
		if change.Column != "" {
			builder.WriteString(fmt.Sprintf(" column %s", change.Column))
		}
		if change.OldType != "" || change.NewType != "" {
			builder.WriteString(fmt.Sprintf(" type %s -> %s", change.OldType, change.NewType))
		}
		if change.OldValue != "" || change.NewValue != "" {
			builder.WriteString(fmt.Sprintf(" value %s -> %s", change.OldValue, change.NewValue))
		}
		if len(change.IndexFields) > 0 {
			builder.WriteString(fmt.Sprintf(" fields %s", strings.Join(change.IndexFields, " ")))
		}
		builder.WriteString("\n")
	}
	if len(result.Steps) > 0 {
		builder.WriteString("steps\n")
		for _, step := range result.Steps {
			builder.WriteString(fmt.Sprintf("  %d action %q reason %q\n", step.Order, step.Action, step.Reason))
		}
	}
	if len(result.Errors) > 0 {
		builder.WriteString("errors\n")
		for _, diagnostic := range result.Errors {
			builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", diagnostic.Code, diagnostic.File, diagnostic.Line, diagnostic.Column))
		}
	}
	return builder.String()
}
