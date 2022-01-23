FROM gcr.io/distroless/base

COPY dist/app /app/my-site-proxy

ENTRYPOINT ["/app/my-site-proxy"]
