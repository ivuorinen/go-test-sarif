#!/bin/sh
set -e

if [ -z "$INPUT_TEST_RESULTS" ]; then
  echo "Missing test results input file"
  exit 1
fi

OUTPUT_FILE="go-test-results.sarif"

/go-test-sarif "$INPUT_TEST_RESULTS" "$OUTPUT_FILE"

echo "Generated SARIF report: $OUTPUT_FILE"
