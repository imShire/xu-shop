# syntax=docker/dockerfile:1
# context: 仓库根目录（xu-shop/）

# ─── Go Builder ─────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS go-builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api    ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# ─── API ────────────────────────────────────────────────────────────────────────
FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-builder /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]

# ─── Worker ─────────────────────────────────────────────────────────────────────
FROM alpine:3.20 AS worker

RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-builder /out/worker /worker
ENTRYPOINT ["/worker"]

# ─── Migrate ────────────────────────────────────────────────────────────────────
FROM alpine:3.20 AS migrate

RUN apk add --no-cache ca-certificates
COPY --from=go-builder /go/bin/goose /usr/local/bin/goose
COPY --from=go-builder /src/migrations /migrations

# ─── Admin Builder (Vue3 + Vite) ────────────────────────────────────────────────
FROM node:20-alpine AS admin-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app
COPY admin/package.json admin/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY admin/ .

ARG VITE_API_BASE=/api/v1
ARG VITE_APP_TITLE=xu-shop后台
ENV VITE_API_BASE=$VITE_API_BASE \
    VITE_APP_TITLE=$VITE_APP_TITLE

RUN pnpm build

# ─── Client Builder (Taro H5) ───────────────────────────────────────────────────
FROM node:20-alpine AS client-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app
COPY client/package.json client/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY client/ .

ARG VITE_API_BASE=/api/v1
ARG TARO_APP_API=/api/v1
ENV VITE_API_BASE=$VITE_API_BASE \
    TARO_APP_API=$TARO_APP_API

RUN pnpm build:h5

# ─── Frontend Runtime (admin :5273 + client H5 :10086) ──────────────────────────
FROM nginx:1.27-alpine AS frontend

COPY --from=admin-builder  /app/dist      /usr/share/nginx/admin
COPY --from=client-builder /app/dist/h5   /usr/share/nginx/client
COPY deploy/nginx.frontend.conf /etc/nginx/conf.d/default.conf

EXPOSE 5273 10086
