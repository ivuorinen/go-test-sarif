// internal/sarif/v22.go
package sarif

import "encoding/json"

func init() {
	Register(Version22, serializeV22)
}

// SARIF v2.2 JSON structures
// v2.2 is structurally similar to v2.1.0 with minor additions

type sarifV22 struct {
	Schema  string   `json:"$schema"`
	Version string   `json:"version"`
	Runs    []runV22 `json:"runs"`
}

type runV22 struct {
	Tool    toolV22     `json:"tool"`
	Results []resultV22 `json:"results"`
}

type toolV22 struct {
	Driver driverV22 `json:"driver"`
}

type driverV22 struct {
	Name           string    `json:"name"`
	InformationURI string    `json:"informationUri,omitempty"`
	Rules          []ruleV22 `json:"rules,omitempty"`
}

type ruleV22 struct {
	ID               string     `json:"id"`
	ShortDescription messageV22 `json:"shortDescription,omitempty"`
}

type resultV22 struct {
	RuleID           string               `json:"ruleId"`
	Level            string               `json:"level"`
	Message          messageV22           `json:"message"`
	LogicalLocations []logicalLocationV22 `json:"logicalLocations,omitempty"`
}

type messageV22 struct {
	Text string `json:"text"`
}

type logicalLocationV22 struct {
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
	Kind               string `json:"kind,omitempty"`
}

func serializeV22(r *Report) ([]byte, error) {
	doc := sarifV22{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.2/schema/sarif-2.2.json",
		Version: "2.2",
		Runs:    []runV22{buildRunV22(r)},
	}
	return json.Marshal(doc)
}

func buildRunV22(r *Report) runV22 {
	run := runV22{
		Tool: toolV22{
			Driver: driverV22{
				Name:           r.ToolName,
				InformationURI: r.ToolInfoURI,
			},
		},
		Results: make([]resultV22, 0, len(r.Results)),
	}

	// Add rules
	for _, rule := range r.Rules {
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, ruleV22{
			ID:               rule.ID,
			ShortDescription: messageV22{Text: rule.Description},
		})
	}

	// Add results
	for _, result := range r.Results {
		res := resultV22{
			RuleID:  result.RuleID,
			Level:   result.Level,
			Message: messageV22{Text: result.Message},
		}

		if result.Location != nil {
			fqn := result.Location.Module
			if result.Location.Function != "" {
				fqn += "." + result.Location.Function
			}
			res.LogicalLocations = []logicalLocationV22{
				{
					FullyQualifiedName: fqn,
					Kind:               "function",
				},
			}
		}

		run.Results = append(run.Results, res)
	}

	return run
}
