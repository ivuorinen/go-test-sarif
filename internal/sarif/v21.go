// Package sarif provides SARIF report generation.
package sarif

func init() {
	Register(Version210, serializeV21)
}

func serializeV21(r *Report) ([]byte, error) {
	return serializeWithVersion(r,
		"https://json.schemastore.org/sarif-2.1.0.json",
		"2.1.0")
}
