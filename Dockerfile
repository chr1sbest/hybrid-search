# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for a static binary
# -o /app/server builds the binary into the /app directory
# --mount=type=cache... provides a persistent cache for Go modules and build artifacts
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/app

# Stage 2: Create the final, minimal image
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Copy the OpenAPI spec so the application can serve it
COPY api/spec.yaml ./api/spec.yaml

# Expose the port the app runs on
EXPOSE 8080

# The command to run the application
ENTRYPOINT ["/app/server"]
