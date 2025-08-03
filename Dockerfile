FROM gcr.io/distroless/base@sha256:007fbc0e0df2f12b739e9032a45ade4c58be0c9981767c440da6c404418f3144

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
