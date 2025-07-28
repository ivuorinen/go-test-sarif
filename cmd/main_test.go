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
			name:     "valid conversion",
			args:     []string{"go-test-sarif", "input.json", "output.sarif"},
			setupFunc: setupValidTestFiles,
			wantExit: 0,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setupFunc != nil {
				inputFile, outputFile, cleanupFunc := tt.setupFunc()
				cleanup = cleanupFunc
				// Replace placeholders with actual file paths
				for i, arg := range tt.args {
					switch arg {
					case "input.json":
						tt.args[i] = inputFile
					case "output.sarif":
						tt.args[i] = outputFile
					}
				}
			}
			if cleanup != nil {
				defer cleanup()
			}

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := run(tt.args, stdout, stderr)

			if exitCode != tt.wantExit {
				t.Errorf("run() exit code = %v, want %v", exitCode, tt.wantExit)
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
	if !strings.Contains(output, "Usage: go-test-sarif <input.json> <output.sarif>") {
		t.Errorf("printUsage() = %q, want to contain usage information", output)
	}
}

func setupValidTestFiles() (string, string, func()) {
	tmpDir, err := os.MkdirTemp("", "go-test-sarif-test")
	if err != nil {
		panic(err)
	}

	inputFile := filepath.Join(tmpDir, "test-input.json")
	outputFile := filepath.Join(tmpDir, "test-output.sarif")

	// Create a valid test JSON file
	testJSON := `{"Time":"2023-01-01T00:00:00Z","Action":"pass","Package":"example.com/test","Test":"TestExample","Elapsed":0.1}`
	if err := os.WriteFile(inputFile, []byte(testJSON), 0644); err != nil {
		panic(err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return inputFile, outputFile, cleanup
}
