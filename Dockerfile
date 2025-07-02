# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/go:latest AS builder
WORKDIR /app
COPY . .
ARG VERSION=dev
RUN go build -ldflags "-X main.Version=${VERSION}" -o easy-dca ./cmd/easy-dca

FROM cgr.dev/chainguard/static:latest
COPY --from=builder /app/easy-dca /easy-dca
ENTRYPOINT ["/easy-dca"] 