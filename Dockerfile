# Use the official Go image to build the executable
FROM golang:1.22 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules manifests and download dependencies
COPY go.mod go.sum  ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go executable
RUN go build -o main .

# Use a minimal image to run the Go executable
FROM ubuntu

# Set working directory
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/main .

COPY --from=builder /app/config.toml .
# Expose the port your app runs on (optional, if your app listens on a port)
EXPOSE 8080

# Run the Go executable
CMD ["./main"]