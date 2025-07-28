package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ivuorinen/go-test-sarif-action/internal"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "go-test-sarif %s\n", version)
	_, _ = fmt.Fprintf(w, "  commit: %s\n", commit)
	_, _ = fmt.Fprintf(w, "  built at: %s\n", date)
	_, _ = fmt.Fprintf(w, "  built by: %s\n", builtBy)
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: go-test-sarif <input.json> <output.sarif>")
	_, _ = fmt.Fprintln(w, "       go-test-sarif --version")
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("go-test-sarif", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var versionFlag bool
	fs.BoolVar(&versionFlag, "version", false, "Display version information")
	fs.BoolVar(&versionFlag, "v", false, "Display version information (short)")
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

	if err := internal.ConvertToSARIF(inputFile, outputFile); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
