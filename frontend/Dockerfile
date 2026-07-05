# Multi-stage Dockerfile for React + TypeScript Vite SPA

# 1) Build stage
FROM node:24-alpine AS builder

# Some npm deps may require native compilation
RUN apk add --no-cache python3 make g++

WORKDIR /app

# Install deps first (better layer caching)
COPY package.json package-lock.json ./
RUN npm ci

# Copy source and build
COPY tsconfig*.json ./
COPY vite.config.ts ./
COPY src ./src
COPY public ./public
COPY index.html ./
RUN npm run build

# 2) Runtime stage: serve static files with nginx + reverse proxy for /api and /ws
FROM nginx:1.27-alpine AS runtime

# Backend upstream used for /api and /ws reverse proxying (can be overridden at runtime)
ENV BACKEND_URL="http://backend:8080"
ENV NGINX_PORT="8080"

# nginx image supports envsubst templates in /etc/nginx/templates/*.template
RUN rm -f /etc/nginx/conf.d/default.conf
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
