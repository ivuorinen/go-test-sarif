package internal

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConvertToSARIF_Success tests the successful conversion of a valid Go test JSON output to SARIF format.
func TestConvertToSARIF_Success(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "input.json")
	inputContent := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif/internal","Test":"TestExample","Output":"Test failed"}` + "\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(dir, "output.sarif")

	if err := ConvertToSARIF(inputPath, outputPath); err != nil {
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

// TestConvertToSARIF_InvalidInput tests the function's behavior when provided with invalid JSON input.
func TestConvertToSARIF_InvalidInput(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "invalid.json")
	inputContent := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif/internal","Test":"TestExample","Output":` +
		`Test failed}` + "\n" // Missing quotes around 'Test failed'
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(dir, "output.sarif")

	if err := ConvertToSARIF(inputPath, outputPath); err == nil {
		t.Errorf("Expected an error for invalid JSON input, but got none")
	}
}

// TestConvertToSARIF_FileNotFound tests the function's behavior when the input file does not exist.
func TestConvertToSARIF_FileNotFound(t *testing.T) {
	inputFile := "non_existent_file.json"

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "output.sarif")

	if err := ConvertToSARIF(inputFile, outputPath); err == nil {
		t.Errorf("Expected an error for non-existent input file, but got none")
	}
}

// TestConvertToSARIF_PackageFailure ensures package-level failures are included in the SARIF output.
func TestConvertToSARIF_PackageFailure(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "input.json")
	inputContent := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif-action","Output":"FAIL"}` + "\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(dir, "output.sarif")

	if err := ConvertToSARIF(inputPath, outputPath); err != nil {
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
