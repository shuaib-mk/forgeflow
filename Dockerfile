# syntax=docker/dockerfile:1.7
FROM golang:1.24.4-alpine3.21 AS go-builder
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/forgeflow-api ./cmd/forgeflow-api && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/forgeflow-worker ./cmd/forgeflow-worker && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/forgeflow-migrate ./cmd/forgeflow-migrate && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/forgeflow ./cmd/forgeflow

FROM alpine:3.21.3 AS runtime
RUN apk add --no-cache ca-certificates git && addgroup -S forgeflow && adduser -S -G forgeflow forgeflow
WORKDIR /app
COPY deployments/migrations ./deployments/migrations
RUN mkdir -p /data/workspaces && chown -R forgeflow:forgeflow /app /data
USER forgeflow

FROM runtime AS api
COPY --from=go-builder /out/forgeflow-api /usr/local/bin/forgeflow-api
EXPOSE 8080
ENTRYPOINT ["forgeflow-api"]

FROM runtime AS worker
COPY --from=go-builder /out/forgeflow-worker /usr/local/bin/forgeflow-worker
ENTRYPOINT ["forgeflow-worker"]

FROM runtime AS migrate
COPY --from=go-builder /out/forgeflow-migrate /usr/local/bin/forgeflow-migrate
ENTRYPOINT ["forgeflow-migrate"]

FROM scratch AS cli
COPY --from=go-builder /out/forgeflow /forgeflow
ENTRYPOINT ["/forgeflow"]

