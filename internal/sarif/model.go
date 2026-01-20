// Package sarif provides SARIF report generation.
package sarif

// Report is the internal version-agnostic representation of a SARIF report.
type Report struct {
	// ToolName is the name of the tool that produced the results.
	ToolName string
	// ToolInfoURI is a URL for more information about the tool.
	ToolInfoURI string
	// Rules contains the rule definitions referenced by results.
	Rules []Rule
	// Results contains the actual findings/test failures.
	Results []Result
}

// Rule defines a rule that can be violated.
type Rule struct {
	// ID is the unique identifier for this rule.
	ID string
	// Description explains what this rule checks.
	Description string
}

// Result represents a single finding.
type Result struct {
	// RuleID references the rule that produced this result.
	RuleID string
	// Level indicates the severity (error, warning, note).
	Level string
	// Message describes the specific issue found.
	Message string
	// Location identifies where the issue was found.
	Location *LogicalLocation
}

// LogicalLocation identifies where an issue occurred without file coordinates.
type LogicalLocation struct {
	// Module is the Go module or package path.
	Module string
	// Function is the name of the function or test.
	Function string
}
