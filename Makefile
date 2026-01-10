APP_NAME=projectMercury
BINARY_NAME=bin/api
MAIN_PACKAGE=./cmd/api
DOCS_ENTRY=internal/handlers/api.go

.PHONY: all build run test clean fmt vet tidy docs help

all: fmt vet docs build

build:
	@echo "Building..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

run:
	@echo "Running..."
	go run $(MAIN_PACKAGE)

test:
	@echo "Testing..."
	go test ./... -v

clean:
	@echo "Cleaning..."
	go clean
	rm -rf bin docs/docs.go docs/swagger.json docs/swagger.yaml

fmt:
	@echo "Formatting..."
	go fmt ./...

vet:
	@echo "Vetting..."
	go vet ./...

tidy:
	@echo "Tidying..."
	go mod tidy

docs:
	@echo "Generating Swagger docs..."
	swag init -g $(DOCS_ENTRY)

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build   - Build the application binary"
	@echo "  run     - Run the application"
	@echo "  test    - Run tests"
	@echo "  clean   - Remove binary and build artifacts"
	@echo "  fmt     - Format code"
	@echo "  vet     - Run go vet"
	@echo "  tidy    - Run go mod tidy"
	@echo "  docs    - Generate Swagger documentation"