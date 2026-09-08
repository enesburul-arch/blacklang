package main

type Position struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type Diagnostic struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

type VersionResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Name    string       `json:"name"`
	Version string       `json:"version"`
	Errors  []Diagnostic `json:"errors"`
}

type ParseResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Version string       `json:"version"`
	File    string       `json:"file,omitempty"`
	Program Program      `json:"program,omitempty"`
	Errors  []Diagnostic `json:"errors"`
}

type FormatResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Version string       `json:"version"`
	File    string       `json:"file,omitempty"`
	Changed bool         `json:"changed"`
	Check   bool         `json:"check"`
	Stdout  bool         `json:"stdout"`
	Errors  []Diagnostic `json:"errors"`
}

type LintResult struct {
	Success  bool         `json:"success"`
	Command  string       `json:"command"`
	Version  string       `json:"version"`
	File     string       `json:"file,omitempty"`
	Summary  Summary      `json:"summary"`
	Checks   []LintCheck  `json:"checks"`
	Findings []Diagnostic `json:"findings"`
	Errors   []Diagnostic `json:"errors"`
}

type LintCheck struct {
	Name     string `json:"name"`
	Success  bool   `json:"success"`
	Findings int    `json:"findings"`
}

type AccessibilityAuditResult struct {
	Success  bool         `json:"success"`
	Command  string       `json:"command"`
	Version  string       `json:"version"`
	File     string       `json:"file,omitempty"`
	Summary  Summary      `json:"summary"`
	Findings []Diagnostic `json:"findings"`
	Errors   []Diagnostic `json:"errors"`
}

type ValidateResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Version string       `json:"version"`
	File    string       `json:"file,omitempty"`
	Summary Summary      `json:"summary"`
	Errors  []Diagnostic `json:"errors"`
}

type BuildResult struct {
	Success bool            `json:"success"`
	Command string          `json:"command"`
	Version string          `json:"version"`
	File    string          `json:"file,omitempty"`
	OutDir  string          `json:"outDir"`
	Summary Summary         `json:"summary"`
	Files   []GeneratedFile `json:"files"`
	Errors  []Diagnostic    `json:"errors"`
}

type BenchmarkResult struct {
	Success        bool                  `json:"success"`
	Command        string                `json:"command"`
	Version        string                `json:"version"`
	Config         ConfigInfo            `json:"config"`
	Summary        Summary               `json:"summary"`
	Source         BenchmarkTotals       `json:"source"`
	Generated      BenchmarkTotals       `json:"generated"`
	Ratios         BenchmarkRatios       `json:"ratios"`
	SourceFiles    []BenchmarkFileMetric `json:"sourceFiles"`
	GeneratedFiles []BenchmarkFileMetric `json:"generatedFiles"`
	GeneratedKinds []BenchmarkKindMetric `json:"generatedKinds"`
	Errors         []Diagnostic          `json:"errors"`
}

type BenchmarkTotals struct {
	Files int   `json:"files"`
	Lines int   `json:"lines"`
	Bytes int64 `json:"bytes"`
}

type BenchmarkRatios struct {
	GeneratedToSourceLines float64 `json:"generatedToSourceLines"`
	SourceToGeneratedLines float64 `json:"sourceToGeneratedLines"`
}

type BenchmarkFileMetric struct {
	Path  string `json:"path"`
	Kind  string `json:"kind"`
	Lines int    `json:"lines"`
	Bytes int64  `json:"bytes"`
}

type BenchmarkKindMetric struct {
	Kind  string `json:"kind"`
	Files int    `json:"files"`
	Lines int    `json:"lines"`
	Bytes int64  `json:"bytes"`
}

type AITaskBenchmarkResult struct {
	Success   bool                    `json:"success"`
	Command   string                  `json:"command"`
	Version   string                  `json:"version"`
	Config    ConfigInfo              `json:"config"`
	Summary   Summary                 `json:"summary"`
	Baseline  AITaskBenchmarkBaseline `json:"baseline"`
	Scenarios []AITaskScenario        `json:"scenarios"`
	Totals    AITaskBenchmarkTotals   `json:"totals"`
	Errors    []Diagnostic            `json:"errors"`
}

type AITaskBenchmarkBaseline struct {
	SourceFiles            int     `json:"sourceFiles"`
	SourceLines            int     `json:"sourceLines"`
	SourceBytes            int64   `json:"sourceBytes"`
	GeneratedFiles         int     `json:"generatedFiles"`
	GeneratedLines         int     `json:"generatedLines"`
	GeneratedBytes         int64   `json:"generatedBytes"`
	GeneratedToSourceLines float64 `json:"generatedToSourceLines"`
	SourceTokensPerLine    int     `json:"sourceTokensPerLine"`
	GeneratedTokensPerLine int     `json:"generatedTokensPerLine"`
}

type AITaskScenario struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Purpose         string          `json:"purpose"`
	Trigger         string          `json:"trigger"`
	ProjectEvidence []string        `json:"projectEvidence"`
	BlackLangEdits  []string        `json:"blacklangEdits"`
	GeneratedImpact []string        `json:"generatedImpact"`
	Commands        []string        `json:"commands"`
	EstimatedTokens AITokenEstimate `json:"estimatedTokens"`
}

type AITokenEstimate struct {
	BlackLangInput          int      `json:"blacklangInput"`
	BlackLangOutput         int      `json:"blacklangOutput"`
	BlackLangTotal          int      `json:"blacklangTotal"`
	ConventionalInput       int      `json:"conventionalInput"`
	ConventionalOutput      int      `json:"conventionalOutput"`
	ConventionalTotal       int      `json:"conventionalTotal"`
	EstimatedSavingsPercent int      `json:"estimatedSavingsPercent"`
	Basis                   []string `json:"basis"`
}

type AITaskBenchmarkTotals struct {
	ScenarioCount           int `json:"scenarioCount"`
	BlackLangTotal          int `json:"blacklangTotal"`
	ConventionalTotal       int `json:"conventionalTotal"`
	EstimatedSavingsPercent int `json:"estimatedSavingsPercent"`
}

type AIEvalCorpusResult struct {
	Success  bool                    `json:"success"`
	Command  string                  `json:"command"`
	Version  string                  `json:"version"`
	Config   ConfigInfo              `json:"config"`
	Summary  Summary                 `json:"summary"`
	Baseline AITaskBenchmarkBaseline `json:"baseline"`
	Suite    AIEvalSuite             `json:"suite"`
	Cases    []AIEvalCase            `json:"cases"`
	Totals   AIEvalTotals            `json:"totals"`
	Errors   []Diagnostic            `json:"errors"`
}

type AIEvalSuite struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Purpose        string   `json:"purpose"`
	Mode           string   `json:"mode"`
	Repeat         int      `json:"repeat"`
	TimeoutMinutes int      `json:"timeoutMinutes"`
	Policy         string   `json:"policy"`
	Commands       []string `json:"commands"`
}

type AIEvalCase struct {
	ID               string                 `json:"id"`
	ScenarioID       string                 `json:"scenarioId"`
	Name             string                 `json:"name"`
	Prompt           string                 `json:"prompt"`
	ExpectedEvidence []string               `json:"expectedEvidence"`
	RequiredCommands []string               `json:"requiredCommands"`
	Scoring          []AIEvalScoreCriterion `json:"scoring"`
	EstimatedTokens  AITokenEstimate        `json:"estimatedTokens"`
}

type AIEvalScoreCriterion struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Points      int    `json:"points"`
}

type AIEvalTotals struct {
	CaseCount                 int `json:"caseCount"`
	Repeat                    int `json:"repeat"`
	TotalRuns                 int `json:"totalRuns"`
	MaxScorePerRun            int `json:"maxScorePerRun"`
	MaxScoreAcrossRepeats     int `json:"maxScoreAcrossRepeats"`
	EstimatedMinutesPerRepeat int `json:"estimatedMinutesPerRepeat"`
	EstimatedMinutesTotal     int `json:"estimatedMinutesTotal"`
	BlackLangTotal            int `json:"blacklangTotal"`
	ConventionalTotal         int `json:"conventionalTotal"`
	EstimatedSavingsPercent   int `json:"estimatedSavingsPercent"`
}

type AIEvalHistoryResult struct {
	Success bool                 `json:"success"`
	Command string               `json:"command"`
	Version string               `json:"version"`
	History AIEvalHistory        `json:"history"`
	Summary AIEvalHistorySummary `json:"summary"`
	Runs    []AIEvalHistoryRun   `json:"runs"`
	Errors  []Diagnostic         `json:"errors"`
}

type AIEvalHistory struct {
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Status      string   `json:"status"`
	Source      string   `json:"source"`
	PublishedAt string   `json:"publishedAt"`
	Policy      string   `json:"policy"`
	Commands    []string `json:"commands"`
}

type AIEvalHistorySummary struct {
	RunCount        int    `json:"runCount"`
	ModelCount      int    `json:"modelCount"`
	TotalRuns       int    `json:"totalRuns"`
	PassedRuns      int    `json:"passedRuns"`
	FailedRuns      int    `json:"failedRuns"`
	AverageScore    int    `json:"averageScore"`
	CoveragePercent int    `json:"coveragePercent"`
	LatestRun       string `json:"latestRun"`
}

type AIEvalHistoryRun struct {
	ID              string                     `json:"id"`
	SuiteID         string                     `json:"suiteId"`
	Source          string                     `json:"source"`
	Status          string                     `json:"status"`
	StartedAt       string                     `json:"startedAt"`
	CompletedAt     string                     `json:"completedAt"`
	CoveragePercent int                        `json:"coveragePercent"`
	CaseCount       int                        `json:"caseCount"`
	Repeat          int                        `json:"repeat"`
	TotalRuns       int                        `json:"totalRuns"`
	PassedRuns      int                        `json:"passedRuns"`
	FailedRuns      int                        `json:"failedRuns"`
	AverageScore    int                        `json:"averageScore"`
	Evidence        []string                   `json:"evidence"`
	Notes           []string                   `json:"notes"`
	Models          []AIEvalHistoryModelResult `json:"models"`
}

type AIEvalHistoryModelResult struct {
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	Source       string   `json:"source"`
	Runs         int      `json:"runs"`
	Passed       int      `json:"passed"`
	Failed       int      `json:"failed"`
	AverageScore int      `json:"averageScore"`
	MedianTokens int      `json:"medianTokens"`
	Notes        []string `json:"notes"`
}

type CoverageResult struct {
	Success           bool                `json:"success"`
	Command           string              `json:"command"`
	Version           string              `json:"version"`
	Target            string              `json:"target"`
	CompletionPercent int                 `json:"completionPercent"`
	Areas             []CoverageArea      `json:"areas"`
	Issues            []CoverageIssue     `json:"issues"`
	Milestones        []CoverageMilestone `json:"milestones"`
	Errors            []Diagnostic        `json:"errors"`
}

type CoverageArea struct {
	Name        string   `json:"name"`
	Weight      int      `json:"weight"`
	Score       int      `json:"score"`
	Status      string   `json:"status"`
	Implemented []string `json:"implemented"`
	Missing     []string `json:"missing"`
	Next        []string `json:"next"`
}

type CoverageIssue struct {
	ID       string `json:"id"`
	Phase    string `json:"phase"`
	Area     string `json:"area"`
	Title    string `json:"title"`
	Priority string `json:"priority"`
	Status   string `json:"status"`
}

type CoverageMilestone struct {
	Percent int    `json:"percent"`
	Meaning string `json:"meaning"`
}

type CoverageIssuesResult struct {
	Success           bool                  `json:"success"`
	Command           string                `json:"command"`
	Version           string                `json:"version"`
	Target            string                `json:"target"`
	CompletionPercent int                   `json:"completionPercent"`
	Summary           CoverageIssuesSummary `json:"summary"`
	Issues            []CoverageIssue       `json:"issues"`
	OpenIssues        []CoverageIssue       `json:"openIssues"`
	Errors            []Diagnostic          `json:"errors"`
}

type CoverageIssuesSummary struct {
	Total int `json:"total"`
	Open  int `json:"open"`
	Done  int `json:"done"`
}

type IDESupportResult struct {
	Success         bool                `json:"success"`
	Command         string              `json:"command"`
	Version         string              `json:"version"`
	Language        IDELanguageInfo     `json:"language"`
	Capabilities    IDECapabilities     `json:"capabilities"`
	Commands        []IDECommand        `json:"commands"`
	CompletionItems []IDECompletionItem `json:"completionItems"`
	Snippets        []IDESnippet        `json:"snippets"`
	DiagnosticCodes []IDEDiagnosticCode `json:"diagnosticCodes"`
	Errors          []Diagnostic        `json:"errors"`
}

type IDELanguageInfo struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	LanguageVersion string   `json:"languageVersion"`
	Extensions      []string `json:"extensions"`
	CommentPrefix   string   `json:"commentPrefix"`
}

type IDECapabilities struct {
	Diagnostics       string `json:"diagnostics"`
	CompletionItems   bool   `json:"completionItems"`
	Snippets          bool   `json:"snippets"`
	DiagnosticCodes   bool   `json:"diagnosticCodes"`
	PositionBase      string `json:"positionBase"`
	RangeEndExclusive bool   `json:"rangeEndExclusive"`
}

type IDECommand struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Purpose string `json:"purpose"`
}

type IDECompletionItem struct {
	Label         string `json:"label"`
	Kind          string `json:"kind"`
	Context       string `json:"context"`
	Detail        string `json:"detail,omitempty"`
	InsertText    string `json:"insertText"`
	Documentation string `json:"documentation,omitempty"`
}

type IDESnippet struct {
	Prefix      string   `json:"prefix"`
	Context     string   `json:"context"`
	Description string   `json:"description"`
	Body        []string `json:"body"`
}

type IDEDiagnosticCode struct {
	Code          string `json:"code"`
	Severity      string `json:"severity"`
	Documentation string `json:"documentation"`
}

type IDEDiagnosticsResult struct {
	Success     bool                  `json:"success"`
	Command     string                `json:"command"`
	Version     string                `json:"version"`
	File        string                `json:"file,omitempty"`
	Valid       bool                  `json:"valid"`
	Project     Summary               `json:"project"`
	Summary     IDEDiagnosticsSummary `json:"summary"`
	Diagnostics []IDEDiagnostic       `json:"diagnostics"`
	Errors      []Diagnostic          `json:"errors"`
}

type IDEDiagnosticsSummary struct {
	Total    int `json:"total"`
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

type IDEDiagnostic struct {
	File       string   `json:"file"`
	Range      IDERange `json:"range"`
	Severity   string   `json:"severity"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion,omitempty"`
	Source     string   `json:"source"`
}

type IDERange struct {
	Start IDEPosition `json:"start"`
	End   IDEPosition `json:"end"`
}

type IDEPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type EcosystemResult struct {
	Success       bool                   `json:"success"`
	Command       string                 `json:"command"`
	Version       string                 `json:"version"`
	Release       EcosystemRelease       `json:"release"`
	Packages      []EcosystemPackage     `json:"packages"`
	Adapters      []EcosystemAdapter     `json:"adapters"`
	Registries    []EcosystemRegistry    `json:"registries"`
	Marketplaces  []EcosystemMarketplace `json:"marketplaces"`
	PublicIndex   EcosystemPublicIndex   `json:"publicIndex"`
	TrustWorkflow EcosystemTrustWorkflow `json:"trustWorkflow"`
	Policies      []string               `json:"policies"`
	Errors        []Diagnostic           `json:"errors"`
}

type EcosystemRelease struct {
	Channel          string         `json:"channel"`
	ArtifactManifest string         `json:"artifactManifest"`
	ChecksumFile     string         `json:"checksumFile"`
	Scripts          []AgentCommand `json:"scripts"`
	Trust            ReleaseTrust   `json:"trust"`
}

type ReleaseTrust struct {
	SignatureAlgorithm    string                   `json:"signatureAlgorithm"`
	SignatureFilePattern  string                   `json:"signatureFilePattern"`
	PublicKeyEnv          string                   `json:"publicKeyEnv"`
	PublicKeyFileEnv      string                   `json:"publicKeyFileEnv"`
	VerifyCommand         string                   `json:"verifyCommand"`
	RequiredBeforePublish []string                 `json:"requiredBeforePublish"`
	TransparencyLog       ReleaseTransparencyLog   `json:"transparencyLog"`
	KeyRotation           ReleaseKeyRotationPolicy `json:"keyRotation"`
}

type ReleaseTransparencyLog struct {
	PolicyFile            string   `json:"policyFile"`
	ReleaseLogFile        string   `json:"releaseLogFile"`
	Status                string   `json:"status"`
	HashAlgorithm         string   `json:"hashAlgorithm"`
	EntryFields           []string `json:"entryFields"`
	RequiredBeforePublish []string `json:"requiredBeforePublish"`
}

type ReleaseKeyRotationPolicy struct {
	PolicyFile            string   `json:"policyFile"`
	Status                string   `json:"status"`
	KeyIDFormat           string   `json:"keyIdFormat"`
	MinimumOverlapDays    int      `json:"minimumOverlapDays"`
	RevocationFile        string   `json:"revocationFile"`
	RequiredBeforePublish []string `json:"requiredBeforePublish"`
}

type EcosystemPackage struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	Status   string   `json:"status"`
	Commands []string `json:"commands"`
}

type EcosystemAdapter struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Source       string   `json:"source"`
	Docs         []string `json:"docs"`
	Capabilities []string `json:"capabilities"`
}

type EcosystemRegistry struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	Status   string   `json:"status"`
	Packages []string `json:"packages"`
	Docs     []string `json:"docs"`
	Commands []string `json:"commands"`
	Trust    []string `json:"trust"`
}

type EcosystemMarketplace struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"`
	Path      string   `json:"path"`
	Status    string   `json:"status"`
	Adapters  []string `json:"adapters"`
	Providers []string `json:"providers"`
	Docs      []string `json:"docs"`
	Commands  []string `json:"commands"`
	Trust     []string `json:"trust"`
}

type EcosystemPublicIndex struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	Status   string   `json:"status"`
	Formats  []string `json:"formats"`
	Sources  []string `json:"sources"`
	Packages []string `json:"packages"`
	Adapters []string `json:"adapters"`
	Docs     []string `json:"docs"`
	Commands []string `json:"commands"`
	Trust    []string `json:"trust"`
}

type EcosystemTrustWorkflow struct {
	PolicyFiles    []string `json:"policyFiles"`
	RequiredChecks []string `json:"requiredChecks"`
	InstallSteps   []string `json:"installSteps"`
}

type SecurityScanResult struct {
	Success  bool         `json:"success"`
	Command  string       `json:"command"`
	Version  string       `json:"version"`
	File     string       `json:"file,omitempty"`
	Findings []Diagnostic `json:"findings"`
	Errors   []Diagnostic `json:"errors"`
}

type EncryptedSourceResult struct {
	Success          bool         `json:"success"`
	Command          string       `json:"command"`
	Version          string       `json:"version"`
	Status           string       `json:"status"`
	Extension        string       `json:"extension"`
	ProtectedFiles   []string     `json:"protectedFiles"`
	ProductionPolicy string       `json:"productionPolicy"`
	BuildPolicy      string       `json:"buildPolicy"`
	KeyPolicy        string       `json:"keyPolicy"`
	Rules            []string     `json:"rules"`
	Errors           []Diagnostic `json:"errors"`
}

type SecurityEncryptResult struct {
	Success         bool                  `json:"success"`
	Command         string                `json:"command"`
	Version         string                `json:"version"`
	File            string                `json:"file,omitempty"`
	OutFile         string                `json:"outFile,omitempty"`
	KeyEnv          string                `json:"keyEnv"`
	Header          EncryptedSourceHeader `json:"header"`
	PlaintextBytes  int64                 `json:"plaintextBytes"`
	CiphertextBytes int64                 `json:"ciphertextBytes"`
	Errors          []Diagnostic          `json:"errors"`
}

type SecurityDecryptResult struct {
	Success        bool                  `json:"success"`
	Command        string                `json:"command"`
	Version        string                `json:"version"`
	File           string                `json:"file,omitempty"`
	KeyEnv         string                `json:"keyEnv"`
	Header         EncryptedSourceHeader `json:"header"`
	Plaintext      string                `json:"plaintext,omitempty"`
	PlaintextBytes int64                 `json:"plaintextBytes"`
	Errors         []Diagnostic          `json:"errors"`
}

type EncryptedSourceHeader struct {
	Version     string `json:"version"`
	Alg         string `json:"alg"`
	KDF         string `json:"kdf"`
	KeyEnv      string `json:"keyEnv"`
	NonceBase64 string `json:"nonceBase64"`
}

type PackageResult struct {
	Success bool            `json:"success"`
	Command string          `json:"command"`
	Version string          `json:"version"`
	Mode    string          `json:"mode"`
	OutDir  string          `json:"outDir"`
	Files   []GeneratedFile `json:"files"`
	Errors  []Diagnostic    `json:"errors"`
}

type InitResult struct {
	Success bool            `json:"success"`
	Command string          `json:"command"`
	Version string          `json:"version"`
	Root    string          `json:"root"`
	Files   []GeneratedFile `json:"files"`
	Errors  []Diagnostic    `json:"errors"`
}

type InspectResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Version string       `json:"version"`
	Config  ConfigInfo   `json:"config"`
	Summary Summary      `json:"summary"`
	Program Program      `json:"program,omitempty"`
	Errors  []Diagnostic `json:"errors"`
}

type InspectAffectedResult struct {
	Success  bool             `json:"success"`
	Command  string           `json:"command"`
	Version  string           `json:"version"`
	Config   ConfigInfo       `json:"config"`
	Summary  Summary          `json:"summary"`
	Affected AffectedAnalysis `json:"affected"`
	Errors   []Diagnostic     `json:"errors"`
}

type AffectedAnalysis struct {
	Symbol         string         `json:"symbol"`
	Kind           string         `json:"kind"`
	Found          bool           `json:"found"`
	Entity         string         `json:"entity,omitempty"`
	Field          string         `json:"field,omitempty"`
	Entities       []AffectedItem `json:"entities"`
	Migrations     []AffectedItem `json:"migrations"`
	Queries        []AffectedItem `json:"queries"`
	Jobs           []AffectedItem `json:"jobs,omitempty"`
	Actions        []AffectedItem `json:"actions"`
	Transactions   []AffectedItem `json:"transactions,omitempty"`
	Services       []AffectedItem `json:"services,omitempty"`
	Seeds          []AffectedItem `json:"seeds,omitempty"`
	Tests          []AffectedItem `json:"tests,omitempty"`
	Pages          []AffectedItem `json:"pages"`
	Roles          []AffectedItem `json:"roles"`
	Workflows      []AffectedItem `json:"workflows"`
	States         []AffectedItem `json:"states"`
	Components     []AffectedItem `json:"components"`
	APIs           []AffectedItem `json:"apis"`
	GeneratedFiles []AffectedItem `json:"generatedFiles"`
	AgentNotes     []string       `json:"agentNotes"`
}

type AffectedItem struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type SchemaMigrationPlanResult struct {
	Success        bool                    `json:"success"`
	Command        string                  `json:"command"`
	Version        string                  `json:"version"`
	OldFile        string                  `json:"oldFile,omitempty"`
	NewFile        string                  `json:"newFile,omitempty"`
	Safe           bool                    `json:"safe"`
	Destructive    bool                    `json:"destructive"`
	Summary        SchemaMigrationSummary  `json:"summary"`
	Changes        []SchemaMigrationChange `json:"changes"`
	Steps          []SchemaMigrationStep   `json:"steps"`
	GeneratedFiles []AffectedItem          `json:"generatedFiles"`
	AgentNotes     []string                `json:"agentNotes"`
	Errors         []Diagnostic            `json:"errors"`
}

type SchemaMigrationSummary struct {
	OldApp             string `json:"oldApp,omitempty"`
	NewApp             string `json:"newApp,omitempty"`
	EntitiesAdded      int    `json:"entitiesAdded"`
	EntitiesRemoved    int    `json:"entitiesRemoved"`
	FieldsAdded        int    `json:"fieldsAdded"`
	FieldsRemoved      int    `json:"fieldsRemoved"`
	FieldChanges       int    `json:"fieldChanges"`
	IndexesAdded       int    `json:"indexesAdded"`
	IndexesRemoved     int    `json:"indexesRemoved"`
	Renames            int    `json:"renames"`
	SafeChanges        int    `json:"safeChanges"`
	ManualChanges      int    `json:"manualChanges"`
	DestructiveChanges int    `json:"destructiveChanges"`
}

type SchemaMigrationChange struct {
	Type        string   `json:"type"`
	Risk        string   `json:"risk"`
	Entity      string   `json:"entity,omitempty"`
	Field       string   `json:"field,omitempty"`
	Column      string   `json:"column,omitempty"`
	OldType     string   `json:"oldType,omitempty"`
	NewType     string   `json:"newType,omitempty"`
	OldValue    string   `json:"oldValue,omitempty"`
	NewValue    string   `json:"newValue,omitempty"`
	Index       string   `json:"index,omitempty"`
	IndexFields []string `json:"indexFields,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

type SchemaMigrationStep struct {
	Order  int    `json:"order"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type DocsResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Version string       `json:"version"`
	Doc     DocEntry     `json:"doc,omitempty"`
	Errors  []Diagnostic `json:"errors"`
}

type DocsAllResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Version string       `json:"version"`
	Count   int          `json:"count"`
	Docs    []DocEntry   `json:"docs"`
	Errors  []Diagnostic `json:"errors"`
}

type ExplainResult struct {
	Success    bool         `json:"success"`
	Command    string       `json:"command"`
	Version    string       `json:"version"`
	Keyword    string       `json:"keyword,omitempty"`
	Purpose    string       `json:"purpose,omitempty"`
	Syntax     string       `json:"syntax,omitempty"`
	Example    string       `json:"example,omitempty"`
	AgentSteps []string     `json:"agentSteps"`
	AgentNotes []string     `json:"agentNotes"`
	Related    []string     `json:"related"`
	ErrorCodes []string     `json:"errorCodes"`
	Errors     []Diagnostic `json:"errors"`
}

type AgentStartupResult struct {
	Success       bool                 `json:"success"`
	Command       string               `json:"command"`
	Version       string               `json:"version"`
	Config        ConfigInfo           `json:"config"`
	Summary       Summary              `json:"summary"`
	ReadFirst     []AgentReadFile      `json:"readFirst"`
	SourceFiles   []string             `json:"sourceFiles"`
	ThemeFiles    []string             `json:"themeFiles,omitempty"`
	GeneratedDirs []string             `json:"generatedDirs"`
	Checklist     []AgentChecklistItem `json:"checklist"`
	Commands      []AgentCommand       `json:"commands"`
	Policies      []string             `json:"policies"`
	Errors        []Diagnostic         `json:"errors"`
}

type AgentReadFile struct {
	Path    string `json:"path"`
	Purpose string `json:"purpose"`
	Exists  bool   `json:"exists"`
}

type AgentChecklistItem struct {
	Step   int    `json:"step"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type AgentCommand struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Purpose string `json:"purpose"`
}

type DocEntry struct {
	Keyword    string   `json:"keyword"`
	Purpose    string   `json:"purpose"`
	Syntax     string   `json:"syntax"`
	Example    string   `json:"example"`
	AgentNotes []string `json:"agentNotes"`
	Errors     []string `json:"errors"`
}

type ThemeInspectResult struct {
	Success bool         `json:"success"`
	Command string       `json:"command"`
	Version string       `json:"version"`
	File    string       `json:"file,omitempty"`
	Theme   ThemeDecl    `json:"theme,omitempty"`
	Errors  []Diagnostic `json:"errors"`
}

type ThemeMigrationResult struct {
	Success bool                   `json:"success"`
	Command string                 `json:"command"`
	Version string                 `json:"version"`
	OldFile string                 `json:"oldFile,omitempty"`
	NewFile string                 `json:"newFile,omitempty"`
	Safe    bool                   `json:"safe"`
	Summary ThemeMigrationSummary  `json:"summary"`
	Changes []ThemeMigrationChange `json:"changes"`
	Errors  []Diagnostic           `json:"errors"`
}

type ThemeMigrationSummary struct {
	OldTheme          string `json:"oldTheme,omitempty"`
	NewTheme          string `json:"newTheme,omitempty"`
	OldVersion        int    `json:"oldVersion,omitempty"`
	NewVersion        int    `json:"newVersion,omitempty"`
	Target            string `json:"target,omitempty"`
	OldLocked         bool   `json:"oldLocked"`
	NewLocked         bool   `json:"newLocked"`
	OldProfile        string `json:"oldProfile,omitempty"`
	NewProfile        string `json:"newProfile,omitempty"`
	OldProfileVersion int    `json:"oldProfileVersion,omitempty"`
	NewProfileVersion int    `json:"newProfileVersion,omitempty"`
}

type ThemeMigrationChange struct {
	Type     string `json:"type"`
	Profile  string `json:"profile,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Slot     string `json:"slot,omitempty"`
	Index    int    `json:"index,omitempty"`
	OldValue string `json:"oldValue,omitempty"`
	NewValue string `json:"newValue,omitempty"`
}

type ThemeDecl struct {
	Name     string        `json:"name"`
	Version  int           `json:"version"`
	Target   string        `json:"target,omitempty"`
	Locked   bool          `json:"locked"`
	Tokens   []ThemeToken  `json:"tokens"`
	Profile  UIProfileDecl `json:"profile"`
	Position Position      `json:"position"`
}

type ThemeToken struct {
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	Value    string   `json:"value"`
	Position Position `json:"position"`
}

type UIProfileDecl struct {
	Name       string         `json:"name"`
	Version    int            `json:"version"`
	Rules      UIProfileRules `json:"rules"`
	ModeGroups []UIModeGroup  `json:"modeGroups"`
	Baselines  []UIModeDecl   `json:"baselines,omitempty"`
	Modes      []UIModeDecl   `json:"modes"`
	Position   Position       `json:"position"`
}

type UIModeGroup struct {
	Name         string   `json:"name"`
	Purpose      string   `json:"purpose"`
	AppliesTo    []string `json:"appliesTo"`
	DefaultSlots []string `json:"defaultSlots"`
	Required     bool     `json:"required"`
}

type UIProfileRules struct {
	InlineSyntax           string `json:"inlineSyntax"`
	SlotOrder              string `json:"slotOrder"`
	ModeSeparator          string `json:"modeSeparator"`
	MissingTrailingSlots   string `json:"missingTrailingSlots"`
	ExtraValues            string `json:"extraValues"`
	DuplicateSlots         string `json:"duplicateSlots"`
	LockBaseline           string `json:"lockBaseline"`
	ExistingSlotsAfterLock string `json:"existingSlotsAfterLock"`
	NewSlotsAfterLock      string `json:"newSlotsAfterLock"`
}

type UIModeDecl struct {
	Name      string   `json:"name"`
	Standard  bool     `json:"standard"`
	Purpose   string   `json:"purpose,omitempty"`
	AppliesTo []string `json:"appliesTo,omitempty"`
	Slots     []string `json:"slots"`
	Position  Position `json:"position"`
}

type ConfigInfo struct {
	LanguageVersion string `json:"languageVersion,omitempty"`
	Target          string `json:"target,omitempty"`
	Source          string `json:"source"`
	Out             string `json:"out"`
	Theme           string `json:"theme,omitempty"`
}

type GeneratedFile struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type Summary struct {
	App      string `json:"app"`
	Entities int    `json:"entities"`
	Pages    int    `json:"pages"`
}

type Program struct {
	App          AppDecl                    `json:"app"`
	Target       *TargetDecl                `json:"target,omitempty"`
	Auth         *AuthDecl                  `json:"auth,omitempty"`
	Database     *DatabaseDecl              `json:"database,omitempty"`
	Security     *SecurityDecl              `json:"security,omitempty"`
	Deploy       *DeployDecl                `json:"deploy,omitempty"`
	Ops          *OpsDecl                   `json:"ops,omitempty"`
	I18N         *I18NDecl                  `json:"i18n,omitempty"`
	Labels       []LabelTranslationDecl     `json:"labels,omitempty"`
	Placeholders []FieldTextTranslationDecl `json:"placeholders,omitempty"`
	HelpTexts    []FieldTextTranslationDecl `json:"helpTexts,omitempty"`
	Messages     []FieldTextTranslationDecl `json:"messages,omitempty"`
	Entities     []EntityDecl               `json:"entities"`
	Migrations   []MigrationDecl            `json:"migrations,omitempty"`
	Seeds        []SeedDecl                 `json:"seeds,omitempty"`
	Tests        []TestDecl                 `json:"tests,omitempty"`
	Queries      []QueryDecl                `json:"queries,omitempty"`
	Jobs         []JobDecl                  `json:"jobs,omitempty"`
	Actions      []CustomActionDecl         `json:"actions,omitempty"`
	Transactions []TransactionDecl          `json:"transactions,omitempty"`
	Services     []ServiceDecl              `json:"services,omitempty"`
	Roles        []RoleDecl                 `json:"roles,omitempty"`
	APIs         []APIDecl                  `json:"apis,omitempty"`
	Layouts      []LayoutDecl               `json:"layouts,omitempty"`
	Pages        []PageDecl                 `json:"pages"`
	Workflows    []WorkflowDecl             `json:"workflows,omitempty"`
	States       []StateDecl                `json:"states,omitempty"`
	Components   []ComponentDecl            `json:"components,omitempty"`
}

type AppDecl struct {
	Name     string   `json:"name"`
	Position Position `json:"position"`
}

type TargetDecl struct {
	Name     string   `json:"name"`
	Frontend string   `json:"frontend,omitempty"`
	Backend  string   `json:"backend,omitempty"`
	Database string   `json:"database,omitempty"`
	Position Position `json:"position"`
}

type EntityDecl struct {
	Name           string                 `json:"name"`
	Fields         []FieldDecl            `json:"fields"`
	ComputedFields []ComputedFieldDecl    `json:"computedFields,omitempty"`
	Indexes        []EntityIndexDecl      `json:"indexes,omitempty"`
	Policies       []EntityPolicyDecl     `json:"policies,omitempty"`
	Validations    []EntityValidationDecl `json:"validations,omitempty"`
	Position       Position               `json:"position"`
}

type AuthDecl struct {
	Strategy string   `json:"strategy,omitempty"`
	Session  string   `json:"session,omitempty"`
	User     UserDecl `json:"user,omitempty"`
	Position Position `json:"position"`
}

type UserDecl struct {
	Fields   []FieldDecl `json:"fields,omitempty"`
	Position Position    `json:"position"`
}

type DatabaseDecl struct {
	URL      EnvRef   `json:"url,omitempty"`
	Position Position `json:"position"`
}

type MigrationDecl struct {
	Name     string                `json:"name"`
	Renames  []MigrationRenameDecl `json:"renames"`
	Position Position              `json:"position"`
}

type MigrationRenameDecl struct {
	Kind     string   `json:"kind"`
	Entity   string   `json:"entity,omitempty"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Position Position `json:"position"`
}

type EnvRef struct {
	Name     string   `json:"name,omitempty"`
	Position Position `json:"position"`
}

type SecurityDecl struct {
	CORS     *CORSDecl `json:"cors,omitempty"`
	Position Position  `json:"position"`
}

type CORSDecl struct {
	Origins     EnvRef   `json:"origins,omitempty"`
	Credentials string   `json:"credentials,omitempty"`
	Position    Position `json:"position"`
}

type DeployDecl struct {
	Target   string              `json:"target,omitempty"`
	Port     *DeployPortDecl     `json:"port,omitempty"`
	Env      []DeployEnvDecl     `json:"env,omitempty"`
	Preview  *DeployPreviewDecl  `json:"preview,omitempty"`
	Rollback *DeployRollbackDecl `json:"rollback,omitempty"`
	Cloud    *DeployCloudDecl    `json:"cloud,omitempty"`
	Position Position            `json:"position"`
}

type DeployPortDecl struct {
	Env      EnvRef   `json:"env,omitempty"`
	Default  string   `json:"default,omitempty"`
	Position Position `json:"position"`
}

type DeployEnvDecl struct {
	Name     string   `json:"name"`
	Mode     string   `json:"mode"`
	Position Position `json:"position"`
}

type DeployPreviewDecl struct {
	Mode     string   `json:"mode"`
	Position Position `json:"position"`
}

type DeployRollbackDecl struct {
	Strategy string   `json:"strategy"`
	Keep     int      `json:"keep"`
	Position Position `json:"position"`
}

type DeployCloudDecl struct {
	Provider string   `json:"provider"`
	App      EnvRef   `json:"app,omitempty"`
	Region   EnvRef   `json:"region,omitempty"`
	Position Position `json:"position"`
}

type OpsDecl struct {
	Health    *OpsEndpointDecl `json:"health,omitempty"`
	Readiness *OpsEndpointDecl `json:"readiness,omitempty"`
	Metrics   *OpsEndpointDecl `json:"metrics,omitempty"`
	Logging   string           `json:"logging,omitempty"`
	Observe   *OpsObserveDecl  `json:"observe,omitempty"`
	Position  Position         `json:"position"`
}

type OpsEndpointDecl struct {
	Path     string   `json:"path"`
	Position Position `json:"position"`
}

type OpsObserveDecl struct {
	Provider string   `json:"provider"`
	Endpoint EnvRef   `json:"endpoint,omitempty"`
	Position Position `json:"position"`
}

type I18NDecl struct {
	Default  string   `json:"default,omitempty"`
	Locales  []string `json:"locales,omitempty"`
	Position Position `json:"position"`
}

type LabelTranslationDecl struct {
	Target       string             `json:"target"`
	Translations []TranslationValue `json:"translations,omitempty"`
	Position     Position           `json:"position"`
}

type FieldTextTranslationDecl struct {
	Target       string             `json:"target"`
	Translations []TranslationValue `json:"translations,omitempty"`
	Position     Position           `json:"position"`
}

type TranslationValue struct {
	Locale   string   `json:"locale"`
	Text     string   `json:"text"`
	Position Position `json:"position"`
}

type RoleDecl struct {
	Name        string           `json:"name"`
	Permissions []PermissionDecl `json:"permissions,omitempty"`
	Position    Position         `json:"position"`
}

type PermissionDecl struct {
	Effect   string   `json:"effect"`
	Action   string   `json:"action"`
	Resource string   `json:"resource,omitempty"`
	Fields   []string `json:"fields,omitempty"`
	Position Position `json:"position"`
}

type APIDecl struct {
	Name     string         `json:"name"`
	Method   string         `json:"method,omitempty"`
	Path     string         `json:"path,omitempty"`
	Queries  []APIParamDecl `json:"queries,omitempty"`
	Params   []APIParamDecl `json:"params,omitempty"`
	Body     []FieldDecl    `json:"body,omitempty"`
	Update   *APIUpdateDecl `json:"update,omitempty"`
	Respond  string         `json:"respond,omitempty"`
	Access   string         `json:"access,omitempty"`
	Webhook  bool           `json:"webhook,omitempty"`
	Position Position       `json:"position"`
}

type APIParamDecl struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
}

type FieldDecl struct {
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Modifiers    []Modifier `json:"modifiers"`
	RelationLoad []string   `json:"relationLoad,omitempty"`
	UI           []UIIntent `json:"ui,omitempty"`
	Position     Position   `json:"position"`
}

type ComputedFieldDecl struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Expression ComputedExpressionDecl `json:"expression"`
	Modifiers  []Modifier             `json:"modifiers,omitempty"`
	Position   Position               `json:"position"`
}

type ComputedExpressionDecl struct {
	Left     string          `json:"left"`
	Operator string          `json:"operator"`
	Right    string          `json:"right"`
	Tree     *ExpressionDecl `json:"tree,omitempty"`
	Position Position        `json:"position"`
}

type ExpressionDecl struct {
	Kind      string          `json:"kind"`
	ValueKind string          `json:"valueKind,omitempty"`
	Value     string          `json:"value,omitempty"`
	Operator  string          `json:"operator,omitempty"`
	Left      *ExpressionDecl `json:"left,omitempty"`
	Right     *ExpressionDecl `json:"right,omitempty"`
	Position  Position        `json:"position"`
}

type EntityIndexDecl struct {
	Fields   []string `json:"fields"`
	Position Position `json:"position"`
}

type EntityPolicyDecl struct {
	Kind     string   `json:"kind"`
	Field    string   `json:"field"`
	Position Position `json:"position"`
}

type EntityValidationDecl struct {
	Left     string                   `json:"left"`
	Operator string                   `json:"operator,omitempty"`
	Right    string                   `json:"right,omitempty"`
	Required bool                     `json:"required,omitempty"`
	When     *ValidationConditionDecl `json:"when,omitempty"`
	Message  string                   `json:"message,omitempty"`
	Position Position                 `json:"position"`
}

type ValidationConditionDecl struct {
	Left     string   `json:"left"`
	Operator string   `json:"operator"`
	Right    string   `json:"right"`
	Position Position `json:"position"`
}

type Modifier struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type PageDecl struct {
	Name          string           `json:"name"`
	Layout        string           `json:"layout,omitempty"`
	Source        string           `json:"source"`
	Query         string           `json:"query,omitempty"`
	QueryPosition Position         `json:"queryPosition,omitempty"`
	View          *PageViewDecl    `json:"view,omitempty"`
	Table         TableDecl        `json:"table"`
	Form          FormDecl         `json:"form"`
	Actions       []string         `json:"actions"`
	ActionUI      []ActionUIIntent `json:"actionUI,omitempty"`
	Access        []string         `json:"access,omitempty"`
	Position      Position         `json:"position"`
}

type PageViewDecl struct {
	Order    []string          `json:"order,omitempty"`
	Compose  *ViewComposeDecl  `json:"compose,omitempty"`
	Sections []ViewSectionDecl `json:"sections,omitempty"`
	Tabs     []ViewTabDecl     `json:"tabs,omitempty"`
	Groups   []ViewGroupDecl   `json:"groups,omitempty"`
	Triggers []ViewTriggerDecl `json:"triggers,omitempty"`
	Position Position          `json:"position"`
}

type ViewComposeDecl struct {
	Mode     string   `json:"mode"`
	Columns  int      `json:"columns,omitempty"`
	Gap      string   `json:"gap,omitempty"`
	StackAt  string   `json:"stackAt,omitempty"`
	Position Position `json:"position"`
}

type ViewSectionDecl struct {
	Name      string   `json:"name"`
	Component string   `json:"component,omitempty"`
	Bind      string   `json:"bind,omitempty"`
	Span      int      `json:"span,omitempty"`
	Display   string   `json:"display,omitempty"`
	Side      string   `json:"side,omitempty"`
	Title     string   `json:"title,omitempty"`
	Position  Position `json:"position"`
}

type ViewTabDecl struct {
	Name     string   `json:"name"`
	Sections []string `json:"sections"`
	Position Position `json:"position"`
}

type ViewGroupDecl struct {
	Name     string           `json:"name"`
	Sections []string         `json:"sections"`
	Compose  *ViewComposeDecl `json:"compose,omitempty"`
	Span     int              `json:"span,omitempty"`
	Title    string           `json:"title,omitempty"`
	Position Position         `json:"position"`
}

type ViewTriggerDecl struct {
	Section  string   `json:"section"`
	Event    string   `json:"event"`
	Position Position `json:"position"`
}

type LayoutDecl struct {
	Name     string      `json:"name"`
	Sidebar  SidebarDecl `json:"sidebar,omitempty"`
	Position Position    `json:"position"`
}

type SidebarDecl struct {
	Items []string `json:"items,omitempty"`
}

type TableDecl struct {
	Columns  []string    `json:"columns"`
	Search   []string    `json:"search"`
	Sort     SortDecl    `json:"sort,omitempty"`
	Paginate int         `json:"paginate,omitempty"`
	Filters  []string    `json:"filters,omitempty"`
	Identity *UIIdentity `json:"identity,omitempty"`
	UI       []UIIntent  `json:"ui,omitempty"`
}

type SortDecl struct {
	Field     string `json:"field,omitempty"`
	Direction string `json:"direction,omitempty"`
}

type FormDecl struct {
	Fields   []string    `json:"fields"`
	Identity *UIIdentity `json:"identity,omitempty"`
	UI       []UIIntent  `json:"ui,omitempty"`
}

type UIIntent struct {
	Mode     string   `json:"mode"`
	Values   []string `json:"values"`
	Position Position `json:"position"`
}

type UIIdentity struct {
	ID       string   `json:"id,omitempty"`
	Classes  []string `json:"classes,omitempty"`
	Position Position `json:"position"`
}

type ActionUIIntent struct {
	Action   string      `json:"action"`
	Identity *UIIdentity `json:"identity,omitempty"`
	UI       []UIIntent  `json:"ui,omitempty"`
	Position Position    `json:"position"`
}

type WorkflowDecl struct {
	Name        string           `json:"name"`
	Source      string           `json:"source"`
	States      []string         `json:"states"`
	Transitions []TransitionDecl `json:"transitions,omitempty"`
	Position    Position         `json:"position"`
}

type TransitionDecl struct {
	Name     string   `json:"name"`
	From     string   `json:"from,omitempty"`
	To       string   `json:"to,omitempty"`
	Allow    []string `json:"allow,omitempty"`
	Position Position `json:"position"`
}

type StateDecl struct {
	Name     string       `json:"name"`
	Fields   []StateField `json:"fields,omitempty"`
	Modals   []StateModal `json:"modals,omitempty"`
	Position Position     `json:"position"`
}

type StateField struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	List     bool     `json:"list,omitempty"`
	Position Position `json:"position"`
}

type StateModal struct {
	Name     string   `json:"name"`
	Default  string   `json:"default"`
	Position Position `json:"position"`
}

type ComponentDecl struct {
	Name     string             `json:"name"`
	Inputs   []ComponentInput   `json:"inputs,omitempty"`
	Variants []ComponentVariant `json:"variants,omitempty"`
	Position Position           `json:"position"`
}

type ComponentInput struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	List     bool     `json:"list,omitempty"`
	Position Position `json:"position"`
}

type ComponentVariant struct {
	Name      string   `json:"name"`
	Condition string   `json:"condition"`
	Position  Position `json:"position"`
}
