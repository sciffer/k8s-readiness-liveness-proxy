# Use the official Golang base image.
FROM golang:1.23 AS build

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files and download dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code.
COPY . .

# Build the Go application.
RUN go build -o main .

FROM scratch

# Copy relevant files
COPY --from=build app/main .
ADD config .

# Expose port 8080 for the application.
EXPOSE 8080

# Define the command to run the application.
CMD ["./main"]