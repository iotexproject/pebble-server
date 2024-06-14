NAME=$(shell basename $$PWD)
FEATURE=$(shell git rev-parse --abbrev-ref HEAD)
VERSION=$(shell git describe --tags --always)
COMMITID=$(shell git rev-parse --short --show HEAD)
DATE=$(shell TZ=Asia/Shanghai date +%Y%m%d%H%M%S)

VERSION_PATH=main

LDFLAGS="-s -w -X ${VERSION_PATH}.Name=${NAME}    \
	-X ${VERSION_PATH}.Feature=${FEATURE}   \
	-X ${VERSION_PATH}.Version=${VERSION}   \
	-X ${VERSION_PATH}.CommitID=${COMMITID} \
	-X ${VERSION_PATH}.Date=${DATE} "

build: clean
	go build -ldflags ${LDFLAGS} -o ${NAME}

run: build
	./example

clean:
	@rm -rf example
