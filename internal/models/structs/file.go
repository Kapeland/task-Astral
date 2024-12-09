package structs

import (
	jsoniter "github.com/json-iterator/go"
	"time"
)

// GetDocDTO dto
type GetDocDTO struct {
	Mime   string `db:"mime" json:"mime"`
	Name   string `db:"title" json:"title"`
	Public bool   `db:"is_public" json:"public"`
	Owner  string `db:"owner" json:"owner"`
	IsFile bool   `db:"file" json:"is_file"`
}

// FileDTO dto
type FileDTO struct {
	Meta DocMetaDTO
	Json jsoniter.RawMessage
}
type DocMetaDTO struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

type RmDoc struct {
	ID   string `db:"id"`
	Name string `db:"title"`
}

type DocEntry struct {
	ID      string    `json:"id" db:"id"`
	Name    string    `json:"name" db:"title"`
	Mime    string    `json:"mime" db:"mime"`
	IsFile  bool      `json:"file"`
	Public  bool      `json:"public" db:"is_public"`
	Created time.Time `json:"created" db:"created_at"`
	Granted []string  `json:"grant"`
}

type ListInfo struct {
	Token string `json:"token"`
	Login string `json:"login"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Limit int    `json:"limit"`
}
