// Package sarif provides SARIF report generation.
package sarif

func init() {
	Register(Version22, serializeV22)
}

func serializeV22(r *Report) ([]byte, error) {
	return serializeWithVersion(r,
		"https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.2/schema/sarif-2.2.json",
		"2.2")
}
