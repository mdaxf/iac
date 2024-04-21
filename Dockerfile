# Use a lightweight Linux distribution as the base image
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /build

# Copy the Go module files
COPY go.mod go.sum ./
COPY apiconfig.json ./
COPY configuration.json ./
# Download the Go module dependencies
RUN go mod download

# Copy the entire application source
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o iac-linux

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/iac-linux /app/iac-linux
COPY --from=builder /build/apiconfig.json /app/apiconfig.json
COPY --from=builder /build/dockerconfiguration.json /app/configuration.json

# Copy the compiled Go application into the container
#COPY iac-linux /app/    
#COPY apiconfig.json /app/  
#COPY configuration.json /app/  


# Set the working directory inside the container
#WORKDIR /app

# Set permissions on the application (if needed)
#RUN chmod +x iac-linux
RUN chmod +x /app/iac-linux

# Expose additional ports
EXPOSE 8080
# Define an entry point to run the application

CMD ["/app/iac-linux"]