# Build UI stage (Vite + TypeScript + Shadcn UI)
FROM node:18-alpine AS ui-builder

WORKDIR /app

# Copy package manifests for ui
COPY ui/package.json ui/package-lock.json* ./

# Install dependencies
RUN npm install

# Copy all ui source files
COPY ui/ ./

# Build the UI
RUN npm run build

# Build Go stage
FROM golang:1.19-alpine AS go-builder

# Install build dependencies for Go
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the Go source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux go build -o middleware-manager .

# Final stage
FROM alpine:3.16

RUN apk add --no-cache ca-certificates sqlite curl tzdata

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=go-builder /app/middleware-manager /app/middleware-manager

# Copy UI build files from UI builder stage
COPY --from=ui-builder /app/dist /app/ui/dist

# Copy configuration files
COPY --from=go-builder /app/config/templates.yaml /app/config/templates.yaml
COPY --from=go-builder /app/config/templates_services.yaml /app/config/templates_services.yaml

# Copy database migrations file
COPY --from=go-builder /app/database/migrations.sql /app/database/migrations.sql
# Also copy to root as fallback
COPY --from=go-builder /app/database/migrations.sql /app/migrations.sql

# Create directories for data
RUN mkdir -p /data /conf

# Set environment variables
ENV PANGOLIN_API_URL=http://pangolin:3001/api/v1 \
    TRAEFIK_CONF_DIR=/conf \
    DB_PATH=/data/middleware.db \
    PORT=3456

# Expose the port
EXPOSE 3456

# Run the application
CMD ["/app/middleware-manager"]
