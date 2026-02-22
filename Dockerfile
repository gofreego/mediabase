# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o application .

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install certificates for HTTPS requests if needed
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder
COPY --from=builder /app/application .

# Copy configuration and resources
COPY --from=builder /app/dev.yaml .
COPY --from=builder /app/api/docs ./api/docs

# Expose ports
EXPOSE 8095 8096

# Run the application
CMD ["/app/application"]