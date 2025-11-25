# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/server/main.go
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o orders-service ./cmd/server

# Stage 2: Run
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/orders-service .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./orders-service"]

