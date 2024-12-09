package models

import "errors"

var ErrNotFound = errors.New("item not found")

var ErrTokenNotFound = errors.New("token not found")

var ErrInvalidInput = errors.New("invalid input")

var ErrConflict = errors.New("item already exists")

var ErrBadCredentials = errors.New("bad password or login")

var ErrInvalidToken = errors.New("invalid token")

var ErrTokenExpired = errors.New("token expired")

var ErrForbidden = errors.New("forbidden")
