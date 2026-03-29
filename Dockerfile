# Build stage
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum (if exists)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o teeter lb/cmd/lb/main.go

# Run stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/teeter .
COPY --from=builder /app/config.yaml .

# Expose ports (default)
EXPOSE 8080 8081

# Command to run the application
CMD ["./teeter", "-config", "config.yaml"]
