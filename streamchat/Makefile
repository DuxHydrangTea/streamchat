.PHONY: test build dev up down

test:
	cd backend && go test ./...
	cd frontend && npm test -- --run

build:
	cd backend && go build ./...
	cd frontend && npm run build

dev:
	docker compose up --build

up:
	docker compose up -d --build

down:
	docker compose down
