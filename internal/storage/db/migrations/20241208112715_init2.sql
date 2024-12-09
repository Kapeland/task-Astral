-- +goose Up
-- +goose StatementBegin
create schema if not exists users_schema;

drop table if exists users_schema.users cascade;

CREATE TABLE if not exists users_schema.users
(
    id            SERIAL PRIMARY KEY,
    login         TEXT unique not null,
    password_hash TEXT        NOT NULL
);

create schema if not exists auth_schema;

drop table if exists auth_schema.users_auth;
create table if not exists auth_schema.users_auth
(
    login       text      not null references users_schema.users (login),
    valid_until TIMESTAMP NOT NULL DEFAULT NOW(),
    token       text      not null unique
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

drop table if exists Documents cascade;
CREATE TABLE IF NOT EXISTS Documents
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title      TEXT      NOT NULL,
    content    TEXT      NOT NULL,
    mime       TEXT,
    file       BOOL             default true,
    owner      text      NOT NULL REFERENCES users_schema.users (login) ON DELETE CASCADE,
    is_public  BOOLEAN          DEFAULT FALSE,
    created_at TIMESTAMP not null
);
drop table if exists DocumentAccess;
CREATE TABLE IF NOT EXISTS DocumentAccess
(
    id          SERIAL PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES Documents (id) ON DELETE CASCADE,
    login       text NOT NULL REFERENCES users_schema.users (login) ON DELETE CASCADE,
    constraint doc_user UNIQUE (document_id, login)
);

CREATE INDEX IF NOT EXISTS idx_documents_owner ON Documents (owner);
CREATE INDEX IF NOT EXISTS idx_document_access_doc_user ON DocumentAccess (document_id, login);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop schema users_schema cascade;
drop schema auth_schema cascade;

drop table Documents cascade;
drop table DocumentAccess cascade;

drop index idx_documents_owner;
drop index idx_document_access_doc_user;
-- +goose StatementEnd
