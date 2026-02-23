# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate docs if not exists
RUN if [ ! -f "docs/docs.go" ]; then \
  go install github.com/swaggo/swag/cmd/swag@latest && \
  swag init -g cmd/server/main.go -o docs/; \
  fi

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Build the migration tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate cmd/migrate/main.go

# Final stage
FROM alpine:3.19

# Install ca-certificates, timezone data, wget for health check, and docker CLI
RUN apk update && apk --no-cache add ca-certificates tzdata wget docker-cli

WORKDIR /root/

# Copy the binaries
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Create uploads directory
RUN mkdir -p uploads/code

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]