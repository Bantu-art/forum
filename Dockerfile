# Stage 1: Build the Go application
FROM golang:1.23.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Enable CGO and build the application
ENV CGO_ENABLED=1
RUN go build -o forum ./main.go

# Stage 2: Create a minimal image with the built binary
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/forum .
COPY static ./static
COPY templates ./templates

# Expose port 8000
EXPOSE 8000

# Command to run the executable
CMD ["./forum"]