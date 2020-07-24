FROM golang:1.14.2 as builder

COPY . /usr/src

RUN cd /usr/src && go build

FROM golang:1.14.2

COPY --from=builder /usr/src/nebula-console /usr/bin/nebula-console
