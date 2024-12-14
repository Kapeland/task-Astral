package servers

import (
	"context"
	"errors"
	"fmt"
	structs2 "github.com/Kapeland/task-Astral/internal/services/structs"
	"github.com/Kapeland/task-Astral/internal/utils/config"
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

type AuthServer struct {
	A AuthModelManager
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

func (s *AuthServer) Register(c *gin.Context) {
	token := c.Query("token")
	login := c.Query("login")
	pswd := c.Query("pswd")
	cfg := config.GetConfig()

	lgr := logger.GetLogger()

	if cfg.Admin.Token != token { // It's not admin
		lgr.Info("Not admin", "authServer", "Register", "")

		c.JSON(http.StatusForbidden, structs2.ErrResponse{Err: structs2.ErrBody{
			Code: 403,
			Text: "Not admin",
		}})
		return
	}
	if !IsLoginPswdValid(login, pswd) { // bad password or login
		lgr.Info("Bad pass or login", "authServer", "Register", "IsLoginPswdValid")

		c.JSON(http.StatusBadRequest, structs2.ErrResponse{Err: structs2.ErrBody{
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
		lgr.Info("item already exists", "authServer", "Register", "register")

		c.JSON(status, structs2.ErrResponse{Err: structs2.ErrBody{
			Code: status,
			Text: "User already exists",
		}})
		return
	}
	if status == http.StatusInternalServerError {
		lgr.Error("internal server error", "authServer", "Register", "register")
		c.JSON(status, structs2.ErrResponse{Err: structs2.ErrBody{
			Code: status,
			Text: "Internal server error",
		}})
		return
	}

	c.JSON(http.StatusOK, structs2.RegisterResp{structs2.RegisterRespBody{login}})
}

func (s *AuthServer) register(ctx context.Context, info structs.RegisterUserInfo) int {
	lgr := logger.GetLogger()

	err := s.A.RegisterUser(ctx, info)
	if err != nil {
		if errors.Is(err, models.ErrConflict) {
			return http.StatusBadRequest
		}

		lgr.Error(err.Error(), "authServer", "register", "RegisterUser")

		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (s *AuthServer) Auth(c *gin.Context) {
	lgr := logger.GetLogger()

	login := c.PostForm("login")
	pswd := c.PostForm("pswd")

	if !IsLoginPswdValid(login, pswd) { // bad password or login
		lgr.Info("Bad pass or login", "authServer", "Auth", "")

		c.JSON(http.StatusBadRequest, structs2.ErrResponse{Err: structs2.ErrBody{
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

	c.JSON(http.StatusOK, structs2.AuthResp{structs2.AuthRespBody{token}})
}

func (s *AuthServer) auth(ctx context.Context, info structs.AuthUserInfo) (string, int, structs2.ErrResponse) {
	lgr := logger.GetLogger()

	token, err := s.A.LoginUser(ctx, info)
	if err != nil {
		if errors.Is(err, models.ErrBadCredentials) {
			return "", http.StatusBadRequest, structs2.ErrResponse{Err: structs2.ErrBody{
				Code: 400,
				Text: "Wrong pass or login",
			}}
		}
		if errors.Is(err, models.ErrConflict) {
			return "", http.StatusBadRequest, structs2.ErrResponse{Err: structs2.ErrBody{
				Code: 400,
				Text: "Generated duplicated token",
			}}
		}

		lgr.Error(err.Error(), "authServer", "auth", "LoginUser")
		return "", http.StatusInternalServerError, structs2.ErrResponse{Err: structs2.ErrBody{
			Code: 500,
			Text: "Internal server error",
		}}
	}

	return token, http.StatusOK, structs2.ErrResponse{}
}

func (s *AuthServer) Logout(c *gin.Context) {
	lgr := logger.GetLogger()

	tokenP := c.Param("token")

	token, status, errResp := s.logout(c.Request.Context(), tokenP)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}

	newData, err := jsoniter.Marshal(structs2.LogoutResp{jsoniter.RawMessage(fmt.Sprintf("{\"%s\":true}", token))})
	if err != nil {
		lgr.Error(err.Error(), "authServer", "Logout", "soniter.Marshal")

		c.JSON(http.StatusInternalServerError, structs2.ErrResponse{Err: structs2.ErrBody{
			Code: 500,
			Text: "Can't marshal JSON",
		}})
		return
	}
	var tmp map[string]interface{}
	err = jsoniter.Unmarshal(newData, &tmp)

	if err != nil {
		lgr.Error(err.Error(), "authServer", "Logout", "soniter.Unmarshal")

		c.JSON(http.StatusInternalServerError, structs2.ErrResponse{Err: structs2.ErrBody{
			Code: 500,
			Text: "Can't unmarshal JSON",
		}})
		return
	}

	c.JSON(status, tmp)
}

func (s *AuthServer) logout(ctx context.Context, token string) (string, int, structs2.ErrResponse) {
	// Здесь валидировать токен не нужно, поскольку, например, при истечении срока годности пользователь не сможет выйти
	// В противном случае если попытаться удалить несуществующий токен, то это нормально.

	err := s.A.LogoutUser(ctx, token)
	if err != nil {
		if errors.Is(err, models.ErrTokenNotFound) {
			return "", http.StatusNotFound, structs2.ErrResponse{Err: structs2.ErrBody{
				Code: 400,
				Text: "Token not found",
			}}
		}
		return "", http.StatusInternalServerError, structs2.ErrResponse{Err: structs2.ErrBody{
			Code: 500,
			Text: "Internal server error",
		}}
	}
	return token, http.StatusOK, structs2.ErrResponse{}
}
