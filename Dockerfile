
# Stage 1: Build the Go application
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-app

# Stage 2: Setup the final image with Gemini CLI
FROM node:16-alpine

WORKDIR /app

# Install Gemini CLI
RUN npm install -g @gemini-cli/cli

# Copy the built Go application from the builder stage
COPY --from=builder /go-app .

# Set the entrypoint for the container
ENTRYPOINT ["./go-app"]
