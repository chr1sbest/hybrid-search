# Makefile for the Hybrid Search API project

.PHONY: all run generate mocks test tidy vendor docs docker

# Default target: cleans up, generates code, and vendors dependencies.
all: tidy generate vendor

# Run the main application.
run:
	@echo "Starting server..."
	go run cmd/app/main.go

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

# Tidy Go modules.
tidy:
	@echo "Tidying Go modules..."
	go mod tidy

# Vendor dependencies.
vendor:
	@echo "Vendoring dependencies..."
	go mod vendor

# Open the API documentation in the browser.
docs:
	@echo "Opening API documentation..."
	open http://localhost:8080/docs

# Run the application using Docker Compose.
docker:
	@echo "Starting services with Docker Compose..."
	docker-compose up --build
