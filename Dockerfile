# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy dependency manifests first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /spotisafe \
    ./cmd/spotisafe

# Stage 2: Minimal runtime image
FROM scratch

# TLS certificates (required for HTTPS to Spotify API)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binary
COPY --from=builder /spotisafe /spotisafe

ENTRYPOINT ["/spotisafe"]
