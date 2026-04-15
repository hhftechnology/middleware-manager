# Build UI stage (existing middleware-manager UI)
FROM node:18-alpine AS ui-builder

WORKDIR /app

COPY ui/package.json ui/package-lock.json* ./
RUN npm install
COPY ui/ ./
RUN npm run build

# Build Traefik Manager UI stage
FROM node:20-alpine AS traefik-ui-builder

WORKDIR /app

COPY traefik-ui/package.json traefik-ui/package-lock.json* ./
RUN npm install
COPY traefik-ui/ ./
RUN npm run build

# Build Go stage - using Debian for glibc compatibility with go-sqlite3
FROM golang:1.25.8-bookworm AS go-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy && \
    CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w -extldflags '-static'" -o middleware-manager .

# Final stage - minimal runtime image
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=go-builder /app/middleware-manager /app/middleware-manager
COPY --from=ui-builder /app/dist /app/ui/dist
COPY --from=traefik-ui-builder /app/dist /app/traefik-ui/dist

COPY --from=go-builder /app/config/templates.yaml /app/config/templates.yaml
COPY --from=go-builder /app/config/templates_services.yaml /app/config/templates_services.yaml
COPY --from=go-builder /app/database/migrations.sql /app/database/migrations.sql
COPY --from=go-builder /app/database/migrations.sql /app/migrations.sql

RUN mkdir -p /data /conf /app/config /app/backups

ENV MODE=middleware-manager \
    PANGOLIN_API_URL=http://pangolin:3001/api/v1 \
    TRAEFIK_CONF_DIR=/conf \
    DB_PATH=/data/middleware.db \
    PORT=3456 \
    TM_UI_PATH=/app/traefik-ui/dist

EXPOSE 3456

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:3456/health || exit 1

CMD ["/app/middleware-manager"]
