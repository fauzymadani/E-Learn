# Build stage
FROM golang:alpine AS builder

RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO disabled for pure Go build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix nocgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/api \
    cmd/api/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and curl for healthcheck
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/api .

# Copy .env file (optional, will be overridden by docker-compose env vars)
COPY .env* ./

# Create required directories with proper permissions
RUN mkdir -p uploads/avatars uploads/videos uploads/files logs && \
    chmod -R 777 logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Set environment variable to indicate running in container
ENV CONTAINER=true

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"]