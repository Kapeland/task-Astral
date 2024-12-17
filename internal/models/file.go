package models

import (
	"context"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/pkg/errors"
)

type FileStorager interface {
	AddDoc(ctx context.Context, doc structs.File, owner string, logins []string) error
	DeleteDoc(ctx context.Context, docID string, userLogin string) (structs.RmDoc, error)
	GetDocsByOwner(ctx context.Context, listInfo structs.ListInfo, ownerLogin string, own bool) ([]structs.DocEntry, error)
	GetDoc(ctx context.Context, docID string) (structs.GetDoc, error)
}

func (m *ModelFiles) AddNewDoc(ctx context.Context, doc structs.File) error {
	login, err := m.as.GetUserLoginBySecret(ctx, doc.Meta.Token)
	if err != nil {
		return err
	}

	err = m.fs.AddDoc(ctx, doc, login, doc.Meta.Grant)
	if err != nil {
		return err
	}

	return nil
}

func (m *ModelFiles) DeleteDoc(ctx context.Context, token string, docID string) (structs.RmDoc, error) {
	login, err := m.as.GetUserLoginBySecret(ctx, token)
	if err != nil {
		return structs.RmDoc{}, err
	}

	doc, err := m.fs.DeleteDoc(ctx, docID, login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return structs.RmDoc{}, ErrNotFound
		}
		return structs.RmDoc{}, err
	}

	return doc, nil
}

func (m *ModelFiles) GetDocs(ctx context.Context, listInfo structs.ListInfo) ([]structs.DocEntry, error) {
	login, err := m.as.GetUserLoginBySecret(ctx, listInfo.Token)
	if err != nil {
		return []structs.DocEntry{}, err
	}
	if listInfo.Login == "" || listInfo.Login == login { // Это значит, что нужно вернуть только наши документы
		docs, err := m.fs.GetDocsByOwner(ctx, listInfo, login, true)
		if err != nil {
			return []structs.DocEntry{}, err
		}

		return docs, nil
	}

	// Значит ищем документы другого человека
	docs, err := m.fs.GetDocsByOwner(ctx, listInfo, login, false)
	if err != nil {
		return []structs.DocEntry{}, err
	}

	return docs, nil

}

func (m *ModelFiles) GetDoc(ctx context.Context, token string, docID string) (structs.GetDoc, error) {
	login, err := m.as.GetUserLoginBySecret(ctx, token)
	if err != nil {
		return structs.GetDoc{}, err
	}

	doc, err := m.fs.GetDoc(ctx, docID)
	if err != nil {
		return structs.GetDoc{}, err
	}
	if doc.Owner == login { // Это наш документ
		return doc, nil
	} else {
		if !doc.Public { // Значит не наш документ и закрытый
			return structs.GetDoc{}, ErrForbidden
		}
	}

	return doc, nil
}
