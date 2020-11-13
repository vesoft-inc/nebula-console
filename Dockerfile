FROM golang:1.13.2-alpine as builder

COPY . /usr/src

RUN cd /usr/src && go build

FROM alpine

COPY --from=builder /usr/src/nebula-console /usr/local/bin/nebula-console

COPY --from=builder /usr/src/data/ /data/

ENTRYPOINT ["nebula-console"]
