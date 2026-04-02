FROM gcr.io/distroless/base@sha256:b0510424f0c7c1d6fdae75ef5c1d349fa72d312e96f69728fad6beb04755b8b4

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
