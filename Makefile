LOCAL_BIN:=$(CURDIR)/bin
DEPLOY_DIR:=$(CURDIR)/deployments

CONFIG_PATH:=$(CURDIR)/configs/local-config.yaml
MIGRATION_FOLDER=$(CURDIR)/internal/storage/db/migrations

ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=postgres password=password dbname=db_Astral host=localhost port=5432 sslmode=disable
endif

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "migration" sql


.PHONY: db-migration-up
db-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

.PHONY: db-migration-down
db-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down


# run in docker
.PHONY: docker-run
docker-run:
	docker compose up --build postgres project

# build app
.PHONY: build
build:
	go mod download && CGO_ENABLED=0  go build \
		-o ./bin/main$(shell go env GOEXE) ./cmd/main.go
