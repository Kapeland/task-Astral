package files

import (
	"context"
	"database/sql"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/db"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"log/slog"
	"math"
	"time"
)

type Repo struct {
	db db.DBops
}

func New(db db.DBops) *Repo {
	return &Repo{db: db}
}

func (m *Repo) AddGrants(ctx context.Context, docID string, userLogin string) error {
	tx, err := m.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO documentaccess(document_id, login)
				VALUES($1,$2) returning id;`, docID, userLogin)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		if pgErr.Code == "23505" {
			return repository.ErrDuplicateKey
		}
		if pgErr.Code == "23503" {
			return repository.ErrAddGrantToLogin
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return err
	}

	return nil
}

// GetAllDocsByOwner returns all docs belonging to an owner from postgres
func (m *Repo) GetAllDocsByOwner(ctx context.Context, listInfo structs.ListInfo, ownerLogin string, own bool) ([]structs.DocEntry, error) {
	filter := listInfo.Value

	lmt := math.MaxInt64
	if listInfo.Limit != 0 {
		lmt = listInfo.Limit
	}

	var docs []*structs.DocEntry
	var err error

	tx, err := m.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return []structs.DocEntry{}, err
	}
	defer tx.Rollback()

	if own {
		switch listInfo.Key {
		case "":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 ORDER BY title, created_at limit $2;`, ownerLogin, lmt)
		case "id":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and id=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "name":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and title=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "mime":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and mime=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "file":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and file=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "public":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and is_public=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "created":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and created_at=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		default:
			err = repository.ErrObjectNotFound
		}
	} else {
		switch listInfo.Key {
		case "":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and is_public=true ORDER BY title, created_at limit $2;`, ownerLogin, lmt)
		case "id":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and id=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "name":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and title=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "mime":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and mime=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "file":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and file=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "public":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "created":
			err = tx.SelectContext(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and created_at=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		default:
			err = repository.ErrObjectNotFound
		}
	}

	if err != nil {
		return nil, err
	}
	docsOut := make([]structs.DocEntry, len(docs))
	for i, doc := range docs {
		docsOut[i] = *doc
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return nil, err
	}

	return docsOut, nil
}

// DelDoc  deletes doc in postgres
func (m *Repo) DelDoc(ctx context.Context, docID string, userLogin string) (structs.RmDoc, error) {
	tmpID, tmpTitle := "", ""

	tx, err := m.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return structs.RmDoc{}, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`DELETE FROM documents WHERE id = $1 and owner = $2 returning id, title;`, docID, userLogin).Scan(&tmpID, &tmpTitle)

	switch {
	case err != nil && (errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows)):
		return structs.RmDoc{}, repository.ErrObjectNotFound
	case err != nil:
		return structs.RmDoc{}, err
	default:
		if err := tx.Commit(); err != nil {
			slog.Info("Looks like the context has been closed")
			slog.Error(err.Error())
			return structs.RmDoc{}, err
		}
		return structs.RmDoc{
			ID:   tmpID,
			Name: tmpTitle,
		}, nil

	}

}

// PostNewDoc add doc info to postgres
func (m *Repo) PostNewDoc(ctx context.Context, file *structs.File, owner string) (string, error) {
	//TODO: По сути тут не учитывается есть ли такой документ. Хотя дальше это предполагается.
	id := ""

	tx, err := m.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO documents(title, content, mime, owner, is_public, created_at, file)
				VALUES($1,$2,$3,$4,$5,$6,$7) returning id;`, file.Meta.Name, string(file.Json), file.Meta.Mime, owner, file.Meta.Public, time.Now(), file.Meta.File).Scan(&id)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		if pgErr.Code == "23505" {
			return "", repository.ErrDuplicateKey
		}
		return "", err
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return "", err
	}

	return id, nil
}

func (m *Repo) GetGrantsByDocID(ctx context.Context, docID string) ([]string, error) {
	var logins []*string

	tx, err := m.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.SelectContext(ctx, &logins,
		`SELECT login
		FROM documentaccess
		WHERE document_id=$1;`, docID)

	if err != nil {
		return nil, err
	}
	loginsOut := make([]string, len(logins))
	for i, doc := range logins {
		loginsOut[i] = *doc
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return nil, err
	}

	return loginsOut, nil
}

func (m *Repo) GetDoc(ctx context.Context, docID string) (*structs.GetDoc, error) {
	doc := structs.GetDoc{}

	tx, err := m.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.GetContext(ctx, &doc,
		`SELECT mime, title, owner, is_public, file FROM documents
				WHERE id=$1;`, docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrObjectNotFound
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return nil, err
	}

	return &doc, nil
}
