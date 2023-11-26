# Use the official Golang image
FROM golang:1.21

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to the container
COPY go.mod .
COPY go.sum .

# Download and install Go dependencies
RUN go mod download

# Copy the rest of the application code to the container
COPY . .

# Build the application
RUN go build -o main .

# Expose the port on which the application will run
EXPOSE 8080

# Command to run the application
CMD ["./main"]
