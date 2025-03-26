# go-test-sarif

`go-test-sarif` is a CLI tool and GitHub Action for converting `go test -json` output into SARIF format,
making it compatible with GitHub Security Tab and other SARIF consumers.

## 🚀 Features

- Converts `go test -json` output to **SARIF format**.
- **GitHub Action integration** for CI/CD pipelines.
- Generates structured test failure reports for **security and compliance tools**.
- Works as a **standalone CLI tool**.

## 📦 Installation

### Using `go install`

```sh
go install github.com/ivuorinen/go-test-sarif@latest
```

### Using Docker

```sh
docker pull ghcr.io/ivuorinen/go-test-sarif:latest
```

## 🛠️ Usage

### CLI Usage

```sh
go test -json ./... > go-test-results.json
go-test-sarif go-test-results.json go-test-results.sarif
```

### Docker Usage

```sh
docker run --rm -v $(pwd):/workspace ghcr.io/ivuorinen/go-test-sarif go-test-results.json go-test-results.sarif
```

### GitHub Action Usage

Add the following step to your GitHub Actions workflow:

```yaml
- name: Convert JSON to SARIF
  uses: ivuorinen/go-test-sarif@v1
  with:
    test_results: go-test-results.json
```

To upload the SARIF file to GitHub Security Tab, add:

```yaml
- name: Upload SARIF report
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: go-test-results.sarif
```

## 📜 Output Example

SARIF report example:
```json
{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Go Test",
          "informationUri": "https://golang.org/cmd/go/#hdr-Test_packages",
          "version": "1.0.0"
        }
      },
      "results": [
        {
          "ruleId": "go-test-failure",
          "level": "error",
          "message": {
            "text": "Test failed"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "github.com/example/package"
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## 🏗 Development

Clone the repository and build the project:
```sh
git clone https://github.com/ivuorinen/go-test-sarif.git
cd go-test-sarif
go build -o go-test-sarif ./cmd/main.go
```

Run tests:

```sh
go test ./...
```

## 📄 License

This project is licensed under the **MIT License**.

## 🤝 Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss the changes.
