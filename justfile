# Project Mercury justfile
# See https://just.systems for documentation

# Default recipe
default: build

# Build the application binary
build:
    @echo "Building..."
    go build -o bin/api ./cmd/api

# Run the application
run:
    @echo "Running..."
    go run ./cmd/api

# Run tests
test:
    @echo "Testing..."
    go test ./... -v

# Clean build artifacts
clean:
    @echo "Cleaning..."
    go clean
    rm -rf bin docs/docs.go docs/swagger.json docs/swagger.yaml

# Format code
fmt:
    @echo "Formatting..."
    go fmt ./...

# Run go vet
vet:
    @echo "Vetting..."
    go vet ./...

# Tidy go modules
tidy:
    @echo "Tidying..."
    go mod tidy

# Generate Swagger documentation
docs:
    @echo "Generating Swagger docs..."
    swag init -g internal/adapters/input/router.go

# Show available recipes
help:
    @echo "Usage: just [recipe]"
    @echo ""
    @echo "Recipes:"
    @echo "  build   - Build the application binary"
    @echo "  run     - Run the application"
    @echo "  test    - Run tests"
    @echo "  clean   - Remove binary and build artifacts"
    @echo "  fmt     - Format code"
    @echo "  vet     - Run go vet"
    @echo "  tidy    - Run go mod tidy"
    @echo "  docs    - Generate Swagger documentation"
