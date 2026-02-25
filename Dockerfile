# Build stage
FROM docker.io/library/golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM docker.io/library/alpine:latest

# Install ca-certificates and curl for healthcheck
RUN apk --no-cache add ca-certificates curl

# Create non-root user and group
RUN addgroup -g 1001 appuser && \
    adduser -D -u 1001 -G appuser appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER 1001:1001

# Expose port
EXPOSE 3003

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3003/api/v1/health || exit 1

# Run the application
CMD ["./main"]
