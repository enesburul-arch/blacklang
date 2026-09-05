package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func InspectTheme(args []string) ThemeInspectResult {
	config := LoadConfig(".")
	file := firstNonOptionArg(args)
	if file == "" {
		file = config.Theme
	}
	if file == "" {
		file = "theme.blackthm"
	}

	source, err := os.ReadFile(file)
	if err != nil {
		return ThemeInspectResult{
			Success: false,
			Command: "theme inspect",
			Version: version,
			File:    file,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "FILE_READ_ERROR",
				Message:    err.Error(),
				Suggestion: "Pass a readable .blackthm file path or set theme in blacklang.toml.",
			}},
		}
	}

	theme, diagnostics := ParseTheme(file, string(source))
	return ThemeInspectResult{
		Success: len(diagnostics) == 0,
		Command: "theme inspect",
		Version: version,
		File:    file,
		Theme:   theme,
		Errors:  diagnostics,
	}
}

func ParseTheme(file string, source string) (ThemeDecl, []Diagnostic) {
	statements, lexDiagnostics := tokenizeSource(file, source)
	diagnostics := append([]Diagnostic{}, lexDiagnostics...)
	theme := ThemeDecl{
		Version: 0,
		Target:  "",
		Locked:  false,
		Tokens:  []ThemeToken{},
		Profile: UIProfileDecl{
			Version:    0,
			Rules:      defaultUIProfileRules(),
			ModeGroups: standardUIModeGroups(),
			Baselines:  []UIModeDecl{},
			Modes:      []UIModeDecl{},
		},
	}
	section := ""
	tokenKeys := map[string]Position{}
	baselineNames := map[string]Position{}
	modeNames := map[string]Position{}

	addDiagnostic := func(pos Position, code string, message string, suggestion string) {
		diagnostics = append(diagnostics, Diagnostic{
			File:       file,
			Line:       pos.Line,
			Column:     pos.Column,
			Code:       code,
			Message:    message,
			Suggestion: suggestion,
		})
	}

	for _, statement := range statements {
		parts := statement.Parts()
		if len(parts) == 0 {
			continue
		}
		if len(parts) == 1 && parts[0] == "}" {
			if section == "profile" {
				section = ""
			}
			continue
		}
		hasOpenBrace := hasOpeningBrace(parts)
		parts = withoutOpeningBrace(parts)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "blackthm":
			if len(parts) != 2 || !hasOpenBrace || !isThemeIdentifier(parts[1]) {
				addDiagnostic(statement.Position, "INVALID_THEME_DECLARATION", "Theme file must start with `blackthm <Name> {`.", "Use `blackthm WarehouseTheme {`.")
				continue
			}
			theme.Name = parts[1]
			theme.Position = statement.Position
		case "version":
			if len(parts) != 2 {
				addDiagnostic(statement.Position, "INVALID_THEME_VERSION", "Theme version must use `version <positiveNumber>`.", "Use `version 1`.")
				continue
			}
			value, ok := parsePositiveThemeInt(parts[1])
			if !ok {
				addDiagnostic(statement.Position, "INVALID_THEME_VERSION", "Theme version must be a positive number.", "Use `version 1`.")
				continue
			}
			if section == "profile" {
				theme.Profile.Version = value
			} else {
				theme.Version = value
			}
		case "target":
			if len(parts) != 2 || parts[1] != "web" {
				addDiagnostic(statement.Position, "INVALID_THEME_TARGET", "Theme target must be `web` in draft v0.2.", "Use `target web`.")
				continue
			}
			theme.Target = parts[1]
		case "locked":
			if len(parts) != 2 || (parts[1] != "true" && parts[1] != "false") {
				addDiagnostic(statement.Position, "INVALID_THEME_LOCKED", "Theme lock state must be `locked true` or `locked false`.", "Use `locked false` before the profile is frozen.")
				continue
			}
			theme.Locked = parts[1] == "true"
		case "token":
			if len(parts) != 4 || !isThemeIdentifier(parts[1]) || !isThemeIdentifier(parts[2]) || strings.TrimSpace(parts[3]) == "" {
				addDiagnostic(statement.Position, "INVALID_THEME_TOKEN", "Theme tokens must use `token <kind> <name> <value>`.", "Use `token color primary \"#2563eb\"`.")
				continue
			}
			key := parts[1] + "." + parts[2]
			if previous, ok := tokenKeys[key]; ok {
				addDiagnostic(statement.Position, "DUPLICATE_THEME_TOKEN", fmt.Sprintf("Theme token %s is already declared.", key), fmt.Sprintf("Keep one declaration for %s. First declaration is at line %d.", key, previous.Line))
				continue
			}
			tokenKeys[key] = statement.Position
			theme.Tokens = append(theme.Tokens, ThemeToken{
				Kind:     parts[1],
				Name:     parts[2],
				Value:    parts[3],
				Position: statement.Position,
			})
		case "profile":
			if len(parts) != 2 || !hasOpenBrace || !isThemeIdentifier(parts[1]) {
				addDiagnostic(statement.Position, "INVALID_UI_PROFILE", "UI profile must use `profile <Name> {`.", "Use `profile UICompact {`.")
				continue
			}
			theme.Profile.Name = parts[1]
			theme.Profile.Position = statement.Position
			section = "profile"
		case "baseline":
			if section != "profile" || len(parts) < 3 || !isThemeIdentifier(parts[1]) {
				addDiagnostic(statement.Position, "INVALID_UI_BASELINE", "UI baselines must be declared inside a profile with `baseline <mode> <slot...>`.", "Use `baseline box color width style pt pr pb pl radius place` inside `profile UICompact`.")
				continue
			}
			slots := parts[2:]
			validSlots, duplicateSlot := validateUISlotList(slots)
			if !validSlots {
				addDiagnostic(statement.Position, "INVALID_UI_BASELINE", "UI baseline slots must be simple identifiers.", "Use slot names such as color, width, style, pt, pr, pb, pl, radius, or place.")
				continue
			}
			if duplicateSlot != "" {
				addDiagnostic(statement.Position, "DUPLICATE_UI_SLOT", fmt.Sprintf("UI baseline %s repeats slot %s.", parts[1], duplicateSlot), "Keep each slot once per baseline so positional values map to one property.")
				continue
			}
			if previous, ok := baselineNames[parts[1]]; ok {
				addDiagnostic(statement.Position, "DUPLICATE_UI_BASELINE", fmt.Sprintf("UI baseline %s is already declared.", parts[1]), fmt.Sprintf("Keep one baseline declaration. First declaration is at line %d.", previous.Line))
				continue
			}
			baselineNames[parts[1]] = statement.Position
			theme.Profile.Baselines = append(theme.Profile.Baselines, newUIModeDecl(parts[1], slots, statement.Position))
		case "mode":
			if section != "profile" || len(parts) < 3 || !isThemeIdentifier(parts[1]) {
				addDiagnostic(statement.Position, "INVALID_UI_MODE", "UI modes must be declared inside a profile with `mode <name> <slot...>`.", "Use `mode box color width style pt pr pb pl radius place` inside `profile UICompact`.")
				continue
			}
			slots := parts[2:]
			validSlots, duplicateSlot := validateUISlotList(slots)
			if !validSlots {
				addDiagnostic(statement.Position, "INVALID_UI_MODE", "UI mode slots must be simple identifiers.", "Use slot names such as color, width, style, pt, pr, pb, pl, radius, or place.")
				continue
			}
			if duplicateSlot != "" {
				addDiagnostic(statement.Position, "DUPLICATE_UI_SLOT", fmt.Sprintf("UI mode %s repeats slot %s.", parts[1], duplicateSlot), "Keep each slot once per mode so positional values map to one property.")
				continue
			}
			if previous, ok := modeNames[parts[1]]; ok {
				addDiagnostic(statement.Position, "DUPLICATE_UI_MODE", fmt.Sprintf("UI mode %s is already declared.", parts[1]), fmt.Sprintf("Keep one mode declaration. First declaration is at line %d.", previous.Line))
				continue
			}
			modeNames[parts[1]] = statement.Position
			theme.Profile.Modes = append(theme.Profile.Modes, newUIModeDecl(parts[1], slots, statement.Position))
		default:
			addDiagnostic(statement.Position, "INVALID_THEME_DECLARATION", fmt.Sprintf("Unsupported .blackthm statement %q.", parts[0]), "Use blackthm, version, target, locked, token, profile, baseline, or mode.")
		}
	}

	if theme.Name == "" {
		addDiagnostic(Position{File: file, Line: 1, Column: 1}, "MISSING_THEME_DECLARATION", "Theme file must declare one theme.", "Start the file with `blackthm <Name> {`.")
	}
	if theme.Version == 0 {
		addDiagnostic(Position{File: file, Line: 1, Column: 1}, "MISSING_THEME_VERSION", "Theme file must declare a version.", "Add `version 1` inside the theme block.")
	}
	if theme.Profile.Name == "" {
		addDiagnostic(Position{File: file, Line: 1, Column: 1}, "MISSING_UI_PROFILE", "Theme file must declare one UI profile.", "Add `profile UICompact { ... }`.")
	}
	if theme.Profile.Name != "" && theme.Profile.Version == 0 {
		addDiagnostic(theme.Profile.Position, "MISSING_UI_PROFILE_VERSION", "UI profile must declare a version.", "Add `version 1` inside the profile block.")
	}
	if theme.Profile.Name != "" && len(theme.Profile.Modes) == 0 {
		addDiagnostic(theme.Profile.Position, "MISSING_UI_MODES", "UI profile must declare at least one mode.", "Add a mode such as `mode box color width style pt pr pb pl radius place`.")
	}
	if theme.Profile.Name != "" && len(theme.Profile.Modes) > 0 {
		validateStandardUIModeGroups(theme.Profile, theme.Locked, addDiagnostic)
	}
	if theme.Locked && theme.Profile.Name != "" {
		validateLockedUIProfile(theme.Profile, addDiagnostic)
	}
	if theme.Target == "" {
		theme.Target = "web"
	}
	return theme, diagnostics
}

func defaultUIProfileRules() UIProfileRules {
	return UIProfileRules{
		InlineSyntax:           "ui <mode> <values...> [| <mode> <values...>...]",
		SlotOrder:              "left-to-right",
		ModeSeparator:          "|",
		MissingTrailingSlots:   "default",
		ExtraValues:            "error",
		DuplicateSlots:         "error",
		LockBaseline:           "required-when-locked",
		ExistingSlotsAfterLock: "immutable",
		NewSlotsAfterLock:      "append-only",
	}
}

func standardUIModeGroups() []UIModeGroup {
	return []UIModeGroup{
		{
			Name:         "box",
			Purpose:      "Container box styling for border, spacing, radius, and placement.",
			AppliesTo:    []string{"form", "table", "component", "panel"},
			DefaultSlots: []string{"color", "width", "style", "pt", "pr", "pb", "pl", "radius", "place"},
			Required:     true,
		},
		{
			Name:         "text",
			Purpose:      "Typography styling for labels, headings, helper text, and body copy.",
			AppliesTo:    []string{"label", "heading", "help", "body"},
			DefaultSlots: []string{"color", "size", "weight", "align"},
			Required:     true,
		},
		{
			Name:         "table",
			Purpose:      "Table-specific styling for borders, density, and row patterns.",
			AppliesTo:    []string{"table", "columns", "rows"},
			DefaultSlots: []string{"color", "width", "style", "density", "zebra"},
			Required:     true,
		},
		{
			Name:         "button",
			Purpose:      "Action control styling for generated and explicit buttons.",
			AppliesTo:    []string{"button", "actions", "submit"},
			DefaultSlots: []string{"bg", "color", "radius", "size", "variant"},
			Required:     true,
		},
	}
}

func standardUIModeGroup(name string) (UIModeGroup, bool) {
	for _, group := range standardUIModeGroups() {
		if group.Name == name {
			return group, true
		}
	}
	return UIModeGroup{}, false
}

func newUIModeDecl(name string, slots []string, position Position) UIModeDecl {
	mode := UIModeDecl{
		Name:     name,
		Standard: false,
		Slots:    slots,
		Position: position,
	}
	if group, ok := standardUIModeGroup(name); ok {
		mode.Standard = true
		mode.Purpose = group.Purpose
		mode.AppliesTo = append([]string(nil), group.AppliesTo...)
	}
	return mode
}

func validateStandardUIModeGroups(profile UIProfileDecl, locked bool, addDiagnostic func(Position, string, string, string)) {
	modes := map[string]bool{}
	for _, mode := range profile.Modes {
		modes[mode.Name] = true
	}
	for _, group := range profile.ModeGroups {
		if !group.Required || modes[group.Name] {
			continue
		}
		suggestion := fmt.Sprintf("Add `mode %s %s` to the profile.", group.Name, strings.Join(group.DefaultSlots, " "))
		if locked {
			suggestion = fmt.Sprintf("Add `baseline %s %s` and `mode %s %s` to the locked profile.", group.Name, strings.Join(group.DefaultSlots, " "), group.Name, strings.Join(group.DefaultSlots, " "))
		}
		addDiagnostic(profile.Position, "MISSING_STANDARD_UI_MODE", fmt.Sprintf("UI profile is missing required standard mode %s.", group.Name), suggestion)
	}
}

func validateUISlotList(slots []string) (bool, string) {
	seenSlots := map[string]bool{}
	for _, slot := range slots {
		if !isThemeIdentifier(slot) {
			return false, ""
		}
		if seenSlots[slot] {
			return true, slot
		}
		seenSlots[slot] = true
	}
	return true, ""
}

func validateLockedUIProfile(profile UIProfileDecl, addDiagnostic func(Position, string, string, string)) {
	if len(profile.Baselines) == 0 {
		addDiagnostic(profile.Position, "MISSING_UI_LOCK_BASELINE", "Locked UI profiles must declare baseline slot order.", "Add `baseline <mode> <slot...>` lines matching the frozen mode prefixes.")
		return
	}

	modes := map[string]UIModeDecl{}
	for _, mode := range profile.Modes {
		modes[mode.Name] = mode
	}
	baselines := map[string]UIModeDecl{}
	for _, baseline := range profile.Baselines {
		baselines[baseline.Name] = baseline
		mode, ok := modes[baseline.Name]
		if !ok {
			addDiagnostic(baseline.Position, "LOCKED_UI_MODE_REMOVED", fmt.Sprintf("Locked UI baseline %s has no matching mode.", baseline.Name), fmt.Sprintf("Restore `mode %s ...` or remove the baseline through a migration.", baseline.Name))
			continue
		}
		if !slotsHavePrefix(mode.Slots, baseline.Slots) {
			addDiagnostic(mode.Position, "NON_APPEND_ONLY_UI_SLOT", fmt.Sprintf("UI mode %s changed locked slot order.", mode.Name), "Keep baseline slots unchanged at the start of the mode and append new slots only at the end.")
		}
	}
	for _, mode := range profile.Modes {
		if _, ok := baselines[mode.Name]; !ok {
			addDiagnostic(mode.Position, "MISSING_UI_LOCK_BASELINE", fmt.Sprintf("Locked UI mode %s has no baseline.", mode.Name), fmt.Sprintf("Add `baseline %s %s` before changing this mode.", mode.Name, strings.Join(mode.Slots, " ")))
		}
	}
}

func slotsHavePrefix(slots []string, prefix []string) bool {
	if len(slots) < len(prefix) {
		return false
	}
	for index, slot := range prefix {
		if slots[index] != slot {
			return false
		}
	}
	return true
}

func hasOpeningBrace(parts []string) bool {
	for _, part := range parts {
		if part == "{" {
			return true
		}
	}
	return false
}

func withoutOpeningBrace(parts []string) []string {
	clean := []string{}
	for _, part := range parts {
		if part == "{" {
			continue
		}
		clean = append(clean, part)
	}
	return clean
}

func parsePositiveThemeInt(value string) (int, bool) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}

func isThemeIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for index, char := range value {
		if index == 0 {
			if !unicode.IsLetter(char) && char != '_' {
				return false
			}
			continue
		}
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return false
		}
	}
	return true
}
