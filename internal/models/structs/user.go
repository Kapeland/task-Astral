package structs

import "time"

type RegisterUserInfo struct {
	Login string `json:"login"`
	Pswd  string `json:"pswd"`
}

// UserDTO dto
type UserDTO struct {
	ID           int    `db:"id"`
	Login        string `db:"login"`
	PasswordHash string `db:"password_hash"`
}

type AuthUserInfo struct {
	Login string `json:"login"`
	Pswd  string `json:"pswd"`
}

// UserSecretDTO dto
type UserSecretDTO struct {
	Login      string    `db:"login"`
	ValidUntil time.Time `db:"valid_until"`
	Token      string    `db:"token"`
}
