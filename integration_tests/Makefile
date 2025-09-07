.PHONY: init up down logs tests verify

up:
	docker compose up --build -d

down:
	docker compose down --volumes

logs:
	docker compose logs -f

init:
	./scripts/setup_keycloak.sh

verify:
	./scripts/verify_user_token.sh

tests:
	docker compose run --rm tests

clean:
	docker compose down --volumes --rmi all
