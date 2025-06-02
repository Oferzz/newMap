# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files from the monorepo structure
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
WORKDIR /build/apps/api

# Download dependencies
RUN go mod download

# Copy the entire API directory
COPY apps/api/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/apps/api/server .

# Copy migrations
COPY --from=builder /build/apps/api/migrations ./migrations

# Copy health check script
COPY healthcheck.sh .

# Create directory for media storage
RUN mkdir -p /data/media && chown -R appuser:appuser /data

# Make health check script executable
RUN chmod +x healthcheck.sh

# Switch to non-root user
USER appuser

# Expose port (Render will override this)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./healthcheck.sh

# Run the application
CMD ["./server"]