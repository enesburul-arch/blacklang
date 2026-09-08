package main

import "strings"

var supportedAppUILabelTargets = setOf(
	"title",
	"menu",
	"language",
	"logout",
	"closeNavigation",
	"primaryNavigation",
	"checkingSession",
	"noPages",
	"source",
	"recordForm",
)

var supportedActionUILabelTargets = setOf(
	"view",
	"create",
	"edit",
	"delete",
	"deleteSelected",
	"archive",
	"restore",
	"cancel",
	"close",
	"saveChanges",
	"saving",
	"run",
	"running",
)

var supportedTableUILabelTargets = setOf(
	"search",
	"showArchived",
	"columns",
	"filter",
	"status",
	"actions",
	"selectVisibleRecords",
	"selectRecord",
	"empty",
	"previous",
	"next",
	"page",
	"of",
)

var supportedStatusUILabelTargets = setOf(
	"active",
	"archived",
	"loadingRecords",
	"loadingDetails",
	"selectRecord",
)

func isUILabelTarget(program Program, entityIndex map[string]EntityDecl, target string) (string, bool) {
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		return "", false
	}
	for _, part := range parts {
		if part == "" || !isThemeIdentifier(part) {
			return "", false
		}
	}

	switch parts[0] {
	case "app":
		if len(parts) == 2 && supportedAppUILabelTargets[parts[1]] {
			return target, true
		}
	case "page":
		if len(parts) == 2 && programHasUILabelPage(program, parts[1]) {
			return target, true
		}
	case "action":
		if len(parts) == 2 {
			if supportedActionUILabelTargets[parts[1]] {
				return target, true
			}
			if _, ok := findCustomAction(program, parts[1]); ok {
				return target, true
			}
			if programHasWorkflowTransition(program, parts[1]) {
				return target, true
			}
		}
		if len(parts) == 3 && (parts[1] == "new" || parts[1] == "create" || parts[1] == "edit" || parts[1] == "view") {
			if _, ok := entityIndex[parts[2]]; ok {
				return target, true
			}
		}
	case "table":
		if len(parts) == 2 && supportedTableUILabelTargets[parts[1]] {
			return target, true
		}
	case "status":
		if len(parts) == 2 && supportedStatusUILabelTargets[parts[1]] {
			return target, true
		}
	}

	return "", false
}

func programHasUILabelPage(program Program, name string) bool {
	for _, page := range program.Pages {
		if page.Name == name {
			return true
		}
	}
	return program.Auth != nil && len(program.Roles) > 0 && (name == "Users" || name == "Audit")
}

func programHasWorkflowTransition(program Program, name string) bool {
	for _, workflow := range program.Workflows {
		for _, transition := range workflow.Transitions {
			if transition.Name == name {
				return true
			}
		}
	}
	return false
}

func programEntityIndex(program Program) map[string]EntityDecl {
	entities := map[string]EntityDecl{}
	for _, entity := range program.Entities {
		entities[entity.Name] = entity
	}
	return entities
}
