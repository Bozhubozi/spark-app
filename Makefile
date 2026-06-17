.PHONY: dev dev-backend dev-flutter test test-all lint docker-up docker-down clean

# Start full development environment
dev:
	@echo "Starting PostgreSQL + Redis..."
	docker compose up -d postgres redis
	@echo "Starting backend..."
	cd backend && go run cmd/server/main.go &
	@sleep 2
	@echo "Starting Flutter web..."
	cd flutter && flutter run -d web-server --web-port=8081 \
		--dart-define=API_BASE_URL=http://localhost:8080 \
		--dart-define=WS_URL=ws://localhost:8080/ws &
	@echo ""
	@echo "Backend:  http://localhost:8080"
	@echo "Flutter:  http://localhost:8081"
	@echo "Health:   http://localhost:8080/health"
	@echo "Metrics:  http://localhost:8080/metrics"

dev-backend:
	cd backend && go run cmd/server/main.go

dev-flutter:
	cd flutter && flutter run -d web-server --web-port=8081 \
		--dart-define=API_BASE_URL=http://localhost:8080 \
		--dart-define=WS_URL=ws://localhost:8080/ws

# Run all tests
test:
	cd backend && go test ./internal/... -count=1
	cd flutter && flutter test

# Backend tests only
test-backend:
	cd backend && go test ./internal/... -count=1 -v

# Flutter tests only
test-flutter:
	cd flutter && flutter test

# Lint check
lint:
	cd backend && go vet ./internal/...
	cd flutter && flutter analyze

# Docker full stack
docker-up:
	docker compose up -d

docker-down:
	docker compose down

# Seed test data
seed:
	cd backend && go run scripts/generate_bots.go

# Clean build artifacts
clean:
	rm -f /tmp/spark-server /tmp/spark-cover.out
	cd flutter && flutter clean
