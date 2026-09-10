.PHONY: run build test tidy up down logs

run:
	go run ./cmd/app

build:
	go build -o bin/app ./cmd/app

test:
	go test ./... -v -race -cover

tidy:
	go mod tidy

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f app
