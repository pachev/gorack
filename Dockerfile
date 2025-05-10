# Build stage
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o gorack .

# Run stage
FROM alpine:latest

# Install CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user to run the application
RUN adduser -D -H -h /app appuser

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/gorack /app/

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose API port
EXPOSE 8080

# Set environment variables
ENV API_PORT=8080

# Run the application
CMD ["/app/gorack"]
