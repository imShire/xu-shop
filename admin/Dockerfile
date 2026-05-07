# syntax=docker/dockerfile:1

FROM node:20-alpine AS builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app
COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY . .

ARG VITE_API_BASE=/api/v1
ARG VITE_APP_TITLE=xu-shop后台
ENV VITE_API_BASE=$VITE_API_BASE
ENV VITE_APP_TITLE=$VITE_APP_TITLE

RUN pnpm build

# ─── Runtime ────────────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
