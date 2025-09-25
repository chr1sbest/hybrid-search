# Makefile for the Hybrid Search API project

.PHONY: all generate mocks test docker

# Default target: generates code and mocks.
all: generate mocks

# Generate Go code from the OpenAPI specification.
generate:
	@echo "Generating OpenAPI server code..."
	go run -mod=mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=api/config.yaml api/spec.yaml

# Generate mocks for all interfaces.
mocks:
	@echo "Generating mocks..."
	go run -mod=mod github.com/vektra/mockery/v2 --name=VectorStore --dir=pkg/storage --output=pkg/storage/mocks --outpkg=mocks --case=underscore
	go run -mod=mod github.com/vektra/mockery/v2 --name=TextStore --dir=pkg/storage --output=pkg/storage/mocks --outpkg=mocks --case=underscore
	go run -mod=mod github.com/vektra/mockery/v2 --name=EmbeddingClient --dir=pkg/embeddings --output=pkg/embeddings/mocks --outpkg=mocks --case=underscore
	go run -mod=mod github.com/vektra/mockery/v2 --name=Service --dir=pkg/search --output=pkg/search/mocks --outpkg=mocks --case=underscore

# Run all tests.
test:
	@echo "Running tests..."
	go test ./...

# Run the application using Docker Compose.
docker:
	@echo "Starting services with Docker Compose..."
	docker compose up --build
