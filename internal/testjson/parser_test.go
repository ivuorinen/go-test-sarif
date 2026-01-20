// Package testjson provides parsing utilities for go test -json output.
package testjson

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_ValidInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")

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
