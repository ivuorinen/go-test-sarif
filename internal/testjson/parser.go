// Package testjson provides parsing utilities for go test -json output.
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
	// Time is when the event occurred.
	Time time.Time `json:"Time"`
	// Action is the event type (run, pass, fail, output, etc.).
	Action string `json:"Action"`
	// Package is the Go package being tested.
	Package string `json:"Package"`
	// Test is the name of the specific test, if applicable.
	Test string `json:"Test,omitempty"`
	// Elapsed is the duration in seconds for pass/fail events.
	Elapsed float64 `json:"Elapsed,omitempty"`
	// Output contains any text output from the test.
	Output string `json:"Output,omitempty"`
	// FailedBuild indicates if this was a build failure.
	FailedBuild bool `json:"FailedBuild,omitempty"`
}

// ParseFile reads and parses a go test -json output file.
// Returns an error with line number if any line contains invalid JSON.
func ParseFile(path string) ([]TestEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var events []TestEvent
	scanner := bufio.NewScanner(f)
	// Increase buffer size for large JSON lines (e.g., verbose test output)
	// Default is 64KB; allow up to 4MB per line
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
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
