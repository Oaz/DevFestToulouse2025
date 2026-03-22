# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /opt

COPY fulcrum/go.mod ./fulcrum/
COPY game/go.mod ./game/

RUN cat > go.work <<'EOF'
go 1.21

use (
    ./fulcrum
    ./game
)
EOF

# Download dependencies
RUN go work sync

# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy the source code
COPY fulcrum/ ./fulcrum/
COPY game/ ./game/

WORKDIR /opt/fulcrum

# Generate swagger docs
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o fulcrum .

# Final stage
FROM alpine:latest

# Add ca certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -S fulcrumgroup && adduser -S fulcrumuser -G fulcrumgroup

# Set working directory
WORKDIR /opt

# Copy the binary from builder
COPY --from=builder /opt/fulcrum/fulcrum .

# Copy generated swagger docs
COPY --from=builder /opt/fulcrum/docs ./docs

# Copy any necessary configuration files
COPY --from=builder /opt/fulcrum/.env* ./

# Use non-root user
USER fulcrumuser

# Expose the application port
EXPOSE 8086

# Command to run the application
CMD ["./fulcrum"]