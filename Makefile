# Pull in .env so make and the Go app read the exact same values.
# The leading "-" means "don't blow up if .env is missing" (fresh clone, CI).
-include .env

ifeq ($(wildcard .env),)
$(warning .env not found - falling back to defaults. Run: cp .env.example .env)
endif

# "?=" means "only set this if it isn't already set", so .env always wins.
DB_HOST     ?= 127.0.0.1
DB_PORT     ?= 5432
DB_NAME     ?= mydb
DB_USER     ?= user
DB_PASSWORD ?= 12345
DB_SSLMODE  ?= disable
DB_TIMEZONE ?= UTC

DB_URL ?= postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)&timezone=$(DB_TIMEZONE)

MIGRATION_FOLDER ?= db/migrations
# DB_URL="host=127.0.0.1 user=user password=12345 dbname=mydb port=5432 sslmode=disable TimeZone=UTC"
# DB_URL="postgresql://user:12345@127.0.0.1:5432/mydb?sslmode=disable&timezone=UTC"

# create a new migration 
migrate-create:  # command: gmake migrate-create name="create_entity_table"
	goose -dir $(MIGRATION_FOLDER) create $(name) sql
migrate-up:      # command: gmake migrate-up
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" up
migrate-down:    # command: gmake migrate-down
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" down
migrate-status:  # command: gmake migrate-status
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" status
migrate-reset:   # command: gmake migrate-reset
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" reset
migrate-version: # command: gmake migrate-version version=1
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" version $(version)
migrate-up-to:   # command: gmake migrate-up-to version=1
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" up-to $(version)
migrate-down-to: # command: gmake migrate-down-to version=1
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" down-to $(version)
migrate-fix:     # command: gmake migrate-fix version=1
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" fix $(version)
migrate-validate: # command: gmake migrate-validate
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" validate
migrate-verbose: # command: gmake migrate-verbose
	goose -dir $(MIGRATION_FOLDER) -v postgres "$(DB_URL)" up
migrate-redo:    # command: gmake migrate-redo
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" redo
migrate-to:	  # command: gmake migrate-to version=1
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" to $(version)
migrate-down-to: # command: gmake migrate-down-to version=1
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" down-to $(version)
migrate-force:    # command: gmake migrate-force version=1
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" force $(version)
migrate-help:     # command: gmake migrate-help
	goose -h
migrate-up-file:   # command: gmake migrate-up-file file=20260212174735_create_user_table.sql
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" up-to $(shell basename "$(file)" | grep -oE '^[0-9]+')
migrate-down-file: # command: gmake migrate-down-file file=20260212174735_create_user_table.sql
	goose -dir $(MIGRATION_FOLDER) postgres "$(DB_URL)" down-to $(shell basename "$(file)" | grep -oE '^[0-9]+')

