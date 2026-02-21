# Build Dashboard (Frontend)
FROM node:20-alpine AS frontend-builder
WORKDIR /app
RUN npm install -g pnpm

# Only copy what's needed for dependencies first
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json ./web/
COPY dashboard/package.json ./dashboard/

RUN pnpm install --frozen-lockfile

# Copy source and build
COPY . .
RUN cd dashboard && pnpm build


# Build Go Server (Backend)
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app

# Enable CGO for SQLite
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-w -s" -o iris-server ./cmd/server


# Final Production Image
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# Copy built backend binary
COPY --from=backend-builder /app/iris-server ./

# Copy built frontend static files
COPY --from=frontend-builder /app/dashboard/dist ./dashboard/dist

# Set environment variables
ENV PORT=8080
ENV DB_PATH=/app/data/iris.db
ENV DASHBOARD_DIR=/app/dashboard/dist

# Expose port and declare volume for persistent SQLite DB
EXPOSE 8080
VOLUME ["/app/data"]

CMD ["./iris-server"]
