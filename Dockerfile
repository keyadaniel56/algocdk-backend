# Use the official Go image
FROM golang:1.22-alpine

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go app
RUN go build -o main .

# Expose the port your app listens on (change if different)
EXPOSE 8080

# Command to run the app
CMD ["./main"]
