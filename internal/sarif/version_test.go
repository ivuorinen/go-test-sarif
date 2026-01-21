// internal/sarif/version_test.go
package sarif

import (
	"bytes"
	"strings"
	"testing"
)

func TestSupportedVersions(t *testing.T) {
	versions := SupportedVersions()

	if len(versions) < 2 {
		t.Errorf("expected at least 2 versions, got %d", len(versions))
	}

	// Should contain 2.1.0 and 2.2
	found210 := false
	found22 := false
	for _, v := range versions {
		if v == "2.1.0" {
			found210 = true
		}
		if v == "2.2" {
			found22 = true
		}
	}

	if !found210 {
		t.Error("SupportedVersions should contain 2.1.0")
	}
	if !found22 {
		t.Error("SupportedVersions should contain 2.2")
	}
}

func TestSerialize_UnknownVersion(t *testing.T) {
	report := &Report{ToolName: "test"}

	_, err := Serialize(report, "9.9.9", false)
	if err == nil {
		t.Fatal("expected error for unknown version, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "unsupported")
	}
}

func TestDefaultVersion(t *testing.T) {
	if DefaultVersion != Version210 {
		t.Errorf("DefaultVersion = %q, want %q", DefaultVersion, Version210)
	}
}

func TestSerialize_PrettyOutput(t *testing.T) {
	report := &Report{
		ToolName: "test-tool",
	}

	compact, err := Serialize(report, Version210, false)
	if err != nil {
		t.Fatalf("Serialize compact returned error: %v", err)
	}

	pretty, err := Serialize(report, Version210, true)
	if err != nil {
		t.Fatalf("Serialize pretty returned error: %v", err)
	}

	// Pretty should be longer due to whitespace
	if len(pretty) <= len(compact) {
		t.Errorf("pretty output (%d bytes) should be longer than compact (%d bytes)", len(pretty), len(compact))
	}

	// Pretty should contain newlines and indentation
	if !bytes.Contains(pretty, []byte("\n")) {
		t.Error("pretty output should contain newlines")
	}
	if !bytes.Contains(pretty, []byte("  ")) {
		t.Error("pretty output should contain indentation")
	}
}
