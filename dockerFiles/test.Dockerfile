# Use the official Golang 1.22 image as a base
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o /app/main .

# Expose the port the application runs on
EXPOSE 8080

# Run the application
CMD ["/app/main"]