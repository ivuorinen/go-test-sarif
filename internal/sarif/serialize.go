package sarif

import "encoding/json"

// Internal SARIF document structure (version-agnostic)
type sarifDoc struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []run  `json:"runs"`
}

type run struct {
	Tool    tool     `json:"tool"`
	Results []result `json:"results"`
}

type tool struct {
	Driver driver `json:"driver"`
}

type driver struct {
	Name           string `json:"name"`
	InformationURI string `json:"informationUri,omitempty"`
	Rules          []rule `json:"rules,omitempty"`
}

type rule struct {
	ID               string  `json:"id"`
	ShortDescription message `json:"shortDescription,omitempty"`
}

type result struct {
	RuleID           string            `json:"ruleId"`
	Level            string            `json:"level"`
	Message          message           `json:"message"`
	LogicalLocations []logicalLocation `json:"logicalLocations,omitempty"`
}

type message struct {
	Text string `json:"text"`
}

type logicalLocation struct {
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
	Kind               string `json:"kind,omitempty"`
}

// serializeWithVersion creates SARIF JSON with specified schema and version
func serializeWithVersion(r *Report, schema, version string) ([]byte, error) {
	doc := sarifDoc{
		Schema:  schema,
		Version: version,
		Runs:    []run{buildRun(r)},
	}
	return json.Marshal(doc)
}

func buildRun(r *Report) run {
	rn := run{
		Tool: tool{
			Driver: driver{
				Name:           r.ToolName,
				InformationURI: r.ToolInfoURI,
			},
		},
		Results: make([]result, 0, len(r.Results)),
	}

	for _, rl := range r.Rules {
		rn.Tool.Driver.Rules = append(rn.Tool.Driver.Rules, rule{
			ID:               rl.ID,
			ShortDescription: message{Text: rl.Description},
		})
	}

	for _, res := range r.Results {
		r := result{
			RuleID:  res.RuleID,
			Level:   res.Level,
			Message: message{Text: res.Message},
		}

		if res.Location != nil {
			var fqn, kind string
			switch {
			case res.Location.Module != "" && res.Location.Function != "":
				fqn = res.Location.Module + "." + res.Location.Function
				kind = "function"
			case res.Location.Function != "":
				fqn = res.Location.Function
				kind = "function"
			case res.Location.Module != "":
				fqn = res.Location.Module
				kind = "module"
			}
			if fqn != "" {
				r.LogicalLocations = []logicalLocation{
					{FullyQualifiedName: fqn, Kind: kind},
				}
			}
		}

		rn.Results = append(rn.Results, r)
	}

	return rn
}
