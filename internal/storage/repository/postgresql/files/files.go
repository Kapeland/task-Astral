package files

import (
	"context"
	"database/sql"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/db"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"math"
	"time"
)

type Repo struct {
	db db.DBops
}

func (m *Repo) AddGrants(ctx context.Context, docID string, userLogin string) error {
	tmpID := 0
	err := m.db.ExecQueryRow(ctx,
		`INSERT INTO documentaccess(document_id, login)
				VALUES($1,$2) returning id;`, docID, userLogin).Scan(&tmpID)

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

	return nil
}

func New(db db.DBops) *Repo {
	return &Repo{db: db}
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

	if own {
		switch listInfo.Key {
		case "":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 ORDER BY title, created_at limit $2;`, ownerLogin, lmt)
		case "id":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and id=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "name":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and title=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "mime":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and mime=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "file":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and file=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "public":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and is_public=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "created":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and created_at=$2 ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		default:
			err = repository.ErrObjectNotFound
		}
	} else {
		switch listInfo.Key {
		case "":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and is_public=true ORDER BY title, created_at limit $2;`, ownerLogin, lmt)
		case "id":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and id=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "name":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and title=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "mime":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and mime=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "file":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and file=$2 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "public":
			err = m.db.Select(ctx, &docs,
				`SELECT id, title, mime, is_public, created_at
		FROM documents
		WHERE owner=$1 and is_public=true ORDER BY title, created_at limit $3;`, ownerLogin, filter, lmt)
		case "created":
			err = m.db.Select(ctx, &docs,
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

	return docsOut, nil
}

// DelDoc  deletes doc in postgres
func (m *Repo) DelDoc(ctx context.Context, docID string, userLogin string) (structs.RmDoc, error) {
	tmpID, tmpTitle := "", ""
	err := m.db.ExecQueryRow(ctx,
		`DELETE FROM documents WHERE id = $1 and owner = $2 returning id, title;`, docID, userLogin).Scan(&tmpID, &tmpTitle)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return structs.RmDoc{}, repository.ErrObjectNotFound
		}
		return structs.RmDoc{}, err
	}

	return structs.RmDoc{
		ID:   tmpID,
		Name: tmpTitle,
	}, nil
}

// PostNewDoc add doc info to postgres
func (m *Repo) PostNewDoc(ctx context.Context, fileDTO *structs.FileDTO, owner string) (string, error) {
	id := ""
	err := m.db.ExecQueryRow(ctx,
		`INSERT INTO documents(title, content, mime, owner, is_public, created_at, file)
				VALUES($1,$2,$3,$4,$5,$6,$7) returning id;`, fileDTO.Meta.Name, string(fileDTO.Json), fileDTO.Meta.Mime, owner, fileDTO.Meta.Public, time.Now(), fileDTO.Meta.File).Scan(&id)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		if pgErr.Code == "23505" {
			return "", repository.ErrDuplicateKey
		}
		return "", err
	}

	return id, nil
}

func (m *Repo) GetGrantsByDocID(ctx context.Context, docID string) ([]string, error) {
	var logins []*string
	err := m.db.Select(ctx, &logins,
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

	return loginsOut, nil
}

func (m *Repo) GetDoc(ctx context.Context, docID string) (*structs.GetDocDTO, error) {
	secretDTO := structs.GetDocDTO{}
	err := m.db.Get(ctx, &secretDTO,
		`SELECT mime, title, owner, is_public, file FROM documents
				WHERE id=$1;`, docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrObjectNotFound
		}
		return nil, err
	}
	return &secretDTO, nil
}
