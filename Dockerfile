# Build stage
FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix nocgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/api \
    cmd/api/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/bin/api .
COPY .env* ./

RUN mkdir -p uploads/avatars uploads/videos uploads/files logs && \
    chmod -R 777 uploads logs

ENV CONTAINER=true

EXPOSE 8080

CMD ["./api"]