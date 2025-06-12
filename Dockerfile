# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o videocraft cmd/server/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    tzdata \
    curl

# Create user
RUN addgroup -g 1000 videocraft && \
    adduser -D -u 1000 -G videocraft videocraft

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/videocraft .

# Copy config
COPY config/config.yaml ./config/

# Create directories and set permissions
RUN mkdir -p /app/generated_videos /app/temp /app/whisper_cache && \
    chown -R videocraft:videocraft /app

USER videocraft

EXPOSE 3002

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3002/health || exit 1

CMD ["./videocraft"]