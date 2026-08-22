FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

# Install build dependencies
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

FROM alpine:3.19

# Security: Install ca-certificates and curl for health probes
RUN apk add --no-cache ca-certificates curl tzdata

# Security: Create non-root runtime user and group
RUN addgroup -g 10001 -S appuser && \
    adduser -u 10001 -S appuser -G appuser

WORKDIR /app

# Create required persistent storage and artifact directories with appropriate ownership
RUN mkdir -p /app/state /app/artifacts /app/model /app/model/backup /app/model/candidates && \
    chown -R appuser:appuser /app

COPY --from=builder --chown=appuser:appuser /app/api /app/api
COPY --from=builder --chown=appuser:appuser /app/migrations /app/migrations

USER appuser:appuser

EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["/app/api"]
