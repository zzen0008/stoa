# Top-level Makefile to manage the docker-compose setup

.PHONY: up down logs

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f
