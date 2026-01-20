// internal/sarif/model.go
package sarif

// Report is the internal version-agnostic representation of a SARIF report.
type Report struct {
	ToolName    string
	ToolInfoURI string
	Rules       []Rule
	Results     []Result
}

// Rule defines a rule that can be violated.
type Rule struct {
	ID          string
	Description string
}

// Result represents a single finding.
type Result struct {
	RuleID   string
	Level    string // "error", "warning", "note"
	Message  string
	Location *LogicalLocation
}

// LogicalLocation identifies where an issue occurred without file coordinates.
type LogicalLocation struct {
	Module   string // Package name (e.g., "github.com/foo/bar")
	Function string // Test or function name (e.g., "TestExample")
}
