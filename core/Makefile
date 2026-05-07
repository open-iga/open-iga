.PHONY: run test lint pretty migration-generate sqlc-generate

run:
	air

test:
	go test ./...

lint:
	golangci-lint run

pretty:
	gofumpt -l -w .

migration-generate:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 create -ext sql -dir internal/repository/migrations -seq $(name)

sqlc-generate:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 && \
	sqlc generate