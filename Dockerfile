FROM gcr.io/distroless/base@sha256:8c8b7cf2a01e2d1c683128b2488d77139fa90ec8cb807f0ae260d57f7022dedd

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
