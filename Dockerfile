# Stage 1: Build
FROM golang:1.25-alpine AS builder

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
FROM alpine:3.21

RUN apk add --no-cache su-exec ca-certificates

COPY --from=builder /spotisafe /spotisafe
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/spotisafe"]
