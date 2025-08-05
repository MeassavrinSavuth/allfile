# Start from the official Go image
FROM golang:1.21-alpine

# Set working directory inside the container
WORKDIR /app

# Install git and build tools
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of your code
COPY . .

# Build the Go app
RUN go build -o server main.go

# Expose port (update if your server uses a different one)
EXPOSE 8080

# Run the executable
CMD ["./server"]
