run:
	go run cmd/main.go

up:
	docker-compose up -d --build

.PHONY: run up