# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o magento magento.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cli cli.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/magento .
COPY --from=builder /app/cli .

# Copy static assets and templates
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/html ./html
COPY --from=builder /app/input.css ./input.css
COPY --from=builder /app/tailwind.config.js ./tailwind.config.js

# Create var directory for logs
RUN mkdir -p /app/var

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./magento"]
