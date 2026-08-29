# =========================
# Build Stage
# =========================
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Copy application source code
COPY . .

# Build arguments for Vite environment variables
ARG VITE_AUTH_BASE_URL=http://localhost:8082
ARG VITE_PORT=4000
ARG VITE_DOCS_BASE_URL=https://docs.roled.io

# Set as environment variables during build
ENV VITE_AUTH_BASE_URL=$VITE_AUTH_BASE_URL \
    VITE_PORT=$VITE_PORT \
    VITE_DOCS_BASE_URL=$VITE_DOCS_BASE_URL

RUN npm run build

# =========================
# Runtime Stage
# =========================
FROM nginx:alpine

WORKDIR /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from builder stage
COPY --from=builder /app/dist .

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
