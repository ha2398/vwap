GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOFMT=gofmt
GOTEST=$(GOCMD) test
GOLINT=golangci-lint run
GOTOOL=$(GOCMD) tool

IMAGE_NAME=ha2398/vwap
EXEC_NAME=vwap

# Run parameters.
FEED_ENDPOINT?=wss://ws-feed.exchange.coinbase.com
TRADING_PAIRS?=BTC-USD,ETH-USD,ETH-BTC
WINDOW_SIZE?=200

all: format install test

format:
	$(GOFMT) -w .
	$(GOLINT)

build:
	$(GOBUILD) -o $(EXEC_NAME) .

install: build
	cp ./$(EXEC_NAME) $(GOPATH)/bin/

test:
	$(GOTEST) ./... --tags=unit,integration -v -race -count=1 -coverprofile cover.out
	$(GOTOOL) cover -html=cover.out -o coverage.html

run:
	./$(EXEC_NAME) --feed-endpoint $(FEED_ENDPOINT) \
		--trading-pairs $(TRADING_PAIRS) \
		--window-size $(WINDOW_SIZE)

docker/build:
	docker build -t $(IMAGE_NAME) .

docker/run:
	docker run -i -t --name vwap --rm $(IMAGE_NAME) \
		--feed-endpoint $(FEED_ENDPOINT) \
		--trading-pairs $(TRADING_PAIRS) \
		--window-size $(WINDOW_SIZE)

clean: 
	rm -f ./$(EXEC_NAME)
	$(GOCLEAN) -i github.com/ha2398/vwap/...
	docker rmi -f $(IMAGE_NAME)