# Stage 1: Builder
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy dependencies first (for local replace directives in go.mod)
COPY scale-tests/sdk/ /sdk/
COPY runai-clients/ /runai-clients/

# Copy go mod files and vendor directory
COPY scale-tests/unified_tests/go.mod scale-tests/unified_tests/go.sum ./
COPY scale-tests/unified_tests/vendor/ ./vendor/

# Copy source code
COPY scale-tests/unified_tests/cmd/ ./cmd/
COPY scale-tests/unified_tests/internal/ ./internal/

# Build binaries using vendor directory
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w" -o bin/workload-scale cmd/workload-scale/main.go && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w" -o bin/report-generator cmd/report-generator/main.go && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w" -o bin/clean-cluster cmd/clean-cluster/main.go

# Stage 2: Runtime
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata bash

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/bin/ ./bin/

# Copy SLO defaults and configuration
COPY scale-tests/unified_tests/metrics/processed/slo_defaults.json ./config/

# Copy entrypoint script
COPY --chmod=755 scale-tests/unified_tests/docker/entrypoint.sh ./

# Create results directory
RUN mkdir -p /app/results

# Run as non-root user
RUN addgroup -S testrunner && adduser -S testrunner -G testrunner && \
    chown -R testrunner:testrunner /app

USER testrunner

ENTRYPOINT ["/app/entrypoint.sh"]
