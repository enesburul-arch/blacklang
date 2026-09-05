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

type DocEntry struct {
	Keyword    string   `json:"keyword"`
	Purpose    string   `json:"purpose"`
	Syntax     string   `json:"syntax"`
	Example    string   `json:"example"`
	AgentNotes []string `json:"agentNotes"`
	Errors     []string `json:"errors"`
}

type ConfigInfo struct {
	LanguageVersion string `json:"languageVersion,omitempty"`
	Target          string `json:"target,omitempty"`
	Source          string `json:"source"`
	Out             string `json:"out"`
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
	App        AppDecl         `json:"app"`
	Auth       *AuthDecl       `json:"auth,omitempty"`
	Database   *DatabaseDecl   `json:"database,omitempty"`
	Entities   []EntityDecl    `json:"entities"`
	Roles      []RoleDecl      `json:"roles,omitempty"`
	APIs       []APIDecl       `json:"apis,omitempty"`
	Layouts    []LayoutDecl    `json:"layouts,omitempty"`
	Pages      []PageDecl      `json:"pages"`
	Workflows  []WorkflowDecl  `json:"workflows,omitempty"`
	States     []StateDecl     `json:"states,omitempty"`
	Components []ComponentDecl `json:"components,omitempty"`
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
	Name     string    `json:"name"`
	Layout   string    `json:"layout,omitempty"`
	Source   string    `json:"source"`
	Table    TableDecl `json:"table"`
	Form     FormDecl  `json:"form"`
	Actions  []string  `json:"actions"`
	Access   []string  `json:"access,omitempty"`
	Position Position  `json:"position"`
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
	Columns  []string `json:"columns"`
	Search   []string `json:"search"`
	Sort     SortDecl `json:"sort,omitempty"`
	Paginate int      `json:"paginate,omitempty"`
	Filters  []string `json:"filters,omitempty"`
}

type SortDecl struct {
	Field     string `json:"field,omitempty"`
	Direction string `json:"direction,omitempty"`
}

type FormDecl struct {
	Fields []string `json:"fields"`
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
