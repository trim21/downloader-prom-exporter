FROM gcr.io/distroless/base@sha256:97406725e9ca912013f59ae49fa3362d44f2745c07eba00705247216225b810c

COPY dist/app /app/downloader-prom-exporter

ENTRYPOINT ["/app/downloader-prom-exporter"]
