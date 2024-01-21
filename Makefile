include .env

MG_PATH= db/migrations
DB_DSN =postgresql://${PG_USER}:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/${PG_DB}?sslmode=disable

.PHONY: migrate-up 
migrate-up: 
	migrate -path $(MG_PATH) -database $(DB_DSN) up

.PHONY: migrate-down 
migrate-down:
	migrate -path $(MG_PATH) -database $(DB_DSN) down

.PHONY: migrate-create 
migrate-create:
	migrate create -dir $(MG_PATH) -seq -ext .sql $(name)

.PHONY: up
up:
	docker compose --env-file ./.env up -d
	migrate -path $(MG_PATH) -database $(DB_DSN) up

.PHONY: down
down: 
	docker compose --env-file ./.env down 