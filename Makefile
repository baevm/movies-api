include .env
 
## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: runs app
.PHONY: run
run:
	go run ./cmd/api -db-dsn=${DB_DSN}

## db/migrate: executes all migrations from migrations folder
.PHONY: db/migrate
db/migrate:
	migrate -path=./migrations -database=${DB_DSN} up

## db/migration name=$1: creates new migration with given name
.PHONY: db/migration
db/migration:
	migrate create -seq -ext .sql -dir ./migrations ${name}

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...