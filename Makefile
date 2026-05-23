.PHONY: test vet fmt lint build clean

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./

lint:
	golangci-lint run

build:
	go build ./...

clean:
	go clean ./...
