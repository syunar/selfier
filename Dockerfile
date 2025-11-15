# Multi-stage build for Selfier application
# Stage 1: Build Go backend
FROM golang:1.25.1-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/selfier cmd/api/main.go

# Stage 2: Build Next.js frontend
FROM node:20-alpine AS web-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./
RUN npm ci

# Copy web source
COPY web/ ./

# Generate Prisma client
RUN npx prisma generate

# Build Next.js app
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# Stage 3: Production image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates nodejs npm curl

# Copy Go binary from builder
COPY --from=go-builder /app/bin/selfier /app/selfier

# Copy Next.js app from builder
COPY --from=web-builder /app/web/.next /app/web/.next
COPY --from=web-builder /app/web/public /app/web/public
COPY --from=web-builder /app/web/node_modules /app/web/node_modules
COPY --from=web-builder /app/web/package.json /app/web/package.json
COPY --from=web-builder /app/web/next.config.* /app/web/

# Copy Prisma files for runtime
COPY --from=web-builder /app/web/prisma /app/web/prisma
COPY --from=web-builder /app/web/node_modules/.prisma /app/web/node_modules/.prisma
COPY --from=web-builder /app/web/node_modules/@prisma /app/web/node_modules/@prisma

# Expose ports
EXPOSE 8080 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Default command (can be overridden in docker-compose)
CMD ["/app/selfier"]
