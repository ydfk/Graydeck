# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS web-builder
WORKDIR /workspace

COPY pnpm-workspace.yaml package.json pnpm-lock.yaml ./
COPY web/package.json ./web/package.json

RUN corepack enable && pnpm install --frozen-lockfile

COPY web ./web
RUN pnpm --filter graydeck-web build

FROM golang:1.24-alpine AS server-builder
WORKDIR /workspace

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
COPY --from=web-builder /workspace/web/dist ./internal/webui/dist

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=0.0.0-dev
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags "-X mihomo-manager/internal/buildinfo.Version=${VERSION}" -o /out/managerd ./cmd/managerd

FROM alpine:3.21 AS runtime
WORKDIR /opt/graydeck

RUN apk add --no-cache ca-certificates tzdata

COPY --from=server-builder /out/managerd /usr/local/bin/managerd
COPY config /config

ENV GRAYDECK_SECRET=graydeck-secret

EXPOSE 8080
EXPOSE 17890/tcp
EXPOSE 17890/udp
EXPOSE 17891/tcp
EXPOSE 17892/tcp
EXPOSE 17893/tcp
EXPOSE 17893/udp
VOLUME ["/data", "/config"]

ENTRYPOINT ["/usr/local/bin/managerd"]
