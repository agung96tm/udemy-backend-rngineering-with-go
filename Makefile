SHELL := /bin/bash

ifneq (,$(wildcard .env))
	include .env
	export
endif

MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: migration
migration:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migration name=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(name)

.PHONY: migrate-up
migrate-up:
	@echo "Running migrations UP..."
	@set -a; source .env; set +a; \
	migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@echo "Running migrations DOWN..."
	@set -a; source .env; set +a; \
	if [ -n "$(steps)" ]; then \
		migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(steps); \
	else \
		migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down; \
	fi


.PHONY: migrate-force
migrate-force:
	@echo "Running migrations Force..."
	@set -a; source .env; set +a; \
	if [ -n "$(force)" ]; then \
		migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) force $(force); \
	else \
		migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) force; \
	fi

.PHONY: seed
seed:
	@echo "Running Seed..."
	@go run cmd/migrate/seeds/main.go


.PHONY: gen-docs
gen-docs:
	@echo "Generate Docs..."
	@swag init -g ./api/main.go -d cmd,internal && swag fmt


.PHONY: test
test:
	@go test -count=1 -v ./...
