// Package sarif provides SARIF report generation.
package sarif

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// Version represents a SARIF specification version.
type Version string

const (
	// Version210 is SARIF version 2.1.0.
	Version210 Version = "2.1.0"
	// Version22 is SARIF version 2.2.
	Version22 Version = "2.2"
)

// DefaultVersion is the default SARIF version used when not specified.
const DefaultVersion = Version210

// Serializer converts an internal Report to version-specific JSON.
type Serializer func(*Report) ([]byte, error)

var serializers = map[Version]Serializer{}

// Register adds a serializer for a SARIF version.
// Called by version-specific files in their init() functions.
func Register(v Version, s Serializer) {
	serializers[v] = s
}

// Serialize converts a Report to JSON for the specified SARIF version.
func Serialize(r *Report, v Version, pretty bool) ([]byte, error) {
	s, ok := serializers[Version(v)]
	if !ok {
		return nil, fmt.Errorf("unsupported SARIF version: %s", v)
	}

	data, err := s(r)
	if err != nil {
		return nil, err
	}

	if pretty {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	return data, nil
}

// SupportedVersions returns all registered SARIF versions, sorted.
func SupportedVersions() []string {
	versions := make([]string, 0, len(serializers))
	for v := range serializers {
		versions = append(versions, string(v))
	}
	sort.Strings(versions)
	return versions
}
