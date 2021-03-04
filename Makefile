# build with verison infos
buildDate = $(shell TZ=UTC date +%FT%T%z)
gitCommit = $(shell git log --pretty=format:'%h' -1)

ldflags="-w -X main.gitTag=${gitTag} -X main.buildDate=${buildDate} -X main.gitCommit=${gitCommit}"

.PHONY: build vendorbuild clean fmt

default: build

build: clean fmt
	go build -o nebula-console --tags netgo -ldflags ${ldflags}

vendorbuild: clean fmt
	@go mod vendor && go build -mod vendor -o nebula-console --tags netgo -ldflags ${ldflags}

clean:
	@rm -rf nebula-console vendor

fmt:
	@go mod tidy && find . -path vendor -prune -o -type f -iname '*.go' -exec go fmt {} \;
