# HyperFleet E2E Testing Framework
#
# Build: podman build -t quay.io/hyperfleet/hyperfleet-e2e:latest .
# Build with commit: podman build --build-arg GIT_COMMIT=$(git rev-parse HEAD) -t quay.io/hyperfleet/hyperfleet-e2e:latest .
# Run:   podman run --rm -e HYPERFLEET_API_URL=<url> quay.io/hyperfleet/hyperfleet-e2e:latest test

ARG BASE_IMAGE=registry.access.redhat.com/ubi9/go-toolset

# Build stage
FROM golang:1.25 AS builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends make curl && rm -rf /var/lib/apt/lists/*

# Download kubectl (stable version)
RUN mkdir -p /build/bin && \
 curl -fsSL "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
 -o /build/bin/kubectl

# Copy source code
COPY . .

# Build binary using make to include commit and build date
ARG GIT_COMMIT=unknown
RUN make build GIT_COMMIT=${GIT_COMMIT}

RUN chmod +x  /build/bin/*  

# Runtime stage
FROM ${BASE_IMAGE}

# Install runtime dependencies
USER root
RUN dnf -y install jq gettext && dnf clean all

WORKDIR /e2e

# Copy binary from builder (make build outputs to bin/)
COPY --from=builder /build/bin/* /usr/local/bin/

# Copy test payloads and fixtures
COPY --from=builder /build/testdata /e2e/testdata

# Copy default config (fallback if ConfigMap is not mounted)
COPY --from=builder /build/configs /e2e/configs

ENTRYPOINT ["/usr/local/bin/hyperfleet-e2e"]
CMD ["test", "--help"]

LABEL name="hyperfleet-e2e" \
      vendor="Red Hat" \
      summary="HyperFleet E2E Testing Framework" \
      description="End to end testing for HyperFleet cluster lifecycle management"
