package main

import (
	"fmt"
	"strconv"
	"strings"
)

type JobDecl struct {
	Name             string          `json:"name"`
	Schedule         JobScheduleDecl `json:"schedule,omitempty"`
	Run              JobRunDecl      `json:"run,omitempty"`
	Position         Position        `json:"position"`
	SchedulePosition Position        `json:"schedulePosition,omitempty"`
	RunPosition      Position        `json:"runPosition,omitempty"`
}

type JobScheduleDecl struct {
	Kind     string   `json:"kind,omitempty"`
	Every    int      `json:"every,omitempty"`
	Unit     string   `json:"unit,omitempty"`
	Position Position `json:"position,omitempty"`
}

type JobRunDecl struct {
	Kind     string   `json:"kind,omitempty"`
	Query    string   `json:"query,omitempty"`
	Position Position `json:"position,omitempty"`
}

var supportedJobScheduleUnits = setOf("minutes", "hours", "days")

func (p *parser) parseJob(start int, parts []string) int {
	line := p.lineNumber(start)
	statement := p.lines[start]
	if len(parts) != 3 || len(statement.Tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || statement.Tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_JOB_DECLARATION", "Job declaration must be `job Name {`.", "Example: `job LowStockMonitor {`.")
		return start
	}

	job := JobDecl{Name: parts[1], Position: p.position(line, 1)}
	seen := map[string]bool{}
	for index := start + 1; index < len(p.lines); index++ {
		statement := p.lines[index]
		tokens := statement.Tokens
		line := p.lineNumber(index)
		if len(tokens) == 1 && tokens[0].Kind == tokenSymbol && tokens[0].Value == "}" {
			p.program.Jobs = append(p.program.Jobs, job)
			return index
		}
		if len(tokens) == 0 {
			continue
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(line, 1, "UNEXPECTED_JOB_TOKEN", "Job clauses must begin with schedule or run.", "Use one job clause per line.")
			continue
		}

		keyword := tokens[0].Value
		if keyword == "schedule" || keyword == "run" {
			if seen[keyword] {
				p.addError(line, 1, "DUPLICATE_JOB_"+strings.ToUpper(keyword), fmt.Sprintf("Job %s already declares %s.", job.Name, keyword), "Keep one "+keyword+" clause inside each job.")
				continue
			}
			seen[keyword] = true
		}

		switch keyword {
		case "schedule":
			if len(tokens) != 4 || tokens[1].Kind != tokenIdentifier || tokens[1].Value != "every" || tokens[2].Kind != tokenIdentifier || tokens[3].Kind != tokenIdentifier {
				p.addError(line, 1, "INVALID_JOB_SCHEDULE", "Job schedule must be `schedule every <integer> minutes|hours|days`.", "Example: `schedule every 15 minutes`.")
				continue
			}
			every, err := strconv.Atoi(tokens[2].Value)
			if err != nil || tokens[2].Value != strconv.Itoa(every) {
				p.addError(line, tokens[2].Position.Column, "INVALID_JOB_SCHEDULE", "Job schedule interval must be a positive integer.", "Example: `schedule every 15 minutes`.")
				continue
			}
			job.Schedule = JobScheduleDecl{Kind: "every", Every: every, Unit: tokens[3].Value, Position: tokens[0].Position}
			job.SchedulePosition = tokens[0].Position
		case "run":
			if len(tokens) != 3 || tokens[1].Kind != tokenIdentifier || tokens[1].Value != "query" || !queryStatementIdentifiers(statement, 2) {
				p.addError(line, 1, "INVALID_JOB_RUN", "Job run clause must be `run query QueryName`.", "Example: `run query LowStockProducts`.")
				continue
			}
			job.Run = JobRunDecl{Kind: "query", Query: tokens[2].Value, Position: tokens[0].Position}
			job.RunPosition = tokens[0].Position
		default:
			p.addError(line, 1, "UNEXPECTED_JOB_TOKEN", fmt.Sprintf("Unexpected job token %q.", keyword), "Use schedule or run inside a job.")
		}
	}
	p.addError(line, 1, "UNCLOSED_JOB", fmt.Sprintf("Job %s is missing a closing brace.", job.Name), "Add `}` after the job body.")
	return len(p.lines) - 1
}

func (v *semanticValidator) validateJobs(entityIndex map[string]EntityDecl) map[string]JobDecl {
	jobs := map[string]JobDecl{}
	normalizedNames := map[string]string{}
	queryIndex := map[string]QueryDecl{}
	for _, query := range v.program.Queries {
		if _, exists := queryIndex[query.Name]; !exists {
			queryIndex[query.Name] = query
		}
	}
	symbols := jobOtherSymbols(v.program)

	for _, job := range v.program.Jobs {
		if existing, ok := jobs[job.Name]; ok {
			v.addDiagnostic(job.Position, "DUPLICATE_JOB", fmt.Sprintf("Job %s is already defined.", job.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		jobs[job.Name] = job

		if !queryNamePattern.MatchString(job.Name) {
			v.addDiagnostic(job.Position, "INVALID_JOB_NAME", fmt.Sprintf("Job name %q must use PascalCase letters and digits.", job.Name), "Use a name such as LowStockMonitor.")
		}
		normalized := strings.ToLower(job.Name)
		if existing, ok := normalizedNames[normalized]; ok {
			v.addDiagnostic(job.Position, "JOB_NAME_COLLISION", fmt.Sprintf("Job %s conflicts with job %s after name normalization.", job.Name, existing), "Choose distinct job names, including after lowercasing.")
		}
		normalizedNames[normalized] = job.Name
		if symbols[job.Name] || symbols[normalized] {
			v.addDiagnostic(job.Position, "JOB_NAME_COLLISION", fmt.Sprintf("Job %s conflicts with another application symbol.", job.Name), "Choose a unique job name for unambiguous inspect --affected output.")
		}

		if job.Schedule.Kind == "" {
			v.addDiagnostic(job.Position, "MISSING_JOB_SCHEDULE", fmt.Sprintf("Job %s is missing a schedule.", job.Name), "Add `schedule every 15 minutes` inside the job.")
		} else {
			v.validateJobSchedule(job)
		}
		if job.Run.Kind == "" {
			v.addDiagnostic(job.Position, "MISSING_JOB_RUN", fmt.Sprintf("Job %s is missing a run clause.", job.Name), "Add `run query QueryName` inside the job.")
			continue
		}
		if job.Run.Kind != "query" {
			v.addDiagnostic(job.RunPosition, "UNSUPPORTED_JOB_RUN", fmt.Sprintf("Job %s uses unsupported run mode %q.", job.Name, job.Run.Kind), "Use `run query QueryName` in this MVP.")
			continue
		}
		if !queryNamePattern.MatchString(job.Run.Query) {
			v.addDiagnostic(job.RunPosition, "INVALID_JOB_QUERY", fmt.Sprintf("Job %s references invalid query name %q.", job.Name, job.Run.Query), "Use an existing PascalCase query name.")
			continue
		}
		query, ok := queryIndex[job.Run.Query]
		if !ok {
			v.addDiagnostic(job.RunPosition, "UNKNOWN_JOB_QUERY", fmt.Sprintf("Job %s references unknown query %s.", job.Name, job.Run.Query), "Declare the query at the top level or change the job run clause.")
			continue
		}
		if _, ok := entityIndex[query.Source]; !ok && query.Source != "" {
			v.addDiagnostic(job.RunPosition, "UNKNOWN_JOB_QUERY_SOURCE", fmt.Sprintf("Job %s references query %s with unknown source %s.", job.Name, query.Name, query.Source), "Fix the query source before running it from a job.")
		}
	}
	return jobs
}

func (v *semanticValidator) validateJobSchedule(job JobDecl) {
	if job.Schedule.Kind != "every" {
		v.addDiagnostic(job.SchedulePosition, "UNSUPPORTED_JOB_SCHEDULE", fmt.Sprintf("Job %s uses unsupported schedule kind %q.", job.Name, job.Schedule.Kind), "Use `schedule every <integer> minutes|hours|days`.")
		return
	}
	if !supportedJobScheduleUnits[job.Schedule.Unit] {
		v.addDiagnostic(job.SchedulePosition, "UNSUPPORTED_JOB_SCHEDULE_UNIT", fmt.Sprintf("Job %s uses unsupported schedule unit %q.", job.Name, job.Schedule.Unit), "Use minutes, hours, or days.")
		return
	}
	max := map[string]int{"minutes": 1440, "hours": 168, "days": 365}[job.Schedule.Unit]
	if job.Schedule.Every < 1 || job.Schedule.Every > max {
		v.addDiagnostic(job.SchedulePosition, "INVALID_JOB_SCHEDULE_INTERVAL", fmt.Sprintf("Job %s schedule interval %d %s is outside the supported range.", job.Name, job.Schedule.Every, job.Schedule.Unit), "Use 1..1440 minutes, 1..168 hours, or 1..365 days.")
	}
}

func jobOtherSymbols(program Program) map[string]bool {
	symbols := queryOtherSymbols(program)
	for _, query := range program.Queries {
		symbols[query.Name] = true
		symbols[strings.ToLower(query.Name)] = true
	}
	return symbols
}

func findJob(program Program, name string) (JobDecl, bool) {
	for _, job := range program.Jobs {
		if job.Name == name {
			return job, true
		}
	}
	return JobDecl{}, false
}

func programHasJobs(program Program) bool {
	return len(program.Jobs) > 0
}

func jobScheduleLabel(job JobDecl) string {
	if job.Schedule.Kind == "" {
		return ""
	}
	return fmt.Sprintf("%s %d %s", job.Schedule.Kind, job.Schedule.Every, job.Schedule.Unit)
}
