# Development stage with Air for hot reloading
FROM golang:1.26-alpine

# Install Air, curl, Node.js and npm
RUN apk add --no-cache curl go nodejs npm && \
    curl -L https://github.com/cosmtrek/air/releases/download/v1.52.3/air_1.52.3_linux_amd64.tar.gz | tar -xz && \
    mv air /usr/local/bin/ && \
    chmod +x /usr/local/bin/air

# Install opencli globally
RUN npm install -g @jackwener/opencli

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy Air configuration and source code
COPY .air.toml ./
COPY . .

# Expose port
EXPOSE 8080

# Command to run Air
CMD ["air"]