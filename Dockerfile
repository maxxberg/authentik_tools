# Build stage
FROM golang:1.24.5-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o authentik-group-manager

# Final stage
FROM alpine:3.19

# Add CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -H -h /app appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/authentik-group-manager .

# Set ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Required environment variables
ENV AGM_API_TOKEN=""
ENV AGM_API_HOST=""

# Run the binary
ENTRYPOINT ["./authentik-group-manager"]
