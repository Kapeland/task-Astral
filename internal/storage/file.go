package storage

import (
	"context"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/repository"

	"github.com/pkg/errors"
)

type FileRepo interface {
	GetAllDocsByOwner(ctx context.Context, listInfo structs.ListInfo, ownerLogin string, own bool) ([]structs.DocEntry, error)
	DelDoc(ctx context.Context, docID string, userLogin string) (structs.RmDoc, error)
	PostNewDoc(ctx context.Context, file *structs.File, owner string) error
	GetGrantsByDocID(ctx context.Context, docID string) ([]string, error)
	GetDoc(ctx context.Context, docID string) (*structs.GetDoc, error)
}

type FileStorage struct {
	fr FileRepo
	fp FileProvider
}

type FileProvider interface {
	GetFileByte(file string) ([]byte, error)
}

func NewFileStorage(fileRepo FileRepo, fp FileProvider) FileStorage {
	return FileStorage{fr: fileRepo, fp: fp}
}

func (m *FileStorage) DeleteDoc(ctx context.Context, docID string, userLogin string) (structs.RmDoc, error) {
	doc, err := m.fr.DelDoc(ctx, docID, userLogin)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.RmDoc{}, models.ErrNotFound
		}
		return structs.RmDoc{}, err
	}
	return doc, nil
}

func (m *FileStorage) AddDoc(ctx context.Context, doc structs.File, owner string, logins []string) error {
	err := m.fr.PostNewDoc(ctx, &doc, owner)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return models.ErrConflict
		}
		if errors.Is(err, repository.ErrAddGrantToLogin) {
			return models.ErrInvalidInput
		}
		return err
	}

	return nil
}

func (m *FileStorage) GetDocsByOwner(ctx context.Context, listInfo structs.ListInfo, ownerLogin string, own bool) ([]structs.DocEntry, error) {
	docs, err := m.fr.GetAllDocsByOwner(ctx, listInfo, ownerLogin, own)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return []structs.DocEntry{}, models.ErrNotFound
		}
		return []structs.DocEntry{}, err
	}
	for i, doc := range docs {
		grants, err := m.fr.GetGrantsByDocID(ctx, doc.ID)
		if err != nil {
			if errors.Is(err, repository.ErrObjectNotFound) {
				return []structs.DocEntry{}, models.ErrNotFound
			}
			return docs, nil // It's OK if no grants
		}
		docs[i].Granted = make([]string, len(grants))
		for j, grant := range grants {
			docs[i].Granted[j] = grant
		}
	}

	return docs, nil
}

func (m *FileStorage) GetDoc(ctx context.Context, docID string) (structs.GetDoc, error) {
	doc, err := m.fr.GetDoc(ctx, docID)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.GetDoc{}, models.ErrNotFound
		}
		return structs.GetDoc{}, err
	}

	return *doc, nil
}
