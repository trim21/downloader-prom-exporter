FROM gcr.io/distroless/base@sha256:9e9b50d2048db3741f86a48d939b4e4cc775f5889b3496439343301ff54cdba8

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
