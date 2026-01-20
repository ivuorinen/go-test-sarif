// internal/sarif/model_test.go
package sarif

import "testing"

const (
	testToolName   = "test-tool"
	testModuleName = "example.com/foo"
)

func TestReport_Structure(t *testing.T) {
	report := &Report{
		ToolName:    testToolName,
		ToolInfoURI: "https://example.com",
		Rules: []Rule{
			{ID: "rule-1", Description: "Test rule"},
		},
		Results: []Result{
			{
				RuleID:  "rule-1",
				Level:   "error",
				Message: "Test failed",
				Location: &LogicalLocation{
					Module:   testModuleName,
					Function: "TestBar",
				},
			},
		},
	}

	if report.ToolName != testToolName {
		t.Errorf("ToolName = %q, want %q", report.ToolName, testToolName)
	}
	if len(report.Rules) != 1 {
		t.Errorf("len(Rules) = %d, want %d", len(report.Rules), 1)
	}
	if len(report.Results) != 1 {
		t.Errorf("len(Results) = %d, want %d", len(report.Results), 1)
	}
	if report.Results[0].Location.Module != testModuleName {
		t.Errorf("Location.Module = %q, want %q", report.Results[0].Location.Module, testModuleName)
	}
}
