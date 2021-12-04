FROM alpine

ENV production=1

COPY dist/app /work/my-site-proxy

ENTRYPOINT ["/work/my-site-proxy"]
