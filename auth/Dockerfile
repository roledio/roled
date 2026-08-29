# =========================
# Build Stage
# =========================
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

ARG GIT_COMMIT_HASH

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build


# =========================
# Runtime Stage
# =========================
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates wget

# The binary is built into tmp dir in the builder stage
COPY --from=builder /app/tmp/app .

# The configs and migrations directories are outside the tmp dir, 
# so we need to copy them separately
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/system/health || exit 1

CMD ["./app"]
