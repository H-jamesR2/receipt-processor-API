# Stage 1: Build the Go application using the official Go Alpine image
FROM golang:1.22.4-alpine AS builder

# Install necessary build dependencies
RUN apk add --no-cache git curl build-base \
    && go install github.com/pressly/goose/v3/cmd/goose@latest

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go modules dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o main .

# Stage 2: Create a minimal runtime image using Alpine
FROM alpine:3.18.3

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates curl

# Set the working directory in the final image
WORKDIR /app

# Copy the compiled binary and goose from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/main /app/main
COPY --from=builder /app/db/migrations /app/db/migrations

# Expose the port the app will run on
EXPOSE 8080

# Command to run migrations and start the server
CMD ["./main"]

