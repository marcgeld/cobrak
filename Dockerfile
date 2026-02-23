FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go modules
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version injection
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o cobrak main.go

# Multi-stage build - final image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cobrak /app/cobrak

# Create non-root user
RUN addgroup -g 1000 cobrak && \
    adduser -D -u 1000 -G cobrak cobrak && \
    chown -R cobrak:cobrak /app

USER cobrak

ENTRYPOINT ["/app/cobrak"]
CMD ["--help"]

LABELS \
    org.opencontainers.image.title="cobrak" \
    org.opencontainers.image.description="Kubernetes cluster analysis CLI" \
    org.opencontainers.image.url="https://github.com/marcgeld/cobrak" \
    org.opencontainers.image.source="https://github.com/marcgeld/cobrak" \
    org.opencontainers.image.vendor="Marcus Gelderman" \

