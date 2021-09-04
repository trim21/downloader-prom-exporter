FROM alpine

ENV production=1

WORKDIR /work/

COPY dist/app /work/app

ENTRYPOINT ["/work/app"]
