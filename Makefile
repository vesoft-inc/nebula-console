name ?= nebula-console

# build with verison infos
buildDate = $(shell TZ=UTC date +%FT%T%z)
gitCommit = ${GITHUB_SHA::7}
gitCommit ?= $(shell git log --pretty=format:'%h' -1)

ldflags="-w -X main.gitTag=${gitTag} -X main.buildDate=${buildDate} -X main.gitCommit=${gitCommit}"

.PHONY: build vendorbuild clean fmt gen

default: build

build: clean fmt
	@CGO_ENABLED=0 go build -o ${name} -ldflags ${ldflags}

vendorbuild: clean fmt
	@CGO_ENABLED=0 go mod vendor && go build -mod vendor -o ${name} --tags netgo -ldflags ${ldflags}

clean:
	@rm -rf ${name} vendor

clean-all:
	@rm -rf ${name} vendor box/blob.go

fmt:
	@go mod tidy && find . -path vendor -prune -o -type f -iname '*.go' -exec go fmt {} \;

# generate box/blob.go(NOTE: `go generate` command may doesn't support cross-platform)
gen:
	@CGO_ENABLED=0 go generate ./...

