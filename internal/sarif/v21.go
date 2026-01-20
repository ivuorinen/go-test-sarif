// internal/sarif/v21.go
package sarif

import "encoding/json"

func init() {
	Register(Version210, serializeV21)
}

// SARIF v2.1.0 JSON structures

type sarifV21 struct {
	Schema  string   `json:"$schema"`
	Version string   `json:"version"`
	Runs    []runV21 `json:"runs"`
}

type runV21 struct {
	Tool    toolV21     `json:"tool"`
	Results []resultV21 `json:"results"`
}

type toolV21 struct {
	Driver driverV21 `json:"driver"`
}

type driverV21 struct {
	Name           string    `json:"name"`
	InformationURI string    `json:"informationUri,omitempty"`
	Rules          []ruleV21 `json:"rules,omitempty"`
}

type ruleV21 struct {
	ID               string     `json:"id"`
	ShortDescription messageV21 `json:"shortDescription,omitempty"`
}

type resultV21 struct {
	RuleID           string               `json:"ruleId"`
	Level            string               `json:"level"`
	Message          messageV21           `json:"message"`
	LogicalLocations []logicalLocationV21 `json:"logicalLocations,omitempty"`
}

type messageV21 struct {
	Text string `json:"text"`
}

type logicalLocationV21 struct {
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
	Kind               string `json:"kind,omitempty"`
}

func serializeV21(r *Report) ([]byte, error) {
	doc := sarifV21{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs:    []runV21{buildRunV21(r)},
	}
	return json.Marshal(doc)
}

func buildRunV21(r *Report) runV21 {
	run := runV21{
		Tool: toolV21{
			Driver: driverV21{
				Name:           r.ToolName,
				InformationURI: r.ToolInfoURI,
			},
		},
		Results: make([]resultV21, 0, len(r.Results)),
	}

	// Add rules
	for _, rule := range r.Rules {
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, ruleV21{
			ID:               rule.ID,
			ShortDescription: messageV21{Text: rule.Description},
		})
	}

	// Add results
	for _, result := range r.Results {
		res := resultV21{
			RuleID:  result.RuleID,
			Level:   result.Level,
			Message: messageV21{Text: result.Message},
		}

		if result.Location != nil {
			fqn := result.Location.Module
			if result.Location.Function != "" {
				fqn += "." + result.Location.Function
			}
			res.LogicalLocations = []logicalLocationV21{
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
