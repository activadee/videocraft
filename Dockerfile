# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev
# Install python3 and pip for Whisper
RUN apk add --no-cache python3 py3-pip

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE} -X main.gitCommit=${VCS_REF}" \
    -o videocraft cmd/server/main.go

# Final stage
FROM alpine:latest

# Add metadata
LABEL org.opencontainers.image.title="VideoCraft" \
      org.opencontainers.image.description="Advanced video generation platform with progressive subtitles" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.vendor="Activadee" \
      org.opencontainers.image.source="https://github.com/activadee/videocraft" \
      org.opencontainers.image.url="https://github.com/activadee/videocraft" \
      org.opencontainers.image.documentation="https://github.com/activadee/videocraft#readme" \
      org.opencontainers.image.licenses="MIT"

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