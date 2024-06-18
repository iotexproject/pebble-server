FROM golang:1.22 AS builder

ENV GO111MODULE=on

WORKDIR /go/src
COPY ./ ./

RUN cd ./cmd/sequencer && make build

FROM --platform=linux/amd64 scratch AS runtime

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/cmd/sequencer/pebble-sequencer /go/bin/pebble-sequencer
EXPOSE 9001

WORKDIR /go/bin
ENTRYPOINT ["/go/bin/pebble-sequencer", "run"]
