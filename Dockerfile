FROM gcr.io/distroless/base@sha256:74ddbf52d93fafbdd21b399271b0b4aac1babf8fa98cab59e5692e01169a1348

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
