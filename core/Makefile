.PHONY: run test lint pretty

run:
	air

test:
	go test ./...

lint:
	golangci-lint run

pretty:
	gofumpt -l -w .