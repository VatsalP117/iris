# Build dashboard frontend only.
FROM node:20-alpine AS frontend-builder
WORKDIR /app
RUN npm install -g pnpm

COPY package.json pnpm-lock.yaml ./
COPY docker/pnpm-workspace.dashboard.yaml ./pnpm-workspace.yaml
COPY dashboard/package.json ./dashboard/package.json
RUN pnpm install --frozen-lockfile --filter @iris/dashboard...

COPY dashboard ./dashboard
RUN pnpm --filter @iris/dashboard build


# Build Go API server only.
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app

ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg
RUN go build -ldflags="-w -s" -o iris-server ./cmd/server


# Final production image.
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=backend-builder /app/iris-server ./
COPY --from=frontend-builder /app/dashboard/dist ./dashboard/dist

ENV PORT=8080
ENV DB_PATH=/app/data/iris.db
ENV DASHBOARD_DIR=/app/dashboard/dist

EXPOSE 8080
VOLUME ["/app/data"]

CMD ["./iris-server"]
