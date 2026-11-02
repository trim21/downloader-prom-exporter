FROM gcr.io/distroless/base@sha256:075a33d9c49026ca7ee72738b4c4a1caea55d0f44227356e46a611a211050e20

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
