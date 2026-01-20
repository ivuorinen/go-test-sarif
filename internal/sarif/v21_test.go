// internal/sarif/v21_test.go
package sarif

import (
	"encoding/json"
	"testing"
)

func TestSerializeV21_Schema(t *testing.T) {
	report := &Report{
		ToolName:    "test-tool",
		ToolInfoURI: "https://example.com",
	}

	data, err := Serialize(report, Version210, false)
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["$schema"] != "https://json.schemastore.org/sarif-2.1.0.json" {
		t.Errorf("$schema = %v, want %v", result["$schema"], "https://json.schemastore.org/sarif-2.1.0.json")
	}
	if result["version"] != "2.1.0" {
		t.Errorf("version = %v, want %v", result["version"], "2.1.0")
	}
}

func TestSerializeV21_WithResults(t *testing.T) {
	report := &Report{
		ToolName:    "go-test-sarif",
		ToolInfoURI: "https://golang.org/cmd/go/",
		Rules: []Rule{
			{ID: "test-failure", Description: "Test failure"},
		},
		Results: []Result{
			{
				RuleID:  "test-failure",
				Level:   "error",
				Message: "TestFoo failed",
			},
		},
	}

	data, err := Serialize(report, Version210, false)
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	runs, ok := result["runs"].([]interface{})
	if !ok || len(runs) != 1 {
		t.Fatalf("expected 1 run, got %v", result["runs"])
	}

	run := runs[0].(map[string]interface{})
	results, ok := run["results"].([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("expected 1 result, got %v", run["results"])
	}

	res := results[0].(map[string]interface{})
	if res["ruleId"] != "test-failure" {
		t.Errorf("ruleId = %v, want %v", res["ruleId"], "test-failure")
	}
	if res["level"] != "error" {
		t.Errorf("level = %v, want %v", res["level"], "error")
	}
}

func TestSerializeV21_LogicalLocation(t *testing.T) {
	report := &Report{
		ToolName: "go-test-sarif",
		Rules: []Rule{
			{ID: "test-failure", Description: "Test failure"},
		},
		Results: []Result{
			{
				RuleID:  "test-failure",
				Level:   "error",
				Message: "TestBar failed",
				Location: &LogicalLocation{
					Module:   "example.com/foo",
					Function: "TestBar",
				},
			},
		},
	}

	data, err := Serialize(report, Version210, false)
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	runs := result["runs"].([]interface{})
	run := runs[0].(map[string]interface{})
	results := run["results"].([]interface{})
	res := results[0].(map[string]interface{})

	locs, ok := res["logicalLocations"].([]interface{})
	if !ok || len(locs) != 1 {
		t.Fatalf("expected 1 logicalLocation, got %v", res["logicalLocations"])
	}

	loc := locs[0].(map[string]interface{})
	if loc["fullyQualifiedName"] != "example.com/foo.TestBar" {
		t.Errorf("fullyQualifiedName = %v, want %v", loc["fullyQualifiedName"], "example.com/foo.TestBar")
	}
}
