FROM --platform=linux/amd64 golang:1.22 AS builder

ENV GO111MODULE=on

WORKDIR /go/src
COPY ./ ./

RUN cd ./cmd/sequencer && make build

FROM --platform=linux/amd64 scratch AS runtime

COPY --from=builder /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=builder /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/cmd/sequencer/pebble-sequencer /go/bin/pebble-sequencer
EXPOSE 80

WORKDIR /go/bin
ENTRYPOINT ["/go/bin/pebble-sequencer"]
