# Build context: apps/orbit-demo/
#   docker build --build-arg VERSION=$(git rev-parse --short HEAD) -t orbit-demo:local apps/orbit-demo

# golang:1.27.1 (multi-arch index digest)
FROM golang:1.27.1@sha256:e0174e51e81218523251d85d248a90d24c3d5e81543b4f07a5d66229397db190 AS build

WORKDIR /src

# Module downloads are a cached layer, rebuilt only when dependencies change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION}" \
      -o /out/orbit-demo .

# gcr.io/distroless/static-debian13:nonroot (multi-arch index digest)
FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3

ARG VERSION=dev
LABEL org.opencontainers.image.source="https://github.com/azizmabrouki/Orbit" \
      org.opencontainers.image.revision="${VERSION}" \
      org.opencontainers.image.title="orbit-demo" \
      org.opencontainers.image.description="Orbit sample workload"

COPY --from=build /out/orbit-demo /orbit-demo

# Numeric, so Kubernetes can enforce runAsNonRoot without a name lookup.
USER 65532:65532
EXPOSE 8080
# Exec form: the binary is PID 1 and receives SIGTERM directly.
ENTRYPOINT ["/orbit-demo"]
