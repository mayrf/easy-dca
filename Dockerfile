# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/go:latest AS builder
WORKDIR /app
COPY . .
ARG VERSION=dev
# Ensure a fully static build
ENV CGO_ENABLED=0
RUN go build -ldflags "-X main.Version=${VERSION} -extldflags '-static'" -o easy-dca ./cmd/easy-dca

FROM cgr.dev/chainguard/static:latest
# Add non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/easy-dca /easy-dca
USER appuser
ENTRYPOINT ["/easy-dca"] 