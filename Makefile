
.PHONY: build vendorbuild clean fmt

default: build

build: clean fmt
	go build -o nebula-console --tags netgo

vendorbuild: clean fmt
	@go mod vendor && go build -mod vendor -o nebula-console

clean:
	@rm -rf nebula-console vendor

fmt:
	@go mod tidy && find . -path vendor -prune -o -type f -iname '*.go' -exec go fmt {} \;
