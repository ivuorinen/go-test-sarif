// Package testjson provides parsing utilities for go test -json output.
package testjson

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testInputFile   = "input.json"
	testPackageName = "example.com/foo"
	testTestName    = "TestBar"
)

func TestParseFile_ValidInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, testInputFile)

	content := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example.com/foo","Test":"TestBar"}
{"Time":"2024-01-15T10:30:01Z","Action":"output","Package":"example.com/foo","Test":"TestBar","Output":"=== RUN   TestBar\n"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example.com/foo","Test":"TestBar","Elapsed":0.5}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o600); err != nil {
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
	if events[0].Package != testPackageName {
		t.Errorf("event[0].Package = %q, want %q", events[0].Package, testPackageName)
	}
	if events[0].Test != testTestName {
		t.Errorf("event[0].Test = %q, want %q", events[0].Test, testTestName)
	}

	// Check elapsed on pass event
	if events[2].Elapsed != 0.5 {
		t.Errorf("event[2].Elapsed = %v, want %v", events[2].Elapsed, 0.5)
	}
}

func TestParseFile_AllFields(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, testInputFile)

	// Event with all fields populated
	content := `{"Time":"2024-01-15T10:30:00Z","Action":"fail","Package":"example.com/foo","Test":"TestBar","Elapsed":1.234,"Output":"FAIL\n","FailedBuild":"example.com/broken"}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o600); err != nil {
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
	if e.Package != testPackageName {
		t.Errorf("Package = %q, want %q", e.Package, testPackageName)
	}
	if e.Test != testTestName {
		t.Errorf("Test = %q, want %q", e.Test, testTestName)
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
	inputPath := filepath.Join(dir, testInputFile)

	content := `{"Action":"pass","Package":"example.com/foo"}
{"Action":"fail","Package":broken json here}
{"Action":"skip","Package":"example.com/bar"}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParseFile(inputPath)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	// Error should mention line 2
	if got := err.Error(); !strings.Contains(got, "line 2") {
		t.Errorf("error = %q, want to contain %q", got, "line 2")
	}
}

func TestParseFile_FileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/to/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
