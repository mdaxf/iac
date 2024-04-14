# Use a lightweight Linux distribution as the base image
FROM golang:1.21-alpine

RUN mkdir /app

# Copy the compiled Go application into the container
COPY iac-linux /app/    
COPY apiconfig.json /app/  
COPY configuration.json /app/  


# Set the working directory inside the container
WORKDIR /app

# Set permissions on the application (if needed)
RUN chmod +x iac-linux

# Expose additional ports
EXPOSE 8080
# Define an entry point to run the application

CMD ["./iac-linux"]