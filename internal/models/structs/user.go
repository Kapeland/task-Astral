package structs

import "time"

type RegisterUserInfo struct {
	Login string `json:"login"`
	Pswd  string `json:"pswd"`
}

type User struct {
	ID       int    `db:"id"`
	Login    string `db:"login"`
	PswdHash string `db:"password_hash"`
}

type AuthUserInfo struct {
	Login string `json:"login"`
	Pswd  string `json:"pswd"`
}

type UserSecret struct {
	Login      string    `db:"login"`
	ValidUntil time.Time `db:"valid_until"`
	Token      string    `db:"token"`
}
