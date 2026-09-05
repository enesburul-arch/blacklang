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
	App        AppDecl                `json:"app"`
	Auth       *AuthDecl              `json:"auth,omitempty"`
	Database   *DatabaseDecl          `json:"database,omitempty"`
	Security   *SecurityDecl          `json:"security,omitempty"`
	Deploy     *DeployDecl            `json:"deploy,omitempty"`
	I18N       *I18NDecl              `json:"i18n,omitempty"`
	Labels     []LabelTranslationDecl `json:"labels,omitempty"`
	Entities   []EntityDecl           `json:"entities"`
	Roles      []RoleDecl             `json:"roles,omitempty"`
	APIs       []APIDecl              `json:"apis,omitempty"`
	Layouts    []LayoutDecl           `json:"layouts,omitempty"`
	Pages      []PageDecl             `json:"pages"`
	Workflows  []WorkflowDecl         `json:"workflows,omitempty"`
	States     []StateDecl            `json:"states,omitempty"`
	Components []ComponentDecl        `json:"components,omitempty"`
}

type AppDecl struct {
	Name     string   `json:"name"`
	Position Position `json:"position"`
}

type EntityDecl struct {
	Name        string                 `json:"name"`
	Fields      []FieldDecl            `json:"fields"`
	Validations []EntityValidationDecl `json:"validations,omitempty"`
	Position    Position               `json:"position"`
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
	Target   string          `json:"target,omitempty"`
	Port     *DeployPortDecl `json:"port,omitempty"`
	Env      []DeployEnvDecl `json:"env,omitempty"`
	Position Position        `json:"position"`
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
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Modifiers []Modifier `json:"modifiers"`
	UI        []UIIntent `json:"ui,omitempty"`
	Position  Position   `json:"position"`
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
	Name     string           `json:"name"`
	Layout   string           `json:"layout,omitempty"`
	Source   string           `json:"source"`
	Table    TableDecl        `json:"table"`
	Form     FormDecl         `json:"form"`
	Actions  []string         `json:"actions"`
	ActionUI []ActionUIIntent `json:"actionUI,omitempty"`
	Access   []string         `json:"access,omitempty"`
	Position Position         `json:"position"`
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
