.PHONY: run migrate test test-cover lint docker-build clean

run:
	go run cmd/server/main.go

migrate:
	psql -h localhost -U spark -d spark -f migrations/001_init.sql

test:
	go test ./internal/... -count=1

test-cover:
	go test ./internal/... -coverprofile=/tmp/spark-cover.out -count=1
	go tool cover -func=/tmp/spark-cover.out

test-html: test-cover
	go tool cover -html=/tmp/spark-cover.out

lint:
	go vet ./...

docker-build:
	docker build -t spark-server .

docker-run:
	docker run -p 8080:8080 \
		-e DB_HOST=host.docker.internal \
		-e REDIS_ADDR=host.docker.internal:6379 \
		spark-server

seed-bots:
	go run scripts/generate_bots.go

seed-testdata:
	go run scripts/seed_test_data.go

fmt:
	go fmt ./...

clean:
	rm -f /tmp/spark-server /tmp/spark-cover.out
