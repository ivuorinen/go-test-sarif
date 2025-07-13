module github.com/ivuorinen/go-test-sarif-action

go 1.24.1

require (
	github.com/owenrumney/go-sarif/v3 v3.2.1
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

replace golang.org/x/crypto => golang.org/x/crypto v0.40.0

replace golang.org/x/net => golang.org/x/net v0.42.0

replace golang.org/x/text => golang.org/x/text v0.27.0

replace gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1
