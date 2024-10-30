FROM --platform=linux/amd64 golang:1.22-alpine AS builder

ENV GO111MODULE=on

RUN apk update && apk upgrade && apk add --no-cache ca-certificates tzdata musl-dev gcc && update-ca-certificates

WORKDIR /go/src
COPY ./ ./

RUN cd ./cmd/sequencer && CGO_ENABLED=1 CGO_LDFLAGS='-L./lib/linux-x86_64 -lioConnectCore' go build -ldflags "-s -w -extldflags '-static'" -o pebble-sequencer

FROM --platform=linux/amd64 alpine:3.20 AS runtime

ENV LANG en_US.UTF-8

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /go/src/cmd/sequencer/pebble-sequencer /go/bin/pebble-sequencer
EXPOSE 9000

WORKDIR /go/bin
ENTRYPOINT ["/go/bin/pebble-sequencer"]
