.PHONY: test vet visual fmt fmt-check lint race coverage api-snapshot api-check deps-verify vuln bench build clean

test:
	go test ./...

vet:
	go vet ./...

visual:
	go test -tags visual ./examples/material3_showcase -run TestMaterial3ShowcaseScreenshots -count=1

fmt:
	go run ./tools/gofmtcheck -write

fmt-check:
	go run ./tools/gofmtcheck

lint:
	golangci-lint run

race:
	go test -race -count=1 ./event ./internal ./internal/collection ./internal/fieldstate ./internal/overlay ./internal/testkit ./router ./state ./ui ./widget

coverage:
	mkdir -p coverage
	while IFS= read -r package; do case "$$package" in \#*|'') continue;; esac; name=$$(printf '%s' "$$package" | sed 's|^\./||; s|/|-|g'); go test -covermode=atomic -coverprofile "coverage/$$name.out" "$$package" || exit $$?; done < ci/core-packages.txt
	go run ./tools/coveragecheck -profiles coverage -baseline ci/coverage-baseline.json -report coverage/summary.json

api-snapshot:
	go run ./tools/api-snapshot -write

api-check:
	go run ./tools/api-snapshot -check

deps-verify:
	go mod verify

vuln:
	govulncheck ./...

bench:
	go test ./widget ./internal/perf -run '^$$' -bench . -benchmem -benchtime=1s -count=5

build:
	go build ./...

clean:
	go clean ./...
