# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod ./

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy && go mod verify

# Copy source code
COPY . .

# Build the application
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy any additional files if needed
# COPY --from=builder /app/README.md .

# Make binary executable
RUN chmod +x ./main

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
