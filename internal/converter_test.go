package internal

import (
	"os"
	"testing"
)

// TestConvertToSARIF_Success tests the successful conversion of a valid Go test JSON output to SARIF format.
func TestConvertToSARIF_Success(t *testing.T) {
	// Create a temporary JSON input file with valid test data
	inputFile, err := os.CreateTemp("", "test_input_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp input file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove temp input file: %v", err)
		}
	}(inputFile.Name())

	inputContent := `[{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif/internal","Output":"Test failed"}]`
	if _, err := inputFile.WriteString(inputContent); err != nil {
		t.Fatalf("Failed to write to temp input file: %v", err)
	}

	// Create a temporary SARIF output file
	outputFile, err := os.CreateTemp("", "test_output_*.sarif")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove temp output file: %v", err)
		}
	}(outputFile.Name())

	// Run the ConvertToSARIF function
	err = ConvertToSARIF(inputFile.Name(), outputFile.Name())
	if err != nil {
		t.Errorf("ConvertToSARIF returned an error: %v", err)
	}

	// Read and validate the SARIF output
	outputContent, err := os.ReadFile(outputFile.Name())
	if err != nil {
		t.Fatalf("Failed to read SARIF output file: %v", err)
	}

	// Perform basic validation on the SARIF output
	if len(outputContent) == 0 {
		t.Errorf("SARIF output is empty")
	}

	// Additional validations can be added here to verify the correctness of the SARIF content
}

// TestConvertToSARIF_InvalidInput tests the function's behavior when provided with invalid JSON input.
func TestConvertToSARIF_InvalidInput(t *testing.T) {
	// Create a temporary JSON input file with invalid test data
	inputFile, err := os.CreateTemp("", "test_input_invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp input file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove temp input file: %v", err)
		}
	}(inputFile.Name())

	inputContent := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif/internal","Output":Test failed}` // Missing quotes around 'Test failed'
	if _, err := inputFile.WriteString(inputContent); err != nil {
		t.Fatalf("Failed to write to temp input file: %v", err)
	}

	// Create a temporary SARIF output file
	outputFile, err := os.CreateTemp("", "test_output_invalid_*.sarif")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove temp output file: %v", err)
		}
	}(outputFile.Name())

	// Run the ConvertToSARIF function
	err = ConvertToSARIF(inputFile.Name(), outputFile.Name())
	if err == nil {
		t.Errorf("Expected an error for invalid JSON input, but got none")
	}
}

// TestConvertToSARIF_FileNotFound tests the function's behavior when the input file does not exist.
func TestConvertToSARIF_FileNotFound(t *testing.T) {
	// Define a non-existent input file path
	inputFile := "non_existent_file.json"

	// Create a temporary SARIF output file
	outputFile, err := os.CreateTemp("", "test_output_notfound_*.sarif")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove temp output file: %v", err)
		}
	}(outputFile.Name())

	// Run the ConvertToSARIF function
	err = ConvertToSARIF(inputFile, outputFile.Name())
	if err == nil {
		t.Errorf("Expected an error for non-existent input file, but got none")
	}
}
