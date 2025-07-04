# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/go:latest AS builder
WORKDIR /app
COPY . .
ARG VERSION=dev
# Ensure a fully static build
ENV CGO_ENABLED=0
RUN go build -ldflags "-X main.Version=${VERSION} -extldflags '-static'" -o easy-dca ./cmd/easy-dca

FROM cgr.dev/chainguard/static:latest
COPY --from=builder /app/easy-dca /easy-dca
USER nonroot
ENTRYPOINT ["/easy-dca"] 