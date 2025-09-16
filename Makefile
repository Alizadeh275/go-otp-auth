.PHONY: build run docker
build:
	go build -o bin/server ./cmd/server
run:
	go run ./cmd/server
docker:
	docker compose up --build