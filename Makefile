BIN=bin
BIN_NAME=gonaomi

all: darwin

fmt:
	go fmt ./...

install:
	env CGO_ENABLED=0 go install

vendor:
	glide install -v --strip-vcs

build: vendor $(BIN) $(shell find . -name "*.go")
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $(BIN)/$(BIN_NAME) .

darwin: vendor $(BIN) $(shell find . -name "*.go")
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $(BIN)/$(BIN_NAME)-darwin .

freebsd: vendor $(BIN) $(shell find . -name "*.go")
	env CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $(BIN)/$(BIN_NAME)-freebsd .

clean:
	go clean -i
	rm -rf $(BIN)
	rm -rf vendor

test:
	go test -v ./...

$(BIN):
	mkdir -p $(BIN)

.PHONY: fmt install clean test all release build darwin freebsd
