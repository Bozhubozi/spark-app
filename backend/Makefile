.PHONY: run migrate test

run:
	go run cmd/server/main.go

migrate:
	psql -h localhost -U spark -d spark -f migrations/001_init.sql

test:
	go test ./...
