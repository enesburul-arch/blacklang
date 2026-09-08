package main

func BenchmarkEvalCorpus(args []string) AIEvalCorpusResult {
	tasks := BenchmarkTasks(args)
	result := AIEvalCorpusResult{
		Success: false,
		Command: "benchmark eval",
		Version: version,
		Config:  tasks.Config,
		Summary: tasks.Summary,
		Errors:  []Diagnostic{},
	}
	if !tasks.Success {
		result.Errors = tasks.Errors
		return result
	}

	repeat := 3
	timeoutMinutes := 120
	result.Baseline = tasks.Baseline
	result.Suite = AIEvalSuite{
		ID:             "BLACKLANG-WEB-AI-EVAL-v0.2",
		Name:           "BlackLang Web AI Eval Corpus",
		Purpose:        "Measure repeated AI editing cost, validation discipline, generated-output safety, and deterministic web behavior across common BlackLang web tasks.",
		Mode:           "long-running",
		Repeat:         repeat,
		TimeoutMinutes: timeoutMinutes,
		Policy:         "The corpus is read-only metadata. It does not call an AI model, mutate generated output, commit, push, deploy, or store secrets. A harness can run each case repeatedly and score evidence from source diffs, CLI JSON, generated tests, and worklog notes.",
		Commands: []string{
			"black benchmark eval --json",
			"black benchmark eval --ir",
			"black benchmark tasks --json",
			"black benchmark coverage --json",
		},
	}
	result.Cases = buildAIEvalCases(tasks.Scenarios)
	result.Totals = summarizeAIEvalCases(result.Cases, repeat)
	result.Success = true
	return result
}

func buildAIEvalCases(scenarios []AITaskScenario) []AIEvalCase {
	cases := make([]AIEvalCase, 0, len(scenarios))
	for _, scenario := range scenarios {
		cases = append(cases, AIEvalCase{
			ID:         "AI-EVAL-" + scenario.ID,
			ScenarioID: scenario.ID,
			Name:       scenario.Name,
			Prompt:     evalPromptForScenario(scenario),
			ExpectedEvidence: []string{
				"source diff touches .black/.blackthm or compiler-owned source files, not generated output by hand",
				"JSON diagnostics are clean or included with stable diagnostic codes",
				"inspect --affected output lists impacted symbols and generated files",
				"generated output is rebuilt by black build when behavior affects generated files",
				"required generated app commands pass or failures are recorded as deterministic JSON diagnostics",
			},
			RequiredCommands: append([]string(nil), scenario.Commands...),
			Scoring: []AIEvalScoreCriterion{
				{ID: "source-intent", Description: "Implements the task through the smallest deterministic BlackLang source/compiler change that matches the current capability boundary.", Points: 30},
				{ID: "agent-learning", Description: "Uses docs/explain/inspect JSON outputs before changing behavior and records affected symbols.", Points: 20},
				{ID: "validation", Description: "Runs required CLI and generated app validation commands without hiding diagnostics.", Points: 30},
				{ID: "generated-boundary", Description: "Does not hand-edit generated output as the solution and keeps secrets outside .black source.", Points: 20},
			},
			EstimatedTokens: scenario.EstimatedTokens,
		})
	}
	return cases
}

func evalPromptForScenario(scenario AITaskScenario) string {
	return "You are editing a BlackLang web project. " + scenario.Trigger + " Read local docs first, use JSON/IR CLI outputs, keep syntax deterministic, do not hand-edit generated output as the solution, run the listed validation commands, and record evidence for this eval case."
}

func summarizeAIEvalCases(cases []AIEvalCase, repeat int) AIEvalTotals {
	totals := AIEvalTotals{
		CaseCount:                 len(cases),
		Repeat:                    repeat,
		TotalRuns:                 len(cases) * repeat,
		MaxScorePerRun:            len(cases) * 100,
		MaxScoreAcrossRepeats:     len(cases) * repeat * 100,
		EstimatedMinutesPerRepeat: len(cases) * 12,
		EstimatedMinutesTotal:     len(cases) * repeat * 12,
	}
	for _, item := range cases {
		totals.BlackLangTotal += item.EstimatedTokens.BlackLangTotal
		totals.ConventionalTotal += item.EstimatedTokens.ConventionalTotal
	}
	totals.EstimatedSavingsPercent = tokenSavingsPercent(totals.BlackLangTotal, totals.ConventionalTotal)
	return totals
}
