# Use the official Golang base image.
FROM --platform=$BUILDPLATFORM golang:1.23 AS build
ARG TARGETOS
ARG TARGETARCH

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files and download dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code.
COPY . .

# Build the Go application.
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o main .

FROM scratch

# Copy relevant files
COPY --from=build app/main .
ADD config .

# Expose port 8080 for the application.
EXPOSE 8080

# Define the command to run the application.
CMD ["./main"]