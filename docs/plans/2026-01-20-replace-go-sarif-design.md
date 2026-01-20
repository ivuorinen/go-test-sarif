# Design: Replace go-sarif with Internal SARIF Implementation

## Context

The `gopkg.in/yaml.v3` dependency is archived and unmaintained. It enters our dependency graph through:

```
go-sarif/v2 → testify → gopkg.in/yaml.v3
```

This project uses a minimal subset of go-sarif. Replacing it with an internal implementation eliminates all external dependencies and the yaml.v3 vulnerability.

## Goals

- Remove go-sarif dependency entirely
- Support SARIF v2.1.0 and v2.2 with extensible version system
- Capture all `go test -json` fields for future use
- Add logical location info (package/test name) to results
- Zero external dependencies after migration

## Package Structure

```
internal/
├── sarif/
│   ├── model.go       # Internal SARIF data model
│   ├── version.go     # Version enum and registry
│   ├── writer.go      # Common writing logic
│   ├── v21.go         # SARIF 2.1.0 serializer
│   └── v22.go         # SARIF 2.2 serializer
├── testjson/
│   └── parser.go      # Go test JSON parser (all 7 fields)
└── converter.go       # Orchestrates parsing → model → SARIF output
```

## Internal Data Model

Version-agnostic model that captures all relevant data:

```go
// internal/sarif/model.go

type Report struct {
    ToolName    string
    ToolInfoURI string
    Rules       []Rule
    Results     []Result
}

type Rule struct {
    ID          string
    Description string
}

type Result struct {
    RuleID   string
    Level    string  // "error", "warning", "note"
    Message  string
    Location *LogicalLocation
}

type LogicalLocation struct {
    Module   string  // Package name
    Function string  // Test name
}
```

## Test Event Parser

Captures all 7 fields from `go test -json`:

```go
// internal/testjson/parser.go

type TestEvent struct {
    Time        time.Time `json:"Time"`
    Action      string    `json:"Action"`
    Package     string    `json:"Package"`
    Test        string    `json:"Test,omitempty"`
    Elapsed     float64   `json:"Elapsed,omitempty"`
    Output      string    `json:"Output,omitempty"`
    FailedBuild string    `json:"FailedBuild,omitempty"`
}

func ParseFile(path string) ([]TestEvent, error)
```

Fails fast on malformed JSON with line numbers in error messages.

## Version Registry

Extensible system for adding SARIF versions:

```go
// internal/sarif/version.go

type Version string

const (
    Version210 Version = "2.1.0"
    Version22  Version = "2.2"
)

const DefaultVersion = Version210

type Serializer func(*Report) ([]byte, error)

var serializers = map[Version]Serializer{}

func Register(v Version, s Serializer)
func Serialize(r *Report, v Version, pretty bool) ([]byte, error)
func SupportedVersions() []string
```

Adding a new version requires:
1. Create version file (e.g., `v23.go`) with serializer function
2. Add version constant
3. Register in `init()`

## Version-Specific Serializers

Each version has its own JSON schema structs:

```go
// internal/sarif/v21.go

type sarifV21 struct {
    Schema  string   `json:"$schema"`
    Version string   `json:"version"`
    Runs    []runV21 `json:"runs"`
}

func serializeV21(r *Report) ([]byte, error)
```

SARIF v2.2 follows the same pattern with its schema differences.

## CLI Interface

```
go-test-sarif <input.json> <output.sarif>
go-test-sarif --sarif-version 2.2 <input.json> <output.sarif>
go-test-sarif --pretty <input.json> <output.sarif>
go-test-sarif --version
```

Flags:
- `--sarif-version`: SARIF output version (default: 2.1.0)
- `--pretty`: Pretty-print JSON output with indentation
- `--version`, `-v`: Display tool version

Help text for `--sarif-version` dynamically lists registered versions.

## Go API

```go
// internal/converter.go

type ConvertOptions struct {
    SARIFVersion sarif.Version
    Pretty       bool
}

func ConvertToSARIF(inputFile, outputFile string, opts ConvertOptions) error
```

## Output Format

- Compact JSON by default (no indentation)
- `--pretty` flag enables 2-space indented output

## Error Handling

- Fail fast on malformed JSON input
- Error messages include line numbers
- Return error for unsupported SARIF version

## Testing Strategy

```
internal/testjson/parser_test.go
- TestParseFile_ValidInput
- TestParseFile_AllFields
- TestParseFile_MalformedJSON
- TestParseFile_FileNotFound

internal/sarif/version_test.go
- TestSupportedVersions
- TestSerialize_UnknownVersion
- TestSerialize_PrettyOutput

internal/sarif/v21_test.go
- TestSerializeV21_Schema
- TestSerializeV21_WithResults
- TestSerializeV21_LogicalLocation

internal/sarif/v22_test.go
- TestSerializeV22_Schema
- TestSerializeV22_WithResults

internal/converter_test.go
- TestConvertToSARIF_Success (update existing)
- TestConvertToSARIF_Options
```

## Migration Steps

1. Implement `internal/testjson/` package
2. Implement `internal/sarif/` package with v2.1.0 and v2.2 serializers
3. Update `internal/converter.go` to use new packages
4. Update `cmd/main.go` with new CLI flags
5. Update existing tests, add new tests
6. Remove go-sarif import
7. Run `go mod tidy`
8. Verify: `go mod graph | grep yaml` returns nothing
9. Run full test suite

## Result

After migration:
- Zero external dependencies
- No yaml.v3 in dependency graph
- Extensible SARIF version support
- Richer output with logical location info
