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

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/managerd ./cmd/managerd

FROM alpine:3.21 AS runtime
WORKDIR /opt/graydeck

RUN apk add --no-cache ca-certificates tzdata

COPY --from=server-builder /out/managerd /usr/local/bin/managerd
COPY --from=web-builder /workspace/web/dist /opt/graydeck/web

ENV GRAYDECK_SECRET=graydeck-secret

EXPOSE 18080
EXPOSE 7890/tcp
EXPOSE 7890/udp
EXPOSE 7891/tcp
EXPOSE 7892/tcp
EXPOSE 7893/tcp
EXPOSE 7893/udp
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/managerd"]
