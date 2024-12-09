# Builder

FROM golang:1.22-alpine AS builder
RUN apk add --update make git curl
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

ARG MODULE_NAME=github.com/Kapeland/task-Astral

COPY Makefile /home/${MODULE_NAME}/Makefile
COPY go.mod /home/${MODULE_NAME}/go.mod
COPY go.sum /home/${MODULE_NAME}/go.sum

WORKDIR /home/${MODULE_NAME}

COPY . /home/${MODULE_NAME}

RUN make build

# Service

FROM alpine:latest AS server
ARG MODULE_NAME=github.com/Kapeland/task-Astral
WORKDIR /root/

COPY --from=builder /home/${MODULE_NAME}/bin/main .
COPY --from=builder /home/${MODULE_NAME}/configs ./configs

RUN mkdir ./file-storage
RUN mkdir ./file-storage/json

COPY --from=builder /go/bin/goose ./goose
COPY --from=builder /home/${MODULE_NAME}/internal/storage/db/migrations ./postgres/migrations

RUN chown root:root main

ARG MIGRATION_FOLDER=postgres/migrations
ENV MIGRATION_FOLDER=${MIGRATION_FOLDER}
ARG POSTGRES_SETUP="user=postgres password=password dbname=db_Astral host=postgres port=5432 sslmode=disable"
ENV POSTGRES_SETUP=${POSTGRES_SETUP}


CMD ["sh", "-c", "./goose -dir ${MIGRATION_FOLDER} postgres \"${POSTGRES_SETUP}\" up && \
    ./main && \
    ./goose -dir ${MIGRATION_FOLDER} postgres \"${POSTGRES_SETUP}\" down"]