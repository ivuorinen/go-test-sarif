# Replace go-sarif Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the go-sarif dependency with a minimal internal SARIF implementation, eliminating the yaml.v3 transitive dependency.

**Architecture:** Three-layer design: (1) testjson parser captures all go test -json fields, (2) internal SARIF model is version-agnostic, (3) version-specific serializers output SARIF v2.1.0 or v2.2. Registry pattern enables adding future versions.

**Tech Stack:** Go standard library only (encoding/json, bufio, os, time, sort, bytes, fmt)

**Note:** The design document mentioned "SARIF v3.0" but this specification doesn't exist. We implement v2.1.0 and v2.2 (the actual supported versions).

---

## Task 1: Test JSON Parser - Types and Basic Test

**Files:**
- Create: `internal/testjson/parser.go`
- Create: `internal/testjson/parser_test.go`

**Step 1: Write the failing test for TestEvent struct and ParseFile**

```go
// internal/testjson/parser_test.go
package testjson

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseFile_ValidInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")

	content := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example.com/foo","Test":"TestBar"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example.com/foo","Test":"TestBar","Output":"=== RUN   TestBar\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example.com/foo","Test":"TestBar","Elapsed":0.5}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	events, err := ParseFile(inputPath)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// Check first event
	if events[0].Action != "run" {
		t.Errorf("event[0].Action = %q, want %q", events[0].Action, "run")
	}
	if events[0].Package != "example.com/foo" {
		t.Errorf("event[0].Package = %q, want %q", events[0].Package, "example.com/foo")
	}
	if events[0].Test != "TestBar" {
		t.Errorf("event[0].Test = %q, want %q", events[0].Test, "TestBar")
	}

	// Check elapsed on pass event
	if events[2].Elapsed != 0.5 {
		t.Errorf("event[2].Elapsed = %v, want %v", events[2].Elapsed, 0.5)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/testjson/... -v`
Expected: Build failure - package doesn't exist

**Step 3: Write minimal implementation**

```go
// internal/testjson/parser.go
package testjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TestEvent captures all fields from go test -json output.
type TestEvent struct {
	Time        time.Time `json:"Time"`
	Action      string    `json:"Action"`
	Package     string    `json:"Package"`
	Test        string    `json:"Test,omitempty"`
	Elapsed     float64   `json:"Elapsed,omitempty"`
	Output      string    `json:"Output,omitempty"`
	FailedBuild string    `json:"FailedBuild,omitempty"`
}

// ParseFile reads and parses a go test -json output file.
// Returns an error with line number if any line contains invalid JSON.
func ParseFile(path string) ([]TestEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []TestEvent
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		var event TestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("line %d: invalid JSON: %w", lineNum, err)
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/testjson/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/testjson/
git commit -m "feat(testjson): add parser for go test -json output"
```

---

## Task 2: Test JSON Parser - Additional Tests

**Files:**
- Modify: `internal/testjson/parser_test.go`

**Step 1: Add tests for all fields, malformed JSON, and file not found**

```go
// Add to internal/testjson/parser_test.go

func TestParseFile_AllFields(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")

	// Event with all fields populated
	content := `{"Time":"2024-01-15T10:30:00Z","Action":"fail","Package":"example.com/foo","Test":"TestBar","Elapsed":1.234,"Output":"FAIL\n","FailedBuild":"example.com/broken"}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	events, err := ParseFile(inputPath)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")

	if !e.Time.Equal(expectedTime) {
		t.Errorf("Time = %v, want %v", e.Time, expectedTime)
	}
	if e.Action != "fail" {
		t.Errorf("Action = %q, want %q", e.Action, "fail")
	}
	if e.Package != "example.com/foo" {
		t.Errorf("Package = %q, want %q", e.Package, "example.com/foo")
	}
	if e.Test != "TestBar" {
		t.Errorf("Test = %q, want %q", e.Test, "TestBar")
	}
	if e.Elapsed != 1.234 {
		t.Errorf("Elapsed = %v, want %v", e.Elapsed, 1.234)
	}
	if e.Output != "FAIL\n" {
		t.Errorf("Output = %q, want %q", e.Output, "FAIL\n")
	}
	if e.FailedBuild != "example.com/broken" {
		t.Errorf("FailedBuild = %q, want %q", e.FailedBuild, "example.com/broken")
	}
}

func TestParseFile_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")

	content := `{"Action":"pass","Package":"example.com/foo"}
{"Action":"fail","Package":broken json here}
{"Action":"skip","Package":"example.com/bar"}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParseFile(inputPath)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	// Error should mention line 2
	if got := err.Error(); !contains(got, "line 2") {
		t.Errorf("error = %q, want to contain %q", got, "line 2")
	}
}

func TestParseFile_FileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/to/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/testjson/... -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add internal/testjson/parser_test.go
git commit -m "test(testjson): add tests for all fields, malformed JSON, file not found"
```

---

## Task 3: SARIF Model

**Files:**
- Create: `internal/sarif/model.go`
- Create: `internal/sarif/model_test.go`

**Step 1: Write test for model types**

```go
// internal/sarif/model_test.go
package sarif

import "testing"

func TestReport_Structure(t *testing.T) {
	report := &Report{
		ToolName:    "test-tool",
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
					Module:   "example.com/foo",
					Function: "TestBar",
				},
			},
		},
	}

	if report.ToolName != "test-tool" {
		t.Errorf("ToolName = %q, want %q", report.ToolName, "test-tool")
	}
	if len(report.Rules) != 1 {
		t.Errorf("len(Rules) = %d, want %d", len(report.Rules), 1)
	}
	if len(report.Results) != 1 {
		t.Errorf("len(Results) = %d, want %d", len(report.Results), 1)
	}
	if report.Results[0].Location.Module != "example.com/foo" {
		t.Errorf("Location.Module = %q, want %q", report.Results[0].Location.Module, "example.com/foo")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/sarif/... -v`
Expected: Build failure - package doesn't exist

**Step 3: Write implementation**

```go
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/sarif/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/sarif/
git commit -m "feat(sarif): add internal SARIF data model"
```

---

## Task 4: SARIF Version Registry

**Files:**
- Create: `internal/sarif/version.go`
- Create: `internal/sarif/version_test.go`

**Step 1: Write failing tests for version registry**

```go
// internal/sarif/version_test.go
package sarif

import (
	"strings"
	"testing"
)

func TestSupportedVersions(t *testing.T) {
	versions := SupportedVersions()

	if len(versions) < 2 {
		t.Errorf("expected at least 2 versions, got %d", len(versions))
	}

	// Should contain 2.1.0 and 2.2
	found210 := false
	found22 := false
	for _, v := range versions {
		if v == "2.1.0" {
			found210 = true
		}
		if v == "2.2" {
			found22 = true
		}
	}

	if !found210 {
		t.Error("SupportedVersions should contain 2.1.0")
	}
	if !found22 {
		t.Error("SupportedVersions should contain 2.2")
	}
}

func TestSerialize_UnknownVersion(t *testing.T) {
	report := &Report{ToolName: "test"}

	_, err := Serialize(report, "9.9.9", false)
	if err == nil {
		t.Fatal("expected error for unknown version, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "unsupported")
	}
}

func TestDefaultVersion(t *testing.T) {
	if DefaultVersion != Version210 {
		t.Errorf("DefaultVersion = %q, want %q", DefaultVersion, Version210)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/sarif/... -v`
Expected: Build failure - functions not defined

**Step 3: Write implementation**

```go
// internal/sarif/version.go
package sarif

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// Version represents a SARIF specification version.
type Version string

const (
	// Version210 is SARIF version 2.1.0.
	Version210 Version = "2.1.0"
	// Version22 is SARIF version 2.2.
	Version22 Version = "2.2"
)

// DefaultVersion is the default SARIF version used when not specified.
const DefaultVersion = Version210

// Serializer converts an internal Report to version-specific JSON.
type Serializer func(*Report) ([]byte, error)

var serializers = map[Version]Serializer{}

// Register adds a serializer for a SARIF version.
// Called by version-specific files in their init() functions.
func Register(v Version, s Serializer) {
	serializers[v] = s
}

// Serialize converts a Report to JSON for the specified SARIF version.
func Serialize(r *Report, v Version, pretty bool) ([]byte, error) {
	s, ok := serializers[Version(v)]
	if !ok {
		return nil, fmt.Errorf("unsupported SARIF version: %s", v)
	}

	data, err := s(r)
	if err != nil {
		return nil, err
	}

	if pretty {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	return data, nil
}

// SupportedVersions returns all registered SARIF versions, sorted.
func SupportedVersions() []string {
	versions := make([]string, 0, len(serializers))
	for v := range serializers {
		versions = append(versions, string(v))
	}
	sort.Strings(versions)
	return versions
}
```

**Step 4: Run tests to verify they fail (no serializers registered yet)**

Run: `go test ./internal/sarif/... -v`
Expected: FAIL - no versions registered

**Step 5: Commit partial progress**

```bash
git add internal/sarif/version.go internal/sarif/version_test.go
git commit -m "feat(sarif): add version registry (serializers pending)"
```

---

## Task 5: SARIF v2.1.0 Serializer

**Files:**
- Create: `internal/sarif/v21.go`
- Create: `internal/sarif/v21_test.go`

**Step 1: Write failing tests for v2.1.0 serializer**

```go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/sarif/... -v -run V21`
Expected: FAIL - v2.1.0 serializer not registered

**Step 3: Write implementation**

```go
// internal/sarif/v21.go
package sarif

import "encoding/json"

func init() {
	Register(Version210, serializeV21)
}

// SARIF v2.1.0 JSON structures

type sarifV21 struct {
	Schema  string    `json:"$schema"`
	Version string    `json:"version"`
	Runs    []runV21  `json:"runs"`
}

type runV21 struct {
	Tool    toolV21     `json:"tool"`
	Results []resultV21 `json:"results"`
}

type toolV21 struct {
	Driver driverV21 `json:"driver"`
}

type driverV21 struct {
	Name           string    `json:"name"`
	InformationURI string    `json:"informationUri,omitempty"`
	Rules          []ruleV21 `json:"rules,omitempty"`
}

type ruleV21 struct {
	ID               string     `json:"id"`
	ShortDescription messageV21 `json:"shortDescription,omitempty"`
}

type resultV21 struct {
	RuleID           string              `json:"ruleId"`
	Level            string              `json:"level"`
	Message          messageV21          `json:"message"`
	LogicalLocations []logicalLocationV21 `json:"logicalLocations,omitempty"`
}

type messageV21 struct {
	Text string `json:"text"`
}

type logicalLocationV21 struct {
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
	Kind               string `json:"kind,omitempty"`
}

func serializeV21(r *Report) ([]byte, error) {
	doc := sarifV21{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs:    []runV21{buildRunV21(r)},
	}
	return json.Marshal(doc)
}

func buildRunV21(r *Report) runV21 {
	run := runV21{
		Tool: toolV21{
			Driver: driverV21{
				Name:           r.ToolName,
				InformationURI: r.ToolInfoURI,
			},
		},
		Results: make([]resultV21, 0, len(r.Results)),
	}

	// Add rules
	for _, rule := range r.Rules {
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, ruleV21{
			ID:               rule.ID,
			ShortDescription: messageV21{Text: rule.Description},
		})
	}

	// Add results
	for _, result := range r.Results {
		res := resultV21{
			RuleID:  result.RuleID,
			Level:   result.Level,
			Message: messageV21{Text: result.Message},
		}

		if result.Location != nil {
			fqn := result.Location.Module
			if result.Location.Function != "" {
				fqn += "." + result.Location.Function
			}
			res.LogicalLocations = []logicalLocationV21{
				{
					FullyQualifiedName: fqn,
					Kind:               "function",
				},
			}
		}

		run.Results = append(run.Results, res)
	}

	return run
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/sarif/... -v`
Expected: PASS (v2.1.0 tests pass, version registry tests now have 1 version)

**Step 5: Commit**

```bash
git add internal/sarif/v21.go internal/sarif/v21_test.go
git commit -m "feat(sarif): add SARIF v2.1.0 serializer"
```

---

## Task 6: SARIF v2.2 Serializer

**Files:**
- Create: `internal/sarif/v22.go`
- Create: `internal/sarif/v22_test.go`

**Step 1: Write failing tests for v2.2 serializer**

```go
// internal/sarif/v22_test.go
package sarif

import (
	"encoding/json"
	"testing"
)

func TestSerializeV22_Schema(t *testing.T) {
	report := &Report{
		ToolName:    "test-tool",
		ToolInfoURI: "https://example.com",
	}

	data, err := Serialize(report, Version22, false)
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["$schema"] != "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.2/schema/sarif-2.2.json" {
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
			{ID: "test-failure", Description: "Test failure"},
		},
		Results: []Result{
			{
				RuleID:  "test-failure",
				Level:   "error",
				Message: "TestFoo failed",
				Location: &LogicalLocation{
					Module:   "example.com/foo",
					Function: "TestFoo",
				},
			},
		},
	}

	data, err := Serialize(report, Version22, false)
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

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0].(map[string]interface{})
	if res["ruleId"] != "test-failure" {
		t.Errorf("ruleId = %v, want %v", res["ruleId"], "test-failure")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/sarif/... -v -run V22`
Expected: FAIL - v2.2 serializer not registered

**Step 3: Write implementation**

```go
// internal/sarif/v22.go
package sarif

import "encoding/json"

func init() {
	Register(Version22, serializeV22)
}

// SARIF v2.2 JSON structures
// v2.2 is structurally similar to v2.1.0 with minor additions

type sarifV22 struct {
	Schema  string   `json:"$schema"`
	Version string   `json:"version"`
	Runs    []runV22 `json:"runs"`
}

type runV22 struct {
	Tool    toolV22     `json:"tool"`
	Results []resultV22 `json:"results"`
}

type toolV22 struct {
	Driver driverV22 `json:"driver"`
}

type driverV22 struct {
	Name           string    `json:"name"`
	InformationURI string    `json:"informationUri,omitempty"`
	Rules          []ruleV22 `json:"rules,omitempty"`
}

type ruleV22 struct {
	ID               string     `json:"id"`
	ShortDescription messageV22 `json:"shortDescription,omitempty"`
}

type resultV22 struct {
	RuleID           string               `json:"ruleId"`
	Level            string               `json:"level"`
	Message          messageV22           `json:"message"`
	LogicalLocations []logicalLocationV22 `json:"logicalLocations,omitempty"`
}

type messageV22 struct {
	Text string `json:"text"`
}

type logicalLocationV22 struct {
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
	Kind               string `json:"kind,omitempty"`
}

func serializeV22(r *Report) ([]byte, error) {
	doc := sarifV22{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.2/schema/sarif-2.2.json",
		Version: "2.2",
		Runs:    []runV22{buildRunV22(r)},
	}
	return json.Marshal(doc)
}

func buildRunV22(r *Report) runV22 {
	run := runV22{
		Tool: toolV22{
			Driver: driverV22{
				Name:           r.ToolName,
				InformationURI: r.ToolInfoURI,
			},
		},
		Results: make([]resultV22, 0, len(r.Results)),
	}

	// Add rules
	for _, rule := range r.Rules {
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, ruleV22{
			ID:               rule.ID,
			ShortDescription: messageV22{Text: rule.Description},
		})
	}

	// Add results
	for _, result := range r.Results {
		res := resultV22{
			RuleID:  result.RuleID,
			Level:   result.Level,
			Message: messageV22{Text: result.Message},
		}

		if result.Location != nil {
			fqn := result.Location.Module
			if result.Location.Function != "" {
				fqn += "." + result.Location.Function
			}
			res.LogicalLocations = []logicalLocationV22{
				{
					FullyQualifiedName: fqn,
					Kind:               "function",
				},
			}
		}

		run.Results = append(run.Results, res)
	}

	return run
}
```

**Step 4: Run all SARIF tests to verify they pass**

Run: `go test ./internal/sarif/... -v`
Expected: All PASS (including version registry tests now with 2 versions)

**Step 5: Commit**

```bash
git add internal/sarif/v22.go internal/sarif/v22_test.go
git commit -m "feat(sarif): add SARIF v2.2 serializer"
```

---

## Task 7: Pretty Print Test

**Files:**
- Modify: `internal/sarif/version_test.go`

**Step 1: Add test for pretty printing**

```go
// Add to internal/sarif/version_test.go

func TestSerialize_PrettyOutput(t *testing.T) {
	report := &Report{
		ToolName: "test-tool",
	}

	compact, err := Serialize(report, Version210, false)
	if err != nil {
		t.Fatalf("Serialize compact returned error: %v", err)
	}

	pretty, err := Serialize(report, Version210, true)
	if err != nil {
		t.Fatalf("Serialize pretty returned error: %v", err)
	}

	// Pretty should be longer due to whitespace
	if len(pretty) <= len(compact) {
		t.Errorf("pretty output (%d bytes) should be longer than compact (%d bytes)", len(pretty), len(compact))
	}

	// Pretty should contain newlines and indentation
	if !bytes.Contains(pretty, []byte("\n")) {
		t.Error("pretty output should contain newlines")
	}
	if !bytes.Contains(pretty, []byte("  ")) {
		t.Error("pretty output should contain indentation")
	}
}
```

**Step 2: Add import**

Add `"bytes"` to the imports in version_test.go.

**Step 3: Run test to verify it passes**

Run: `go test ./internal/sarif/... -v -run Pretty`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/sarif/version_test.go
git commit -m "test(sarif): add pretty print test"
```

---

## Task 8: Update Converter

**Files:**
- Modify: `internal/converter.go`
- Modify: `internal/converter_test.go`

**Step 1: Update converter to use new packages**

```go
// internal/converter.go
package internal

import (
	"fmt"
	"os"

	"github.com/ivuorinen/go-test-sarif-action/internal/sarif"
	"github.com/ivuorinen/go-test-sarif-action/internal/testjson"
)

// ConvertOptions configures the conversion behavior.
type ConvertOptions struct {
	SARIFVersion sarif.Version
	Pretty       bool
}

// DefaultConvertOptions returns options with sensible defaults.
func DefaultConvertOptions() ConvertOptions {
	return ConvertOptions{
		SARIFVersion: sarif.DefaultVersion,
		Pretty:       false,
	}
}

// ConvertToSARIF converts Go test JSON events to SARIF format.
func ConvertToSARIF(inputFile, outputFile string, opts ConvertOptions) error {
	// Parse go test JSON
	events, err := testjson.ParseFile(inputFile)
	if err != nil {
		return err
	}

	// Build internal SARIF model
	report := buildReport(events)

	// Serialize to requested version
	data, err := sarif.Serialize(report, opts.SARIFVersion, opts.Pretty)
	if err != nil {
		return err
	}

	// Write output
	if err := os.WriteFile(outputFile, data, 0o644); err != nil {
		return err
	}

	fmt.Printf("SARIF report generated: %s\n", outputFile)
	return nil
}

func buildReport(events []testjson.TestEvent) *sarif.Report {
	report := &sarif.Report{
		ToolName:    "go-test-sarif",
		ToolInfoURI: "https://golang.org/cmd/go/#hdr-Test_packages",
		Rules: []sarif.Rule{{
			ID:          "go-test-failure",
			Description: "go test failure",
		}},
	}

	for _, e := range events {
		if e.Action == "fail" && (e.Test != "" || e.Package != "") {
			result := sarif.Result{
				RuleID:  "go-test-failure",
				Level:   "error",
				Message: e.Output,
			}
			if e.Package != "" || e.Test != "" {
				result.Location = &sarif.LogicalLocation{
					Module:   e.Package,
					Function: e.Test,
				}
			}
			report.Results = append(report.Results, result)
		}
	}

	return report
}
```

**Step 2: Update existing tests**

```go
// internal/converter_test.go
package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/go-test-sarif-action/internal/sarif"
)

func TestConvertToSARIF_Success(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "input.json")
	inputContent := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif/internal","Test":"TestExample","Output":"Test failed"}` + "\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(dir, "output.sarif")

	opts := DefaultConvertOptions()
	if err := ConvertToSARIF(inputPath, outputPath, opts); err != nil {
		t.Errorf("ConvertToSARIF returned an error: %v", err)
	}

	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read SARIF output file: %v", err)
	}

	if len(outputContent) == 0 {
		t.Errorf("SARIF output is empty")
	}
}

func TestConvertToSARIF_InvalidInput(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "invalid.json")
	inputContent := `{"Action":"fail","Package":"example.com","Output":` +
		`Test failed}` + "\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(dir, "output.sarif")

	opts := DefaultConvertOptions()
	if err := ConvertToSARIF(inputPath, outputPath, opts); err == nil {
		t.Errorf("Expected an error for invalid JSON input, but got none")
	}
}

func TestConvertToSARIF_FileNotFound(t *testing.T) {
	inputFile := "non_existent_file.json"

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "output.sarif")

	opts := DefaultConvertOptions()
	if err := ConvertToSARIF(inputFile, outputPath, opts); err == nil {
		t.Errorf("Expected an error for non-existent input file, but got none")
	}
}

func TestConvertToSARIF_PackageFailure(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "input.json")
	inputContent := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif-action","Output":"FAIL"}` + "\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(dir, "output.sarif")

	opts := DefaultConvertOptions()
	if err := ConvertToSARIF(inputPath, outputPath, opts); err != nil {
		t.Errorf("ConvertToSARIF returned an error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read SARIF output file: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("SARIF output is empty")
	}
}

func TestConvertToSARIF_Options(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "input.json")
	inputContent := `{"Action":"fail","Package":"example.com/foo","Test":"TestBar","Output":"failed"}` + "\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	tests := []struct {
		name    string
		opts    ConvertOptions
		wantErr bool
	}{
		{
			name:    "default options",
			opts:    DefaultConvertOptions(),
			wantErr: false,
		},
		{
			name: "v2.1.0 pretty",
			opts: ConvertOptions{
				SARIFVersion: sarif.Version210,
				Pretty:       true,
			},
			wantErr: false,
		},
		{
			name: "v2.2",
			opts: ConvertOptions{
				SARIFVersion: sarif.Version22,
				Pretty:       false,
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			opts: ConvertOptions{
				SARIFVersion: "9.9.9",
				Pretty:       false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(dir, tt.name+".sarif")
			err := ConvertToSARIF(inputPath, outputPath, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToSARIF() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 3: Run tests to verify they pass**

Run: `go test ./internal/... -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add internal/converter.go internal/converter_test.go
git commit -m "refactor(converter): use internal sarif and testjson packages"
```

---

## Task 9: Update CLI

**Files:**
- Modify: `cmd/main.go`
- Modify: `cmd/main_test.go`

**Step 1: Update main.go with new flags**

```go
// cmd/main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ivuorinen/go-test-sarif-action/internal"
	"github.com/ivuorinen/go-test-sarif-action/internal/sarif"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "go-test-sarif %s\n", version)
	_, _ = fmt.Fprintf(w, "  commit: %s\n", commit)
	_, _ = fmt.Fprintf(w, "  built at: %s\n", date)
	_, _ = fmt.Fprintf(w, "  built by: %s\n", builtBy)
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: go-test-sarif [options] <input.json> <output.sarif>")
	_, _ = fmt.Fprintln(w, "       go-test-sarif --version")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Options:")
	_, _ = fmt.Fprintf(w, "  --sarif-version string   SARIF version (%s) (default %q)\n",
		strings.Join(sarif.SupportedVersions(), ", "), sarif.DefaultVersion)
	_, _ = fmt.Fprintln(w, "  --pretty                 Pretty-print JSON output")
	_, _ = fmt.Fprintln(w, "  -v, --version            Display version information")
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("go-test-sarif", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		versionFlag  bool
		sarifVersion string
		prettyOutput bool
	)

	fs.BoolVar(&versionFlag, "version", false, "Display version information")
	fs.BoolVar(&versionFlag, "v", false, "Display version information (short)")
	fs.StringVar(&sarifVersion, "sarif-version", string(sarif.DefaultVersion),
		fmt.Sprintf("SARIF version (%s)", strings.Join(sarif.SupportedVersions(), ", ")))
	fs.BoolVar(&prettyOutput, "pretty", false, "Pretty-print JSON output")

	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	if versionFlag {
		printVersion(stdout)
		return 0
	}

	if fs.NArg() < 2 {
		printUsage(stderr)
		return 1
	}

	inputFile := fs.Arg(0)
	outputFile := fs.Arg(1)

	opts := internal.ConvertOptions{
		SARIFVersion: sarif.Version(sarifVersion),
		Pretty:       prettyOutput,
	}

	if err := internal.ConvertToSARIF(inputFile, outputFile, opts); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
```

**Step 2: Update main_test.go**

```go
// cmd/main_test.go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		setupFunc  func() (string, string, func())
		wantExit   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "version flag long",
			args:       []string{"go-test-sarif", "--version"},
			wantExit:   0,
			wantStdout: "go-test-sarif dev",
		},
		{
			name:       "version flag short",
			args:       []string{"go-test-sarif", "-v"},
			wantExit:   0,
			wantStdout: "go-test-sarif dev",
		},
		{
			name:       "missing arguments",
			args:       []string{"go-test-sarif"},
			wantExit:   1,
			wantStderr: "Usage: go-test-sarif",
		},
		{
			name:       "only one argument",
			args:       []string{"go-test-sarif", "input.json"},
			wantExit:   1,
			wantStderr: "Usage: go-test-sarif",
		},
		{
			name:      "valid conversion",
			args:      []string{"go-test-sarif", "input.json", "output.sarif"},
			setupFunc: setupValidTestFiles,
			wantExit:  0,
		},
		{
			name:       "invalid input file",
			args:       []string{"go-test-sarif", "nonexistent.json", "output.sarif"},
			wantExit:   1,
			wantStderr: "Error:",
		},
		{
			name:       "invalid flag",
			args:       []string{"go-test-sarif", "--invalid"},
			wantExit:   1,
			wantStderr: "flag provided but not defined",
		},
		{
			name:      "with sarif-version flag",
			args:      []string{"go-test-sarif", "--sarif-version", "2.2", "input.json", "output.sarif"},
			setupFunc: setupValidTestFiles,
			wantExit:  0,
		},
		{
			name:      "with pretty flag",
			args:      []string{"go-test-sarif", "--pretty", "input.json", "output.sarif"},
			setupFunc: setupValidTestFiles,
			wantExit:  0,
		},
		{
			name:       "invalid sarif version",
			args:       []string{"go-test-sarif", "--sarif-version", "9.9.9", "input.json", "output.sarif"},
			setupFunc:  setupValidTestFiles,
			wantExit:   1,
			wantStderr: "Error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			args := make([]string, len(tt.args))
			copy(args, tt.args)

			if tt.setupFunc != nil {
				inputFile, outputFile, cleanupFunc := tt.setupFunc()
				cleanup = cleanupFunc
				for i, arg := range args {
					switch arg {
					case "input.json":
						args[i] = inputFile
					case "output.sarif":
						args[i] = outputFile
					}
				}
			}
			if cleanup != nil {
				defer cleanup()
			}

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := run(args, stdout, stderr)

			if exitCode != tt.wantExit {
				t.Errorf("run() exit code = %v, want %v\nstderr: %s", exitCode, tt.wantExit, stderr.String())
			}

			if tt.wantStdout != "" && !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want to contain %q", stdout.String(), tt.wantStdout)
			}

			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	buf := &bytes.Buffer{}
	printVersion(buf)

	output := buf.String()
	if !strings.Contains(output, "go-test-sarif dev") {
		t.Errorf("printVersion() = %q, want to contain %q", output, "go-test-sarif dev")
	}
	if !strings.Contains(output, "commit: none") {
		t.Errorf("printVersion() = %q, want to contain %q", output, "commit: none")
	}
}

func TestPrintUsage(t *testing.T) {
	buf := &bytes.Buffer{}
	printUsage(buf)

	output := buf.String()
	if !strings.Contains(output, "Usage: go-test-sarif") {
		t.Errorf("printUsage() = %q, want to contain usage information", output)
	}
	if !strings.Contains(output, "--sarif-version") {
		t.Errorf("printUsage() = %q, want to contain --sarif-version flag", output)
	}
	if !strings.Contains(output, "--pretty") {
		t.Errorf("printUsage() = %q, want to contain --pretty flag", output)
	}
}

func setupValidTestFiles() (string, string, func()) {
	tmpDir, err := os.MkdirTemp("", "go-test-sarif-test")
	if err != nil {
		panic(err)
	}

	inputFile := filepath.Join(tmpDir, "test-input.json")
	outputFile := filepath.Join(tmpDir, "test-output.sarif")

	testJSON := `{"Time":"2023-01-01T00:00:00Z","Action":"pass","Package":"example.com/test","Test":"TestExample","Elapsed":0.1}`
	if err := os.WriteFile(inputFile, []byte(testJSON), 0o644); err != nil {
		panic(err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return inputFile, outputFile, cleanup
}
```

**Step 3: Run all tests to verify they pass**

Run: `go test ./... -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add cmd/main.go cmd/main_test.go
git commit -m "feat(cli): add --sarif-version and --pretty flags"
```

---

## Task 10: Remove go-sarif Dependency

**Files:**
- Modify: `go.mod`

**Step 1: Run go mod tidy to clean up dependencies**

Run: `go mod tidy`

**Step 2: Verify go-sarif is removed**

Run: `go mod graph | grep sarif`
Expected: No output (go-sarif removed)

**Step 3: Verify yaml.v3 is removed**

Run: `go mod graph | grep yaml`
Expected: No output (yaml.v3 removed)

**Step 4: Run full test suite**

Run: `go test ./... -v`
Expected: All PASS

**Step 5: Verify build works**

Run: `go build ./cmd/...`
Expected: Success

**Step 6: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: remove go-sarif dependency

Replaced external go-sarif library with internal implementation.
This eliminates the transitive dependency on gopkg.in/yaml.v3."
```

---

## Task 11: Final Verification

**Step 1: Verify no external dependencies remain**

Run: `go list -m all | wc -l`
Expected: 1 (only the main module)

**Step 2: Run tests with race detector**

Run: `go test -race ./...`
Expected: PASS

**Step 3: Verify binary works end-to-end**

```bash
# Build
go build -o go-test-sarif ./cmd/

# Create test input
echo '{"Action":"fail","Package":"example.com/test","Test":"TestFail","Output":"assertion failed"}' > /tmp/test.json

# Run with default version
./go-test-sarif /tmp/test.json /tmp/output-v21.sarif

# Run with v2.2
./go-test-sarif --sarif-version 2.2 /tmp/test.json /tmp/output-v22.sarif

# Run with pretty print
./go-test-sarif --pretty /tmp/test.json /tmp/output-pretty.sarif

# Verify outputs exist and contain expected content
cat /tmp/output-v21.sarif | head -5
cat /tmp/output-v22.sarif | head -5
cat /tmp/output-pretty.sarif | head -10

# Cleanup
rm -f go-test-sarif /tmp/test.json /tmp/output-*.sarif
```

**Step 4: Update design document with correction**

Note that we implemented SARIF v2.1.0 and v2.2 (not v3.0 as originally mentioned, since SARIF 3.0 doesn't exist).

---

## Summary

After completing all tasks:

- **Zero external dependencies** - Only Go standard library
- **No yaml.v3** in dependency graph
- **SARIF v2.1.0 and v2.2** support with extensible registry
- **New CLI flags**: `--sarif-version`, `--pretty`
- **Logical locations** included in results (package + test name)
- **All fields captured** from `go test -json` for future use
