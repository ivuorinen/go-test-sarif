# Set the application name
app_name := "go-test-sarif"
binary_path := "./bin/" + app_name
src := "./cmd/main.go"

# Default task
default:
    just build

# Lint Go code
lint:
  echo "Linting..."
  go vet ./...
  golangci-lint run

# Build the Go binary
build:
    echo "Building {{app_name}}..."
    mkdir -p bin
    GOOS=linux GOARCH=amd64 go build -o {{binary_path}} {{src}}
    echo "Binary built at {{binary_path}}"

# Run tests
test:
    echo "Running tests..."
    go test ./... -v

# Run the application
run:
    echo "Running {{app_name}}..."
    {{binary_path}} go-test-results.json go-test-results.sarif

# Clean build artifacts
clean:
    echo "Cleaning up..."
    rm -rf bin go-test-results.sarif

# Build the Docker image
docker-build:
    echo "Building Docker image..."
    docker build -t ghcr.io/ivuorinen/{{app_name}}:latest .

# Run the application inside Docker
docker-run:
    echo "Running {{app_name}} in Docker..."
    docker run --rm -v $(pwd):/workspace ghcr.io/ivuorinen/{{app_name}} go-test-results.json go-test-results.sarif

# Check if goreleaser is installed
check-goreleaser:
    @which goreleaser > /dev/null || (echo "goreleaser not found. Please install from https://goreleaser.com/install/" && exit 1)

# Create a snapshot release (for testing)
release-snapshot: check-goreleaser
    echo "Creating snapshot release..."
    goreleaser release --snapshot --clean

# Create a local release (without publishing)
release-local: check-goreleaser
    echo "Creating local release..."
    goreleaser release --skip=publish --clean

# Create and publish a release (requires GITHUB_TOKEN)
release: check-goreleaser
    echo "Creating and publishing release..."
    GITHUB_TOKEN=$(gh auth token) goreleaser release --clean

# Validate goreleaser configuration
release-check: check-goreleaser
    echo "Checking goreleaser configuration..."
    goreleaser check
