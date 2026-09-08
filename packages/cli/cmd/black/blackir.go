package main

import (
	"fmt"
	"strings"
)

func FormatBlackIR(program Program) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n\n")
	if program.App.Name != "" {
		builder.WriteString(fmt.Sprintf("app %s\n", program.App.Name))
	}

	if program.Target != nil {
		builder.WriteString(fmt.Sprintf("\ntarget %s", program.Target.Name))
		if program.Target.Frontend != "" {
			builder.WriteString(fmt.Sprintf(" frontend %s", program.Target.Frontend))
		}
		if program.Target.Backend != "" {
			builder.WriteString(fmt.Sprintf(" backend %s", program.Target.Backend))
		}
		if program.Target.Database != "" {
			builder.WriteString(fmt.Sprintf(" database %s", program.Target.Database))
		}
		builder.WriteString("\n")
	}

	if program.Auth != nil {
		builder.WriteString(fmt.Sprintf("\nauth strategy %s session %s\n", program.Auth.Strategy, program.Auth.Session))
		if len(program.Auth.User.Fields) > 0 {
			builder.WriteString("  user\n")
			for _, field := range program.Auth.User.Fields {
				builder.WriteString(fmt.Sprintf("    %s %s", field.Name, field.Type))
				for _, modifier := range field.Modifiers {
					builder.WriteString(" ")
					builder.WriteString(formatModifierIR(modifier))
				}
				if len(field.UI) > 0 {
					builder.WriteString(" ")
					builder.WriteString(formatUIIntentLine(field.UI))
				}
				builder.WriteString("\n")
			}
		}
	}

	if program.Database != nil {
		builder.WriteString("\ndatabase\n")
		if program.Database.URL.Name != "" {
			builder.WriteString(fmt.Sprintf("  url env %s\n", program.Database.URL.Name))
		}
	}

	if program.Security != nil {
		builder.WriteString("\nsecurity\n")
		if program.Security.CORS != nil {
			builder.WriteString("  cors")
			if program.Security.CORS.Origins.Name != "" {
				builder.WriteString(fmt.Sprintf(" origins env %s", program.Security.CORS.Origins.Name))
			}
			if program.Security.CORS.Credentials != "" {
				builder.WriteString(fmt.Sprintf(" credentials %s", program.Security.CORS.Credentials))
			}
			builder.WriteString("\n")
		}
	}

	if program.Deploy != nil {
		builder.WriteString("\ndeploy")
		if program.Deploy.Target != "" {
			builder.WriteString(fmt.Sprintf(" target %s", program.Deploy.Target))
		}
		builder.WriteString("\n")
		if program.Deploy.Port != nil {
			builder.WriteString(fmt.Sprintf("  port env %s default %s\n", program.Deploy.Port.Env.Name, program.Deploy.Port.Default))
		}
		for _, env := range program.Deploy.Env {
			builder.WriteString(fmt.Sprintf("  env %s %s\n", env.Name, env.Mode))
		}
		if program.Deploy.Preview != nil {
			builder.WriteString(fmt.Sprintf("  preview %s\n", program.Deploy.Preview.Mode))
		}
		if program.Deploy.Rollback != nil {
			builder.WriteString(fmt.Sprintf("  rollback %s %d\n", program.Deploy.Rollback.Strategy, program.Deploy.Rollback.Keep))
		}
		if program.Deploy.Cloud != nil {
			builder.WriteString(fmt.Sprintf("  cloud %s app env %s", program.Deploy.Cloud.Provider, program.Deploy.Cloud.App.Name))
			if program.Deploy.Cloud.Region.Name != "" {
				builder.WriteString(fmt.Sprintf(" region env %s", program.Deploy.Cloud.Region.Name))
			}
			builder.WriteString("\n")
		}
	}

	if program.Ops != nil {
		builder.WriteString("\nops\n")
		if program.Ops.Health != nil {
			builder.WriteString(fmt.Sprintf("  health path %s\n", quoteBlackString(program.Ops.Health.Path)))
		}
		if program.Ops.Readiness != nil {
			builder.WriteString(fmt.Sprintf("  readiness path %s\n", quoteBlackString(program.Ops.Readiness.Path)))
		}
		if program.Ops.Metrics != nil {
			builder.WriteString(fmt.Sprintf("  metrics path %s\n", quoteBlackString(program.Ops.Metrics.Path)))
		}
		if program.Ops.Logging != "" {
			builder.WriteString(fmt.Sprintf("  logging %s\n", program.Ops.Logging))
		}
		if program.Ops.Observe != nil {
			builder.WriteString(fmt.Sprintf("  observe %s endpoint env %s\n", program.Ops.Observe.Provider, program.Ops.Observe.Endpoint.Name))
		}
	}

	if program.I18N != nil {
		builder.WriteString("\ni18n")
		if program.I18N.Default != "" {
			builder.WriteString(fmt.Sprintf(" default %s", program.I18N.Default))
		}
		if len(program.I18N.Locales) > 0 {
			builder.WriteString(" locales ")
			builder.WriteString(strings.Join(program.I18N.Locales, " "))
		}
		builder.WriteString("\n")
	}
	for _, label := range program.Labels {
		builder.WriteString(fmt.Sprintf("\nlabel %s\n", label.Target))
		for _, translation := range label.Translations {
			builder.WriteString(fmt.Sprintf("  %s %s\n", translation.Locale, quoteBlackString(translation.Text)))
		}
	}
	writeFieldTextTranslationIR(&builder, "placeholder", program.Placeholders)
	writeFieldTextTranslationIR(&builder, "help", program.HelpTexts)
	writeFieldTextTranslationIR(&builder, "message", program.Messages)

	for _, entity := range program.Entities {
		builder.WriteString(fmt.Sprintf("\nentity %s\n", entity.Name))
		for _, field := range entity.Fields {
			builder.WriteString(fmt.Sprintf("  %s %s", field.Name, field.Type))
			for _, modifier := range field.Modifiers {
				builder.WriteString(" ")
				builder.WriteString(formatModifierIR(modifier))
			}
			if len(field.UI) > 0 {
				builder.WriteString(" ")
				builder.WriteString(formatUIIntentLine(field.UI))
			}
			builder.WriteString("\n")
		}
		for _, field := range entity.ComputedFields {
			expression := fmt.Sprintf("%s %s %s", field.Expression.Left, field.Expression.Operator, field.Expression.Right)
			if field.Expression.Tree != nil {
				expression = formatCoreExpression(*field.Expression.Tree)
			}
			builder.WriteString(fmt.Sprintf("  computed %s %s = %s", field.Name, field.Type, expression))
			for _, modifier := range field.Modifiers {
				builder.WriteString(" ")
				builder.WriteString(formatModifierIR(modifier))
			}
			builder.WriteString("\n")
		}
		for _, index := range entity.Indexes {
			builder.WriteString(fmt.Sprintf("  index %s\n", strings.Join(index.Fields, " ")))
		}
		for _, policy := range entity.Policies {
			builder.WriteString(fmt.Sprintf("  policy %s %s\n", policy.Kind, policy.Field))
		}
		for _, validation := range entity.Validations {
			if validation.Required && validation.When != nil {
				builder.WriteString(fmt.Sprintf("  validate %s required when %s %s %s", validation.Left, validation.When.Left, validation.When.Operator, validation.When.Right))
			} else {
				builder.WriteString(fmt.Sprintf("  validate %s %s %s", validation.Left, validation.Operator, validation.Right))
			}
			if validation.Message != "" {
				builder.WriteString(" message ")
				builder.WriteString(validation.Message)
			}
			builder.WriteString("\n")
		}
	}

	for _, migration := range program.Migrations {
		builder.WriteString(fmt.Sprintf("\nmigration %s\n", migration.Name))
		for _, rename := range migration.Renames {
			switch rename.Kind {
			case "entity":
				builder.WriteString(fmt.Sprintf("  rename entity %s to %s\n", rename.From, rename.To))
			case "field":
				builder.WriteString(fmt.Sprintf("  rename field %s.%s to %s\n", rename.Entity, rename.From, rename.To))
			}
		}
	}

	for _, seed := range program.Seeds {
		builder.WriteString(fmt.Sprintf("\nseed %s source %s rows %d\n", seed.Name, seed.Source, len(seed.Rows)))
		for _, row := range seed.Rows {
			builder.WriteString(fmt.Sprintf("  row %s\n", row.Key))
			for _, value := range row.Values {
				builder.WriteString(fmt.Sprintf("    %s %s\n", value.Field, formatSeedLiteral(value.Value)))
			}
		}
	}

	for _, test := range program.Tests {
		builder.WriteString(fmt.Sprintf("\ntest %s page %s expectations %d\n", test.Name, test.Page, len(test.Expectations)))
		for _, expectation := range test.Expectations {
			value := expectation.Value
			if expectation.Kind == "text" {
				value = quoteBlackString(value)
			}
			builder.WriteString(fmt.Sprintf("  expect %s %s\n", expectation.Kind, value))
		}
	}

	for _, query := range program.Queries {
		builder.WriteString(fmt.Sprintf("\nquery %s source %s\n", query.Name, query.Source))
		for _, filter := range query.Where {
			value := filter.Value.Value
			if filter.Value.Kind == "string" {
				value = quoteBlackString(value)
			}
			builder.WriteString(fmt.Sprintf("  where %s %s %s\n", filter.Field, filter.Operator, value))
		}
		for _, aggregate := range query.Aggregates {
			if aggregate.Field != "" {
				builder.WriteString(fmt.Sprintf("  aggregate %s %s %s\n", aggregate.Name, aggregate.Function, aggregate.Field))
			} else {
				builder.WriteString(fmt.Sprintf("  aggregate %s %s\n", aggregate.Name, aggregate.Function))
			}
		}
		if query.Sort.Field != "" {
			builder.WriteString(fmt.Sprintf("  sort %s %s\n", query.Sort.Field, query.Sort.Direction))
		}
		if query.Limit > 0 {
			builder.WriteString(fmt.Sprintf("  limit %d\n", query.Limit))
		}
	}

	for _, job := range program.Jobs {
		builder.WriteString(fmt.Sprintf("\njob %s schedule %s\n", job.Name, jobScheduleLabel(job)))
		if job.Run.Kind == "query" {
			builder.WriteString(fmt.Sprintf("  run query %s\n", job.Run.Query))
		}
	}

	for _, action := range program.Actions {
		builder.WriteString(fmt.Sprintf("\naction %s source %s\n", action.Name, action.Source))
		for _, input := range action.Inputs {
			builder.WriteString(fmt.Sprintf("  input %s %s", input.Name, input.Type))
			for _, modifier := range input.Modifiers {
				builder.WriteString(" ")
				builder.WriteString(modifier.Name)
				if modifier.Value != "" {
					builder.WriteString(" ")
					if modifier.Name == "label" || modifier.Name == "placeholder" || modifier.Name == "help" || modifier.Name == "message" {
						builder.WriteString(quoteBlackString(modifier.Value))
					} else {
						builder.WriteString(modifier.Value)
					}
				}
			}
			builder.WriteString("\n")
		}
		builder.WriteString(formatActionStatementsIR(actionStatements(action), "  "))
		if len(action.Allow) > 0 {
			builder.WriteString("  allow ")
			builder.WriteString(strings.Join(action.Allow, " "))
			builder.WriteString("\n")
		}
		if action.Success != "" {
			builder.WriteString("  success ")
			builder.WriteString(quoteBlackString(action.Success))
			builder.WriteString("\n")
		}
	}

	for _, transaction := range program.Transactions {
		builder.WriteString(fmt.Sprintf("\ntransaction %s\n", transaction.Name))
		for _, target := range transaction.Actions {
			builder.WriteString(fmt.Sprintf("  action %s\n", target.Name))
		}
		for _, target := range transaction.APIs {
			builder.WriteString(fmt.Sprintf("  api %s\n", target.Name))
		}
	}

	for _, service := range program.Services {
		builder.WriteString(fmt.Sprintf("\nservice %s\n", service.Name))
		for _, target := range service.APIs {
			builder.WriteString(fmt.Sprintf("  api %s\n", target.Name))
		}
	}

	for _, role := range program.Roles {
		builder.WriteString(fmt.Sprintf("\nrole %s\n", role.Name))
		for _, permission := range role.Permissions {
			builder.WriteString(fmt.Sprintf("  %s %s", permission.Effect, permission.Action))
			if permission.Resource != "" {
				builder.WriteString(" ")
				builder.WriteString(permission.Resource)
			}
			if len(permission.Fields) > 0 {
				builder.WriteString(" ")
				builder.WriteString(strings.Join(permission.Fields, " "))
			}
			builder.WriteString("\n")
		}
	}

	for _, api := range program.APIs {
		builder.WriteString(fmt.Sprintf("\napi %s method %s path %s", api.Name, strings.ToUpper(api.Method), api.Path))
		if api.Access != "" {
			builder.WriteString(" ")
			builder.WriteString(api.Access)
		}
		if api.Webhook {
			builder.WriteString(" webhook")
		}
		builder.WriteString("\n")
		for _, param := range api.Params {
			builder.WriteString(fmt.Sprintf("  param %s %s\n", param.Name, param.Type))
		}
		for _, query := range api.Queries {
			builder.WriteString(fmt.Sprintf("  query %s %s\n", query.Name, query.Type))
		}
		for _, body := range api.Body {
			modifiers := formatModifiers(body.Modifiers)
			if modifiers != "" {
				modifiers = " " + modifiers
			}
			builder.WriteString(fmt.Sprintf("  body %s %s%s\n", body.Name, body.Type, modifiers))
		}
		if api.Update != nil {
			if apiUpdateUsesBlock(*api.Update) {
				builder.WriteString(fmt.Sprintf("  update %s where %s %s %s\n", api.Update.Source, api.Update.Where.Field, api.Update.Where.Operator, formatAPIOperandIR(api.Update.Where.Value)))
				builder.WriteString(formatAPIStatementsIR(apiUpdateStatements(*api.Update), "    "))
			} else {
				builder.WriteString(fmt.Sprintf("  update %s where %s %s %s set %s\n", api.Update.Source, api.Update.Where.Field, api.Update.Where.Operator, formatAPIOperandIR(api.Update.Where.Value), formatAPISetIR(apiStatementSets(apiUpdateStatements(*api.Update)))))
			}
		}
		if api.Respond != "" {
			builder.WriteString(fmt.Sprintf("  respond %s\n", api.Respond))
		}
	}

	for _, layout := range program.Layouts {
		builder.WriteString(fmt.Sprintf("\nlayout %s\n", layout.Name))
		if len(layout.Sidebar.Items) > 0 {
			builder.WriteString("  sidebar ")
			builder.WriteString(strings.Join(layout.Sidebar.Items, " "))
			builder.WriteString("\n")
		}
	}

	for _, page := range program.Pages {
		builder.WriteString(fmt.Sprintf("\npage %s", page.Name))
		if page.Layout != "" {
			builder.WriteString(fmt.Sprintf(" layout %s", page.Layout))
		}
		builder.WriteString(fmt.Sprintf(" source %s\n", page.Source))
		if page.Query != "" {
			builder.WriteString(fmt.Sprintf("  query %s\n", page.Query))
		}
		if page.View != nil && len(page.View.Order) > 0 {
			builder.WriteString("  view-order ")
			builder.WriteString(strings.Join(page.View.Order, " "))
			builder.WriteString("\n")
		}
		if page.View != nil && page.View.Compose != nil {
			builder.WriteString(fmt.Sprintf("  view-compose %s", page.View.Compose.Mode))
			if page.View.Compose.Columns > 0 {
				builder.WriteString(fmt.Sprintf(" columns %d", page.View.Compose.Columns))
			}
			if page.View.Compose.Gap != "" {
				builder.WriteString(fmt.Sprintf(" gap %s", page.View.Compose.Gap))
			}
			if page.View.Compose.StackAt != "" {
				builder.WriteString(fmt.Sprintf(" stackAt %s", page.View.Compose.StackAt))
			}
			builder.WriteString("\n")
		}
		if page.View != nil {
			for _, section := range page.View.Sections {
				builder.WriteString(formatViewSectionIR("  ", section))
			}
			for _, group := range page.View.Groups {
				builder.WriteString(formatViewGroupIR("  ", group))
			}
			for _, tab := range page.View.Tabs {
				builder.WriteString(fmt.Sprintf("  view-tab %s sections %s\n", tab.Name, strings.Join(tab.Sections, " ")))
			}
			for _, trigger := range page.View.Triggers {
				builder.WriteString(formatViewTriggerIR("  ", trigger))
			}
		}
		if len(page.Table.Columns) > 0 {
			builder.WriteString("  table ")
			builder.WriteString(strings.Join(page.Table.Columns, " "))
			builder.WriteString("\n")
		}
		if len(page.Table.Search) > 0 {
			builder.WriteString("  search ")
			builder.WriteString(strings.Join(page.Table.Search, " "))
			builder.WriteString("\n")
		}
		if len(page.Table.Filters) > 0 {
			builder.WriteString("  filter ")
			builder.WriteString(strings.Join(page.Table.Filters, " "))
			builder.WriteString("\n")
		}
		if page.Table.Sort.Field != "" {
			builder.WriteString(fmt.Sprintf("  sort %s %s\n", page.Table.Sort.Field, page.Table.Sort.Direction))
		}
		if page.Table.Paginate > 0 {
			builder.WriteString(fmt.Sprintf("  paginate %d\n", page.Table.Paginate))
		}
		if page.Table.Identity != nil && page.Table.Identity.ID != "" {
			builder.WriteString(fmt.Sprintf("  table-id %s\n", page.Table.Identity.ID))
		}
		if page.Table.Identity != nil && len(page.Table.Identity.Classes) > 0 {
			builder.WriteString("  table-class ")
			builder.WriteString(strings.Join(page.Table.Identity.Classes, " "))
			builder.WriteString("\n")
		}
		if len(page.Table.UI) > 0 {
			builder.WriteString("  table-ui ")
			builder.WriteString(formatUIIntentSegments(page.Table.UI))
			builder.WriteString("\n")
		}
		if len(page.Form.Fields) > 0 {
			builder.WriteString("  form ")
			builder.WriteString(strings.Join(page.Form.Fields, " "))
			builder.WriteString("\n")
		}
		if page.Form.Identity != nil && page.Form.Identity.ID != "" {
			builder.WriteString(fmt.Sprintf("  form-id %s\n", page.Form.Identity.ID))
		}
		if page.Form.Identity != nil && len(page.Form.Identity.Classes) > 0 {
			builder.WriteString("  form-class ")
			builder.WriteString(strings.Join(page.Form.Identity.Classes, " "))
			builder.WriteString("\n")
		}
		if len(page.Form.UI) > 0 {
			builder.WriteString("  form-ui ")
			builder.WriteString(formatUIIntentSegments(page.Form.UI))
			builder.WriteString("\n")
		}
		if len(page.Actions) > 0 {
			builder.WriteString("  actions ")
			builder.WriteString(strings.Join(page.Actions, " "))
			builder.WriteString("\n")
		}
		for _, actionUI := range page.ActionUI {
			if actionUI.Identity != nil && actionUI.Identity.ID != "" {
				builder.WriteString(fmt.Sprintf("  action-id %s %s\n", actionUI.Action, actionUI.Identity.ID))
			}
			if actionUI.Identity != nil && len(actionUI.Identity.Classes) > 0 {
				builder.WriteString(fmt.Sprintf("  action-class %s ", actionUI.Action))
				builder.WriteString(strings.Join(actionUI.Identity.Classes, " "))
				builder.WriteString("\n")
			}
			if len(actionUI.UI) > 0 {
				builder.WriteString(fmt.Sprintf("  action-ui %s ", actionUI.Action))
				builder.WriteString(formatUIIntentSegments(actionUI.UI))
				builder.WriteString("\n")
			}
		}
		if len(page.Access) > 0 {
			builder.WriteString("  access ")
			builder.WriteString(strings.Join(page.Access, " "))
			builder.WriteString("\n")
		}
	}

	for _, workflow := range program.Workflows {
		builder.WriteString(fmt.Sprintf("\nworkflow %s source %s\n", workflow.Name, workflow.Source))
		if len(workflow.States) > 0 {
			builder.WriteString("  states ")
			builder.WriteString(strings.Join(workflow.States, " "))
			builder.WriteString("\n")
		}
		for _, transition := range workflow.Transitions {
			builder.WriteString(fmt.Sprintf("  transition %s from %s to %s", transition.Name, transition.From, transition.To))
			if len(transition.Allow) > 0 {
				builder.WriteString(" allow ")
				builder.WriteString(strings.Join(transition.Allow, " "))
			}
			builder.WriteString("\n")
		}
	}

	for _, state := range program.States {
		builder.WriteString(fmt.Sprintf("\nstate %s\n", state.Name))
		for _, field := range state.Fields {
			fieldType := field.Type
			if field.List {
				fieldType += "[]"
			}
			builder.WriteString(fmt.Sprintf("  %s %s\n", field.Name, fieldType))
		}
		for _, modal := range state.Modals {
			builder.WriteString(fmt.Sprintf("  modal %s %s\n", modal.Name, modal.Default))
		}
	}

	for _, component := range program.Components {
		builder.WriteString(fmt.Sprintf("\ncomponent %s\n", component.Name))
		for _, input := range component.Inputs {
			inputType := input.Type
			if input.List {
				inputType += "[]"
			}
			builder.WriteString(fmt.Sprintf("  input %s %s\n", input.Name, inputType))
		}
		for _, variant := range component.Variants {
			builder.WriteString(fmt.Sprintf("  variant %s when %s\n", variant.Name, variant.Condition))
		}
	}

	return builder.String()
}

func writeFieldTextTranslationIR(builder *strings.Builder, kind string, blocks []FieldTextTranslationDecl) {
	for _, block := range blocks {
		builder.WriteString(fmt.Sprintf("\n%s %s\n", kind, block.Target))
		for _, translation := range block.Translations {
			builder.WriteString(fmt.Sprintf("  %s %s\n", translation.Locale, quoteBlackString(translation.Text)))
		}
	}
}

func formatUIIntentLine(intents []UIIntent) string {
	if len(intents) == 0 {
		return ""
	}
	return "ui " + formatUIIntentSegments(intents)
}

func formatUIIntentSegments(intents []UIIntent) string {
	segments := []string{}
	for _, intent := range intents {
		parts := []string{intent.Mode}
		parts = append(parts, intent.Values...)
		segments = append(segments, strings.Join(parts, " "))
	}
	return strings.Join(segments, " | ")
}

func FormatValidationIR(result ValidateResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("validate ok\n")
	builder.WriteString(fmt.Sprintf("app %s\n", result.Summary.App))
	builder.WriteString(fmt.Sprintf("entities %d\n", result.Summary.Entities))
	builder.WriteString(fmt.Sprintf("pages %d\n", result.Summary.Pages))
	return builder.String()
}

func FormatBuildIR(result BuildResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("build ok\n")
	builder.WriteString(fmt.Sprintf("out %s\n", result.OutDir))
	for _, file := range result.Files {
		builder.WriteString(fmt.Sprintf("file %s %s\n", file.Kind, file.Path))
	}
	return builder.String()
}

func FormatBenchmarkIR(result BenchmarkResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("benchmark ok\n")
	builder.WriteString(fmt.Sprintf("app %s\n", result.Summary.App))
	builder.WriteString(fmt.Sprintf("source files %d lines %d bytes %d\n", result.Source.Files, result.Source.Lines, result.Source.Bytes))
	builder.WriteString(fmt.Sprintf("generated files %d lines %d bytes %d\n", result.Generated.Files, result.Generated.Lines, result.Generated.Bytes))
	builder.WriteString(fmt.Sprintf("ratio generated_to_source_lines %.2f\n", result.Ratios.GeneratedToSourceLines))
	for _, kind := range result.GeneratedKinds {
		builder.WriteString(fmt.Sprintf("kind %s files %d lines %d bytes %d\n", kind.Kind, kind.Files, kind.Lines, kind.Bytes))
	}
	return builder.String()
}

func FormatAITaskBenchmarkIR(result AITaskBenchmarkResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("benchmark tasks ok\n")
	} else {
		builder.WriteString("benchmark tasks failed\n")
	}
	builder.WriteString(fmt.Sprintf("app %s\n", result.Summary.App))
	builder.WriteString(fmt.Sprintf("baseline source_lines %d generated_lines %d generated_files %d ratio %.2f\n",
		result.Baseline.SourceLines,
		result.Baseline.GeneratedLines,
		result.Baseline.GeneratedFiles,
		result.Baseline.GeneratedToSourceLines,
	))
	builder.WriteString(fmt.Sprintf("scenarios %d\n", len(result.Scenarios)))
	for _, scenario := range result.Scenarios {
		builder.WriteString(fmt.Sprintf("  scenario %s blacklang_tokens %d conventional_tokens %d savings %d\n",
			scenario.ID,
			scenario.EstimatedTokens.BlackLangTotal,
			scenario.EstimatedTokens.ConventionalTotal,
			scenario.EstimatedTokens.EstimatedSavingsPercent,
		))
	}
	builder.WriteString(fmt.Sprintf("totals blacklang_tokens %d conventional_tokens %d savings %d\n",
		result.Totals.BlackLangTotal,
		result.Totals.ConventionalTotal,
		result.Totals.EstimatedSavingsPercent,
	))
	return builder.String()
}

func FormatAIEvalCorpusIR(result AIEvalCorpusResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("benchmark eval ok\n")
	} else {
		builder.WriteString("benchmark eval failed\n")
	}
	builder.WriteString(fmt.Sprintf("app %s\n", result.Summary.App))
	builder.WriteString(fmt.Sprintf("suite %s mode %s repeat %d timeout_minutes %d\n",
		result.Suite.ID,
		result.Suite.Mode,
		result.Suite.Repeat,
		result.Suite.TimeoutMinutes,
	))
	builder.WriteString(fmt.Sprintf("baseline source_lines %d generated_lines %d generated_files %d ratio %.2f\n",
		result.Baseline.SourceLines,
		result.Baseline.GeneratedLines,
		result.Baseline.GeneratedFiles,
		result.Baseline.GeneratedToSourceLines,
	))
	builder.WriteString(fmt.Sprintf("cases %d total_runs %d max_score %d estimated_minutes %d\n",
		result.Totals.CaseCount,
		result.Totals.TotalRuns,
		result.Totals.MaxScoreAcrossRepeats,
		result.Totals.EstimatedMinutesTotal,
	))
	for _, item := range result.Cases {
		builder.WriteString(fmt.Sprintf("  case %s scenario %s score %d blacklang_tokens %d conventional_tokens %d savings %d\n",
			item.ID,
			item.ScenarioID,
			scoreMax(item.Scoring),
			item.EstimatedTokens.BlackLangTotal,
			item.EstimatedTokens.ConventionalTotal,
			item.EstimatedTokens.EstimatedSavingsPercent,
		))
	}
	return builder.String()
}

func FormatAIEvalHistoryIR(result AIEvalHistoryResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("benchmark eval-history ok\n")
	} else {
		builder.WriteString("benchmark eval-history failed\n")
	}
	builder.WriteString(fmt.Sprintf("history %s version %s status %s published_at %s\n",
		result.History.ID,
		result.History.Version,
		result.History.Status,
		result.History.PublishedAt,
	))
	builder.WriteString(fmt.Sprintf("summary runs %d models %d total_runs %d passed %d failed %d average_score %d coverage %d latest %s\n",
		result.Summary.RunCount,
		result.Summary.ModelCount,
		result.Summary.TotalRuns,
		result.Summary.PassedRuns,
		result.Summary.FailedRuns,
		result.Summary.AverageScore,
		result.Summary.CoveragePercent,
		result.Summary.LatestRun,
	))
	for _, run := range result.Runs {
		builder.WriteString(fmt.Sprintf("  run %s suite %s status %s source %s coverage %d score %d total_runs %d passed %d failed %d\n",
			run.ID,
			run.SuiteID,
			run.Status,
			run.Source,
			run.CoveragePercent,
			run.AverageScore,
			run.TotalRuns,
			run.PassedRuns,
			run.FailedRuns,
		))
		for _, model := range run.Models {
			builder.WriteString(fmt.Sprintf("    model %s provider %s source %s runs %d passed %d failed %d score %d median_tokens %d\n",
				model.Name,
				model.Provider,
				model.Source,
				model.Runs,
				model.Passed,
				model.Failed,
				model.AverageScore,
				model.MedianTokens,
			))
		}
		for _, evidence := range run.Evidence {
			builder.WriteString(fmt.Sprintf("    evidence %s\n", evidence))
		}
	}
	return builder.String()
}

func scoreMax(scoring []AIEvalScoreCriterion) int {
	total := 0
	for _, item := range scoring {
		total += item.Points
	}
	return total
}

func FormatInitIR(result InitResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("init ok\n")
	builder.WriteString(fmt.Sprintf("root %s\n", result.Root))
	for _, file := range result.Files {
		builder.WriteString(fmt.Sprintf("file %s %s\n", file.Kind, file.Path))
	}
	return builder.String()
}

func FormatInspectIR(result InspectResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("inspect ok\n")
	if result.Config.LanguageVersion != "" {
		builder.WriteString(fmt.Sprintf("language %s\n", result.Config.LanguageVersion))
	}
	if result.Config.Target != "" {
		builder.WriteString(fmt.Sprintf("target %s\n", result.Config.Target))
	}
	builder.WriteString(fmt.Sprintf("source %s\n", result.Config.Source))
	builder.WriteString(fmt.Sprintf("out %s\n", result.Config.Out))
	builder.WriteString(fmt.Sprintf("app %s\n", result.Summary.App))
	if result.Program.Database != nil && result.Program.Database.URL.Name != "" {
		builder.WriteString(fmt.Sprintf("database url env %s\n", result.Program.Database.URL.Name))
	}
	if result.Program.Ops != nil {
		builder.WriteString("ops")
		if result.Program.Ops.Health != nil {
			builder.WriteString(fmt.Sprintf(" health %s", result.Program.Ops.Health.Path))
		}
		if result.Program.Ops.Readiness != nil {
			builder.WriteString(fmt.Sprintf(" readiness %s", result.Program.Ops.Readiness.Path))
		}
		if result.Program.Ops.Metrics != nil {
			builder.WriteString(fmt.Sprintf(" metrics %s", result.Program.Ops.Metrics.Path))
		}
		if result.Program.Ops.Logging != "" {
			builder.WriteString(fmt.Sprintf(" logging %s", result.Program.Ops.Logging))
		}
		if result.Program.Ops.Observe != nil {
			builder.WriteString(fmt.Sprintf(" observe %s endpoint env %s", result.Program.Ops.Observe.Provider, result.Program.Ops.Observe.Endpoint.Name))
		}
		builder.WriteString("\n")
	}
	if result.Program.I18N != nil {
		builder.WriteString(fmt.Sprintf("i18n default %s locales %s labels %d placeholders %d helpTexts %d messages %d\n",
			result.Program.I18N.Default,
			strings.Join(result.Program.I18N.Locales, " "),
			len(result.Program.Labels),
			len(result.Program.Placeholders),
			len(result.Program.HelpTexts),
			len(result.Program.Messages),
		))
	}
	builder.WriteString(fmt.Sprintf("entities %d\n", result.Summary.Entities))
	for _, entity := range result.Program.Entities {
		builder.WriteString(fmt.Sprintf("  entity %s fields %d", entity.Name, len(entity.Fields)))
		if len(entity.ComputedFields) > 0 {
			builder.WriteString(fmt.Sprintf(" computed %d", len(entity.ComputedFields)))
		}
		if len(entity.Indexes) > 0 {
			builder.WriteString(fmt.Sprintf(" indexes %d", len(entity.Indexes)))
		}
		if len(entity.Policies) > 0 {
			builder.WriteString(fmt.Sprintf(" policies %d", len(entity.Policies)))
		}
		builder.WriteString("\n")
	}
	if len(result.Program.Migrations) > 0 {
		builder.WriteString(fmt.Sprintf("migrations %d\n", len(result.Program.Migrations)))
		for _, migration := range result.Program.Migrations {
			builder.WriteString(fmt.Sprintf("  migration %s renames %d\n", migration.Name, len(migration.Renames)))
			for _, rename := range migration.Renames {
				switch rename.Kind {
				case "entity":
					builder.WriteString(fmt.Sprintf("    rename entity %s to %s\n", rename.From, rename.To))
				case "field":
					builder.WriteString(fmt.Sprintf("    rename field %s.%s to %s\n", rename.Entity, rename.From, rename.To))
				}
			}
		}
	}
	builder.WriteString(fmt.Sprintf("pages %d\n", result.Summary.Pages))
	for _, page := range result.Program.Pages {
		builder.WriteString(fmt.Sprintf("  page %s source %s actions %s\n", page.Name, page.Source, strings.Join(page.Actions, " ")))
		if page.Query != "" {
			builder.WriteString(fmt.Sprintf("    query %s\n", page.Query))
		}
		if page.View != nil && page.View.Compose != nil {
			builder.WriteString(fmt.Sprintf("    view-compose %s", page.View.Compose.Mode))
			if page.View.Compose.Columns > 0 {
				builder.WriteString(fmt.Sprintf(" columns %d", page.View.Compose.Columns))
			}
			if page.View.Compose.Gap != "" {
				builder.WriteString(fmt.Sprintf(" gap %s", page.View.Compose.Gap))
			}
			if page.View.Compose.StackAt != "" {
				builder.WriteString(fmt.Sprintf(" stackAt %s", page.View.Compose.StackAt))
			}
			builder.WriteString("\n")
		}
		if page.View != nil {
			for _, section := range page.View.Sections {
				builder.WriteString(formatViewSectionIR("    ", section))
			}
			for _, group := range page.View.Groups {
				builder.WriteString(formatViewGroupIR("    ", group))
			}
			for _, tab := range page.View.Tabs {
				builder.WriteString(fmt.Sprintf("    view-tab %s sections %s\n", tab.Name, strings.Join(tab.Sections, " ")))
			}
			for _, trigger := range page.View.Triggers {
				builder.WriteString(formatViewTriggerIR("    ", trigger))
			}
		}
	}
	if len(result.Program.Queries) > 0 {
		builder.WriteString(fmt.Sprintf("queries %d\n", len(result.Program.Queries)))
		for _, query := range result.Program.Queries {
			builder.WriteString(fmt.Sprintf("  query %s source %s filters %d aggregates %d\n", query.Name, query.Source, len(query.Where), len(query.Aggregates)))
		}
	}
	if len(result.Program.Jobs) > 0 {
		builder.WriteString(fmt.Sprintf("jobs %d\n", len(result.Program.Jobs)))
		for _, job := range result.Program.Jobs {
			builder.WriteString(fmt.Sprintf("  job %s schedule %s run %s %s\n", job.Name, jobScheduleLabel(job), job.Run.Kind, job.Run.Query))
		}
	}
	if len(result.Program.Seeds) > 0 {
		builder.WriteString(fmt.Sprintf("seeds %d\n", len(result.Program.Seeds)))
		for _, seed := range result.Program.Seeds {
			builder.WriteString(fmt.Sprintf("  seed %s source %s rows %d\n", seed.Name, seed.Source, len(seed.Rows)))
		}
	}
	if len(result.Program.Tests) > 0 {
		builder.WriteString(fmt.Sprintf("tests %d\n", len(result.Program.Tests)))
		for _, test := range result.Program.Tests {
			builder.WriteString(fmt.Sprintf("  test %s page %s expectations %d\n", test.Name, test.Page, len(test.Expectations)))
		}
	}
	if len(result.Program.Actions) > 0 {
		builder.WriteString(fmt.Sprintf("actions %d\n", len(result.Program.Actions)))
		for _, action := range result.Program.Actions {
			transactionName := ""
			if bindings := transactionActionBindings(result.Program); bindings != nil {
				transactionName = bindings[action.Name]
			}
			builder.WriteString(fmt.Sprintf("  action %s source %s inputs %d sets %d transaction %s\n", action.Name, action.Source, len(action.Inputs), len(actionStatementSets(actionStatements(action))), transactionName))
		}
	}
	if len(result.Program.Transactions) > 0 {
		builder.WriteString(fmt.Sprintf("transactions %d\n", len(result.Program.Transactions)))
		for _, transaction := range result.Program.Transactions {
			builder.WriteString(fmt.Sprintf("  transaction %s actions %d apis %d\n", transaction.Name, len(transaction.Actions), len(transaction.APIs)))
		}
	}
	if len(result.Program.Services) > 0 {
		builder.WriteString(fmt.Sprintf("services %d\n", len(result.Program.Services)))
		for _, service := range result.Program.Services {
			builder.WriteString(fmt.Sprintf("  service %s apis %d\n", service.Name, len(service.APIs)))
		}
	}
	if len(result.Program.APIs) > 0 {
		builder.WriteString(fmt.Sprintf("apis %d\n", len(result.Program.APIs)))
		for _, api := range result.Program.APIs {
			handler := "declared"
			if api.Update != nil {
				handler = "update"
			}
			transactionName := ""
			if bindings := transactionAPIBindings(result.Program); bindings != nil {
				transactionName = bindings[api.Name]
			}
			serviceName := ""
			if bindings := serviceAPIBindings(result.Program); bindings != nil {
				serviceName = bindings[api.Name]
			}
			builder.WriteString(fmt.Sprintf("  api %s method %s path %s body %d handler %s transaction %s service %s\n", api.Name, strings.ToUpper(api.Method), api.Path, len(api.Body), handler, transactionName, serviceName))
		}
	}
	if len(result.Program.Workflows) > 0 {
		builder.WriteString(fmt.Sprintf("workflows %d\n", len(result.Program.Workflows)))
		for _, workflow := range result.Program.Workflows {
			builder.WriteString(fmt.Sprintf("  workflow %s source %s states %s\n", workflow.Name, workflow.Source, strings.Join(workflow.States, " ")))
		}
	}
	if len(result.Program.States) > 0 {
		builder.WriteString(fmt.Sprintf("states %d\n", len(result.Program.States)))
		for _, state := range result.Program.States {
			builder.WriteString(fmt.Sprintf("  state %s fields %d modals %d\n", state.Name, len(state.Fields), len(state.Modals)))
		}
	}
	if len(result.Program.Components) > 0 {
		builder.WriteString(fmt.Sprintf("components %d\n", len(result.Program.Components)))
		for _, component := range result.Program.Components {
			builder.WriteString(fmt.Sprintf("  component %s inputs %d variants %d\n", component.Name, len(component.Inputs), len(component.Variants)))
		}
	}
	return builder.String()
}

func FormatAffectedIR(result InspectAffectedResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("inspect affected ok\n")
	builder.WriteString(fmt.Sprintf("symbol %s\n", result.Affected.Symbol))
	builder.WriteString(fmt.Sprintf("kind %s\n", result.Affected.Kind))
	if result.Affected.Entity != "" {
		builder.WriteString(fmt.Sprintf("entity %s\n", result.Affected.Entity))
	}
	if result.Affected.Field != "" {
		builder.WriteString(fmt.Sprintf("field %s\n", result.Affected.Field))
	}
	writeAffectedIRItems(&builder, "entities", result.Affected.Entities)
	writeAffectedIRItems(&builder, "migrations", result.Affected.Migrations)
	writeAffectedIRItems(&builder, "pages", result.Affected.Pages)
	writeAffectedIRItems(&builder, "queries", result.Affected.Queries)
	writeAffectedIRItems(&builder, "jobs", result.Affected.Jobs)
	writeAffectedIRItems(&builder, "actions", result.Affected.Actions)
	writeAffectedIRItems(&builder, "transactions", result.Affected.Transactions)
	writeAffectedIRItems(&builder, "services", result.Affected.Services)
	writeAffectedIRItems(&builder, "seeds", result.Affected.Seeds)
	writeAffectedIRItems(&builder, "tests", result.Affected.Tests)
	writeAffectedIRItems(&builder, "roles", result.Affected.Roles)
	writeAffectedIRItems(&builder, "workflows", result.Affected.Workflows)
	writeAffectedIRItems(&builder, "states", result.Affected.States)
	writeAffectedIRItems(&builder, "components", result.Affected.Components)
	writeAffectedIRItems(&builder, "apis", result.Affected.APIs)
	writeAffectedIRItems(&builder, "generated-files", result.Affected.GeneratedFiles)
	return builder.String()
}

func formatViewSectionIR(indent string, section ViewSectionDecl) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%sview-section %s", indent, section.Name))
	if section.Component != "" {
		builder.WriteString(fmt.Sprintf(" component %s", section.Component))
	}
	if section.Bind != "" {
		builder.WriteString(fmt.Sprintf(" bind %s", section.Bind))
	}
	if section.Span > 0 {
		builder.WriteString(fmt.Sprintf(" span %d", section.Span))
	}
	if section.Display != "" {
		builder.WriteString(fmt.Sprintf(" display %s", section.Display))
	}
	if section.Side != "" {
		builder.WriteString(fmt.Sprintf(" side %s", section.Side))
	}
	if section.Title != "" {
		builder.WriteString(fmt.Sprintf(" title %q", section.Title))
	}
	builder.WriteString("\n")
	return builder.String()
}

func formatViewGroupIR(indent string, group ViewGroupDecl) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%sview-group %s sections %s", indent, group.Name, strings.Join(group.Sections, " ")))
	if group.Compose != nil {
		builder.WriteString(fmt.Sprintf(" compose %s", group.Compose.Mode))
		if group.Compose.Columns > 0 {
			builder.WriteString(fmt.Sprintf(" columns %d", group.Compose.Columns))
		}
		if group.Compose.Gap != "" {
			builder.WriteString(fmt.Sprintf(" gap %s", group.Compose.Gap))
		}
	}
	if group.Span > 0 {
		builder.WriteString(fmt.Sprintf(" span %d", group.Span))
	}
	if group.Title != "" {
		builder.WriteString(fmt.Sprintf(" title %q", group.Title))
	}
	builder.WriteString("\n")
	return builder.String()
}

func formatViewTriggerIR(indent string, trigger ViewTriggerDecl) string {
	return fmt.Sprintf("%sview-trigger %s on %s\n", indent, trigger.Section, trigger.Event)
}

func writeAffectedIRItems(builder *strings.Builder, label string, items []AffectedItem) {
	if len(items) == 0 {
		return
	}
	builder.WriteString(fmt.Sprintf("%s %d\n", label, len(items)))
	for _, item := range items {
		builder.WriteString(fmt.Sprintf("  %s reason %q\n", item.Name, item.Reason))
	}
}

func formatActionOperand(operand ActionOperandDecl) string {
	if operand.Kind == "string" {
		return quoteBlackString(operand.Value)
	}
	return operand.Value
}

func formatAPIOperandIR(operand APIOperandDecl) string {
	switch operand.Kind {
	case "string":
		return quoteBlackString(operand.Value)
	case "body", "param":
		return operand.Kind + "." + operand.Value
	default:
		return operand.Value
	}
}

func formatSeedLiteral(value SeedLiteral) string {
	if value.Kind == "string" {
		return quoteBlackString(value.Value)
	}
	if value.Kind == "ref" {
		return "ref " + value.Value
	}
	return value.Value
}

func formatAPISetIR(sets []APISetDecl) string {
	parts := []string{}
	for _, set := range sets {
		value := fmt.Sprintf("%s = %s", set.Field, formatAPIExpressionIR(set.Expr))
		parts = append(parts, value)
	}
	return strings.Join(parts, ", ")
}

func formatActionStatementsIR(statements []ActionStatementDecl, indent string) string {
	var builder strings.Builder
	for _, statement := range statements {
		builder.WriteString(formatActionStatementIR(statement, indent))
	}
	return builder.String()
}

func formatActionStatementIR(statement ActionStatementDecl, indent string) string {
	var builder strings.Builder
	switch statement.Kind {
	case "value":
		if statement.Value != nil {
			builder.WriteString(fmt.Sprintf("%svalue %s = %s\n", indent, statement.Value.Name, formatActionExpressionIR(statement.Value.Expression)))
		}
	case "set":
		if statement.Set != nil {
			builder.WriteString(fmt.Sprintf("%sset %s = %s\n", indent, statement.Set.Field, formatActionExpressionIR(statement.Set.Expression)))
		}
	case "if":
		if statement.If != nil {
			builder.WriteString(fmt.Sprintf("%sif %s\n", indent, formatActionConditionIR(statement.If.Condition)))
			builder.WriteString(formatActionStatementsIR(statement.If.Then, indent+"  "))
			if len(statement.If.Else) > 0 {
				builder.WriteString(fmt.Sprintf("%selse\n", indent))
				builder.WriteString(formatActionStatementsIR(statement.If.Else, indent+"  "))
			}
		}
	}
	return builder.String()
}

func formatActionConditionIR(condition ActionConditionDecl) string {
	if condition.Tree != nil {
		return formatConditionExpressionIR(*condition.Tree, formatActionExpressionNodeIR, 0)
	}
	return fmt.Sprintf("%s %s %s", formatActionExpressionIR(condition.Left), condition.Operator, formatActionExpressionIR(condition.Right))
}

func formatAPIStatementsIR(statements []APIStatementDecl, indent string) string {
	var builder strings.Builder
	for _, statement := range statements {
		builder.WriteString(formatAPIStatementIR(statement, indent))
	}
	return builder.String()
}

func formatAPIStatementIR(statement APIStatementDecl, indent string) string {
	var builder strings.Builder
	switch statement.Kind {
	case "value":
		if statement.Value != nil {
			builder.WriteString(fmt.Sprintf("%svalue %s = %s\n", indent, statement.Value.Name, formatAPIExpressionIR(statement.Value.Expression)))
		}
	case "set":
		if statement.Set != nil {
			builder.WriteString(fmt.Sprintf("%sset %s = %s\n", indent, statement.Set.Field, formatAPIExpressionIR(statement.Set.Expr)))
		}
	case "if":
		if statement.If != nil {
			builder.WriteString(fmt.Sprintf("%sif %s\n", indent, formatAPIConditionIR(statement.If.Condition)))
			builder.WriteString(formatAPIStatementsIR(statement.If.Then, indent+"  "))
			if len(statement.If.Else) > 0 {
				builder.WriteString(fmt.Sprintf("%selse\n", indent))
				builder.WriteString(formatAPIStatementsIR(statement.If.Else, indent+"  "))
			}
		}
	}
	return builder.String()
}

func formatAPIConditionIR(condition APIConditionDecl) string {
	if condition.Tree != nil {
		return formatConditionExpressionIR(*condition.Tree, formatAPIExpressionNodeIR, 0)
	}
	return fmt.Sprintf("%s %s %s", formatAPIExpressionIR(condition.Left), condition.Operator, formatAPIExpressionIR(condition.Right))
}

func apiUpdateUsesBlock(update APIUpdateDecl) bool {
	for _, statement := range apiUpdateStatements(update) {
		if statement.Kind != "set" {
			return true
		}
	}
	return false
}

func formatActionExpressionIR(expression ActionExpressionDecl) string {
	if expression.Tree != nil {
		return formatCoreExpression(*expression.Tree)
	}
	value := formatActionOperand(expression.Left)
	if expression.Operator != "" && expression.Right != nil {
		value += fmt.Sprintf(" %s %s", expression.Operator, formatActionOperand(*expression.Right))
	}
	return value
}

func formatAPIExpressionIR(expression APIExpressionDecl) string {
	if expression.Tree != nil {
		return formatAPIExpressionTreeIR(*expression.Tree)
	}
	value := formatAPIOperandIR(expression.Left)
	if expression.Operator != "" && expression.Right != nil {
		value += fmt.Sprintf(" %s %s", expression.Operator, formatAPIOperandIR(*expression.Right))
	}
	return value
}

func formatAPIExpressionTreeIR(expression ExpressionDecl) string {
	return formatAPIExpressionTreeIRWithParent(expression, 0, "", false)
}

func formatAPIExpressionTreeIRWithParent(expression ExpressionDecl, parentPrecedence int, parentOperator string, rightChild bool) string {
	if expression.Kind != "binary" {
		return formatAPIExpressionOperandIR(expression)
	}
	precedence := expressionOperatorPrecedence(expression.Operator)
	left := ""
	right := ""
	if expression.Left != nil {
		left = formatAPIExpressionTreeIRWithParent(*expression.Left, precedence, expression.Operator, false)
	}
	if expression.Right != nil {
		right = formatAPIExpressionTreeIRWithParent(*expression.Right, precedence, expression.Operator, true)
	}
	rendered := fmt.Sprintf("%s %s %s", left, expression.Operator, right)
	if precedence < parentPrecedence || (rightChild && precedence == parentPrecedence && (parentOperator == "-" || parentOperator == "/")) {
		return "(" + rendered + ")"
	}
	return rendered
}

func formatAPIExpressionOperandIR(expression ExpressionDecl) string {
	switch expression.ValueKind {
	case "string":
		return quoteBlackString(expression.Value)
	case "body", "param":
		return expression.ValueKind + "." + expression.Value
	default:
		return expression.Value
	}
}

func formatActionExpressionNodeIR(expression ExpressionDecl) string {
	return formatActionExpressionIR(actionExpressionFromCore(expression))
}

func formatAPIExpressionNodeIR(expression ExpressionDecl) string {
	return formatAPIExpressionIR(apiExpressionFromCore(expression))
}

func formatConditionExpressionIR(condition ConditionExpressionDecl, expressionFormatter func(ExpressionDecl) string, parentPrecedence int) string {
	switch condition.Kind {
	case "comparison":
		if condition.Comparison == nil {
			return ""
		}
		return fmt.Sprintf("%s %s %s", expressionFormatter(condition.Comparison.Left), condition.Comparison.Operator, expressionFormatter(condition.Comparison.Right))
	case "not":
		if condition.Left == nil {
			return "not"
		}
		rendered := "not " + formatConditionExpressionIR(*condition.Left, expressionFormatter, 3)
		if parentPrecedence > 3 {
			return "(" + rendered + ")"
		}
		return rendered
	case "and", "or":
		if condition.Left == nil || condition.Right == nil {
			return ""
		}
		precedence := conditionOperatorPrecedence(condition.Kind)
		left := formatConditionExpressionIR(*condition.Left, expressionFormatter, precedence)
		right := formatConditionExpressionIR(*condition.Right, expressionFormatter, precedence+1)
		rendered := left + " " + condition.Kind + " " + right
		if precedence < parentPrecedence {
			return "(" + rendered + ")"
		}
		return rendered
	default:
		return ""
	}
}

func conditionOperatorPrecedence(operator string) int {
	switch operator {
	case "or":
		return 1
	case "and":
		return 2
	case "not":
		return 3
	default:
		return 4
	}
}

func formatModifiers(modifiers []Modifier) string {
	parts := []string{}
	for _, modifier := range modifiers {
		if modifier.Value == "" {
			parts = append(parts, modifier.Name)
			continue
		}
		value := modifier.Value
		if modifierTakesList(modifier.Name) {
			value = strings.ReplaceAll(value, ",", " ")
		}
		if modifier.Name == "label" || modifier.Name == "placeholder" || modifier.Name == "help" || modifier.Name == "message" {
			value = quoteBlackString(value)
		}
		parts = append(parts, modifier.Name+" "+value)
	}
	return strings.Join(parts, " ")
}

func formatModifierIR(modifier Modifier) string {
	if modifier.Value == "" {
		return modifier.Name
	}
	value := modifier.Value
	if modifierTakesList(modifier.Name) {
		value = strings.ReplaceAll(value, ",", " ")
	}
	return modifier.Name + " " + value
}

func FormatDocsIR(result DocsResult) string {
	doc := result.Doc
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("docs ok\n")
	builder.WriteString(fmt.Sprintf("keyword %s\n", doc.Keyword))
	builder.WriteString(fmt.Sprintf("purpose %q\n", doc.Purpose))
	builder.WriteString(fmt.Sprintf("syntax %q\n", doc.Syntax))
	builder.WriteString("example\n")
	for _, line := range strings.Split(doc.Example, "\n") {
		if line == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("  %s\n", line))
	}
	if len(doc.AgentNotes) > 0 {
		builder.WriteString("agentNotes\n")
		for _, note := range doc.AgentNotes {
			builder.WriteString(fmt.Sprintf("  - %s\n", note))
		}
	}
	if len(doc.Errors) > 0 {
		builder.WriteString("errors ")
		builder.WriteString(strings.Join(doc.Errors, " "))
		builder.WriteString("\n")
	}
	return builder.String()
}

func FormatDocsAllIR(result DocsAllResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("docs all ok\n")
	builder.WriteString(fmt.Sprintf("count %d\n", result.Count))
	for _, doc := range result.Docs {
		builder.WriteString(fmt.Sprintf("keyword %s\n", doc.Keyword))
		builder.WriteString(fmt.Sprintf("  purpose %q\n", doc.Purpose))
		builder.WriteString(fmt.Sprintf("  syntax %q\n", doc.Syntax))
		if len(doc.Errors) > 0 {
			builder.WriteString("  errors ")
			builder.WriteString(strings.Join(doc.Errors, " "))
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func FormatExplainIR(result ExplainResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("explain ok\n")
	builder.WriteString(fmt.Sprintf("keyword %s\n", result.Keyword))
	builder.WriteString(fmt.Sprintf("purpose %q\n", result.Purpose))
	builder.WriteString(fmt.Sprintf("syntax %q\n", result.Syntax))
	if len(result.AgentSteps) > 0 {
		builder.WriteString("agentSteps\n")
		for _, step := range result.AgentSteps {
			builder.WriteString(fmt.Sprintf("  - %s\n", step))
		}
	}
	if len(result.Related) > 0 {
		builder.WriteString("related ")
		builder.WriteString(strings.Join(result.Related, " "))
		builder.WriteString("\n")
	}
	if len(result.ErrorCodes) > 0 {
		builder.WriteString("errors ")
		builder.WriteString(strings.Join(result.ErrorCodes, " "))
		builder.WriteString("\n")
	}
	return builder.String()
}

func FormatAgentStartupIR(result AgentStartupResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("agent startup ok\n")
	if result.Config.LanguageVersion != "" {
		builder.WriteString(fmt.Sprintf("language %s\n", result.Config.LanguageVersion))
	}
	if result.Config.Target != "" {
		builder.WriteString(fmt.Sprintf("target %s\n", result.Config.Target))
	}
	builder.WriteString(fmt.Sprintf("source %s\n", result.Config.Source))
	builder.WriteString(fmt.Sprintf("out %s\n", result.Config.Out))
	if result.Config.Theme != "" {
		builder.WriteString(fmt.Sprintf("theme %s\n", result.Config.Theme))
	}
	if result.Summary.App != "" {
		builder.WriteString(fmt.Sprintf("app %s\n", result.Summary.App))
	}
	if len(result.ReadFirst) > 0 {
		builder.WriteString("readFirst\n")
		for _, file := range result.ReadFirst {
			exists := "missing"
			if file.Exists {
				exists = "exists"
			}
			builder.WriteString(fmt.Sprintf("  %s %s purpose %q\n", exists, file.Path, file.Purpose))
		}
	}
	if len(result.Checklist) > 0 {
		builder.WriteString("checklist\n")
		for _, item := range result.Checklist {
			builder.WriteString(fmt.Sprintf("  %d action %q reason %q\n", item.Step, item.Action, item.Reason))
		}
	}
	if len(result.Commands) > 0 {
		builder.WriteString("commands\n")
		for _, command := range result.Commands {
			builder.WriteString(fmt.Sprintf("  %s %q\n", command.Name, command.Command))
		}
	}
	return builder.String()
}

func FormatThemeIR(result ThemeInspectResult) string {
	theme := result.Theme
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("theme inspect ok\n")
	builder.WriteString(fmt.Sprintf("file %s\n", result.File))
	builder.WriteString(fmt.Sprintf("theme %s version %d target %s locked %t\n", theme.Name, theme.Version, theme.Target, theme.Locked))
	if len(theme.Tokens) > 0 {
		builder.WriteString("tokens\n")
		for _, token := range theme.Tokens {
			builder.WriteString(fmt.Sprintf("  %s %s %q\n", token.Kind, token.Name, token.Value))
		}
	}
	if theme.Profile.Name != "" {
		builder.WriteString(fmt.Sprintf("profile %s version %d\n", theme.Profile.Name, theme.Profile.Version))
		builder.WriteString(fmt.Sprintf("  rules syntax %q order %s separator %q missing %s extra %s duplicate %s baseline %s locked %s new %s\n",
			theme.Profile.Rules.InlineSyntax,
			theme.Profile.Rules.SlotOrder,
			theme.Profile.Rules.ModeSeparator,
			theme.Profile.Rules.MissingTrailingSlots,
			theme.Profile.Rules.ExtraValues,
			theme.Profile.Rules.DuplicateSlots,
			theme.Profile.Rules.LockBaseline,
			theme.Profile.Rules.ExistingSlotsAfterLock,
			theme.Profile.Rules.NewSlotsAfterLock,
		))
		if len(theme.Profile.ModeGroups) > 0 {
			builder.WriteString("  groups\n")
			for _, group := range theme.Profile.ModeGroups {
				builder.WriteString(fmt.Sprintf("    %s required %t slots %s applies %s\n", group.Name, group.Required, strings.Join(group.DefaultSlots, " "), strings.Join(group.AppliesTo, " ")))
			}
		}
		for _, baseline := range theme.Profile.Baselines {
			builder.WriteString(fmt.Sprintf("  baseline %s slots %s\n", baseline.Name, strings.Join(baseline.Slots, " ")))
		}
		for _, mode := range theme.Profile.Modes {
			builder.WriteString(fmt.Sprintf("  mode %s standard %t slots %s\n", mode.Name, mode.Standard, strings.Join(mode.Slots, " ")))
		}
	}
	return builder.String()
}

func FormatThemeMigrationIR(result ThemeMigrationResult) string {
	var builder strings.Builder
	status := "failed"
	if result.Success {
		status = "ok"
	}
	builder.WriteString("blackir 0.1\n")
	builder.WriteString(fmt.Sprintf("theme migrate %s\n", status))
	builder.WriteString(fmt.Sprintf("old %s\n", result.OldFile))
	builder.WriteString(fmt.Sprintf("new %s\n", result.NewFile))
	builder.WriteString(fmt.Sprintf("safe %t\n", result.Safe))
	builder.WriteString(fmt.Sprintf("theme %s version %d -> %s version %d target %s\n",
		result.Summary.OldTheme,
		result.Summary.OldVersion,
		result.Summary.NewTheme,
		result.Summary.NewVersion,
		result.Summary.Target,
	))
	builder.WriteString(fmt.Sprintf("profile %s version %d locked %t -> %s version %d locked %t\n",
		result.Summary.OldProfile,
		result.Summary.OldProfileVersion,
		result.Summary.OldLocked,
		result.Summary.NewProfile,
		result.Summary.NewProfileVersion,
		result.Summary.NewLocked,
	))
	if len(result.Changes) > 0 {
		builder.WriteString("changes\n")
		for _, change := range result.Changes {
			switch change.Type {
			case "slot-appended":
				builder.WriteString(fmt.Sprintf("  slot-appended profile %s mode %s slot %s index %d\n", change.Profile, change.Mode, change.Slot, change.Index))
			case "mode-added":
				builder.WriteString(fmt.Sprintf("  mode-added profile %s mode %s\n", change.Profile, change.Mode))
			default:
				builder.WriteString(fmt.Sprintf("  %s %s -> %s\n", change.Type, change.OldValue, change.NewValue))
			}
		}
	}
	if len(result.Errors) > 0 {
		builder.WriteString("errors\n")
		for _, diagnostic := range result.Errors {
			builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", diagnostic.Code, diagnostic.File, diagnostic.Line, diagnostic.Column))
			builder.WriteString(fmt.Sprintf("    message %q\n", diagnostic.Message))
			if diagnostic.Suggestion != "" {
				builder.WriteString(fmt.Sprintf("    suggestion %q\n", diagnostic.Suggestion))
			}
		}
	}
	return builder.String()
}

func printDiagnosticsIR(command string, diagnostics []Diagnostic) {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString(fmt.Sprintf("%s failed\n", command))
	for _, diagnostic := range diagnostics {
		builder.WriteString(fmt.Sprintf("error %s %s:%d:%d\n", diagnostic.Code, diagnostic.File, diagnostic.Line, diagnostic.Column))
		builder.WriteString(fmt.Sprintf("  message %q\n", diagnostic.Message))
		if diagnostic.Suggestion != "" {
			builder.WriteString(fmt.Sprintf("  suggestion %q\n", diagnostic.Suggestion))
		}
	}
	fmt.Print(builder.String())
}
