FROM golang:1.22 AS builder

ENV GO111MODULE=on

WORKDIR /go/src
COPY ./ ./

RUN cd ./cmd/sequencer && make build

FROM scratch AS runtime

COPY --from=builder /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/cmd/sequencer/pebble-sequencer /go/bin/pebble-sequencer
EXPOSE 80

WORKDIR /go/bin
ENTRYPOINT ["/go/bin/pebble-sequencer"]
