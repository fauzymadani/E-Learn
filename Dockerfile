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
# This creates a static binary that doesn't require C libraries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix nocgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/api \
    cmd/api/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/api .

# Copy .env file (optional, can be overridden by docker-compose)
COPY .env .env

# Create uploads directory
RUN mkdir -p uploads/avatars uploads/videos uploads/files && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"]

