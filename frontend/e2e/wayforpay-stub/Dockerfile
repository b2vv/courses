FROM golang:1.27.1-bookworm AS build
WORKDIR /src
COPY frontend/e2e/wayforpay-stub/main.go .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/wayforpay-stub main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/* && \
    useradd --uid 10001 --user-group --no-create-home stub
USER 10001:10001
COPY --from=build /out/wayforpay-stub /wayforpay-stub
EXPOSE 8080
ENTRYPOINT ["/wayforpay-stub"]
