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
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/asr-worker ./cmd/asr-worker

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

FROM runtime AS asr-worker
USER root
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg python3 python3-pip python3-venv && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /media/source /media/hls /media/thumbnails /media/temporary && \
    chown -R 10001:10001 /media
RUN python3 -m venv /opt/whisper && \
    /opt/whisper/bin/pip install --no-cache-dir faster-whisper==1.2.1
ENV HF_HOME=/opt/whisper/models \
    WHISPER_MODEL_DIR=/opt/whisper/models/large-v3-turbo
RUN /opt/whisper/bin/python3 -c \
    "from faster_whisper.utils import download_model; download_model('large-v3-turbo', output_dir='$WHISPER_MODEL_DIR')" && \
    test -f "$WHISPER_MODEL_DIR/model.bin" && \
    chmod -R a+rX /opt/whisper
USER 10001:10001
COPY --from=build /out/asr-worker /app/asr-worker
COPY internal/transcript/adapters/whisper/transcribe.py /opt/whisper/transcribe.py
EXPOSE 8083
ENTRYPOINT ["/app/asr-worker"]
