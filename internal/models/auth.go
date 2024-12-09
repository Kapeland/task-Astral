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
		logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: RegisterUser: CreateUser error: %s", err.Error()))
		return err
	}

	return nil
}

// Returns ErrInvalidToken or ErrTokenExpired or err
func (m *ModelAuth) ValidateToken(ctx context.Context, token string) (bool, error) {
	userSecret, err := m.as.GetUserSecretBySecret(ctx, token)

	if err != nil {
		if errors.Is(err, ErrNotFound) { //It's OK. Bad token or not authenticated
			return false, ErrInvalidToken
		}
		logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: ValidateToken: GetUserSecretBySecret error: %s", err.Error()))
		return false, err
	}
	if userSecret.ValidUntil.Before(time.Now()) { // То есть он был, но просрочен.
		return false, ErrTokenExpired
	}

	return true, nil
}

func (m *ModelAuth) LoginUser(ctx context.Context, info structs.AuthUserInfo) (string, error) {
	userDTO, err := m.us.GetUserByLogin(ctx, info.Login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			logger.Log(logger.InfoPrefix, fmt.Sprintf("ModelAuth: GetUserByLogin: login not found: %s", info.Login))
			return "", ErrBadCredentials
		}

		logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: LoginUser: GetUserByLogin error: %s", err.Error()))
		return "", err
	}
	if userDTO.PasswordHash != getHash(info.Pswd) {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("ModelAuth: LoginUser: password is incorrect"))
		return "", ErrBadCredentials
	}

	userSecret, err := m.as.GetUserSecretByLogin(ctx, info.Login)

	if err != nil {
		if errors.Is(err, ErrNotFound) { // Токена (секрета) не было, значит нужно создать

			userSecret, err := genUserSecret(info.Login)
			if err != nil {
				logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: LoginUser: genKey error: %s", err.Error()))
				return "", err
			}

			err = m.as.CreateUserSecret(ctx, userSecret)
			if err != nil {
				logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: LoginUser: CreateUserSecret error: %s", err.Error()))
				return "", err
			}

			return userSecret.Token, nil
		}

		logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: GetUserSecretByLogin: error: %s", err.Error()))
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
			logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: LoginUser: UpdateUserSecret error: %s", err.Error()))
			return "", err
		}
	}

	return userSecret.Token, nil
}

func (m *ModelAuth) LogoutUser(ctx context.Context, token string) error {
	err := m.as.DeleteUserSecret(ctx, token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			logger.Log(logger.InfoPrefix, fmt.Sprintf("ModelAuth: LogoutUser: DeleteUserSecret: token not found: %s", token))
			return ErrTokenNotFound
		}
		logger.Log(logger.ErrPrefix, fmt.Sprintf("ModelAuth: LogoutUser: DeleteUserSecret: %s", err.Error()))
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
