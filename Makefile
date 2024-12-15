MIGRATION_FOLDER=$(CURDIR)/internal/storage/db/migrations

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "migration" sql


# run in docker
.PHONY: docker-run
docker-run:
	docker compose up --build

# build app
.PHONY: build
build:
	go mod download && CGO_ENABLED=0  go build \
		-o ./bin/main$(shell go env GOEXE) ./cmd/main.go
