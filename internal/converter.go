package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// TestResult represents a single test result from 'go test -json' output.
type TestResult struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Output  string `json:"Output"`
}

// ConvertToSARIF converts Go test JSON results to SARIF format.
func ConvertToSARIF(inputFile, outputFile string) error {
	// Read the input file
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse the JSON data
	var testResults []TestResult
	if err := json.Unmarshal(data, &testResults); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Convert test results to SARIF format
	sarifData := map[string]interface{}{
		"version": "2.1.0",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":    "go-test-sarif",
						"version": "1.0.0",
					},
				},
				"results": convertResults(testResults),
			},
		},
	}

	// Marshal SARIF data to JSON
	sarifJSON, err := json.MarshalIndent(sarifData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SARIF data: %w", err)
	}

	// Write the SARIF JSON to the output file
	if err := ioutil.WriteFile(outputFile, sarifJSON, 0644); err != nil {
		return fmt.Errorf("failed to write SARIF output file: %w", err)
	}

	fmt.Printf("SARIF report generated: %s\n", outputFile)
	return nil
}

// convertResults transforms test results into SARIF result objects.
func convertResults(testResults []TestResult) []map[string]interface{} {
	var results []map[string]interface{}
	for _, tr := range testResults {
		if tr.Action == "fail" {
			results = append(results, map[string]interface{}{
				"ruleId":   "go-test-failure",
				"message":  map[string]string{"text": tr.Output},
				"level":    "error",
				"locations": []map[string]interface{}{},
			})
		}
	}
	return results
}
