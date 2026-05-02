FROM gcr.io/distroless/base@sha256:c83f022002fc917a92501a8c30c605efdad3010157ba2c8998a2cbf213299201

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
