# syntax=docker/dockerfile:1
FROM golang:1.27.1-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/media-worker ./cmd/media-worker && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/upload ./cmd/upload && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && \
    rm -rf /var/lib/apt/lists/* && groupadd --gid 10001 courses && useradd --uid 10001 --gid courses --no-create-home courses
USER 10001:10001
WORKDIR /app

FROM runtime AS api
COPY --from=build /out/api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]

FROM runtime AS migrate
COPY --from=build /out/migrate /app/migrate
COPY --chown=10001:10001 migrations /app/migrations
ENV MIGRATIONS_DIR=/app/migrations
ENTRYPOINT ["/app/migrate"]

FROM runtime AS upload
USER root
RUN mkdir -p /media/source /media/hls /media/thumbnails /media/temporary && chown -R 10001:10001 /media
USER 10001:10001
COPY --from=build /out/upload /app/upload
EXPOSE 8082
ENTRYPOINT ["/app/upload"]

FROM runtime AS worker
USER root
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg && rm -rf /var/lib/apt/lists/* && \
    mkdir -p /media/source /media/hls /media/thumbnails /media/temporary && chown -R 10001:10001 /media
USER 10001:10001
COPY --from=build /out/media-worker /app/media-worker
EXPOSE 8081
ENTRYPOINT ["/app/media-worker"]
