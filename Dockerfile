FROM golang:1.18-alpine as builder

COPY . /usr/src

RUN cd /usr/src && apk add --no-cache git make && make

FROM alpine

COPY --from=builder /usr/src/nebula-console /usr/local/bin/nebula-console

COPY --from=builder /usr/src/data/ /data/

ENTRYPOINT ["nebula-console"]
