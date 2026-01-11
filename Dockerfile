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
FROM golang:1.21-alpine AS go-builder

# Install build dependencies for Go with CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum files first for better layer caching
COPY go.mod go.sum ./

# Download Go dependencies (this layer is cached if go.mod/go.sum don't change)
RUN go mod download

# Copy the Go source code
COPY . .

# Ensure go.sum is up to date and build the application
# Using CGO for SQLite support
RUN go mod tidy && \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o middleware-manager .

# Final stage - minimal runtime image
FROM alpine:3.18

# Install runtime dependencies
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

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3456/health || exit 1

# Run the application
CMD ["/app/middleware-manager"]
