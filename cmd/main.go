// main package contains the cli functionality
package main

import (
	"fmt"
	"os"

	"github.com/ivuorinen/go-test-sarif-action/internal"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go-test-sarif <input.json> <output.sarif>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	err := internal.ConvertToSARIF(inputFile, outputFile)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
