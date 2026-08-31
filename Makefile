.PHONY: help build run migrate-up migrate-down docker-up docker-down clean

help:
	@echo "Available targets:"
	@echo "  make build         - Build the API binary"
	@echo "  make run           - Run the API server (requires MySQL running)"
	@echo "  make docker-up     - Start MySQL container with docker-compose"
	@echo "  make docker-down   - Stop MySQL container"
	@echo "  make migrate-up    - Apply database migrations"
	@echo "  make migrate-down  - Rollback database migrations"
	@echo "  make clean         - Remove binary and build artifacts"

build:
	go build -o bin/api ./cmd/api

run: build
	./bin/api

docker-up:
	docker-compose up -d
	@echo "Waiting for MySQL to be ready..."
	sleep 5

docker-down:
	docker-compose down

migrate-up: docker-up
	@for f in migrations/*.up.sql; do \
		echo "Applying $$f ..."; \
		mysql -h 127.0.0.1 -u root -proot -D mangea < $$f || exit 1; \
	done
	@echo "Migrations applied successfully"

migrate-down:
	@for f in $(shell ls -r migrations/*.down.sql); do \
		echo "Rolling back $$f ..."; \
		mysql -h 127.0.0.1 -u root -proot -D mangea < $$f || exit 1; \
	done
	@echo "Migrations rolled back successfully"

clean:
	rm -rf bin/
	go clean
