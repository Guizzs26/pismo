.PHONY: run down restart build test test-unit test-integration logs logs-app logs-db db docs

run:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

test:
	go test ./... -v -count=1

test-unit:
	go test ./internal/... -v -count=1

test-integration:
	go test ./tests/integration/... -v -count=1

logs:
	docker compose logs -f

logs-app:
	docker compose logs -f app

logs-db:
	docker compose logs -f db

db:
	docker exec -it pismo_db psql -U pismousr -d pismodb

docs:
	swag init -g cmd/api/main.go -o docs

build: 
	docker compose build