# Multi-stage build for Go backend
# Builder stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Build a static binary (defaults to linux/amd64, but supports other targets with buildx)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0

# Install build tools (git is sometimes needed for go modules)
RUN apk add --no-cache ca-certificates git

# Copy go module files and download dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source
COPY . .

# Build the server binary
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /bin/server ./cmd/server

# Final minimal runtime image
FROM alpine:3.20

# Add CA certificates for TLS connections (MongoDB, MinIO, etc.)
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/server /app/server

 # Default HTTP port (app also defaults to 8080 if PORT is unset)
 ENV PORT=8080

# Expose HTTP port
EXPOSE 8080

# Use non-root user
USER app

# Run the server
ENTRYPOINT ["/app/server"]
