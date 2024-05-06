name ?= nebula-console

# Set the default GOPROXY
GOPROXY ?= https://proxy.golang.org,direct

# build with version info
buildDate = $(shell TZ=UTC date +%FT%T%z)
gitCommit = ${GITHUB_SHA::7}
gitCommit ?= $(shell git log --pretty=format:'%h' -1)

ldflags="-w -X main.gitTag=${gitTag} -X main.buildDate=${buildDate} -X main.gitCommit=${gitCommit}"

.PHONY: build vendorbuild clean fmt gen

default: build

build: clean fmt
	@GOPROXY=$(GOPROXY) CGO_ENABLED=0 go build -o ${name} -ldflags ${ldflags}

vendorbuild: clean fmt
	@GOPROXY=$(GOPROXY) CGO_ENABLED=0 go mod vendor && go build -mod vendor -o ${name} --tags netgo -ldflags ${ldflags}

clean:
	@rm -rf ${name} vendor

clean-all:
	@rm -rf ${name} vendor box/blob.go

fmt:
	@GOPROXY=$(GOPROXY) go mod tidy && find . -path vendor -prune -o -type f -iname '*.go' -exec go fmt {} \;

# generate box/blob.go (NOTE: `go generate` command may not support cross-platform)
gen:
	@GOPROXY=$(GOPROXY) CGO_ENABLED=0 go generate ./...
