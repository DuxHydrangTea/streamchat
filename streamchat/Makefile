.PHONY: test build dev up down migrate

COMPOSE ?= docker compose

test:
	cd backend && go test ./...
	cd frontend && npm test -- --run

build:
	cd backend && go build ./...
	cd frontend && npm run build

dev:
	$(COMPOSE) up --build

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

migrate:
	$(COMPOSE) run --build --rm migrate
