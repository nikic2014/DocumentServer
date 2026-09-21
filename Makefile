COMPOSE = docker compose
PSQL    = $(COMPOSE) exec -T postgres psql -U user -d document_server -v ON_ERROR_STOP=1

.PHONY: up down migrate rollback psql

up: ## Поднять БД
	$(COMPOSE) up -d postgres

down: ## Остановить
	$(COMPOSE) down

migrate: ## Применить миграции
	$(PSQL) -f /migrations/001_init.sql

rollback: ## Откатить миграции
	$(PSQL) -f /migrations/001_init_down.sql

psql: ## Зайти в psql
	$(COMPOSE) exec postgres psql -U user -d document_server