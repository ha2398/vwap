GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOFMT=gofmt
GOTEST=$(GOCMD) test
GOLINT=golangci-lint run
GOTOOL=$(GOCMD) tool

all: install test

build:
	$(GOFMT) -w .
	$(GOLINT)
	$(GOBUILD) -o vwap .

install: build
	cp ./vwap $(GOPATH)/bin/

test:
	$(GOTEST) ./... -v -race -count=1 -coverprofile cover.out
	$(GOTOOL) cover -html=cover.out -o coverage.html

clean: 
	rm -f ./vwap
	$(GOCLEAN) -i github.com/cyralinc/vwap/...
