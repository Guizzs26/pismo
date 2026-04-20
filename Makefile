.PHONY: run down build test logs logs-app logs-db db

build:
	docker compose build

run:
	docker compose up -d

down:
	docker compose down

test:
	go test ./... -v -count=1

logs:
	docker compose logs -f

logs-app:
	docker compose logs -f app

logs-db:
	docker compose logs -f db

db:
	docker exec -it pismo_db psql -U pismousr -d pismodb