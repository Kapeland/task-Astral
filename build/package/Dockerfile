# Builder

FROM golang:1.22-alpine AS builder
RUN apk add --update curl git make && apk cache clean

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

RUN mkdir -p file-storage/json

COPY --from=builder /home/${MODULE_NAME}/internal/storage/db/migrations ./postgres/migrations

RUN chown root:root main

CMD ["sh", "-c", " ./main"]