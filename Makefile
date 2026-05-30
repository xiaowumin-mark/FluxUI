.PHONY: test vet visual fmt lint build clean

test:
	go test ./...

vet:
	go vet ./...

visual:
	go test -tags visual ./examples/material3_showcase -run TestMaterial3ShowcaseScreenshots -count=1

fmt:
	gofmt -w ./

lint:
	golangci-lint run

build:
	go build ./...

clean:
	go clean ./...
