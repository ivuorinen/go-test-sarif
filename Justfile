# Set the application name
app_name := "go-test-sarif"
binary_path := "./bin/" + app_name
src := "./cmd/main.go"

# Default task
default:
    just build

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
