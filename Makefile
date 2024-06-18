#PACKAGES=$(shell go list ./... | grep -E -v 'example|proto|testdata|mock')
PACKAGES=$(shell go list ./... | grep -E -v 'pb$|testdata|mock|proto|example')

debug:
	@echo ${PACKAGES}

xgo:
	@if [ "${XGO_OK}" != "0" ]; then \
		echo "installing xgo for unit test"; \
		go install github.com/xhd2015/xgo/cmd/xgo@latest; \
	fi

tidy:
	@go mod tidy

cover: xgo tidy vet
	@xgo test -failfast -parallel 1 -gcflags="all=-N -l" ${PACKAGES} -covermode=count -coverprofile=cover.out

test: tidy vet
	@xgo test -race -failfast -parallel 1 -gcflags="all=-N -l" ${PACKAGES}

vet:
	@go vet ${PACKAGES}

report:
	@echo ">>>static checking"
	@go vet ./...
	@echo "done\n"
	@echo ">>>detecting ineffectual assignments"
	@ineffassign ./...
	@echo "done\n"
	@echo ">>>detecting icyclomatic complexities over 10 and average"
	@gocyclo -over 10 -avg -ignore '_test|vendor' . || true
	@echo "done\n"

check: tidy vet test

sequencer_image:
	@cd cmd/sequencer && make image

