# HyperFleet E2E Testing Framework
#
# Build: podman build -t quay.io/hyperfleet/hyperfleet-e2e:latest .
# Build with commit: podman build --build-arg GIT_COMMIT=$(git rev-parse HEAD) -t quay.io/hyperfleet/hyperfleet-e2e:latest .
# Run:   podman run --rm -e HYPERFLEET_API_URL=<url> quay.io/hyperfleet/hyperfleet-e2e:latest test

ARG BASE_IMAGE=gcr.io/distroless/static-debian12:nonroot

# Build stage
FROM golang:1.25-alpine AS builder

# Build arguments passed from build machine
ARG GIT_COMMIT=unknown

# Install build dependencies
RUN apk add --no-cache make curl

WORKDIR /build

# Copy source code
COPY . .

# Build binary using make to include commit and build date
RUN make build GIT_COMMIT=${GIT_COMMIT}

# Runtime stage
FROM ${BASE_IMAGE}

WORKDIR /app

# Copy binary from builder (make build outputs to bin/)
COPY --from=builder /build/bin/hyperfleet-e2e /app/hyperfleet-e2e

# Copy test payloads and fixtures
COPY --from=builder /build/testdata /app/testdata

# Copy default config (fallback if ConfigMap is not mounted)
COPY --from=builder /build/configs /app/configs

ENTRYPOINT ["/app/hyperfleet-e2e"]
CMD ["test", "--help"]

LABEL name="hyperfleet-e2e" \
      vendor="Red Hat" \
      summary="HyperFleet E2E Testing Framework" \
      description="End to end testing for HyperFleet cluster lifecycle management"
