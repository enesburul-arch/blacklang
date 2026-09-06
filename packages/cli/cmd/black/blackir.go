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
					builder.WriteString(modifier.Name)
					if modifier.Value != "" {
						builder.WriteString(" ")
						builder.WriteString(modifier.Value)
					}
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

	for _, entity := range program.Entities {
		builder.WriteString(fmt.Sprintf("\nentity %s\n", entity.Name))
		for _, field := range entity.Fields {
			builder.WriteString(fmt.Sprintf("  %s %s", field.Name, field.Type))
			for _, modifier := range field.Modifiers {
				builder.WriteString(" ")
				builder.WriteString(modifier.Name)
				if modifier.Value != "" {
					builder.WriteString(" ")
					builder.WriteString(modifier.Value)
				}
			}
			if len(field.UI) > 0 {
				builder.WriteString(" ")
				builder.WriteString(formatUIIntentLine(field.UI))
			}
			builder.WriteString("\n")
		}
		for _, field := range entity.ComputedFields {
			builder.WriteString(fmt.Sprintf("  computed %s %s = %s %s %s", field.Name, field.Type, field.Expression.Left, field.Expression.Operator, field.Expression.Right))
			for _, modifier := range field.Modifiers {
				builder.WriteString(" ")
				builder.WriteString(modifier.Name)
				if modifier.Value != "" {
					builder.WriteString(" ")
					builder.WriteString(modifier.Value)
				}
			}
			builder.WriteString("\n")
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

	for _, query := range program.Queries {
		builder.WriteString(fmt.Sprintf("\nquery %s source %s\n", query.Name, query.Source))
		for _, filter := range query.Where {
			value := filter.Value.Value
			if filter.Value.Kind == "string" {
				value = quoteBlackString(value)
			}
			builder.WriteString(fmt.Sprintf("  where %s %s %s\n", filter.Field, filter.Operator, value))
		}
		if query.Sort.Field != "" {
			builder.WriteString(fmt.Sprintf("  sort %s %s\n", query.Sort.Field, query.Sort.Direction))
		}
		if query.Limit > 0 {
			builder.WriteString(fmt.Sprintf("  limit %d\n", query.Limit))
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
	builder.WriteString(fmt.Sprintf("entities %d\n", result.Summary.Entities))
	for _, entity := range result.Program.Entities {
		builder.WriteString(fmt.Sprintf("  entity %s fields %d", entity.Name, len(entity.Fields)))
		if len(entity.ComputedFields) > 0 {
			builder.WriteString(fmt.Sprintf(" computed %d", len(entity.ComputedFields)))
		}
		builder.WriteString("\n")
	}
	builder.WriteString(fmt.Sprintf("pages %d\n", result.Summary.Pages))
	for _, page := range result.Program.Pages {
		builder.WriteString(fmt.Sprintf("  page %s source %s actions %s\n", page.Name, page.Source, strings.Join(page.Actions, " ")))
		if page.Query != "" {
			builder.WriteString(fmt.Sprintf("    query %s\n", page.Query))
		}
	}
	if len(result.Program.Queries) > 0 {
		builder.WriteString(fmt.Sprintf("queries %d\n", len(result.Program.Queries)))
		for _, query := range result.Program.Queries {
			builder.WriteString(fmt.Sprintf("  query %s source %s filters %d\n", query.Name, query.Source, len(query.Where)))
		}
	}
	if len(result.Program.APIs) > 0 {
		builder.WriteString(fmt.Sprintf("apis %d\n", len(result.Program.APIs)))
		for _, api := range result.Program.APIs {
			builder.WriteString(fmt.Sprintf("  api %s method %s path %s\n", api.Name, strings.ToUpper(api.Method), api.Path))
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
	writeAffectedIRItems(&builder, "pages", result.Affected.Pages)
	writeAffectedIRItems(&builder, "queries", result.Affected.Queries)
	writeAffectedIRItems(&builder, "roles", result.Affected.Roles)
	writeAffectedIRItems(&builder, "workflows", result.Affected.Workflows)
	writeAffectedIRItems(&builder, "states", result.Affected.States)
	writeAffectedIRItems(&builder, "components", result.Affected.Components)
	writeAffectedIRItems(&builder, "apis", result.Affected.APIs)
	writeAffectedIRItems(&builder, "generated-files", result.Affected.GeneratedFiles)
	return builder.String()
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
