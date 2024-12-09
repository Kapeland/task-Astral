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
	PostNewDoc(ctx context.Context, fileDTO *structs.FileDTO, owner string) (string, error)
	AddGrants(ctx context.Context, docID string, userLogin string) error
	GetGrantsByDocID(ctx context.Context, docID string) ([]string, error)
	GetDoc(ctx context.Context, docID string) (*structs.GetDocDTO, error)
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

func (m *FileStorage) AddDoc(ctx context.Context, docDTO structs.FileDTO, owner string, logins []string) error {
	docID, err := m.fr.PostNewDoc(ctx, &docDTO, owner)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return models.ErrConflict
		}

		return err
	}

	for _, login := range logins {
		err := m.fr.AddGrants(ctx, docID, login)
		if err != nil {
			if errors.Is(err, repository.ErrDuplicateKey) {
				return models.ErrConflict
			}
			if errors.Is(err, repository.ErrAddGrantToLogin) {
				m.fr.DelDoc(ctx, docID, owner)
				// По идее не может вернуть ошибку, поскольку мы только что точно успешно добавили документ

				return models.ErrInvalidInput
			}

			return err
		}

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

func (m *FileStorage) GetDoc(ctx context.Context, docID string) (structs.GetDocDTO, error) {
	doc, err := m.fr.GetDoc(ctx, docID)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.GetDocDTO{}, models.ErrNotFound
		}
		return structs.GetDocDTO{}, err
	}

	return *doc, nil
}
