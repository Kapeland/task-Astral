package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/utils/configer"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"unicode"

	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
)

type AuthModelManager interface {
	RegisterUser(ctx context.Context, info structs.RegisterUserInfo) error
	LoginUser(ctx context.Context, info structs.AuthUserInfo) (string, error)
	LogoutUser(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (bool, error)
}

type authServer struct {
	a AuthModelManager
}

func isPasswordValid(s string) bool {
	if len(s) < 8 {
		return false
	}
	symbols := 0
	number, upper, lower, special := false, false, false, false
	for _, c := range s {
		switch {
		case unicode.IsDigit(c):
			number = true
			symbols++
		case unicode.IsUpper(c):
			upper = true
			symbols++
		case unicode.IsLower(c):
			lower = true
			symbols++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
			symbols++
		case unicode.IsLetter(c):
			symbols++
		}
	}
	if symbols < 8 || !(number && special && upper && lower) {
		return false
	}

	return true
}

func isLoginValid(s string) bool {
	if len(s) < 8 {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) {
			return false
		}
	}

	return true
}

func IsLoginPswdValid(login string, password string) bool {
	return isLoginValid(login) && isPasswordValid(password)
}

func (s *authServer) Register(c *gin.Context) {
	token := c.Query("token")
	login := c.Query("login")
	pswd := c.Query("pswd")
	cfg, err := configer.GetConfig()
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("authServer: Register: configer.GetConfig error: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "GetConfig error",
		}})
		return
	}

	if cfg.Admin.Token != token { // It's not admin
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: Register: Not admin"))
		c.JSON(http.StatusForbidden, ErrResponse{Err: ErrBody{
			Code: 403,
			Text: "Not admin",
		}})
		return
	}
	if !IsLoginPswdValid(login, pswd) { // bad password or login
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: Register: Bad pass or login"))
		c.JSON(http.StatusBadRequest, ErrResponse{Err: ErrBody{
			Code: 400,
			Text: "Bad pass or login",
		}})
		return
	}

	userInfo := structs.RegisterUserInfo{
		Login: login,
		Pswd:  pswd,
	}
	status := s.register(c.Request.Context(), userInfo)

	if status == http.StatusBadRequest {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: register: item already exists"))
		c.JSON(status, ErrResponse{Err: ErrBody{
			Code: status,
			Text: "item already exists",
		}})
		return
	}
	if status == http.StatusInternalServerError {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: register: internal server error"))
		c.JSON(status, ErrResponse{Err: ErrBody{
			Code: status,
			Text: "internal server error",
		}})
		return
	}

	c.JSON(http.StatusOK, RegisterResp{RegisterRespBody{login}})
}

func (s *authServer) register(ctx context.Context, info structs.RegisterUserInfo) int {
	err := s.a.RegisterUser(ctx, info)
	if err != nil {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: RegisterUser: Error: %s", err.Error()))
		if errors.Is(err, models.ErrConflict) {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (s *authServer) Auth(c *gin.Context) {
	login := c.PostForm("login")
	pswd := c.PostForm("pswd")

	if !IsLoginPswdValid(login, pswd) { // bad password or login
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: Auth: Bad pass or login"))
		c.JSON(http.StatusBadRequest, ErrResponse{Err: ErrBody{
			Code: 400,
			Text: "Bad pass or login",
		}})
		return
	}

	userInfo := structs.AuthUserInfo{
		Login: login,
		Pswd:  pswd,
	}
	token, status, errResp := s.auth(c.Request.Context(), userInfo)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusOK, AuthResp{AuthRespBody{token}})
}

func (s *authServer) auth(ctx context.Context, info structs.AuthUserInfo) (string, int, ErrResponse) {
	token, err := s.a.LoginUser(ctx, info)
	if err != nil {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: LoginUser: Error: %s", err.Error()))

		if errors.Is(err, models.ErrBadCredentials) {
			return "", http.StatusBadRequest, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Wrong pass or login",
			}}
		}
		if errors.Is(err, models.ErrConflict) {
			return "", http.StatusBadRequest, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Generated duplicated token",
			}}
		}
		return "", http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server error",
		}}
	}

	return token, http.StatusOK, ErrResponse{}
}

func (s *authServer) Logout(c *gin.Context) {
	tokenP := c.Param("token")

	token, status, errResp := s.logout(c.Request.Context(), tokenP)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}

	newData, err := jsoniter.Marshal(LogoutResp{jsoniter.RawMessage(fmt.Sprintf("{\"%s\":true}", token))})
	if err != nil {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: Logout: soniter.Marshal err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Can't marshal JSON",
		}})
		return
	}
	var tmp map[string]interface{}
	err = jsoniter.Unmarshal(newData, &tmp)

	if err != nil {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("authServer: Logout: soniter.Unmarshal err:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Can't unmarshal JSON",
		}})
		return
	}

	c.JSON(status, tmp)
}

func (s *authServer) logout(ctx context.Context, token string) (string, int, ErrResponse) {
	// Здесь валидировать токен не нужно, поскольку, например, при истечении срока годности пользователь не сможет выйти
	// В противном случае если попытаться удалить несуществующий токен, то это нормально.

	err := s.a.LogoutUser(ctx, token)
	if err != nil {
		if errors.Is(err, models.ErrTokenNotFound) {
			return "", http.StatusNotFound, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Token not found",
			}}
		}
		return "", http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server error",
		}}
	}
	return token, http.StatusOK, ErrResponse{}
}
