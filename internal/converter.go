// Package internal provides the SARIF conversion utilities.
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
