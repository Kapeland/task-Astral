package file

import (
	"path"
)

const filePath = "file-storage"

type FileProvider interface {
	GetFile(path string) ([]byte, error)
}

type Repository struct {
	f FileProvider
}

func NewRepository(f FileProvider) *Repository {
	return &Repository{f: f}
}

func (r *Repository) GetFileByte(file string) ([]byte, error) {
	bytes, err := r.f.GetFile(path.Join(filePath, file))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
