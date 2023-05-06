include .env
 
NAME = "migration"

migrate:
	migrate -path=./migrations -database=${DB_DSN} up

run:
	go run ./cmd/api

migration:
	migrate create -seq -ext .sql -dir ./migrations "${NAME}"