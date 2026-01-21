package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/go-test-sarif-action/internal/testutil"
)

type runTestCase struct {
	name       string
	args       []string
	setupFunc  func() (string, string, func())
	wantExit   int
	wantStdout string
	wantStderr string
}

// runTestCaseHelper executes a single test case and validates results.
func runTestCaseHelper(t *testing.T, tc runTestCase) {
	t.Helper()

	args := make([]string, len(tc.args))
	copy(args, tc.args)

	var cleanup func()
	if tc.setupFunc != nil {
		inputFile, outputFile, cleanupFunc := tc.setupFunc()
		cleanup = cleanupFunc
		args = replaceFilePlaceholders(args, inputFile, outputFile)
	}
	if cleanup != nil {
		defer cleanup()
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run(args, stdout, stderr)

	if exitCode != tc.wantExit {
		t.Errorf("exit code = %v, want %v", exitCode, tc.wantExit)
	}
	if tc.wantStdout != "" && !strings.Contains(stdout.String(), tc.wantStdout) {
		t.Errorf("stdout = %q, want to contain %q", stdout.String(), tc.wantStdout)
	}
	if tc.wantStderr != "" && !strings.Contains(stderr.String(), tc.wantStderr) {
		t.Errorf("stderr = %q, want to contain %q", stderr.String(), tc.wantStderr)
	}
}

// replaceFilePlaceholders replaces placeholder file names with actual paths.
func replaceFilePlaceholders(args []string, inputFile, outputFile string) []string {
	for i, arg := range args {
		switch arg {
		case testutil.InputJSON:
			args[i] = inputFile
		case testutil.OutputSARIF:
			args[i] = outputFile
		}
	}
	return args
}

func TestRun(t *testing.T) {
	tests := []runTestCase{
		{
			name:       "version flag long",
			args:       []string{testutil.AppName, "--version"},
			wantExit:   0,
			wantStdout: testutil.VersionOutput,
		},
		{
			name:       "version flag short",
			args:       []string{testutil.AppName, "-v"},
			wantExit:   0,
			wantStdout: testutil.VersionOutput,
		},
		{
			name:       "missing arguments",
			args:       []string{testutil.AppName},
			wantExit:   1,
			wantStderr: "Usage: " + testutil.AppName,
		},
		{
			name:       "only one argument",
			args:       []string{testutil.AppName, testutil.InputJSON},
			wantExit:   1,
			wantStderr: "Usage: " + testutil.AppName,
		},
		{
			name:      "valid conversion",
			args:      []string{testutil.AppName, testutil.InputJSON, testutil.OutputSARIF},
			setupFunc: setupValidTestFiles,
			wantExit:  0,
		},
		{
			name:       "invalid input file",
			args:       []string{testutil.AppName, "nonexistent.json", testutil.OutputSARIF},
			wantExit:   1,
			wantStderr: "Error:",
		},
		{
			name:       "invalid flag",
			args:       []string{testutil.AppName, "--invalid"},
			wantExit:   1,
			wantStderr: "flag provided but not defined",
		},
		{
			name:      "with sarif-version flag",
			args:      []string{testutil.AppName, "--sarif-version", "2.2", testutil.InputJSON, testutil.OutputSARIF},
			setupFunc: setupValidTestFiles,
			wantExit:  0,
		},
		{
			name:      "with pretty flag",
			args:      []string{testutil.AppName, "--pretty", testutil.InputJSON, testutil.OutputSARIF},
			setupFunc: setupValidTestFiles,
			wantExit:  0,
		},
		{
			name:       "invalid sarif version",
			args:       []string{testutil.AppName, "--sarif-version", "9.9.9", testutil.InputJSON, testutil.OutputSARIF},
			setupFunc:  setupValidTestFiles,
			wantExit:   1,
			wantStderr: "Error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestCaseHelper(t, tt)
		})
	}
}

func TestPrintVersion(t *testing.T) {
	buf := &bytes.Buffer{}
	printVersion(buf)

	output := buf.String()
	if !strings.Contains(output, testutil.VersionOutput) {
		t.Errorf("printVersion() = %q, want to contain %q", output, testutil.VersionOutput)
	}
	if !strings.Contains(output, "commit: none") {
		t.Errorf("printVersion() = %q, want to contain %q", output, "commit: none")
	}
}

func TestPrintUsage(t *testing.T) {
	buf := &bytes.Buffer{}
	printUsage(buf)

	output := buf.String()
	if !strings.Contains(output, "Usage: "+testutil.AppName) {
		t.Errorf("printUsage() = %q, want to contain usage information", output)
	}
	if !strings.Contains(output, "--sarif-version") {
		t.Errorf("printUsage() = %q, want to contain --sarif-version flag", output)
	}
	if !strings.Contains(output, "--pretty") {
		t.Errorf("printUsage() = %q, want to contain --pretty flag", output)
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
