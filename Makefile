.PHONY: up down logs run test clean help

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

run:
	go run cmd/api/main.go

test:
	go test ./... -v

clean:
	docker-compose down -v

help:
	@echo "StoreFlow API - Makefile Commands"
	@echo ""
	@echo "  make up       - Start Docker containers (PostgreSQL, Redis, API)"
	@echo "  make down     - Stop Docker containers"
	@echo "  make logs     - View Docker container logs"
	@echo "  make run      - Run API locally (requires Go)"
	@echo "  make test     - Run all tests"
	@echo "  make clean    - Stop containers and remove volumes"
	@echo ""
	@echo "For migrations on Windows, see README.md for PowerShell commands"
