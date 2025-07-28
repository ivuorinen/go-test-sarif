// Package internal provides the SARIF conversion utilities.
package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/owenrumney/go-sarif/v2/sarif"
)

// TestEvent represents a single line of `go test -json` output.
type TestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test,omitempty"`
	Output  string `json:"Output,omitempty"`
}

// ConvertToSARIF converts Go test JSON events to the SARIF format.
func ConvertToSARIF(inputFile, outputFile string) error {
	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	report, err := sarif.New(sarif.Version210)
	if err != nil {
		return err
	}

	run := sarif.NewRunWithInformationURI("go-test-sarif", "https://golang.org/cmd/go/#hdr-Test_packages")
	rule := run.AddRule("go-test-failure").WithDescription("go test failure")

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event TestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		if event.Action == "fail" && (event.Test != "" || event.Package != "") {
			result := sarif.NewRuleResult(rule.ID).
				WithLevel("error").
				WithMessage(sarif.NewTextMessage(event.Output))
			run.AddResult(result)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	report.AddRun(run)
	if err := report.WriteFile(outputFile); err != nil {
		return err
	}

	fmt.Printf("SARIF report generated: %s\n", outputFile)
	return nil
}
