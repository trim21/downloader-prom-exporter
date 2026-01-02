FROM gcr.io/distroless/base@sha256:f5a3067027c2b322cd71b844f3d84ad3deada45ceb8a30f301260a602455070e

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
