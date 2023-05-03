include .env

migrate:
	migrate -path=./migrations -database=${DB_DSN} up

run:
	go run ./cmd/api