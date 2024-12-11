package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
)

type AuthStorager interface {
	GetUserSecretByLogin(ctx context.Context, login string) (structs.UserSecretDTO, error)
	CreateUserSecret(ctx context.Context, secretDTO structs.UserSecretDTO) error
	UpdateUserSecret(ctx context.Context, secretDTO structs.UserSecretDTO) error
	DeleteUserSecret(ctx context.Context, token string) error
	GetUserLoginBySecret(ctx context.Context, secret string) (string, error)
	GetUserSecretBySecret(ctx context.Context, token string) (structs.UserSecretDTO, error)
}

const validHoursNum = 24

func (m *ModelAuth) RegisterUser(ctx context.Context, info structs.RegisterUserInfo) error {
	lgr := logger.GetLogger()

	passHash := getHash(info.Pswd)
	userDTO := structs.UserDTO{
		Login:        info.Login,
		PasswordHash: passHash,
	}
	_, err := m.us.CreateUser(ctx, userDTO)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		lgr.Error(err.Error(), "ModelAuth", "RegisterUser", "CreateUser")

		return err
	}

	return nil
}

// ValidateToken checks whether token bad, expired or not authenticated.
// Returns ErrInvalidToken or ErrTokenExpired or err
func (m *ModelAuth) ValidateToken(ctx context.Context, token string) (bool, error) {
	lgr := logger.GetLogger()

	userSecret, err := m.as.GetUserSecretBySecret(ctx, token)

	if err != nil {
		if errors.Is(err, ErrNotFound) { //It's OK. Bad token or not authenticated
			return false, ErrInvalidToken
		}
		lgr.Error(err.Error(), "ModelAuth", "ValidateToken", "GetUserSecretBySecret")

		return false, err
	}
	if userSecret.ValidUntil.Before(time.Now()) { // То есть он был, но просрочен.
		return false, ErrTokenExpired
	}

	return true, nil
}

func (m *ModelAuth) LoginUser(ctx context.Context, info structs.AuthUserInfo) (string, error) {
	lgr := logger.GetLogger()

	userDTO, err := m.us.GetUserByLogin(ctx, info.Login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			lgr.Info(fmt.Sprintf("login not found: %s", info.Login), "ModelAuth", "LoginUser", "GetUserByLogin")

			return "", ErrBadCredentials
		}
		lgr.Error(err.Error(), "ModelAuth", "LoginUser", "GetUserByLogin")

		return "", err
	}
	if userDTO.PasswordHash != getHash(info.Pswd) {
		lgr.Info("password is incorrect", "ModelAuth", "LoginUser", "GetUserByLogin")

		return "", ErrBadCredentials
	}

	userSecret, err := m.as.GetUserSecretByLogin(ctx, info.Login)

	if err != nil {
		if errors.Is(err, ErrNotFound) { // Токена (секрета) не было, значит нужно создать

			userSecret, err := genUserSecret(info.Login)
			if err != nil {
				lgr.Error(err.Error(), "ModelAuth", "LoginUser", "genKey")
				return "", err
			}

			err = m.as.CreateUserSecret(ctx, userSecret)
			if err != nil {
				lgr.Error(err.Error(), "ModelAuth", "LoginUser", "CreateUserSecret")

				return "", err
			}

			return userSecret.Token, nil
		}
		lgr.Error(err.Error(), "ModelAuth", "LoginUser", "GetUserSecretByLogin")

		return "", err
	}

	if userSecret.ValidUntil.Before(time.Now()) { // То есть он был, но просрочен. Значит нужно обновить
		userSecret.ValidUntil = time.Now().Add(time.Hour * validHoursNum)

		err = m.as.UpdateUserSecret(ctx, userSecret)
		if err != nil {
			if errors.Is(err, ErrConflict) {
				return "", ErrConflict
				// Но в теории такую ошибку не получить. Разве что токен повторится
			}
			lgr.Error(err.Error(), "ModelAuth", "LoginUser", "UpdateUserSecret")

			return "", err
		}
	}

	return userSecret.Token, nil
}

func (m *ModelAuth) LogoutUser(ctx context.Context, token string) error {
	lgr := logger.GetLogger()

	err := m.as.DeleteUserSecret(ctx, token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			lgr.Info(fmt.Sprintf("token not found: %s", token), "ModelAuth", "LogoutUser", "DeleteUserSecret")

			return ErrTokenNotFound
		}
		lgr.Error(err.Error(), "ModelAuth", "LogoutUser", "DeleteUserSecret")

		return err
	}
	return nil
}

func getHash(pass string) string {
	hash := sha256.New()
	hash.Write([]byte(pass))
	return hex.EncodeToString(hash.Sum(nil))
}

func genKey(length int) (string, error) {
	result := ""
	for {
		if len(result) >= length {
			return result, nil
		}
		num, err := rand.Int(rand.Reader, big.NewInt(int64(127)))
		if err != nil {
			return "", err
		}
		n := num.Int64()
		if (n >= 48 && n <= 57) || (n >= 65 && n <= 90) || (n >= 97 && n <= 122) {
			result += string(n)
		}
	}
}

func genUserSecret(login string) (structs.UserSecretDTO, error) {
	key, err := genKey(64)
	if err != nil {
		return structs.UserSecretDTO{}, err
	}

	userSecret := structs.UserSecretDTO{
		Login:      login,
		Token:      key,
		ValidUntil: time.Now().Add(time.Hour * validHoursNum),
	}
	return userSecret, nil
}
