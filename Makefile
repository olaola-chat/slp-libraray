# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

fmt:
	gofmt -l -s -w .

lint:
	golangci-lint run ./...
	go vet ./...

clean:
	$(GOCLEAN)

.PHONY: test
test:
	$(GOTEST) -v -count=1 ./...
