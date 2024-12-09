package file_provider

import (
	"os"
	"path/filepath"
)

type FileProvider struct {
}

func NewFileProvider() *FileProvider {
	return &FileProvider{}
}

func (f *FileProvider) GetFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *FileProvider) GetAllFileNames(path string) ([]string, error) {
	names, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, err
	}
	namesClean := make([]string, 0)
	for _, name := range names {
		namesClean = append(namesClean, filepath.Base(name))
	}
	return namesClean, nil
}
