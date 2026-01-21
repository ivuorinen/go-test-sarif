// Package main provides the CLI for converting go test JSON output to SARIF format.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ivuorinen/go-test-sarif-action/internal"
	"github.com/ivuorinen/go-test-sarif-action/internal/sarif"
)

// Build-time variables set via ldflags.
var (
	// version is the application version, set at build time.
	version = "dev"
	// commit is the git commit hash, set at build time.
	commit = "none"
	// date is the build date, set at build time.
	date = "unknown"
	// builtBy is the builder identifier, set at build time.
	builtBy = "unknown"
)

func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "go-test-sarif %s\n", version)
	_, _ = fmt.Fprintf(w, "  commit: %s\n", commit)
	_, _ = fmt.Fprintf(w, "  built at: %s\n", date)
	_, _ = fmt.Fprintf(w, "  built by: %s\n", builtBy)
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: go-test-sarif [options] <input.json> <output.sarif>")
	_, _ = fmt.Fprintln(w, "       go-test-sarif --version")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Options:")
	_, _ = fmt.Fprintf(w, "  --sarif-version string   SARIF version (%s) (default %q)\n",
		strings.Join(sarif.SupportedVersions(), ", "), sarif.DefaultVersion)
	_, _ = fmt.Fprintln(w, "  --pretty                 Pretty-print JSON output")
	_, _ = fmt.Fprintln(w, "  -v, --version            Display version information")
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("go-test-sarif", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		versionFlag  bool
		sarifVersion string
		prettyOutput bool
	)

	fs.BoolVar(&versionFlag, "version", false, "Display version information")
	fs.BoolVar(&versionFlag, "v", false, "Display version information (short)")
	fs.StringVar(&sarifVersion, "sarif-version", string(sarif.DefaultVersion),
		fmt.Sprintf("SARIF version (%s)", strings.Join(sarif.SupportedVersions(), ", ")))
	fs.BoolVar(&prettyOutput, "pretty", false, "Pretty-print JSON output")

	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	if versionFlag {
		printVersion(stdout)
		return 0
	}

	if fs.NArg() < 2 {
		printUsage(stderr)
		return 1
	}

	inputFile := fs.Arg(0)
	outputFile := fs.Arg(1)

	opts := internal.ConvertOptions{
		SARIFVersion: sarif.Version(sarifVersion),
		Pretty:       prettyOutput,
	}

	if err := internal.ConvertToSARIF(inputFile, outputFile, opts); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
