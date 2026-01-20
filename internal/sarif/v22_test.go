// internal/sarif/v22_test.go
package sarif

import (
	"encoding/json"
	"testing"
)

func TestSerializeV22_Schema(t *testing.T) {
	report := &Report{
		ToolName:    testToolName,
		ToolInfoURI: "https://example.com",
	}

	data, err := Serialize(report, Version22, false)
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["$schema"] != "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/2.2-prerelease-2024-08-08/sarif-2.2/schema/sarif-2-2.schema.json" {
		t.Errorf("$schema = %v", result["$schema"])
	}
	if result["version"] != "2.2" {
		t.Errorf("version = %v, want %v", result["version"], "2.2")
	}
}

func TestSerializeV22_WithResults(t *testing.T) {
	report := &Report{
		ToolName: "go-test-sarif",
		Rules: []Rule{
			{ID: testRuleID, Description: "Test failure"},
		},
		Results: []Result{
			{
				RuleID:  testRuleID,
				Level:   testLevelError,
				Message: "TestFoo failed",
				Location: &LogicalLocation{
					Module:   testModuleName,
					Function: "TestFoo",
				},
			},
		},
	}

	data, err := Serialize(report, Version22, false)
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0].(map[string]any)
	if res["ruleId"] != testRuleID {
		t.Errorf("ruleId = %v, want %v", res["ruleId"], testRuleID)
	}
}
