FROM gcr.io/distroless/base@sha256:57c1e4c72feb5925c4763ae4f6bd2013ad3854f57eff5b60dd9acb1ce0abc66e

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
