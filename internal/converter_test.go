package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/go-test-sarif-action/internal/sarif"
)

const testInputFile = "input.json"

// testConvertHelper sets up input/output files and runs conversion
func testConvertHelper(t *testing.T, inputJSON string, opts ConvertOptions) ([]byte, error) {
	t.Helper()
	dir := t.TempDir()

	inputPath := filepath.Join(dir, testInputFile)
	if err := os.WriteFile(inputPath, []byte(inputJSON), 0o600); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(dir, "output.sarif")
	if err := ConvertToSARIF(inputPath, outputPath, opts); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	return data, nil
}

func TestConvertToSARIF_Success(t *testing.T) {
	inputJSON := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif/internal","Test":"TestExample","Output":"Test failed"}` + "\n"

	data, err := testConvertHelper(t, inputJSON, DefaultConvertOptions())
	if err != nil {
		t.Errorf("ConvertToSARIF returned an error: %v", err)
	}
	if len(data) == 0 {
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
	inputJSON := `{"Action":"fail","Package":"github.com/ivuorinen/go-test-sarif-action","Output":"FAIL"}` + "\n"

	data, err := testConvertHelper(t, inputJSON, DefaultConvertOptions())
	if err != nil {
		t.Errorf("ConvertToSARIF returned an error: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("SARIF output is empty")
	}
}

func TestConvertToSARIF_Options(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, testInputFile)
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
